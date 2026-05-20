package ui

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"

	"git.sr.ht/~icikowski/account-center/internal/auth"
	"git.sr.ht/~icikowski/account-center/internal/evaluator"
	"git.sr.ht/~icikowski/account-center/internal/i18n"
	"git.sr.ht/~icikowski/account-center/internal/knowledgebase"
	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
	"git.sr.ht/~icikowski/account-center/internal/shared/xmiddleware"
	"git.sr.ht/~icikowski/account-center/internal/web/layouts"
	"git.sr.ht/~icikowski/account-center/internal/web/templates"
)

const (
	routeRoot                     = "/"
	routeCatalog                  = "/catalog"
	routeKnowledgeBase            = "/kb"
	routeKnowledgeBaseAttachments = routeKnowledgeBase + "/attachments"
)

type uiHandler struct {
	bundle *goi18n.Bundle

	instanceName        string
	catalogSource       model.Reloader[model.Catalog]
	knowledgeBaseSource model.Reloader[model.KnowledgeBase]
	evaluator           evaluator.Evaluator
}

// NewHandler creates a new web UI handler with the provided instance name, catalog source, and knowledge base source.
func NewHandler(
	instanceName string,
	catalogSource model.Reloader[model.Catalog],
	knowledgeBaseSource model.Reloader[model.KnowledgeBase],
	evaluator evaluator.Evaluator,
) xhttp.RouteBinder {
	return &uiHandler{
		bundle:              i18n.NewBundle(),
		instanceName:        instanceName,
		catalogSource:       catalogSource,
		knowledgeBaseSource: knowledgeBaseSource,
		evaluator:           evaluator,
	}
}

// Bind implements [xhttp.RouteBinder].
func (h *uiHandler) Bind(r chi.Router) {
	r.Use(
		xmiddleware.I18n(h.bundle),
		xmiddleware.SidebarState,
		h.baseDataMiddleware,
	)

	r.HandleFunc(routeRoot, h.splash)
	r.HandleFunc(routeCatalog, h.catalog)
	if h.knowledgeBaseSource != nil {
		r.Route(routeKnowledgeBaseAttachments, knowledgebase.NewAttachmentsHandler(h.knowledgeBaseSource).Bind)
		r.Route(routeKnowledgeBase, knowledgebase.NewArticleHandler(
			h.knowledgeBaseSource,
			routeKnowledgeBase,
			routeKnowledgeBaseAttachments,
			h.knowledgeBaseListing,
			h.knowledgeBaseArticle,
			knowledgebase.WithNotFoundHandler(h.notFound),
		).Bind)
	} else {
		r.HandleFunc(routeKnowledgeBase, h.knowledgeBaseDisabled)
	}

	r.NotFound(h.notFound)
	r.MethodNotAllowed(h.methodNotAllowed)
}

func (h *uiHandler) splash(w http.ResponseWriter, r *http.Request) {
	session := auth.FromContext(r.Context())
	if session != nil {
		http.Redirect(w, r, routeCatalog, http.StatusSeeOther)
		return
	}

	renderTemplate(w, r, http.StatusOK, templates.Login())
}

func (h *uiHandler) catalog(w http.ResponseWriter, r *http.Request) {
	session := auth.FromContext(r.Context())

	es := h.evaluator.Evaluate(h.catalogSource, session.User)
	sort.SliceStable(es, func(i, j int) bool {
		return es[i].Name < es[j].Name
	})

	renderTemplate(w, r, http.StatusOK, templates.Catalog(es))
}

func (h *uiHandler) knowledgeBaseListing(
	w http.ResponseWriter,
	r *http.Request,
	articles []model.KnowledgeBaseListingArticle,
) {
	renderTemplate(w, r, http.StatusOK, templates.KnowledgeBase(articles))
}

func (h *uiHandler) knowledgeBaseArticle(
	w http.ResponseWriter,
	r *http.Request,
	article model.KnowledgeBaseArticle,
	body template.HTML,
) {
	renderTemplate(w, r, http.StatusOK, templates.KnowledgeBaseArticle(article, body))
}

func (h *uiHandler) knowledgeBaseDisabled(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, http.StatusOK, templates.KnowledgeBaseDisabled())
}

func (h *uiHandler) notFound(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, http.StatusNotFound, templates.ErrorNotFound())
}

func (h *uiHandler) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, http.StatusMethodNotAllowed, templates.ErrorMethodNotAllowed())
}

func (h *uiHandler) baseDataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx      = r.Context()
			session  = auth.FromContext(ctx)
			baseData = layouts.BaseData{InstanceName: h.instanceName}
		)

		if session != nil {
			effectiveServices := len(h.evaluator.Evaluate(h.catalogSource, session.User))
			effectiveKnowledgeBaseArticles := 0
			if h.knowledgeBaseSource != nil {
				effectiveKnowledgeBaseArticles = len(h.knowledgeBaseSource.Snapshot().Articles)
			}

			baseData.Counters = layouts.Counters{
				Services:              effectiveServices,
				KnowledgeBaseArticles: effectiveKnowledgeBaseArticles,
			}
			baseData.User = &layouts.User{
				FullName: session.User.Name,
				Email:    session.User.Email,
			}
		}

		ctx = layouts.NewBaseDataContext(ctx, baseData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
