# Configuration overview

**Account Center** reads configuration from three places:

1. `AC_*` environment variables for runtime behavior and integrations;
2. a services catalog YAML file that defines services and access mappings;
3. a knowledge base directory containing Markdown articles and assets (optionally, if enabled).

Startup fails fast if required values are missing or invalid.

The application serves HTTP only. Deploy it behind a reverse proxy that terminates TLS/SSL and forwards trusted headers.

## Routes

| Route                   | Purpose                                                               |
| ----------------------- | --------------------------------------------------------------------- |
| `/`                     | Public landing page; authenticated users are redirected to `/catalog` |
| `/catalog`              | Role-aware services catalog                                           |
| `/kb`                   | Knowledge base root when enabled                                      |
| `/kb/attachments/...`   | Assets referenced from KB articles                                    |
| `/login`                | Starts the OIDC authorization flow                                    |
| `/oidc-callback`        | OIDC callback endpoint                                                |
| `/refresh`              | Forces a session refresh                                              |
| `/logout`               | Clears the local session and attempts token revocation                |
| `/assets/...`           | Static frontend assets                                                |
| `/manifest.webmanifest` | Web app manifest                                                      |
| `/health/live`          | Liveness probe                                                        |
| `/health/ready`         | Readiness probe for catalog, KB, and storage                          |

## Startup

At startup the application:

1. loads and validates environment variables;
2. parses trusted proxy CIDRs (if configured);
3. loads the catalog and starts its watcher;
4. loads the knowledge base and starts its watcher (if configured);
5. initializes session storage in memory or Redis;
6. discovers the OIDC provider and starts the HTTP server.

Successful startup requires:

- valid OIDC settings,
- a reachable OIDC provider,
- a valid catalog file,
- a valid knowledge base directory (if enabled).

## Live reload

The catalog and knowledge base both support live reload.

- `AC_CATALOG_RELOAD_DEBOUNCE` controls catalog reload debounce.
- `AC_KB_RELOAD_DEBOUNCE` controls KB reload debounce.

After a catalog reload, visibility is recalculated on the next request. Users do not need a restart for catalog updates to take effect.

## Session storage

By default, login state and sessions are stored in memory. That works well for:

- local development;
- single-instance deployments;
- short-lived environments.

Use Redis when you need:

- session persistence across restarts;
- a separate app and state layer;
- shared sessions across multiple instances.

Redis keys use this shape:

```text
{AC_REDIS_KEY_PREFIX}:{kind}:{id}
```

Where `kind` is `login-state` or `session`.

## Reverse proxies

OIDC redirect URLs must match the public URL exactly, so reverse-proxy deployments should usually set both:

- `AC_INSTANCE_BASE_URL=https://account.example.com`
- `AC_SERVER_TRUSTED_PROXIES=<CIDR(s) of your reverse proxy>`

`AC_INSTANCE_BASE_URL` gives the app a stable public callback URL. `AC_SERVER_TRUSTED_PROXIES` allows trusted forwarded headers to be used when deriving host and scheme.

If you leave both unset behind reverse proxy, the app may generate internal `http://...` callback URLs and OIDC login will fail.

See [Deployment](deployment.md) and [OIDC configuration](oidc.md) for examples.

## Authentication model

**Account Center** assumes that:

- users authenticate with an OIDC provider;
- user identity includes `sub`, `name`, `email`, and `groups`;
- group membership drives service visibility.

The `groups` claim name is fixed and not configurable. Providers must expose group membership under that exact claim.

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
