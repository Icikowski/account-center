# API

**Account Center** exposes a small, read-only JSON API under `/api/v1`. It is meant for non-HTML clients such as the [browser extension](browser-extension.md) or a PWA service worker.

The machine-readable description is available as an OpenAPI 3.1 document: [`internal/web/api/openapi.yaml`](../internal/web/api/openapi.yaml). Every instance serves it at `/api/openapi.yaml`.

## Design

- **Stateless:** authentication reuses the existing session cookie.
- **Read-only:** only `GET` is accepted, which is why no CSRF protection is needed; other methods return `405`.
- **No tokens exposed:** responses never contain OIDC tokens or session identifiers.
- **Uncacheable:** every `/api/v1` response carries `Cache-Control: no-store`.
- **No CORS:** cross-origin web pages cannot call the API with credentials; browser extensions can call it only for hosts they were granted access to; without that grant the browser reports a CORS error.
- **Errors** are always JSON: `{"error": "message"}` (including `404` for unknown paths and `405` for other methods); the API never redirects to the login page.

Signing in is done through the regular `/login` route; the API only reports state.

## Endpoints

### `GET /api/v1/session`

Always answers `200`, so clients can cheaply check whether the user is signed in.

```json
{
  "authenticated": true,
  "user": { "name": "Jane Doe", "email": "jane@example.com", "picture": "..." }
}
```

For anonymous callers the response is `{"authenticated": false}`.

### `GET /api/v1/services`

Returns the services visible to the caller, sorted by name, using the same evaluation rules as the catalog page.

```json
{
  "services": [
    {
      "name": "Wiki",
      "url": "https://wiki.example.com",
      "icon": "...",
      "role": "editor"
    }
  ]
}
```

`icon` is omitted when the service has none. `role` is the effective role identifier (for example `system_administrator`), not a translated label. Anonymous callers, including those with an expired session, get `401`.

## See also

- [Configuration overview](configuration.md)
- [Browser extension](browser-extension.md)
- [Services catalog](services-catalog.md)
