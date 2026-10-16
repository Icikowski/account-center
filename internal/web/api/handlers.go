package api

import (
	_ "embed" // Required for embedding the OpenAPI document.
	"encoding/json"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"git.sr.ht/~icikowski/account-center/internal/auth"
	"git.sr.ht/~icikowski/account-center/internal/consts"
	"git.sr.ht/~icikowski/account-center/internal/evaluator"
	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
	"git.sr.ht/~icikowski/account-center/internal/shared/xmiddleware"
)

//go:embed openapi.yaml
var openAPISpec []byte

type apiHandler struct {
	catalogSource model.Reloader[model.Catalog]
	evaluator     evaluator.Evaluator
}

// NewHandler creates a new JSON API handler. The routes expect the session (if any) to be already placed in the
// request context by the authentication middleware.
func NewHandler(
	catalogSource model.Reloader[model.Catalog],
	evaluator evaluator.Evaluator,
) xhttp.RouteBinder {
	return &apiHandler{
		catalogSource: catalogSource,
		evaluator:     evaluator,
	}
}

// Bind implements [xhttp.RouteBinder].
func (h *apiHandler) Bind(r chi.Router) {
	r.With(
		xmiddleware.NoStore,
	).Route(consts.RouteAPIV1, func(r chi.Router) {
		r.Get(consts.RouteAPISession, h.session)
		r.Get(consts.RouteAPIServices, h.services)
	})

	r.Get(consts.RouteAPIOpenAPI, h.openAPI)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusNotFound, "not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed")
	})
}

func (h *apiHandler) openAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(consts.HeaderContentType, consts.MIMEApplicationYAML)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(openAPISpec); err != nil {
		l := zerolog.Ctx(r.Context())
		l.Error().Err(err).Msg("failed to write OpenAPI document response")
	}
}

func (h *apiHandler) session(w http.ResponseWriter, r *http.Request) {
	session := auth.FromContext(r.Context())
	if session == nil {
		writeJSON(w, r, http.StatusOK, SessionResponse{})
		return
	}

	writeJSON(w, r, http.StatusOK, SessionResponse{
		Authenticated: true,
		User: &User{
			Name:    session.User.Name,
			Email:   session.User.Email,
			Picture: session.User.Picture,
		},
	})
}

func (h *apiHandler) services(w http.ResponseWriter, r *http.Request) {
	session := auth.FromContext(r.Context())
	if session == nil {
		writeError(w, r, http.StatusUnauthorized, "authentication required")
		return
	}

	effective := h.evaluator.Evaluate(r.Context(), h.catalogSource, session.User)
	sort.SliceStable(effective, func(i, j int) bool {
		return effective[i].Name < effective[j].Name
	})

	services := make([]Service, 0, len(effective))
	for _, es := range effective {
		svc := Service{
			Name: es.Name,
			URL:  es.URL.String(),
			Role: string(es.EffectiveRole),
		}
		if es.Icon != nil {
			svc.Icon = es.Icon.String()
		}
		services = append(services, svc)
	}

	writeJSON(w, r, http.StatusOK, ServicesResponse{Services: services})
}

func writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	writeJSON(w, r, status, ErrorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	w.Header().Set(consts.HeaderContentType, consts.MIMEApplicationJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		l := zerolog.Ctx(r.Context())
		l.Error().Err(err).Msg("failed to encode API response")
	}
}
