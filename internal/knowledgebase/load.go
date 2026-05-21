package knowledgebase

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	goldmarkast "github.com/yuin/goldmark/ast"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.yaml.in/yaml/v3"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"git.sr.ht/~icikowski/account-center/internal/model"
	"git.sr.ht/~icikowski/account-center/internal/shared/xerror"
)

const (
	extensionMarkdown = ".md"

	documentIndex = "index.md"
)

var (
	errBasePathNotDirectory = errors.New("knowledge base path is not a directory")

	markdownRenderer = goldmark.New(goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()))
)

type discoveredArticle struct {
	sourcePath    string
	sourceRelPath string
	slug          string
	isIndex       bool
	content       []byte
	frontMatter   articleFrontMatter
}

type renderContext struct {
	baseDir            string
	articleBySourceRel map[string]discoveredArticle
	assetBySource      map[string]model.KnowledgeBaseAsset
}

type articleFrontMatter struct {
	Title         string `yaml:"title"`
	Description   string `yaml:"description"`
	FeaturedImage string `yaml:"featured_image"`
}

// Load loads and validates a [model.KnowledgeBases] from the given directory.
func Load(baseDir string) (*model.KnowledgeBase, error) {
	baseDir = filepath.Clean(baseDir)

	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", errBasePathNotDirectory, baseDir)
	}

	articles, err := discoverArticles(baseDir)
	if err != nil {
		return nil, err
	}

	knowledgeBase, err := renderKnowledgeBase(baseDir, articles)
	if err != nil {
		return nil, err
	}

	return knowledgeBase, nil
}

func discoverArticles(baseDir string) ([]discoveredArticle, error) {
	articles := make([]discoveredArticle, 0)
	articlesBySlug := make(map[string]discoveredArticle)
	errs := make([]error, 0)

	err := filepath.WalkDir(baseDir, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != extensionMarkdown {
			return nil
		}

		sourceRelPath, slug, isIndex, err := determineArticleLocation(baseDir, currentPath)
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		content, err := os.ReadFile(currentPath)
		if err != nil {
			return err
		}
		markdownContent, frontMatter, err := parseArticleContent(content)
		if err != nil {
			errs = append(errs, xerror.NewValidationError(
				fmt.Sprintf("invalid article '%s'", sourceRelPath),
				err,
			))
			return nil
		}

		article := discoveredArticle{
			sourcePath:    currentPath,
			sourceRelPath: sourceRelPath,
			slug:          slug,
			isIndex:       isIndex,
			content:       markdownContent,
			frontMatter:   frontMatter,
		}
		if err := validateArticleMetadata(article.frontMatter); err != nil {
			errs = append(errs, xerror.NewValidationError(
				fmt.Sprintf("invalid article '%s' metadata", sourceRelPath),
				err,
			))
			return nil
		}

		if existing, ok := articlesBySlug[article.slug]; ok {
			errs = append(errs, duplicateSlugError(existing, article))
			return nil
		}

		articlesBySlug[article.slug] = article
		articles = append(articles, article)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(errs) != 0 {
		return nil, xerror.NewValidationError("invalid knowledge base", errs...)
	}

	sort.SliceStable(articles, func(i, j int) bool {
		return articles[i].sourceRelPath < articles[j].sourceRelPath
	})

	return articles, nil
}

func determineArticleLocation(
	baseDir string,
	sourcePath string,
) (string, string, bool, error) {
	sourceRelPath, err := filepath.Rel(baseDir, sourcePath)
	if err != nil {
		return "", "", false, err
	}
	sourceRelPath = filepath.ToSlash(sourceRelPath)

	fileName := path.Base(sourceRelPath)
	isIndex := fileName == documentIndex

	switch {
	case isIndex && sourceRelPath == documentIndex:
		return "", "", false, xerror.NewValidationError(
			fmt.Sprintf("root index article '%s' is not allowed", sourceRelPath),
		)
	case isIndex:
		parentDir := path.Dir(sourceRelPath)
		return sourceRelPath, "/" + path.Clean(parentDir), true, nil
	default:
		return sourceRelPath, "/" + strings.TrimSuffix(sourceRelPath, extensionMarkdown), false, nil
	}
}

func duplicateSlugError(existing discoveredArticle, current discoveredArticle) error {
	switch {
	case existing.isIndex && !current.isIndex:
		return xerror.NewValidationError(fmt.Sprintf(
			"article '%s' duplicates slug '%s' from index article '%s'",
			current.sourceRelPath,
			current.slug,
			existing.sourceRelPath,
		))
	case !existing.isIndex && current.isIndex:
		return xerror.NewValidationError(fmt.Sprintf(
			"article '%s' duplicates slug '%s' from index article '%s'",
			existing.sourceRelPath,
			existing.slug,
			current.sourceRelPath,
		))
	default:
		return xerror.NewValidationError(fmt.Sprintf(
			"article '%s' duplicates slug '%s' already provided by '%s'",
			current.sourceRelPath,
			current.slug,
			existing.sourceRelPath,
		))
	}
}

func renderKnowledgeBase(baseDir string, discoveredArticles []discoveredArticle) (*model.KnowledgeBase, error) {
	articleBySourceRel := make(map[string]discoveredArticle, len(discoveredArticles))
	for _, article := range discoveredArticles {
		articleBySourceRel[article.sourceRelPath] = article
	}

	ctx := renderContext{
		baseDir:            baseDir,
		articleBySourceRel: articleBySourceRel,
		assetBySource:      make(map[string]model.KnowledgeBaseAsset),
	}

	articles := make([]model.KnowledgeBaseArticle, 0, len(discoveredArticles))
	errs := make([]error, 0)

	for _, article := range discoveredArticles {
		renderedArticle, err := renderArticle(ctx, article)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		articles = append(articles, renderedArticle)
	}

	if len(errs) != 0 {
		return nil, xerror.NewValidationError("invalid knowledge base", errs...)
	}

	sort.SliceStable(articles, func(i, j int) bool {
		return articles[i].Slug < articles[j].Slug
	})

	assets := make([]model.KnowledgeBaseAsset, 0, len(ctx.assetBySource))
	for _, asset := range ctx.assetBySource {
		assets = append(assets, asset)
	}
	sort.SliceStable(assets, func(i, j int) bool {
		return assets[i].SourceRelPath < assets[j].SourceRelPath
	})

	knowledgeBase := model.NewKnowledgeBase(baseDir, articles, assets)
	return &knowledgeBase, nil
}

func renderArticle(ctx renderContext, article discoveredArticle) (model.KnowledgeBaseArticle, error) {
	document := markdownRenderer.Parser().Parse(text.NewReader(article.content))
	title := strings.TrimSpace(article.frontMatter.Title)
	if title == "" {
		title = findArticleTitle(document, article.content, article.slug)
	}

	var rendered bytes.Buffer
	if err := markdownRenderer.Renderer().Render(&rendered, article.content, document); err != nil {
		return model.KnowledgeBaseArticle{}, err
	}

	renderedBody, linkedArticleSlugs, assetIDs, err := rewriteRenderedHTML(
		ctx,
		article,
		rendered.String(),
	)
	if err != nil {
		return model.KnowledgeBaseArticle{}, err
	}

	featuredImage, featuredImageAssetID, err := resolveFeaturedImage(ctx, article)
	if err != nil {
		return model.KnowledgeBaseArticle{}, xerror.NewValidationError(
			fmt.Sprintf("invalid article '%s'", article.sourceRelPath),
			err,
		)
	}
	if featuredImageAssetID != "" {
		assetIDs = appendUnique(assetIDs, featuredImageAssetID)
	}

	return model.KnowledgeBaseArticle{
		Title:              title,
		Description:        strings.TrimSpace(article.frontMatter.Description),
		Slug:               article.slug,
		SourcePath:         article.sourcePath,
		SourceRelPath:      article.sourceRelPath,
		FeaturedImage:      featuredImage,
		RenderedBody:       renderedBody,
		LinkedArticleSlugs: linkedArticleSlugs,
		AssetIDs:           assetIDs,
	}, nil
}

func rewriteRenderedHTML(
	ctx renderContext,
	article discoveredArticle,
	renderedBody string,
) (string, []string, []string, error) {
	container := &html.Node{
		Type:     html.ElementNode,
		Data:     atom.Body.String(),
		DataAtom: atom.Body,
	}

	nodes, err := html.ParseFragment(strings.NewReader(renderedBody), container)
	if err != nil {
		return "", nil, nil, err
	}

	linkedArticles := make(map[string]struct{})
	referencedAssets := make(map[string]struct{})
	errs := make([]error, 0)

	for _, node := range nodes {
		walkHTML(node, func(current *html.Node) {
			if current.Type != html.ElementNode {
				return
			}

			for i := range current.Attr {
				attr := &current.Attr[i]
				if attr.Key != "href" && attr.Key != "src" {
					continue
				}

				rewrittenValue, targetSlug, assetID, err := rewriteLinkAttribute(
					ctx,
					article,
					current.Data,
					attr.Key,
					attr.Val,
				)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				attr.Val = rewrittenValue
				if targetSlug != "" {
					linkedArticles[targetSlug] = struct{}{}
				}
				if assetID != "" {
					referencedAssets[assetID] = struct{}{}
				}
			}
		})
	}

	if len(errs) != 0 {
		return "", nil, nil, xerror.NewValidationError(
			fmt.Sprintf("invalid article '%s'", article.sourceRelPath),
			errs...,
		)
	}

	var buffer bytes.Buffer
	for _, node := range nodes {
		if err := html.Render(&buffer, node); err != nil {
			return "", nil, nil, err
		}
	}

	linkedArticleSlugs := mapKeys(linkedArticles)
	assetIDs := mapKeys(referencedAssets)

	return buffer.String(), linkedArticleSlugs, assetIDs, nil
}

func rewriteLinkAttribute(
	ctx renderContext,
	article discoveredArticle,
	tagName string,
	attrName string,
	rawValue string,
) (string, string, string, error) {
	parsedURL, err := url.Parse(rawValue)
	if err != nil {
		return "", "", "", xerror.NewValidationError(
			fmt.Sprintf("invalid %s on <%s>: %v", attrName, tagName, err),
		)
	}

	if parsedURL.Scheme != "" || parsedURL.Host != "" || strings.HasPrefix(parsedURL.Path, "/") {
		return rawValue, "", "", nil
	}
	if parsedURL.Path == "" {
		return rawValue, "", "", nil
	}

	if attrName == "href" && strings.EqualFold(path.Ext(parsedURL.Path), extensionMarkdown) {
		targetArticle, err := resolveLinkedArticle(ctx, article, parsedURL.Path)
		if err != nil {
			return "", "", "", err
		}

		targetURL := buildURL(
			model.KnowledgeBaseArticleURL(targetArticle.slug),
			parsedURL.RawQuery,
			parsedURL.Fragment,
		)
		return targetURL, targetArticle.slug, "", nil
	}

	asset, err := resolveAsset(ctx, article, parsedURL.Path)
	if err != nil {
		return "", "", "", err
	}

	assetURL := buildURL(
		model.KnowledgeBaseAssetURL(asset.ID),
		parsedURL.RawQuery,
		parsedURL.Fragment,
	)
	return assetURL, "", asset.ID, nil
}

func resolveLinkedArticle(
	ctx renderContext,
	article discoveredArticle,
	rawPath string,
) (discoveredArticle, error) {
	_, sourceRelPath, err := resolveRelativeTarget(ctx.baseDir, article.sourcePath, rawPath)
	if err != nil {
		return discoveredArticle{}, err
	}

	targetArticle, ok := ctx.articleBySourceRel[sourceRelPath]
	if !ok {
		return discoveredArticle{}, xerror.NewValidationError(fmt.Sprintf(
			"linked article '%s' does not exist",
			rawPath,
		))
	}

	return targetArticle, nil
}

func resolveAsset(
	ctx renderContext,
	article discoveredArticle,
	rawPath string,
) (model.KnowledgeBaseAsset, error) {
	sourcePath, sourceRelPath, err := resolveRelativeTarget(ctx.baseDir, article.sourcePath, rawPath)
	if err != nil {
		return model.KnowledgeBaseAsset{}, err
	}

	if existing, ok := ctx.assetBySource[sourceRelPath]; ok {
		return existing, nil
	}

	info, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return model.KnowledgeBaseAsset{}, xerror.NewValidationError(fmt.Sprintf(
				"referenced asset '%s' does not exist",
				rawPath,
			))
		}
		return model.KnowledgeBaseAsset{}, err
	}
	if info.IsDir() {
		return model.KnowledgeBaseAsset{}, xerror.NewValidationError(fmt.Sprintf(
			"referenced asset '%s' is a directory",
			rawPath,
		))
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return model.KnowledgeBaseAsset{}, err
	}

	asset := model.KnowledgeBaseAsset{
		ID:            assetID(sourceRelPath),
		SourcePath:    sourcePath,
		SourceRelPath: sourceRelPath,
		ContentType:   detectContentType(sourceRelPath, content),
		Content:       content,
		LastModified:  info.ModTime(),
	}
	ctx.assetBySource[sourceRelPath] = asset
	return asset, nil
}

func resolveFeaturedImage(ctx renderContext, article discoveredArticle) (string, string, error) {
	rawValue := strings.TrimSpace(article.frontMatter.FeaturedImage)
	if rawValue == "" {
		return "", "", nil
	}

	parsedURL, err := url.Parse(rawValue)
	if err != nil {
		return "", "", xerror.NewValidationError(
			fmt.Sprintf("invalid featured image '%s': %v", rawValue, err),
		)
	}

	if parsedURL.Scheme != "" || parsedURL.Host != "" || strings.HasPrefix(parsedURL.Path, "/") {
		return rawValue, "", nil
	}
	if parsedURL.Path == "" {
		return "", "", xerror.NewValidationError("featured image path is empty")
	}
	if strings.EqualFold(path.Ext(parsedURL.Path), extensionMarkdown) {
		return "", "", xerror.NewValidationError(
			fmt.Sprintf("featured image '%s' cannot point to a markdown article", rawValue),
		)
	}

	asset, err := resolveAsset(ctx, article, parsedURL.Path)
	if err != nil {
		return "", "", err
	}

	return buildURL(
		model.KnowledgeBaseAssetURL(asset.ID),
		parsedURL.RawQuery,
		parsedURL.Fragment,
	), asset.ID, nil
}

func resolveRelativeTarget(baseDir string, sourcePath string, rawPath string) (string, string, error) {
	unescapedPath, err := url.PathUnescape(rawPath)
	if err != nil {
		return "", "", xerror.NewValidationError(fmt.Sprintf(
			"invalid relative path '%s': %v",
			rawPath,
			err,
		))
	}

	sourcePathResolved := filepath.Clean(
		filepath.Join(filepath.Dir(sourcePath), filepath.FromSlash(unescapedPath)),
	)
	sourceRelPath, err := filepath.Rel(baseDir, sourcePathResolved)
	if err != nil {
		return "", "", err
	}
	if sourceRelPath == ".." || strings.HasPrefix(sourceRelPath, ".."+string(filepath.Separator)) {
		return "", "", xerror.NewValidationError(fmt.Sprintf(
			"relative path '%s' points outside the knowledge base",
			rawPath,
		))
	}

	return sourcePathResolved, filepath.ToSlash(sourceRelPath), nil
}

func buildURL(base string, rawQuery string, fragment string) string {
	var builder strings.Builder
	builder.WriteString(base)
	if rawQuery != "" {
		builder.WriteByte('?')
		builder.WriteString(rawQuery)
	}
	if fragment != "" {
		builder.WriteByte('#')
		builder.WriteString(fragment)
	}
	return builder.String()
}

func assetID(sourceRelPath string) string {
	sum := sha256.Sum256([]byte(sourceRelPath))
	return hex.EncodeToString(sum[:])
}

func detectContentType(sourceRelPath string, content []byte) string {
	if contentType := mime.TypeByExtension(
		strings.ToLower(filepath.Ext(sourceRelPath)),
	); contentType != "" {
		return contentType
	}
	return http.DetectContentType(content)
}

func findArticleTitle(document goldmarkast.Node, source []byte, slug string) string {
	var title string
	_ = goldmarkast.Walk(
		document,
		func(node goldmarkast.Node, entering bool) (goldmarkast.WalkStatus, error) {
			if !entering {
				return goldmarkast.WalkContinue, nil
			}

			heading, ok := node.(*goldmarkast.Heading)
			if !ok || heading.Level != 1 {
				return goldmarkast.WalkContinue, nil
			}

			title = strings.TrimSpace(extractNodeText(heading, source))
			return goldmarkast.WalkStop, nil
		},
	)

	if title != "" {
		return title
	}

	return path.Base(slug)
}

func extractNodeText(node goldmarkast.Node, source []byte) string {
	var builder strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch typed := child.(type) {
		case *goldmarkast.Text:
			builder.Write(typed.Segment.Value(source))
		default:
			builder.WriteString(extractNodeText(child, source))
		}
	}
	return builder.String()
}

func walkHTML(node *html.Node, visit func(node *html.Node)) {
	visit(node)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkHTML(child, visit)
	}
}

func mapKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for value := range values {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	return keys
}

func appendUnique(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}

	values = append(values, value)
	sort.Strings(values)
	return values
}

func parseArticleContent(content []byte) ([]byte, articleFrontMatter, error) {
	firstLine, nextOffset, ok := nextLine(content, 0)
	if !ok || firstLine != "---" {
		return nil, articleFrontMatter{}, xerror.NewValidationError("missing YAML front matter")
	}

	offset := nextOffset
	yamlStart := offset
	for {
		line, nextOffset, ok := nextLine(content, offset)
		if !ok {
			return nil, articleFrontMatter{}, xerror.NewValidationError(
				"unclosed YAML front matter",
			)
		}
		if line == "---" {
			var frontMatter articleFrontMatter
			if err := yaml.Unmarshal(content[yamlStart:offset], &frontMatter); err != nil {
				return nil, articleFrontMatter{}, err
			}
			return content[nextOffset:], frontMatter, nil
		}

		offset = nextOffset
	}
}

func validateArticleMetadata(frontMatter articleFrontMatter) error {
	errs := make([]error, 0, 2)

	if strings.TrimSpace(frontMatter.Title) == "" {
		errs = append(errs, xerror.NewValidationError("title is required"))
	}
	if strings.TrimSpace(frontMatter.Description) == "" {
		errs = append(errs, xerror.NewValidationError("description is required"))
	}

	if len(errs) == 0 {
		return nil
	}

	return xerror.NewValidationError("invalid metadata", errs...)
}

func nextLine(content []byte, offset int) (string, int, bool) {
	if offset >= len(content) {
		return "", offset, false
	}

	end := bytes.IndexByte(content[offset:], '\n')
	if end == -1 {
		line := string(bytes.TrimSuffix(content[offset:], []byte{'\r'}))
		return line, len(content), true
	}

	lineBytes := content[offset : offset+end]
	lineBytes = bytes.TrimSuffix(lineBytes, []byte{'\r'})
	return string(lineBytes), offset + end + 1, true
}
