package consts

// Application routes.
const (
	RouteRoot = "/"

	RouteHealth = "/health"
	RouteLive   = "/live"
	RouteReady  = "/ready"

	RouteLogin        = "/login"
	RouteLogout       = "/logout"
	RouteRefresh      = "/refresh"
	RouteOIDCCallback = "/oidc-callback"

	RouteAssets = "/assets"

	RouteWebManifest = "/manifest.webmanifest"

	RouteCatalog                  = "/catalog"
	RouteKnowledgeBase            = "/kb"
	RouteKnowledgeBaseAttachments = "/kb/attachments"
)
