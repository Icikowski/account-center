package health

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"git.sr.ht/~icikowski/account-center/internal/consts"
	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
	"git.sr.ht/~icikowski/account-center/internal/store"
)

type healthHandler struct {
	catalogSource       model.Reloader[model.Catalog]
	knowledgeBaseSource model.Reloader[model.KnowledgeBase]
	storage             store.StorageBackend
}

// NewHandler creates a new health check handler with the provided dependencies to check.
func NewHandler(
	catalogSource model.Reloader[model.Catalog],
	knowledgeBaseSource model.Reloader[model.KnowledgeBase],
	storage store.StorageBackend,
) xhttp.RouteBinder {
	return &healthHandler{
		catalogSource:       catalogSource,
		knowledgeBaseSource: knowledgeBaseSource,
		storage:             storage,
	}
}

func (h *healthHandler) Bind(r chi.Router) {
	r.HandleFunc(consts.RouteLive, h.live)
	r.HandleFunc(consts.RouteReady, h.ready)
}

func (h *healthHandler) live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *healthHandler) ready(w http.ResponseWriter, r *http.Request) {
	checks := make([]string, 0, 3)
	if h.catalogSource == nil || h.catalogSource.Current().Revision == 0 {
		checks = append(checks, "catalog not ready")
	}

	if h.knowledgeBaseSource != nil && h.knowledgeBaseSource.Current().Revision == 0 {
		checks = append(checks, "knowledge base not ready")
	}

	if err := h.storage.Ping(r.Context()); err != nil {
		checks = append(checks, "storage backend not ready: "+err.Error())
	}

	if len(checks) > 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(checks); err != nil {
			l := zerolog.Ctx(r.Context())
			l.Error().Err(err).Msg("failed to encode health check response")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
