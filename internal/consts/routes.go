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

	RouteAPI         = "/api"
	RouteAPIV1       = "/v1"
	RouteAPISession  = "/session"
	RouteAPIServices = "/services"

	RouteAPIOpenAPI = "/openapi.yaml"

	RouteFallbackProfilePicture = "/assets/img/user.jpg"
)
