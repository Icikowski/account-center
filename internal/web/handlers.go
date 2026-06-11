package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"git.sr.ht/~icikowski/account-center/internal/auth"
	"git.sr.ht/~icikowski/account-center/internal/consts"
	"git.sr.ht/~icikowski/account-center/internal/evaluator"
	"git.sr.ht/~icikowski/account-center/internal/i18n"
	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xmiddleware"
	"git.sr.ht/~icikowski/account-center/internal/web/assets"
	"git.sr.ht/~icikowski/account-center/internal/web/ui"
	"git.sr.ht/~icikowski/account-center/internal/web/webmanifest"
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
	log zerolog.Logger,
) http.Handler {
	assetsHandler := assets.NewHandler(consts.RouteAssets)
	webManifestHandler := webmanifest.NewHandler(instanceName, consts.RouteAssets)
	uiHandler := ui.NewHandler(instanceName, catalogSource, knowledgeBaseSource, evaluator)

	authMiddleware := auth.NewMiddleware(
		authService,
		sessionCookieName,
		consts.RouteOIDCCallback,
		consts.RouteLogin,
		consts.RouteRefresh,
		consts.RouteLogout,
		trustedProxies,
		consts.RouteRoot,
	)

	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		xmiddleware.RealIP(trustedProxies),
		middleware.CleanPath,
		xmiddleware.Logger(log),
		middleware.Recoverer,
	)

	r.With(
		xmiddleware.StaticResourcesCacheControl,
	).Route(consts.RouteAssets, assetsHandler.Bind)

	r.With(
		xmiddleware.I18n(i18n.NewBundle()),
	).Route(consts.RouteWebManifest, webManifestHandler.Bind)

	r.With(
		authMiddleware.Middleware,
	).Route(consts.RouteRoot, uiHandler.Bind)

	return r
}
