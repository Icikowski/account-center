# Configuration overview

Account Center is configured through three inputs:

1. **Environment variables** (`AC_*`) for runtime behavior and integration settings.
2. **A services catalog YAML file** that defines which services exist and how group membership maps to visible roles.
3. **An optional knowledge base directory** containing Markdown articles and referenced assets.

The application validates all of this on startup. If required values are missing or content is invalid, it fails fast instead of starting with partial behavior.

## Runtime routes

These are the main operator-relevant routes exposed by the application:

| Route                   | Purpose                                                          |
| ----------------------- | ---------------------------------------------------------------- |
| `/`                     | Public landing page; redirects authenticated users to `/catalog` |
| `/health/live`          | Process liveness probe                                           |
| `/health/ready`         | Readiness probe for catalog, KB, and storage backend             |
| `/login`                | Starts the OIDC authorization flow                               |
| `/oidc-callback`        | OIDC authorization-code callback endpoint                        |
| `/refresh`              | Forces a session refresh                                         |
| `/logout`               | Clears the local session and attempts token revocation           |
| `/catalog`              | Role-aware services catalog                                      |
| `/kb`                   | Knowledge base root when enabled                                 |
| `/kb/attachments/...`   | Served assets referenced from KB articles                        |
| `/assets/...`           | Static frontend assets                                           |
| `/manifest.webmanifest` | Web app manifest                                                 |

## Startup model

At startup the application:

1. Loads and validates environment variables.
2. Parses trusted proxy CIDRs, if configured.
3. Loads the catalog and starts a file watcher for it.
4. Optionally loads the knowledge base and starts a watcher for it.
5. Initializes session storage in memory or Redis.
6. Discovers the OIDC provider and starts the HTTP server.

This means the following are required for a successful start:

- valid OIDC settings,
- a reachable OIDC provider,
- a valid catalog file,
- and, if KB is enabled, a valid knowledge base directory.

## Live reload

Both the catalog and the knowledge base support live reload.

- `AC_CATALOG_RELOAD_DEBOUNCE` controls the catalog watcher debounce.
- `AC_KB_RELOAD_DEBOUNCE` controls the knowledge base watcher debounce.

After a catalog reload, service visibility is recalculated automatically on the next request. Existing users do not need a full restart of the app for catalog updates to take effect.

## Session storage

By default, Account Center stores OIDC login state and user sessions **in memory**. This is enough for:

- local development,
- single-instance setups,
- short-lived or disposable environments.

Enable Redis when you need:

- session persistence across restarts,
- a cleaner separation of app and state,
- multiple app instances sharing the same session store.

Redis keys are stored with this shape:

```text
{AC_REDIS_KEY_PREFIX}:{kind}:{id}
```

Where `kind` is either `login-state` or `session`.

## Reverse proxies and public URL handling

Reverse-proxy deployments need special care because OIDC redirect URLs must match exactly.

The safest setup is to configure **both**:

- `AC_INSTANCE_BASE_URL=https://account.example.com`
- `AC_SERVER_TRUSTED_PROXIES=<CIDR(s) of your reverse proxy>`

Why both matter:

- `AC_INSTANCE_BASE_URL` gives Account Center a stable public URL for callback generation.
- `AC_SERVER_TRUSTED_PROXIES` allows the app to trust `X-Forwarded-Host` and `X-Forwarded-Proto` from your proxy.

If you deploy behind Nginx or another reverse proxy and leave both unset, the application may derive `http://...` callback URLs from the internal request instead of the public `https://...` URL, which will break OIDC login.

See [`deployment.md`](deployment.md) and [`oidc.md`](oidc.md) for the full proxy and provider examples.

## Authentication model

Account Center assumes:

- users authenticate with an OIDC provider,
- user identity includes `sub`, `name`, `email`, and `groups`,
- and group membership drives service visibility in the catalog.

The `groups` claim name is **not configurable**. Providers must expose group membership under that exact claim name.

## Operator-facing files

| File or path                      | Purpose                                |
| --------------------------------- | -------------------------------------- |
| environment variables / env file  | Runtime configuration                  |
| `catalog.yaml`                    | Services catalog definition            |
| `kb/`                             | Optional knowledge base directory      |
| `schemas/catalog.schema.json`     | Reference schema for catalog authoring |
| `schemas/frontmatter.schema.json` | Reference schema for KB front matter   |

## Related guides

- [Environment variables](environment-variables.md)
- [Deployment](deployment.md)
- [OIDC configuration](oidc.md)
- [Services catalog](services-catalog.md)
- [Knowledge base](knowledge-base.md)
