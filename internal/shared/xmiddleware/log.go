package xmiddleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// Logger is a middleware that attaches a [zerolog.Logger] to the request context, with request-specific fields.
// It also logs both the start and end of each request.
//
// Logger can be retrieved from the request context in handlers using [zerolog.Ctx].
func Logger(l zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rl := l.With().
				Str("id", middleware.GetReqID(r.Context())).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("client_ip", ClientIP(r)).
				Logger()

			ctx := rl.WithContext(r.Context())

			start := time.Now()
			rl.Trace().Msg("request received")
			next.ServeHTTP(w, r.WithContext(ctx))
			duration := time.Since(start)

			rl.Debug().
				Dur("duration", duration).
				Msg("request completed")
		})
	}
}
