package xmiddleware

import (
	"net/http"

	"git.sr.ht/~icikowski/account-center/internal/consts"
)

// StaticResourcesCacheControl returns a middleware that sets the Cache-Control for static
// resources to public with a max-age of one year (31536000 seconds).
func StaticResourcesCacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(consts.HeaderCacheControl, "public, max-age=31536000")
		next.ServeHTTP(w, r)
	})
}
