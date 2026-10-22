package xmiddleware

import (
	"net/http"

	"git.sr.ht/~icikowski/account-center/internal/consts"
)

const (
	contentSecurityPolicy = "base-uri 'self'; object-src 'none'; frame-ancestors 'none'; form-action 'self'"
	permissionsPolicy     = "camera=(), microphone=(), geolocation=()"
	referrerPolicy        = "no-referrer"
	contentTypeOptions    = "nosniff"
	frameOptions          = "DENY"
)

// SecurityHeaders sets security-related response headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set(consts.HeaderContentSecurityPolicy, contentSecurityPolicy)
		header.Set(consts.HeaderPermissionsPolicy, permissionsPolicy)
		header.Set(consts.HeaderReferrerPolicy, referrerPolicy)
		header.Set(consts.HeaderXContentTypeOptions, contentTypeOptions)
		header.Set(consts.HeaderXFrameOptions, frameOptions)

		next.ServeHTTP(w, r)
	})
}
