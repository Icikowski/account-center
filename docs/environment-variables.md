# Environment variables

All runtime configuration is provided through `AC_*` environment variables.

Durations use Go-style syntax such as `500ms`, `10m`, and `24h`.

## Instance

| Variable               | Default | Required | Purpose                                                                                                   |
| ---------------------- | ------- | -------- | --------------------------------------------------------------------------------------------------------- |
| `AC_INSTANCE_NAME`     | empty   | No       | Optional label used in titles and UI context                                                              |
| `AC_INSTANCE_BASE_URL` | empty   | No       | Public base URL used for OIDC callback generation; required when TLS/SSL is terminated by a reverse proxy |

## Server

| Variable                    | Default | Required | Purpose                                                       |
| --------------------------- | ------- | -------- | ------------------------------------------------------------- |
| `AC_SERVER_ADDRESS`         | empty   | No       | Bind address; empty means all interfaces                      |
| `AC_SERVER_PORT`            | `8080`  | No       | HTTP listen port                                              |
| `AC_SERVER_TRUSTED_PROXIES` | empty   | No       | Comma-separated IPs/CIDRs allowed to supply forwarded headers |

### `AC_SERVER_TRUSTED_PROXIES`

Use this only for addresses you fully trust, such as:

- `127.0.0.1/32` for a local reverse proxy;
- `10.0.0.0/8` for an internal proxy network;
- `172.17.0.0/16` for a Docker network.

If empty, forwarded headers are ignored.

## Services catalog

| Variable                     | Default          | Required | Purpose                                        |
| ---------------------------- | ---------------- | -------- | ---------------------------------------------- |
| `AC_CATALOG_PATH`            | `./catalog.yaml` | No       | Path to the services catalog YAML file         |
| `AC_CATALOG_RELOAD_DEBOUNCE` | `500ms`          | No       | Debounce for live reload after catalog changes |

The catalog file must exist, be valid YAML and contain at least one service.

## Knowledge base

| Variable                | Default | Required | Purpose                                                                                 |
| ----------------------- | ------- | -------- | --------------------------------------------------------------------------------------- |
| `AC_KB_ENABLED`         | `false` | No       | Enables the knowledge base module                                                       |
| `AC_KB_PATH`            | `./kb`  | No       | Path to the knowledge base directory; must resolve to a valid KB directory when enabled |
| `AC_KB_RELOAD_DEBOUNCE` | `500ms` | No       | Debounce for knowledge base live reload                                                 |

If `AC_KB_ENABLED=false`, the KB content path is ignored.

## Auth sessions

| Variable                      | Default                  | Required | Purpose                                                              |
| ----------------------------- | ------------------------ | -------- | -------------------------------------------------------------------- |
| `AC_AUTH_SESSION_TTL`         | `24h`                    | No       | Lifetime of persisted user sessions                                  |
| `AC_AUTH_LOGIN_STATE_TTL`     | `10m`                    | No       | Lifetime of temporary OIDC login state (state, nonce, PKCE verifier) |
| `AC_AUTH_SESSION_COOKIE_NAME` | `account-center-session` | No       | Name of the session cookie                                           |

## OIDC

| Variable                 | Default | Required | Purpose                                      |
| ------------------------ | ------- | -------- | -------------------------------------------- |
| `AC_OIDC_PROVIDER_URL`   | none    | Yes      | OIDC issuer/provider URL                     |
| `AC_OIDC_CLIENT_ID`      | none    | Yes      | OIDC client ID                               |
| `AC_OIDC_CLIENT_SECRET`  | none    | Yes      | OIDC client secret                           |
| `AC_OIDC_REFRESH_BEFORE` | `1m`    | No       | Refreshes tokens before expiry when possible |

These values are required for the application to start.

## Redis

| Variable              | Default          | Required     | Purpose                                   |
| --------------------- | ---------------- | ------------ | ----------------------------------------- |
| `AC_REDIS_ENABLED`    | `false`          | No           | Enables Redis-backed auth/session storage |
| `AC_REDIS_ADDRESS`    | none             | When enabled | Redis host and port                       |
| `AC_REDIS_USERNAME`   | empty            | No           | Redis username                            |
| `AC_REDIS_PASSWORD`   | empty            | No           | Redis password                            |
| `AC_REDIS_DATABASE`   | `0`              | No           | Redis database number                     |
| `AC_REDIS_KEY_PREFIX` | `account-center` | When enabled | Prefix for stored Redis keys              |

If Redis is disabled, **Account Center** uses in-memory storage.

## Logging

| Variable        | Default | Required | Purpose                            |
| --------------- | ------- | -------- | ---------------------------------- |
| `AC_LOG_LEVEL`  | `info`  | No       | Zerolog log level                  |
| `AC_LOG_PRETTY` | `false` | No       | Enables human-friendly pretty logs |

Pretty logging is useful during development but adds overhead.

## Container image overrides

The binary defaults above are not the same as the published container image defaults:

| Variable          | Image value          |
| ----------------- | -------------------- |
| `AC_CATALOG_PATH` | `/data/catalog.yaml` |
| `AC_KB_PATH`      | `/data/kb`           |

Container deployments should either mount content under `/data` or override those two variables explicitly.

## Minimal examples

### Single-instance binary deployment

```dotenv
AC_INSTANCE_NAME="Example Inc."
AC_INSTANCE_BASE_URL="https://account.example.com"

AC_SERVER_PORT=8080
AC_SERVER_TRUSTED_PROXIES="127.0.0.1/32"

AC_CATALOG_PATH="./catalog.yaml"
AC_KB_ENABLED=true
AC_KB_PATH="./kb"

AC_OIDC_PROVIDER_URL="https://sso.example.com"
AC_OIDC_CLIENT_ID="account-center"
AC_OIDC_CLIENT_SECRET="replace-me"
```

### Docker deployment

```dotenv
AC_INSTANCE_NAME="Example Inc."
AC_INSTANCE_BASE_URL="https://account.example.com"

AC_SERVER_TRUSTED_PROXIES="172.17.0.0/16,127.0.0.1/32"

AC_OIDC_PROVIDER_URL="https://sso.example.com"
AC_OIDC_CLIENT_ID="account-center"
AC_OIDC_CLIENT_SECRET="replace-me"

AC_KB_ENABLED=true
```

Add `AC_SERVER_TRUSTED_PROXIES` when the container sits behind a trusted reverse proxy and forwarded headers should affect public URL handling. The app itself does not serve TLS.

## See also

- [Configuration overview](configuration.md)
- [Deployment](deployment.md)
- [OIDC configuration](oidc.md)
