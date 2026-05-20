package knowledgebase

import (
	"bytes"
	"html/template"
	"net/http"
	"path"
	"strconv"

	"github.com/go-chi/chi/v5"

	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xhttp"
)

// ArticleHandlerFunc handles a resolved [model.KnowledgeBaseArticle] and its rendered [template.HTML] body.
type ArticleHandlerFunc func(
	w http.ResponseWriter,
	r *http.Request,
	article model.KnowledgeBaseArticle,
	body template.HTML,
)

// ListingHandlerFunc handles the knowledge base article listing root.
type ListingHandlerFunc func(
	w http.ResponseWriter,
	r *http.Request,
	articles []model.KnowledgeBaseListingArticle,
)

// ArticleHandlerOption configures the behavior of [NewArticleHandler].
type ArticleHandlerOption interface {
	apply(h *articleHandler)
}

type articleHandlerOptionFunc func(h *articleHandler)

func (f articleHandlerOptionFunc) apply(h *articleHandler) {
	f(h)
}

// WithNotFoundHandler configures a custom not found handler for [NewArticleHandler].
func WithNotFoundHandler(fn http.HandlerFunc) ArticleHandlerOption {
	return articleHandlerOptionFunc(func(h *articleHandler) {
		h.notFoundHandler = fn
	})
}

type articleHandler struct {
	source             model.Reloader[model.KnowledgeBase]
	articleBasePath    string
	attachmentBasePath string
	notFoundHandler    http.HandlerFunc
	listingHandler     ListingHandlerFunc
	handler            ArticleHandlerFunc
}

// NewArticleHandler creates a route binder for serving knowledge base articles.
func NewArticleHandler(
	source model.Reloader[model.KnowledgeBase],
	articleBasePath string,
	attachmentBasePath string,
	listingHandler ListingHandlerFunc,
	handler ArticleHandlerFunc,
	opts ...ArticleHandlerOption,
) xhttp.RouteBinder {
	h := &articleHandler{
		source:             source,
		articleBasePath:    articleBasePath,
		attachmentBasePath: attachmentBasePath,
		notFoundHandler:    http.NotFound,
		listingHandler:     listingHandler,
		handler:            handler,
	}

	for _, opt := range opts {
		opt.apply(h)
	}

	return h
}

// Bind implements [xhttp.RouteBinder].
func (h *articleHandler) Bind(r chi.Router) {
	r.Get("/", h.article)
	r.Get("/*", h.article)
}

func (h *articleHandler) article(w http.ResponseWriter, r *http.Request) {
	current := h.source.Current()
	slug, redirected := canonicalSlug(chi.URLParam(r, "*"))
	if redirected {
		http.Redirect(w, r, mountedPath(h.articleBasePath, slug), http.StatusMovedPermanently)
		return
	}
	if slug == "" {
		if h.listingHandler == nil {
			http.NotFound(w, r)
			return
		}

		h.listingHandler(
			w,
			r,
			resolveListingArticles(
				current.Value.Articles,
				h.articleBasePath,
				mountedAssetPath(h.attachmentBasePath, current.Revision),
			),
		)
		return
	}

	article, ok := current.Value.ArticleBySlug(slug)
	if !ok {
		h.notFoundHandler(w, r)
		return
	}

	h.handler(
		w,
		r,
		article,
		article.ResolveRenderedBody(
			h.articleBasePath,
			mountedAssetPath(h.attachmentBasePath, current.Revision),
		),
	)
}

type attachmentsHandler struct {
	source model.Reloader[model.KnowledgeBase]
}

// NewAttachmentsHandler creates a route binder for serving referenced knowledge base assets.
func NewAttachmentsHandler(source model.Reloader[model.KnowledgeBase]) xhttp.RouteBinder {
	return &attachmentsHandler{
		source: source,
	}
}

// Bind implements [xhttp.RouteBinder].
func (h *attachmentsHandler) Bind(r chi.Router) {
	r.Get("/{revision}/{id}", h.asset)
}

func (h *attachmentsHandler) asset(w http.ResponseWriter, r *http.Request) {
	current := h.source.Current()
	if chi.URLParam(r, "revision") != strconv.FormatUint(current.Revision, 10) {
		http.NotFound(w, r)
		return
	}

	asset, ok := current.Value.AssetByID(chi.URLParam(r, "id"))
	if !ok {
		http.NotFound(w, r)
		return
	}

	if asset.ContentType != "" {
		w.Header().Set("Content-Type", asset.ContentType)
	}

	http.ServeContent(
		w,
		r,
		path.Base(asset.SourceRelPath),
		asset.LastModified,
		bytes.NewReader(asset.Content),
	)
}

func canonicalSlug(routeParam string) (string, bool) {
	if routeParam == "" {
		return "", false
	}

	rawSlug := "/" + routeParam
	slug := path.Clean(rawSlug)
	if slug == "/" {
		return "", rawSlug != slug
	}

	return slug, rawSlug != slug
}

func mountedPath(basePath string, slug string) string {
	basePath = path.Clean("/" + basePath)
	switch basePath {
	case ".", "/":
		return slug
	default:
		return basePath + slug
	}
}

func mountedAssetPath(basePath string, revision uint64) string {
	return mountedPath(basePath, "/"+strconv.FormatUint(revision, 10))
}

func resolveListingArticles(
	articles []model.KnowledgeBaseArticle,
	articleBasePath string,
	attachmentBasePath string,
) []model.KnowledgeBaseListingArticle {
	items := make([]model.KnowledgeBaseListingArticle, len(articles))
	for i, article := range articles {
		items[i] = model.KnowledgeBaseListingArticle{
			Article:          article,
			URL:              article.ResolveURL(articleBasePath),
			FeaturedImageURL: article.ResolveFeaturedImage(attachmentBasePath),
		}
	}
	return items
}
