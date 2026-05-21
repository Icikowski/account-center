package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/sync/singleflight"

	"git.sr.ht/~icikowski/account-center/internal/consts"
)

const (
	hintRefreshToken = "refresh_token"
	hintAccessToken  = "access_token"
	hintIDToken      = "id_token"
)

type serviceOptions struct {
	httpClient     *http.Client
	randReader     io.Reader
	trustedProxies *TrustedProxies
}

// Option is a functional option for configuring the auth service.
type Option interface {
	apply(opts *serviceOptions)
}

type serviceOption func(*serviceOptions)

func (f serviceOption) apply(opts *serviceOptions) {
	f(opts)
}

// WithHTTPClient sets the HTTP client used for discovery, token exchange, UserInfo fetches, and revocation.
func WithHTTPClient(client *http.Client) Option {
	return serviceOption(func(opts *serviceOptions) {
		opts.httpClient = client
	})
}

// WithRandomReader overrides the random source used for generated IDs and PKCE verifiers.
func WithRandomReader(reader io.Reader) Option {
	return serviceOption(func(opts *serviceOptions) {
		opts.randReader = reader
	})
}

// WithTrustedProxies sets the trusted proxies allowed to supply forwarded headers.
func WithTrustedProxies(trustedProxies *TrustedProxies) Option {
	return serviceOption(func(opts *serviceOptions) {
		opts.trustedProxies = trustedProxies
	})
}

// Service represents the OIDC authentication service, handling the authorization flow and session management.
type Service interface {
	// AuthorizationRequest prepares a new OIDC authorization request and stores its transient login state.
	AuthorizationRequest(ctx context.Context, r *http.Request, next string) (AuthorizationRequest, error)
	// ExchangeCode completes the authorization code flow for the given login state.
	ExchangeCode(ctx context.Context, loginID, code string) (Session, string, error)
	// GetSession returns the stored session and refreshes it automatically when it is expired or close to expiry.
	GetSession(ctx context.Context, sessionID string) (Session, error)
	// RefreshSession forces a token refresh before returning the updated session.
	RefreshSession(ctx context.Context, sessionID string) (Session, error)
	// SessionTTL returns the configured session lifetime used for persisted sessions.
	SessionTTL() time.Duration
	// Logout removes the stored session and revokes provider-side tokens when supported.
	Logout(ctx context.Context, sessionID string) error
}

type service struct {
	baseURL                                  string
	clientID, clientSecret                   string
	refreshBefore, sessionTTL, loginStateTTL time.Duration

	store        SessionStore
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	refreshGroup singleflight.Group

	revocationEndpoint string

	randReader     io.Reader
	httpClient     *http.Client
	trustedProxies *TrustedProxies
}

// NewService constructs a reusable OIDC auth [Service].
func NewService(
	ctx context.Context,
	providerURL, clientID, clientSecret string,
	baseURL string,
	refreshBefore, sessionTTL, loginStateTTL time.Duration,
	authStore SessionStore,
	opts ...Option,
) (Service, error) {
	options := serviceOptions{
		httpClient: http.DefaultClient,
		randReader: rand.Reader,
	}
	for _, opt := range opts {
		opt.apply(&options)
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, options.httpClient)
	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errOIDCDiscoveryFailed, err)
	}

	var metadata struct {
		RevocationEndpoint string `json:"revocation_endpoint"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return nil, fmt.Errorf("%w: %w", errOIDCMetadataReadFailed, err)
	}

	return &service{
		baseURL:       baseURL,
		clientID:      clientID,
		clientSecret:  clientSecret,
		refreshBefore: refreshBefore,
		sessionTTL:    sessionTTL,
		loginStateTTL: loginStateTTL,
		store:         authStore,
		provider:      provider,
		verifier: provider.Verifier(&oidc.Config{
			ClientID: clientID,
		}),
		revocationEndpoint: metadata.RevocationEndpoint,
		randReader:         options.randReader,
		httpClient:         options.httpClient,
		trustedProxies:     options.trustedProxies,
	}, nil
}

// AuthorizationRequest implements [Service].
func (s *service) AuthorizationRequest(
	ctx context.Context,
	r *http.Request,
	next string,
) (AuthorizationRequest, error) {
	redirectURL, err := s.redirectURL(r)
	if err != nil {
		return AuthorizationRequest{}, err
	}

	loginID, err := s.randomID()
	if err != nil {
		return AuthorizationRequest{}, err
	}
	codeVerifier, err := s.randomID()
	if err != nil {
		return AuthorizationRequest{}, err
	}
	nonce, err := s.randomID()
	if err != nil {
		return AuthorizationRequest{}, err
	}

	state := LoginState{
		ID:           loginID,
		Next:         next,
		RedirectURL:  redirectURL,
		CodeVerifier: codeVerifier,
		Nonce:        nonce,
		CreatedAt:    time.Now(),
	}
	if err := s.store.LoginStates().Set(ctx, loginID, state, s.loginStateTTL); err != nil {
		return AuthorizationRequest{}, fmt.Errorf("%w: %w", errSaveLoginState, err)
	}

	authURL := s.oauth2Config(redirectURL).AuthCodeURL(
		loginID,
		oauth2.AccessTypeOffline,
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(codeVerifier),
	)

	return AuthorizationRequest{
		LoginID: loginID,
		URL:     authURL,
	}, nil
}

// ExchangeCode implements [Service].
func (s *service) ExchangeCode(ctx context.Context, loginID, code string) (Session, string, error) {
	state, err := s.store.LoginStates().Get(ctx, loginID)
	if err != nil {
		return Session{}, "", fmt.Errorf("%w: %w", errLoadLoginState, err)
	}
	if err := s.store.LoginStates().Delete(ctx, loginID); err != nil {
		return Session{}, "", fmt.Errorf("%w: %w", errDeleteLoginState, err)
	}

	token, err := s.oauth2Config(state.RedirectURL).Exchange(
		context.WithValue(ctx, oauth2.HTTPClient, s.httpClient),
		code,
		oauth2.VerifierOption(state.CodeVerifier),
	)
	if err != nil {
		return Session{}, "", fmt.Errorf("%w: %w", errAuthorizationCodeExchangeFailed, err)
	}

	idToken, claims, err := s.verifyInitialIDToken(ctx, token, state.Nonce)
	if err != nil {
		return Session{}, "", err
	}

	user, err := s.resolveUser(ctx, token, claims, User{})
	if err != nil {
		return Session{}, "", err
	}

	sessionID, err := s.randomID()
	if err != nil {
		return Session{}, "", err
	}

	now := time.Now()
	record := StoredSession{
		ID:           sessionID,
		Subject:      claims.Subject,
		User:         user,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      idToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.store.Sessions().Set(ctx, sessionID, record, s.sessionTTL); err != nil {
		return Session{}, "", fmt.Errorf("%w: %w", errSaveSession, err)
	}

	return record.Session(), state.Next, nil
}

// GetSession implements [Service].
func (s *service) GetSession(ctx context.Context, sessionID string) (Session, error) {
	record, err := s.store.Sessions().Get(ctx, sessionID)
	if err != nil {
		return Session{}, fmt.Errorf("%w: %w", errLoadSession, err)
	}

	if !s.shouldRefresh(record.Expiry) {
		return record.Session(), nil
	}

	refreshed, err := s.refreshSession(ctx, sessionID, false)
	if err != nil {
		if !record.Expiry.IsZero() && record.Expiry.After(time.Now()) {
			return record.Session(), nil
		}
		return Session{}, err
	}

	return refreshed.Session(), nil
}

// RefreshSession implements [Service].
func (s *service) RefreshSession(ctx context.Context, sessionID string) (Session, error) {
	record, err := s.refreshSession(ctx, sessionID, true)
	if err != nil {
		return Session{}, err
	}
	return record.Session(), nil
}

// SessionTTL implements [Service].
func (s *service) SessionTTL() time.Duration {
	return s.sessionTTL
}

// Logout implements [Session].
func (s *service) Logout(ctx context.Context, sessionID string) error {
	record, err := s.store.Sessions().Get(ctx, sessionID)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return nil
		}
		return fmt.Errorf("%w: %w", errLoadSession, err)
	}

	var revokeErr error
	if s.revocationEndpoint != "" {
		if record.RefreshToken != "" {
			revokeErr = errors.Join(revokeErr, s.revokeToken(ctx, record.RefreshToken, hintRefreshToken))
		}
		if record.AccessToken != "" {
			revokeErr = errors.Join(revokeErr, s.revokeToken(ctx, record.AccessToken, hintAccessToken))
		}
	}

	deleteErr := s.store.Sessions().Delete(ctx, sessionID)
	if deleteErr != nil {
		deleteErr = fmt.Errorf("%w: %w", errDeleteSession, deleteErr)
	}

	return errors.Join(revokeErr, deleteErr)
}

func (s *service) refreshSession(ctx context.Context, sessionID string, force bool) (StoredSession, error) {
	result, err, _ := s.refreshGroup.Do(sessionID, func() (any, error) {
		now := time.Now()
		record, err := s.store.Sessions().Get(ctx, sessionID)
		if err != nil {
			return StoredSession{}, fmt.Errorf("%w: %w", errLoadSession, err)
		}
		if !force && !s.shouldRefresh(record.Expiry) {
			return record, nil
		}
		if record.RefreshToken == "" {
			if !force && record.Expiry.After(now) {
				return record, nil
			}
			if err := s.store.Sessions().Delete(ctx, sessionID); err != nil {
				return StoredSession{}, fmt.Errorf("%w: %w", errDeleteSession, err)
			}
			return StoredSession{}, errReauthenticationRequired
		}

		tokenSource := s.oauth2Config("").TokenSource(
			context.WithValue(ctx, oauth2.HTTPClient, s.httpClient),
			&oauth2.Token{
				AccessToken:  record.AccessToken,
				RefreshToken: record.RefreshToken,
				TokenType:    record.TokenType,
				Expiry:       now.Add(-time.Second),
			},
		)
		token, err := tokenSource.Token()
		if err != nil {
			if !force && record.Expiry.After(now) {
				return record, nil
			}
			if err := s.store.Sessions().Delete(ctx, sessionID); err != nil {
				return StoredSession{}, fmt.Errorf("%w: %w", errDeleteSession, err)
			}
			return StoredSession{}, fmt.Errorf("%w: %w", errReauthenticationRequired, err)
		}

		idToken, claims, err := s.refreshClaims(ctx, token, record)
		if err != nil {
			return StoredSession{}, err
		}
		user, err := s.resolveUser(ctx, token, claims, record.User)
		if err != nil {
			return StoredSession{}, err
		}

		if token.RefreshToken == "" {
			token.RefreshToken = record.RefreshToken
		}
		if idToken == "" {
			idToken = record.IDToken
		}
		if claims.Subject == "" {
			claims.Subject = record.Subject
		}

		record.AccessToken = token.AccessToken
		record.RefreshToken = token.RefreshToken
		record.IDToken = idToken
		record.TokenType = token.TokenType
		record.Expiry = token.Expiry
		record.Subject = claims.Subject
		record.User = user
		record.UpdatedAt = now

		if err := s.store.Sessions().Set(ctx, sessionID, record, s.sessionTTL); err != nil {
			return StoredSession{}, fmt.Errorf("%w: %w", errSaveSession, err)
		}

		return record, nil
	})
	if err != nil {
		return StoredSession{}, err
	}
	record, ok := result.(StoredSession)
	if !ok {
		return StoredSession{}, errUnexpectedRefreshResultType
	}
	return record, nil
}

func (s *service) verifyInitialIDToken(
	ctx context.Context,
	token *oauth2.Token,
	expectedNonce string,
) (string, profileClaims, error) {
	raw, claims, err := s.verifyIDToken(ctx, token, true)
	if err != nil {
		return "", profileClaims{}, err
	}
	if subtle.ConstantTimeCompare([]byte(claims.Nonce), []byte(expectedNonce)) != 1 {
		return "", profileClaims{}, errIDTokenNonceMismatch
	}
	return raw, claims, nil
}

func (s *service) refreshClaims(
	ctx context.Context,
	token *oauth2.Token,
	record StoredSession,
) (string, profileClaims, error) {
	raw, claims, err := s.verifyIDToken(ctx, token, false)
	if err == nil {
		return raw, claims, nil
	}
	if errors.Is(err, errNotFound) {
		return "", profileClaims{Subject: record.Subject}, nil
	}
	return "", profileClaims{}, err
}

func (s *service) verifyIDToken(
	ctx context.Context,
	token *oauth2.Token,
	require bool,
) (string, profileClaims, error) {
	var claims profileClaims

	raw, ok := token.Extra(hintIDToken).(string)
	if !ok || raw == "" {
		if require {
			return "", profileClaims{}, errIDTokenMissing
		}
		return "", profileClaims{}, errNotFound
	}

	idToken, err := s.verifier.Verify(context.WithValue(ctx, oauth2.HTTPClient, s.httpClient), raw)
	if err != nil {
		return "", profileClaims{}, fmt.Errorf("%w: %w", errIDTokenVerification, err)
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", profileClaims{}, fmt.Errorf("%w: %w", errIDTokenDecode, err)
	}

	return raw, claims, nil
}

func (s *service) resolveUser(
	ctx context.Context,
	token *oauth2.Token,
	fallback profileClaims,
	fallbackUser User,
) (User, error) {
	user := userFromClaims(fallback, fallbackUser)

	info, err := s.provider.UserInfo(
		context.WithValue(ctx, oauth2.HTTPClient, s.httpClient),
		oauth2.StaticTokenSource(token),
	)
	if err != nil {
		// If the provider doesn't support UserInfo or the request fails, return the user data from the ID token claims.
		//
		//nolint:nilerr
		return user, nil
	}

	var claims profileClaims
	if err := info.Claims(&claims); err != nil {
		return User{}, fmt.Errorf("%w: %w", errUserInfoClaimsDecode, err)
	}

	return userFromClaims(claims, user), nil
}

func userFromClaims(claims profileClaims, fallback User) User {
	out := fallback
	out.Subject = claims.Subject
	if claims.Name != "" {
		out.Name = claims.Name
	}
	if claims.Email != "" {
		out.Email = claims.Email
	}
	if len(claims.Groups) > 0 {
		out.Groups = append([]string(nil), claims.Groups...)
	} else if out.Groups != nil {
		out.Groups = append([]string(nil), out.Groups...)
	}
	return out
}

func (s *service) revokeToken(ctx context.Context, token, hint string) error {
	form := url.Values{}
	form.Set("token", token)
	if hint != "" {
		form.Set("token_type_hint", hint)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.revocationEndpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("%w: %w", errRevocationRequestBuild, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.clientID, s.clientSecret)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w %s: %w", errRevocationFailed, hint, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%w: %s, status %s", errRevocationFailed, hint, resp.Status)
	}

	return nil
}

func (s *service) shouldRefresh(expiry time.Time) bool {
	if expiry.IsZero() {
		return false
	}
	return !time.Now().Add(s.refreshBefore).Before(expiry)
}

func (s *service) oauth2Config(redirectURL string) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		Endpoint:     s.provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       defaultScopes,
	}
	return cfg
}

func (s *service) redirectURL(r *http.Request) (string, error) {
	baseURL := s.baseURL
	if baseURL == "" {
		var err error
		baseURL, err = requestBaseURL(r, s.trustedProxies)
		if err != nil {
			return "", err
		}
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errBaseURLParse, err)
	}

	return base.ResolveReference(&url.URL{Path: consts.RouteOIDCCallback}).String(), nil
}

func requestBaseURL(r *http.Request, trustedProxies *TrustedProxies) (string, error) {
	if r == nil {
		return "", fmt.Errorf("%w: request is required", errBaseURLParse)
	}

	scheme := ""
	if trustedProxies.AllowsForwardedHeaders(r) {
		scheme = headerFirstValue(r.Header.Get(consts.HeaderXForwardedProto))
	}
	switch {
	case scheme != "":
	case r.TLS != nil:
		scheme = consts.SchemeHTTPS
	default:
		scheme = consts.SchemeHTTP
	}

	host := ""
	if trustedProxies.AllowsForwardedHeaders(r) {
		host = headerFirstValue(r.Header.Get(consts.HeaderXForwardedHost))
	}
	if host == "" {
		host = r.Host
	}
	if host == "" {
		return "", fmt.Errorf("%w: request host is empty", errBaseURLParse)
	}

	return (&url.URL{
		Scheme: scheme,
		Host:   host,
	}).String(), nil
}

func headerFirstValue(value string) string {
	if value == "" {
		return ""
	}
	if before, _, ok := strings.Cut(value, ","); ok {
		return strings.TrimSpace(before)
	}
	return strings.TrimSpace(value)
}

func (s *service) randomID() (string, error) {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(s.randReader, buf); err != nil {
		return "", fmt.Errorf("%w: %w", errRandomIDGeneration, err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
