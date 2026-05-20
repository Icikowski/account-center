package i18n

import (
	"embed"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"go.yaml.in/yaml/v3"
	"golang.org/x/text/language"
)

//go:embed locales/*.yaml
var localeFS embed.FS

// NewBundle creates a new [goi18n.Bundle] with embedded locale files.
func NewBundle() *goi18n.Bundle {
	bundle := goi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	entries, _ := localeFS.ReadDir("locales")
	for _, e := range entries {
		if !e.IsDir() {
			_, _ = bundle.LoadMessageFileFS(localeFS, "locales/"+e.Name())
		}
	}
	return bundle
}
