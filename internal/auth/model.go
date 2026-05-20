package auth

import (
	"errors"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	pathOIDCCallback = "/oidc-callback"
)

var (
	errNotFound                        = errors.New("not found")
	errOIDCDiscoveryFailed             = errors.New("failed to discover OIDC provider")
	errOIDCMetadataReadFailed          = errors.New("failed to read OIDC provider metadata")
	errBaseURLParse                    = errors.New("failed to parse base URL")
	errReauthenticationRequired        = errors.New("re-authentication required")
	errUnexpectedRefreshResultType     = errors.New("unexpected refresh result type")
	errSaveLoginState                  = errors.New("failed to save login state")
	errLoadLoginState                  = errors.New("failed to load login state")
	errDeleteLoginState                = errors.New("failed to delete login state")
	errAuthorizationCodeExchangeFailed = errors.New("failed to exchange authorization code")
	errSaveSession                     = errors.New("failed to save session")
	errLoadSession                     = errors.New("failed to load session")
	errDeleteSession                   = errors.New("failed to delete session")
	errIDTokenMissing                  = errors.New("id_token missing from token response")
	errIDTokenDecode                   = errors.New("failed to decode id_token")
	errIDTokenVerification             = errors.New("failed to verify id_token")
	errIDTokenNonceMismatch            = errors.New("nonce mismatch when verifying id_token")
	errRevocationFailed                = errors.New("token revocation failed")
	errRevocationRequestBuild          = errors.New("failed to build token revocation request")
	errUserInfoClaimsDecode            = errors.New("failed to decode user info claims")
	errRandomIDGeneration              = errors.New("failed to generate random ID")
	errLoadFailed                      = errors.New("failed to load")
	errUnmarshalFailed                 = errors.New("failed to unmarshal")

	defaultScopes = []string{
		oidc.ScopeOpenID,
		"profile",
		"email",
		"groups",
		oidc.ScopeOfflineAccess,
	}
)

// LoginState represents the transient state of an ongoing login flow, stored by the session store during
// the authentication process.
type LoginState struct {
	ID           string    `json:"id"`
	ReturnTo     string    `json:"next"`
	RedirectURL  string    `json:"redirect_url"`
	CodeVerifier string    `json:"code_verifier"`
	Nonce        string    `json:"nonce"`
	CreatedAt    time.Time `json:"created_at"`
}

// StoredSession represents a persisted authenticated session, stored by the session store after successful
// authentication.
type StoredSession struct {
	ID           string    `json:"id"`
	Subject      string    `json:"subject"`
	User         User      `json:"user"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IDToken      string    `json:"id_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session returns the public [Session] view.
func (r StoredSession) Session() Session {
	return Session{
		ID:   r.ID,
		User: r.User,
	}
}

// Session represents the public view of an authenticated session, without any token data.
type Session struct {
	ID   string
	User User
}

// User contains the identity data fetched from OIDC.
type User struct {
	Subject string
	Name    string
	Email   string
	Groups  []string
}

// AuthorizationRequest describes a newly prepared OIDC login request.
type AuthorizationRequest struct {
	LoginID string
	URL     string
}

type profileClaims struct {
	Subject string   `json:"sub"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Groups  []string `json:"groups"`
	Nonce   string   `json:"nonce"`
}
