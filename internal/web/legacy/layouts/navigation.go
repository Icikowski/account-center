package layouts

import (
	"slices"

	"git.sr.ht/~icikowski/account-center/internal/i18n"
	"git.sr.ht/~icikowski/account-center/internal/model"
)

// NavigationItem represents a single item in the application's navigation menu.
type NavigationItem struct {
	Key    i18n.Key
	URL    string
	Parent *NavigationItem
}

// FullURL returns the full URL for the navigation item, including the URLs of all parent items.
func (n *NavigationItem) FullURL() string {
	if n.Parent == nil {
		return n.URL
	}
	return n.Parent.FullURL() + n.URL
}

// Breadcrumbs returns the breadcrumb trail for the navigation item, starting from the root item down to the current
// item.
func (n *NavigationItem) Breadcrumbs() []*NavigationItem {
	if n.Parent == nil {
		return []*NavigationItem{n}
	}
	return append(n.Parent.Breadcrumbs(), n)
}

// IsActive checks if the navigation item is active based on the current URL.
//
// Makes sense only on root navigation items, as child items are rendered in form of breadcrumbs.
func (n *NavigationItem) IsActive(currentNav *NavigationItem) bool {
	if currentNav == nil {
		return false
	}
	return slices.Contains(currentNav.Breadcrumbs(), n)
}

// Predefined navigation roots that correspond to different sections of the application.
var (
	NavigationCatalog = &NavigationItem{
		Key: i18n.KeySectionCatalog,
		URL: "/catalog",
	}
	NavigationKnowledgeBase = &NavigationItem{
		Key: i18n.KeySectionKnowledgeBase,
		URL: "/kb",
	}
)

// NavigationKnowledgeBaseArticle creates a navigation item for a specific knowledge
// base article, with the appropriate URL and parent item.
func NavigationKnowledgeBaseArticle(article model.KnowledgeBaseArticle) *NavigationItem {
	return &NavigationItem{
		Key:    i18n.Key(article.Title),
		URL:    "/" + article.Slug,
		Parent: NavigationKnowledgeBase,
	}
}
