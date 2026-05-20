package assets

import (
	"embed"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
)

//go:embed css/* js/* img/*
var assetsFS embed.FS

type assetsHandler struct {
	path string
}

// NewHandler creates a new static file handler instance that serves files from the embedded filesystem.
func NewHandler(path string) xhttp.RouteBinder {
	return &assetsHandler{
		path: path,
	}
}

// Bind implements [xhttp.RouteBinder].
func (h *assetsHandler) Bind(r chi.Router) {
	r.With(
		middleware.StripPrefix(h.path),
	).Handle("/*", http.FileServerFS(assetsFS))
}
