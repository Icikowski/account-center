package auth

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"git.sr.ht/~icikowski/account-center/internal/consts"
)

// Middleware defines the interface for the authentication middleware provider.
type Middleware interface {
	Middleware(next http.Handler) http.Handler
}

type authMiddleware struct {
	svc               Service
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
		sessionCookieName: sessionCookieName,
		oidcCallbackPath:  oidcCallbackPath,
		loginPath:         loginPath,
		refreshPath:       refreshPath,
		logoutPath:        logoutPath,
		publicPaths:       publicPaths,
		trustedProxies:    trustedProxies,
	}
}

// Middleware implements [Middleware].
func (m *authMiddleware) Middleware(next http.Handler) http.Handler {
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

		if m.isPublicPath(r.URL.Path) {
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

func (m *authMiddleware) withSessionContext(
	w http.ResponseWriter,
	r *http.Request,
	required bool,
) (*http.Request, bool) {
	sessionID, ok := m.sessionIDFromRequest(r)
	if !ok {
		if required {
			m.redirectToLogin(w, r, m.currentRequestTarget(r))
			return nil, false
		}
		return r, true
	}

	session, err := m.svc.GetSession(r.Context(), sessionID)
	if err != nil {
		m.clearSessionCookie(w)
		if required {
			if errors.Is(err, errNotFound) || errors.Is(err, errReauthenticationRequired) {
				m.redirectToLogin(w, r, m.currentRequestTarget(r))
				return nil, false
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil, false
		}
		return r, true
	}

	ctx := NewContext(r.Context(), &session)
	return r.WithContext(ctx), true
}

func (m *authMiddleware) handleLogin(w http.ResponseWriter, r *http.Request) {
	if sessionID, ok := m.sessionIDFromRequest(r); ok {
		if _, err := m.svc.GetSession(r.Context(), sessionID); err == nil {
			http.Redirect(w, r, consts.RouteRoot, http.StatusSeeOther)
			return
		}
		m.clearSessionCookie(w)
	}

	next := m.sanitizeRedirectTarget(r, r.URL.Query().Get(paramNext))
	if next == "" {
		next = consts.RouteRoot
	}

	authReq, err := m.svc.AuthorizationRequest(r.Context(), r, next)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authReq.URL, http.StatusSeeOther)
}

func (m *authMiddleware) handleCallback(w http.ResponseWriter, r *http.Request) {
	loginID := r.URL.Query().Get(paramState)
	code := r.URL.Query().Get(paramCode)
	if loginID == "" || code == "" {
		http.Redirect(w, r, m.loginPath, http.StatusSeeOther)
		return
	}

	session, next, err := m.svc.ExchangeCode(r.Context(), loginID, code)
	if err != nil {
		http.Redirect(w, r, m.loginPath, http.StatusSeeOther)
		return
	}

	m.setSessionCookie(w, r, session.ID)
	redirectTarget := m.sanitizeRedirectTarget(r, next)
	if redirectTarget == "" {
		redirectTarget = consts.RouteRoot
	}

	http.Redirect(w, r, redirectTarget, http.StatusSeeOther)
}

func (m *authMiddleware) handleRefresh(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := m.sessionIDFromRequest(r)
	if !ok {
		http.Redirect(w, r, consts.RouteRoot, http.StatusSeeOther)
		return
	}

	session, err := m.svc.RefreshSession(r.Context(), sessionID)
	if err != nil {
		m.clearSessionCookie(w)
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

	http.Redirect(w, r, redirectTarget, http.StatusSeeOther)
}

func (m *authMiddleware) handleLogout(w http.ResponseWriter, r *http.Request) {
	if sessionID, ok := m.sessionIDFromRequest(r); ok {
		_ = m.svc.Logout(r.Context(), sessionID)
	}
	m.clearSessionCookie(w)
	http.Redirect(w, r, consts.RouteRoot, http.StatusSeeOther)
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
	http.SetCookie(w, &http.Cookie{
		Name:     m.sessionCookieName,
		Value:    value,
		Path:     consts.RouteRoot,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   requestIsHTTPS(r, m.trustedProxies),
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
		if host == "" {
			return ""
		}
		if !strings.EqualFold(parsed.Host, host) {
			return ""
		}
		target := parsed.EscapedPath()
		if target == "" {
			target = consts.RouteRoot
		}
		if parsed.RawQuery != "" {
			target += "?" + parsed.RawQuery
		}
		return target
	case strings.HasPrefix(candidate, "/") && !strings.HasPrefix(candidate, "//"):
		return candidate
	default:
		return ""
	}
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
