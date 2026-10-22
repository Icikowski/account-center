package auth

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"git.sr.ht/~icikowski/account-center/internal/consts"
	"git.sr.ht/~icikowski/account-center/internal/shared/xlog"
)

// Middleware defines the interface for the authentication middleware provider.
type Middleware interface {
	// Required loads the session into the request context when present, and redirects or rejects the request when not
	// present.
	Required(next http.Handler) http.Handler
	// Optional loads the session into the request context when present, but never redirects or rejects the request.
	Optional(next http.Handler) http.Handler
}

type authMiddleware struct {
	svc               Service
	sessionTTL        time.Duration
	loginStateTTL     time.Duration
	secureCookies     bool
	sessionCookieName string
	oidcCallbackPath  string
	loginPath         string
	refreshPath       string
	logoutPath        string
	publicPaths       []string
	trustedProxies    *TrustedProxies
}

// NewMiddleware creates a new [Middleware] instance with the provided configuration.
func NewMiddleware(
	svc Service,
	sessionCookieName, oidcCallbackPath, loginPath, refreshPath, logoutPath string,
	trustedProxies *TrustedProxies,
	publicPaths ...string,
) Middleware {
	return &authMiddleware{
		svc:               svc,
		sessionTTL:        svc.SessionTTL(),
		loginStateTTL:     svc.LoginStateTTL(),
		secureCookies:     svc.SecureCookies(),
		sessionCookieName: sessionCookieName,
		oidcCallbackPath:  oidcCallbackPath,
		loginPath:         loginPath,
		refreshPath:       refreshPath,
		logoutPath:        logoutPath,
		publicPaths:       publicPaths,
		trustedProxies:    trustedProxies,
	}
}

// Required implements [Middleware].
func (m *authMiddleware) Required(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case m.oidcCallbackPath:
			m.handleCallback(w, r)
			return
		case m.loginPath:
			m.handleLogin(w, r)
			return
		case m.refreshPath:
			m.handleRefresh(w, r)
			return
		case m.logoutPath:
			m.handleLogout(w, r)
			return
		}

		if !m.hasRegisteredHandler(r) || m.isPublicPath(r.URL.Path) {
			if rr, ok := m.withSessionContext(w, r, false); ok {
				r = rr
			}
			next.ServeHTTP(w, r)
			return
		}

		rr, ok := m.withSessionContext(w, r, true)
		if !ok {
			return
		}
		next.ServeHTTP(w, rr)
	})
}

// Optional implements [Middleware].
func (m *authMiddleware) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rr, ok := m.withSessionContext(w, r, false); ok {
			r = rr
		}
		next.ServeHTTP(w, r)
	})
}

func (m *authMiddleware) withSessionContext(
	w http.ResponseWriter,
	r *http.Request,
	required bool,
) (*http.Request, bool) {
	l := zerolog.Ctx(r.Context())

	sessionID, ok := m.sessionIDFromRequest(r)
	if !ok {
		if required {
			l.Debug().Msg("missing session cookie, redirecting to login")
			m.redirectToLogin(w, r, m.currentRequestTarget(r))
			return nil, false
		}
		return r, true
	}

	session, err := m.svc.GetSession(r.Context(), sessionID)
	if err != nil {
		m.clearSessionCookie(w)
		if required {
			l.Warn().Err(err).Msg("failed to load session, redirecting to login")
			if errors.Is(err, errNotFound) || errors.Is(err, errReauthenticationRequired) {
				m.redirectToLogin(w, r, m.currentRequestTarget(r))
				return nil, false
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, false
		}
		return r, true
	}

	m.setSessionCookie(w, r, session.ID)
	ctx := NewContext(r.Context(), &session)
	return r.WithContext(ctx), true
}

func (m *authMiddleware) handleLogin(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())

	if sessionID, ok := m.sessionIDFromRequest(r); ok {
		if _, err := m.svc.GetSession(r.Context(), sessionID); err == nil {
			m.setSessionCookie(w, r, sessionID)
			http.Redirect(w, r, consts.RouteRoot, http.StatusSeeOther)
			return
		}
		l.Debug().Msg("clearing stale session cookie before login")
		m.clearSessionCookie(w)
	}

	next := m.sanitizeRedirectTarget(r, r.URL.Query().Get(paramNext))
	if next == "" {
		next = consts.RouteRoot
	}
	l.Info().Str(xlog.FieldNext, next).Msg("starting login flow")

	authReq, err := m.svc.AuthorizationRequest(r.Context(), r, next)
	if err != nil {
		l.Error().Err(err).Msg("failed to start login flow")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	l.Debug().Str(xlog.FieldNext, next).Msg("redirecting to identity provider")
	m.setLoginStateCookie(w, r, authReq.LoginID)
	http.Redirect(w, r, authReq.URL, http.StatusSeeOther)
}

func (m *authMiddleware) handleCallback(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())

	loginID := r.URL.Query().Get(paramState)
	code := r.URL.Query().Get(paramCode)
	if loginID == "" || code == "" {
		l.Warn().Msg("received invalid oidc callback")
		m.clearLoginStateCookie(w)
		http.Redirect(w, r, m.loginPath, http.StatusSeeOther)
		return
	}

	// The login state is stored server-side, but it must also be bound to the
	// browser that started the flow. Otherwise an attacker can obtain a valid
	// state/code pair for their own account and trick a victim into completing
	// the flow, logging the victim into the attacker's session.
	if !m.loginStateMatchesCookie(r, loginID) {
		l.Warn().Msg("oidc callback state does not match login state cookie")
		m.clearLoginStateCookie(w)
		http.Redirect(w, r, m.loginPath, http.StatusSeeOther)
		return
	}
	m.clearLoginStateCookie(w)

	session, next, err := m.svc.ExchangeCode(r.Context(), loginID, code)
	if err != nil {
		l.Error().Err(err).Msg("failed to complete login callback")
		http.Redirect(w, r, m.loginPath, http.StatusSeeOther)
		return
	}

	m.setSessionCookie(w, r, session.ID)
	redirectTarget := m.sanitizeRedirectTarget(r, next)
	if redirectTarget == "" {
		redirectTarget = consts.RouteRoot
	}
	l.Debug().Str(xlog.FieldNext, redirectTarget).Msg("completed login flow")

	http.Redirect(w, r, redirectTarget, http.StatusSeeOther)
}

func (m *authMiddleware) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}

	l := zerolog.Ctx(r.Context())

	sessionID, ok := m.sessionIDFromRequest(r)
	if !ok {
		l.Debug().Msg("refresh requested without session cookie")
		http.Redirect(w, r, consts.RouteRoot, http.StatusSeeOther)
		return
	}

	session, err := m.svc.RefreshSession(r.Context(), sessionID)
	if err != nil {
		m.clearSessionCookie(w)
		l.Warn().Err(err).Msg("failed to refresh session")
		if errors.Is(err, errNotFound) || errors.Is(err, errReauthenticationRequired) {
			http.Redirect(w, r, consts.RouteRoot, http.StatusSeeOther)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	m.setSessionCookie(w, r, session.ID)
	redirectTarget := m.sanitizeRedirectTarget(r, r.URL.Query().Get(paramNext))
	if redirectTarget == "" {
		redirectTarget = m.sanitizeRedirectTarget(r, r.Referer())
	}
	if redirectTarget == "" {
		redirectTarget = consts.RouteRoot
	}
	l.Debug().Str(xlog.FieldNext, redirectTarget).Msg("refreshed session")

	http.Redirect(w, r, redirectTarget, http.StatusSeeOther)
}

func (m *authMiddleware) handleLogout(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}

	l := zerolog.Ctx(r.Context())

	if sessionID, ok := m.sessionIDFromRequest(r); ok {
		if err := m.svc.Logout(r.Context(), sessionID); err != nil {
			l.Error().Err(err).Msg("failed to log out session")
		} else {
			l.Debug().Msg("logged out session")
		}
	}
	m.clearSessionCookie(w)
	http.Redirect(w, r, consts.RouteRoot, http.StatusSeeOther)
}

func requirePost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}

	w.Header().Set(consts.HeaderAllow, http.MethodPost)
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	return false
}

func (m *authMiddleware) redirectToLogin(w http.ResponseWriter, r *http.Request, next string) {
	target := m.loginPath
	if next != "" {
		target = m.loginPath + "?" + paramNext + "=" + url.QueryEscape(next)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (m *authMiddleware) currentRequestTarget(r *http.Request) string {
	if r == nil || r.URL == nil {
		return consts.RouteRoot
	}
	target := r.URL.Path
	if target == "" {
		target = consts.RouteRoot
	}
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	return target
}

func (m *authMiddleware) sessionIDFromRequest(r *http.Request) (string, bool) {
	if m.sessionCookieName == "" {
		return "", false
	}
	cookie, err := r.Cookie(m.sessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

func (m *authMiddleware) setSessionCookie(w http.ResponseWriter, r *http.Request, value string) {
	expires := time.Now().Add(m.sessionTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     m.sessionCookieName,
		Value:    value,
		Path:     consts.RouteRoot,
		Expires:  expires,
		MaxAge:   int((m.sessionTTL + time.Second - 1) / time.Second),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.secureCookies || requestIsHTTPS(r, m.trustedProxies),
	})
}

func (m *authMiddleware) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.sessionCookieName,
		Value:    "",
		Path:     consts.RouteRoot,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *authMiddleware) loginStateCookieName() string {
	return m.sessionCookieName + "-login"
}

func (m *authMiddleware) setLoginStateCookie(w http.ResponseWriter, r *http.Request, loginID string) {
	expires := time.Now().Add(m.loginStateTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     m.loginStateCookieName(),
		Value:    loginID,
		Path:     consts.RouteRoot,
		Expires:  expires,
		MaxAge:   int((m.loginStateTTL + time.Second - 1) / time.Second),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.secureCookies || requestIsHTTPS(r, m.trustedProxies),
	})
}

func (m *authMiddleware) clearLoginStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.loginStateCookieName(),
		Value:    "",
		Path:     consts.RouteRoot,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *authMiddleware) loginStateMatchesCookie(r *http.Request, loginID string) bool {
	cookie, err := r.Cookie(m.loginStateCookieName())
	if err != nil || cookie.Value == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(loginID)) == 1
}

func (m *authMiddleware) isPublicPath(path string) bool {
	for _, publicPath := range m.publicPaths {
		if publicPath == "" {
			continue
		}
		matched, err := doublestar.Match(publicPath, path)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func (m *authMiddleware) hasRegisteredHandler(r *http.Request) bool {
	rctx := chi.RouteContext(r.Context())
	probe := chi.NewRouteContext()

	probe.Routes = rctx.Routes
	routePath := rctx.RoutePath
	if routePath == "" {
		routePath = r.URL.Path
	}

	return probe.Routes.Match(probe, r.Method, routePath)
}

func requestIsHTTPS(r *http.Request, trustedProxies *TrustedProxies) bool {
	if r == nil {
		return false
	}
	if r.TLS != nil {
		return true
	}
	if !trustedProxies.AllowsForwardedHeaders(r) {
		return false
	}
	return strings.EqualFold(headerFirstValue(r.Header.Get(consts.HeaderXForwardedProto)), consts.SchemeHTTPS)
}

func (m *authMiddleware) sanitizeRedirectTarget(r *http.Request, candidate string) string {
	if candidate == "" {
		return ""
	}

	parsed, err := url.Parse(candidate)
	if err != nil {
		return ""
	}

	switch {
	case parsed.IsAbs():
		if r == nil {
			return ""
		}
		host := m.requestHost(r)
		if host == "" || !strings.EqualFold(parsed.Host, host) {
			return ""
		}
	case strings.HasPrefix(candidate, "/") && !strings.HasPrefix(candidate, "//"):
	default:
		return ""
	}

	target := parsed.EscapedPath()
	if target == "" {
		target = consts.RouteRoot
	}
	if !strings.HasPrefix(target, "/") || strings.HasPrefix(target, "//") || strings.Contains(target, `\`) {
		return ""
	}

	if parsed.RawQuery != "" {
		target += "?" + parsed.RawQuery
	}
	return target
}

func (m *authMiddleware) requestHost(r *http.Request) string {
	if r == nil {
		return ""
	}
	if m.trustedProxies.AllowsForwardedHeaders(r) {
		if host := headerFirstValue(r.Header.Get(consts.HeaderXForwardedHost)); host != "" {
			return host
		}
	}
	return r.Host
}
