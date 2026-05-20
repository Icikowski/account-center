package i18n

import "slices"

// Language represents a language available in the application.
type Language struct {
	Name Key
	Code string
}

// AvailableLanguages returns a slice of all [Language]s available in the application.
func AvailableLanguages() []Language {
	return slices.Clone(availableLanguages)
}

var availableLanguages = []Language{
	{
		Name: KeyLanguageEnglish,
		Code: "en",
	},
	{
		Name: KeyLanguagePolish,
		Code: "pl",
	},
}
