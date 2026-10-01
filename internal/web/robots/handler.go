package robots

import (
	_ "embed"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"git.sr.ht/~icikowski/account-center/internal/consts"
	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
)

//go:embed robots.txt
var robotsContent []byte

type robotsHandler struct{}

// NewHandler creates a new handler for serving the robots.txt file.
func NewHandler() xhttp.RouteBinder {
	return &robotsHandler{}
}

// Bind implements [xhttp.RouteBinder].
func (h *robotsHandler) Bind(r chi.Router) {
	r.HandleFunc("/*", h.robots)
}

func (h *robotsHandler) robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(consts.HeaderContentType, consts.MIMETextPlain)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(robotsContent); err != nil {
		l := zerolog.Ctx(r.Context())
		l.Error().Err(err).Msg("failed to write robots.txt response")
	}
}
