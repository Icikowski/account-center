package xmiddleware

import (
	"context"
	"net/http"
)

const (
	cookieSidebarState string = "sidebar_state"
)

type contextKeySidebarState struct{}

// SidebarState is a middleware that retrieves the sidebar collapsed state from the cookie and stores it in the request
// context.
func SidebarState(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(cookieSidebarState)
		collapsed := cookie != nil && cookie.Value == "false"
		ctx := context.WithValue(r.Context(), contextKeySidebarState{}, collapsed)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsSidebarCollapsed retrieves the sidebar collapsed state from the context.
func IsSidebarCollapsed(ctx context.Context) bool {
	v, _ := ctx.Value(contextKeySidebarState{}).(bool)
	return v
}
