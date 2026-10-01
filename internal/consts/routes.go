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
	RouteRobots      = "/robots.txt"

	RouteCatalog                  = "/catalog"
	RouteKnowledgeBase            = "/kb"
	RouteKnowledgeBaseAttachments = "/kb/attachments"

	RouteFallbackProfilePicture = "/assets/img/user.jpg"
)
