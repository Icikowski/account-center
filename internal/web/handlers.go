package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"git.sr.ht/~icikowski/account-center/internal/auth"
	"git.sr.ht/~icikowski/account-center/internal/evaluator"
	"git.sr.ht/~icikowski/account-center/internal/i18n"
	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xmiddleware"
	"git.sr.ht/~icikowski/account-center/internal/web/assets"
	"git.sr.ht/~icikowski/account-center/internal/web/ui"
	"git.sr.ht/~icikowski/account-center/internal/web/webmanifest"
)

const (
	routeRoot        = "/"
	routeAssets      = "/assets"
	routeWebManifest = "/manifest.webmanifest"
	routeLogin       = "/login"
	routeRefresh     = "/refresh"
	routeLogout      = "/logout"
)

// NewHandler creates a new HTTP handler for the application, setting up routes for static assets and the UI.
func NewHandler(
	instanceName string,
	catalogSource model.Reloader[model.Catalog],
	knowledgeBaseSource model.Reloader[model.KnowledgeBase],
	authService auth.Service,
	sessionCookieName string,
	trustedProxies *auth.TrustedProxies,
	evaluator evaluator.Evaluator,
) http.Handler {
	assetsHandler := assets.NewHandler(routeAssets)
	webManifestHandler := webmanifest.NewHandler(instanceName, routeAssets)
	uiHandler := ui.NewHandler(instanceName, catalogSource, knowledgeBaseSource, evaluator)

	authMiddleware := auth.NewMiddleware(
		authService,
		sessionCookieName,
		routeLogin,
		routeRefresh,
		routeLogout,
		trustedProxies,
		routeRoot,
	)

	r := chi.NewRouter()
	r.Use(
		middleware.RealIP,
		middleware.CleanPath,
		middleware.Recoverer,
	)

	r.With(
		xmiddleware.StaticResourcesCacheControl,
	).Route(routeAssets, assetsHandler.Bind)

	r.With(
		xmiddleware.I18n(i18n.NewBundle()),
	).Route(routeWebManifest, webManifestHandler.Bind)

	r.With(
		authMiddleware.Middleware,
	).Route(routeRoot, uiHandler.Bind)

	return r
}
