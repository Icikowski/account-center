package xmiddleware

import (
	"net/http"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"

	"git.sr.ht/~icikowski/account-center/internal/consts"
	"git.sr.ht/~icikowski/account-center/internal/i18n"
)

// I18n is a middleware that sets up internationalization for the request context.
func I18n(bundle *goi18n.Bundle) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			langs := make([]string, 0, 2)

			if c, err := r.Cookie(consts.CookieUserLanguage); err == nil && c.Value != "" {
				langs = append(langs, c.Value)
			}
			if accept := r.Header.Get(consts.HeaderAcceptLanguage); accept != "" {
				langs = append(langs, accept)
			}

			if len(langs) == 0 {
				langs = append(langs, "en")
			}

			l := goi18n.NewLocalizer(bundle, langs...)
			ctx := i18n.NewContext(r.Context(), l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
