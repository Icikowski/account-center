# Account Center

**Self-hosted, OIDC-authenticated portal for internal services and knowledge base articles.**

**Account Center** is a self-hosted OIDC portal for internal services and optional Markdown knowledge base content.

It serves HTTP only, so place a reverse proxy in front of the container to terminate TLS/SSL.

## Images

- `icikowski/account-center`
- `ghcr.io/icikowski/account-center`

## Required environment variables

- `AC_OIDC_PROVIDER_URL`
- `AC_OIDC_CLIENT_ID`
- `AC_OIDC_CLIENT_SECRET`

Common container settings:

- `AC_INSTANCE_NAME`
- `AC_INSTANCE_BASE_URL`
- `AC_SERVER_PORT`
- `AC_SERVER_TRUSTED_PROXIES`
- `AC_CATALOG_PATH` (`/data/catalog.yaml`)
- `AC_KB_ENABLED`
- `AC_KB_PATH` (`/data/kb`)
- `AC_REDIS_ENABLED`
- `AC_REDIS_ADDRESS`
- `AC_LOG_LEVEL`

The image includes a Docker healthcheck for `/health/live` and `/health/ready`.

## Docker

```bash
docker run -d \
  --name account-center \
  -p 8080:8080 \
  --env-file .env \
  -v "$PWD/catalog.yaml:/data/catalog.yaml:ro" \
  -v "$PWD/kb:/data/kb:ro" \
  icikowski/account-center:latest
```

Omit the `kb` mount if the knowledge base is disabled.

## Docker Compose

```yaml
services:
  account-center:
    image: icikowski/account-center:latest
    restart: unless-stopped
    env_file:
      - .env
    ports:
      - "8080:8080"
    volumes:
      - ./catalog.yaml:/data/catalog.yaml:ro
      - ./kb:/data/kb:ro
```
