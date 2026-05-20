package model

import (
	"html/template"
	"strings"
	"time"
)

const (
	knowledgeBaseArticleURLScheme = "kb-article://"
	knowledgeBaseAssetURLScheme   = "kb-asset://"
)

// KnowledgeBase represents a rendered, immutable knowledge base snapshot.
type KnowledgeBase struct {
	BasePath string
	Articles []KnowledgeBaseArticle
	Assets   []KnowledgeBaseAsset

	articlesBySlug map[string]KnowledgeBaseArticle
	assetsByID     map[string]KnowledgeBaseAsset
}

// KnowledgeBaseArticle represents a single knowledge base article.
type KnowledgeBaseArticle struct {
	Title              string
	Description        string
	Slug               string
	SourcePath         string
	SourceRelPath      string
	FeaturedImage      string
	RenderedBody       string
	LinkedArticleSlugs []string
	AssetIDs           []string
}

// ResolveRenderedBody resolves route-local article and attachment URLs for serving.
func (p KnowledgeBaseArticle) ResolveRenderedBody(
	articleBasePath string,
	attachmentBasePath string,
) template.HTML {
	body := strings.ReplaceAll(
		p.RenderedBody,
		knowledgeBaseArticleURLScheme,
		normalizeMountedBase(articleBasePath),
	)
	body = strings.ReplaceAll(
		body,
		knowledgeBaseAssetURLScheme,
		normalizeMountedBase(attachmentBasePath)+"/",
	)
	return template.HTML(body)
}

// ResolveURL resolves the article URL for the mounted knowledge base root.
func (p KnowledgeBaseArticle) ResolveURL(articleBasePath string) string {
	basePath := normalizeMountedBase(articleBasePath)
	if basePath == "" {
		return p.Slug
	}

	return basePath + p.Slug
}

// ResolveFeaturedImage resolves the featured image URL for the mounted attachment root.
func (p KnowledgeBaseArticle) ResolveFeaturedImage(attachmentBasePath string) string {
	if p.FeaturedImage == "" {
		return ""
	}

	return strings.ReplaceAll(
		p.FeaturedImage,
		knowledgeBaseAssetURLScheme,
		normalizeMountedBase(attachmentBasePath)+"/",
	)
}

// KnowledgeBaseListingArticle represents a article as exposed to a listing handler.
type KnowledgeBaseListingArticle struct {
	Article          KnowledgeBaseArticle
	URL              string
	FeaturedImageURL string
}

// KnowledgeBaseAsset represents a referenced asset served alongside the knowledge base.
type KnowledgeBaseAsset struct {
	ID            string
	SourcePath    string
	SourceRelPath string
	ContentType   string
	Content       []byte
	LastModified  time.Time
}

// NewKnowledgeBase creates a [KnowledgeBase] with lookup indexes.
func NewKnowledgeBase(
	basePath string,
	articles []KnowledgeBaseArticle,
	assets []KnowledgeBaseAsset,
) KnowledgeBase {
	articlesBySlug := make(map[string]KnowledgeBaseArticle, len(articles))
	for _, article := range articles {
		articlesBySlug[article.Slug] = article
	}

	assetsByID := make(map[string]KnowledgeBaseAsset, len(assets))
	for _, asset := range assets {
		assetsByID[asset.ID] = asset
	}

	return KnowledgeBase{
		BasePath:       basePath,
		Articles:       articles,
		Assets:         assets,
		articlesBySlug: articlesBySlug,
		assetsByID:     assetsByID,
	}
}

// ArticleBySlug returns the [KnowledgeBaseArticle] for the given canonical slug.
func (kb KnowledgeBase) ArticleBySlug(slug string) (KnowledgeBaseArticle, bool) {
	if kb.articlesBySlug != nil {
		article, ok := kb.articlesBySlug[slug]
		return article, ok
	}

	for _, article := range kb.Articles {
		if article.Slug == slug {
			return article, true
		}
	}
	return KnowledgeBaseArticle{}, false
}

// AssetByID returns the [KnowledgeBaseAsset] for the given opaque ID.
func (kb KnowledgeBase) AssetByID(id string) (KnowledgeBaseAsset, bool) {
	if kb.assetsByID != nil {
		asset, ok := kb.assetsByID[id]
		return asset, ok
	}

	for _, asset := range kb.Assets {
		if asset.ID == id {
			return asset, true
		}
	}
	return KnowledgeBaseAsset{}, false
}

// KnowledgeBaseArticleURL returns a route-independent article placeholder URL.
func KnowledgeBaseArticleURL(slug string) string {
	return knowledgeBaseArticleURLScheme + slug
}

// KnowledgeBaseAssetURL returns a route-independent asset placeholder URL.
func KnowledgeBaseAssetURL(id string) string {
	return knowledgeBaseAssetURLScheme + id
}

func normalizeMountedBase(basePath string) string {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" || basePath == "/" {
		return ""
	}

	if !strings.Contains(basePath, "://") && !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}

	return strings.TrimRight(basePath, "/")
}
