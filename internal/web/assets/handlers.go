package assets

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"git.sr.ht/~icikowski/account-center/internal/consts"
	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
)

//go:embed css/* js/* img/* fonts/*
var assetsFS embed.FS

// assetVersion changes whenever the mutable assets change, so the long-lived
// cache headers cannot pin an old stylesheet or page script.
var assetVersion = func() string {
	hasher := sha256.New()
	for _, name := range []string{"css/style.css", "js/page.js"} {
		content, err := assetsFS.ReadFile(name)
		if err != nil {
			continue
		}
		_, _ = hasher.Write(content)
	}
	return hex.EncodeToString(hasher.Sum(nil)[:6])
}()

// URL returns the public URL of an embedded asset with a cache-busting version.
func URL(path string) string {
	return consts.RouteAssets + "/" + strings.TrimPrefix(path, "/") + "?v=" + assetVersion
}

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
