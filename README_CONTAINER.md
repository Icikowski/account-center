# Account Center

**Self-hosted, OIDC-authenticated portal for internal services and knowledge base articles.**

Account Center is a lightweight portal for self-hosted environments that need one place to send users after login. It combines:

- an OIDC-based login flow,
- a role-aware internal services catalog,
- an optional knowledge base rendered from Markdown.

It is designed for operators who want something similar to an Okta-style app dashboard, but fully self-hosted and easy to configure.

## Environment variables

Required:

- `AC_OIDC_PROVIDER_URL` - OIDC issuer / provider URL
- `AC_OIDC_CLIENT_ID` - OIDC client ID
- `AC_OIDC_CLIENT_SECRET` - OIDC client secret

Common optional variables:

- `AC_INSTANCE_NAME` - display name for the instance
- `AC_INSTANCE_BASE_URL` - public base URL, recommended behind a reverse proxy
- `AC_SERVER_PORT` - HTTP listen port, default `8080`
- `AC_SERVER_TRUSTED_PROXIES` - trusted proxy IPs/CIDRs for forwarded headers
- `AC_CATALOG_PATH` - path to the services catalog file, default `/data/catalog.yaml`
- `AC_KB_ENABLED` - enable knowledge base, default `false`
- `AC_KB_PATH` - path to the knowledge base directory, default `/data/kb`
- `AC_REDIS_ENABLED` - enable Redis-backed session storage, default `false`
- `AC_REDIS_ADDRESS` - Redis address
- `AC_LOG_LEVEL` - log level, default `info`

## Minimal Docker example

Docker Hub:

```bash
docker run -d \
  --name account-center \
  -p 8080:8080 \
  --env-file .env \
  -v "$PWD/catalog.yaml:/data/catalog.yaml:ro" \
  icikowski/account-center:latest
```

GitHub Container Registry:

```bash
docker run -d \
  --name account-center \
  -p 8080:8080 \
  --env-file .env \
  -v "$PWD/catalog.yaml:/data/catalog.yaml:ro" \
  ghcr.io/icikowski/account-center:latest
```

If you enable the knowledge base, also mount:

```bash
-v "$PWD/kb:/data/kb:ro"
```

## Minimal Docker Compose example

Docker Hub:

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
```

GitHub Container Registry:

```yaml
services:
  account-center:
    image: ghcr.io/icikowski/account-center:latest
    restart: unless-stopped
    env_file:
      - .env
    ports:
      - "8080:8080"
    volumes:
      - ./catalog.yaml:/data/catalog.yaml:ro
```

## Documentation

- Project: https://git.sr.ht/~icikowski/account-center
- Configuration: https://git.sr.ht/~icikowski/account-center/tree/main/item/docs/configuration.md
- Environment variables: https://git.sr.ht/~icikowski/account-center/tree/main/item/docs/environment-variables.md
- Deployment: https://git.sr.ht/~icikowski/account-center/tree/main/item/docs/deployment.md
- OIDC setup: https://git.sr.ht/~icikowski/account-center/tree/main/item/docs/oidc.md
- Services catalog: https://git.sr.ht/~icikowski/account-center/tree/main/item/docs/services-catalog.md
- Knowledge base: https://git.sr.ht/~icikowski/account-center/tree/main/item/docs/knowledge-base.md
