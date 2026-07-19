package xmiddleware

import (
	"context"
	"net"
	"net/http"

	"git.sr.ht/~icikowski/account-center/internal/auth"
	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
)

type contextKeyClientIP struct{}

// RealIP resolves the client IP from trusted proxy headers and stores it in the request context.
func RealIP(trustedProxies *auth.TrustedProxies) xhttp.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := trustedProxies.ClientIP(r)
			if clientIP == nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyClientIP{}, clientIP.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClientIP retrieves the resolved client IP from the request context.
func ClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	if clientIP, ok := r.Context().Value(contextKeyClientIP{}).(string); ok && clientIP != "" {
		return clientIP
	}

	return requestIP(r.RemoteAddr)
}

func requestIP(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		remoteAddr = host
	}
	if ip := net.ParseIP(remoteAddr); ip != nil {
		return ip.String()
	}
	return ""
}
