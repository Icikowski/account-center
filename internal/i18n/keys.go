package i18n

import "git.sr.ht/~icikowski/account-center/internal/model"

//go:generate locales/zz_generate.sh

// Key is a type alias for string, representing the key for an internationalized message.
type Key string

// Constants for message keys used in the application.
const (
	KeyLanguageCode Key = "code"

	KeyGeneralAccountCenter Key = "general.account_center"
	KeyGeneralDescription   Key = "general.description"
	KeyGeneralLanguage      Key = "general.language"
	KeyGeneralTheme         Key = "general.theme"
	KeyGeneralSystem        Key = "general.system"

	KeyWelcomeMessage Key = "welcome.message"

	KeyAboutVersion      Key = "about.version"
	KeyAboutGitReference Key = "about.git_reference"
	KeyAboutBuildTime    Key = "about.build_time"
	KeyAboutGoVersion    Key = "about.go_version"

	KeyAuthLogin       Key = "auth.login"
	KeyAuthLogout      Key = "auth.logout"
	KeyAuthRefresh     Key = "auth.refresh"
	KeyAuthUnknownUser Key = "auth.unknown_user"
	KeyAuthNoEmail     Key = "auth.no_email"

	KeyLanguageEnglish Key = "language.english"
	KeyLanguagePolish  Key = "language.polish"

	KeyThemeLight Key = "theme.light"
	KeyThemeDark  Key = "theme.dark"

	KeyLabelResources Key = "label.resources"

	KeySectionCatalog       Key = "section.catalog"
	KeySectionKnowledgeBase Key = "section.knowledge_base"

	KeyCatalogGoToService       Key = "catalog.go_to_service"
	KeyCatalogNoServices        Key = "catalog.no_services.title"
	KeyCatalogNoServicesMessage Key = "catalog.no_services.description"

	KeyKnowledgeBaseGoToArticle       Key = "knowledge_base.go_to_article"
	KeyKnowledgeBaseNoArticles        Key = "knowledge_base.no_articles.title"
	KeyKnowledgeBaseNoArticlesMessage Key = "knowledge_base.no_articles.description"
	KeyKnowledgeBaseDisabled          Key = "knowledge_base.disabled.title"
	KeyKnowledgeBaseDisabledMessage   Key = "knowledge_base.disabled.description"

	KeyErrorGoHome                  Key = "error.go_home"
	KeyErrorNotFound                Key = "error.not_found"
	KeyErrorNotFoundMessage         Key = "error.not_found.description"
	KeyErrorMethodNotAllowed        Key = "error.method_not_allowed"
	KeyErrorMethodNotAllowedMessage Key = "error.method_not_allowed.description"

	KeyRoleSuperuser           Key = "role.superuser"
	KeyRoleSystemAdministrator Key = "role.system_administrator"
	KeyRoleAdministrator       Key = "role.administrator"
	KeyRoleRedactor            Key = "role.redactor"
	KeyRoleEditor              Key = "role.editor"
	KeyRoleViewer              Key = "role.viewer"
	KeyRoleUser                Key = "role.user"
	KeyRoleGuest               Key = "role.guest"
	KeyRoleGeneralAccess       Key = "role.general_access"
)

// UserRoleKey return a translation [Key] for a given [model.UserRole].
func UserRoleKey(role model.UserRole) Key {
	return Key("role." + string(role))
}
