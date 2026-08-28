# Deployment

**Account Center** can run as:

1. a standalone binary;
2. a container;
3. a Docker Compose stack.

## Published images

- `icikowski/account-center`
- `ghcr.io/icikowski/account-center`

## Reverse proxy for TLS/SSL

**Account Center** does not terminate TLS/SSL itself. Put a reverse proxy (such as `nginx`) in front of it and handle HTTPS there.

## Requirements

Every deployment needs:

- a valid OIDC provider;
- a non-empty catalog file;
- the required OIDC environment variables;
- a knowledge base directory (if enabled).

For production, also set:

- `AC_INSTANCE_BASE_URL` to the public URL;
- `AC_SERVER_TRUSTED_PROXIES` to trusted reverse proxy CIDRs;
- Redis if sessions must survive restarts;
- configuration outside the container image.

## Standalone binary

### Build

```bash
task generate
task build-static
```

### Run

```bash
AC_INSTANCE_BASE_URL="https://account.example.com" \
AC_SERVER_TRUSTED_PROXIES="127.0.0.1/32" \
AC_CATALOG_PATH="./catalog.yaml" \
AC_KB_ENABLED=true \
AC_KB_PATH="./kb" \
AC_OIDC_PROVIDER_URL="https://sso.example.com" \
AC_OIDC_CLIENT_ID="account-center" \
AC_OIDC_CLIENT_SECRET="replace-me" \
./bin/account-center
```

Even in local or bare-binary deployments, keep TLS/SSL termination in a reverse proxy if the app is exposed on the network.

## Docker

The published image uses:

- `AC_CATALOG_PATH=/data/catalog.yaml`
- `AC_KB_PATH=/data/kb`
- a Docker `HEALTHCHECK` for `/health/live` and `/health/ready`

Mount content under `/data` for the simplest setup:

```bash
docker run -d \
  --name account-center \
  -p 8080:8080 \
  --env-file .env \
  -v "$PWD/catalog.yaml:/data/catalog.yaml:ro" \
  -v "$PWD/kb:/data/kb:ro" \
  icikowski/account-center:<tag>
```

If KB is disabled, omit the `kb` mount.

Docker Hub image:

```bash
docker run -d \
  --name account-center \
  -p 8080:8080 \
  --env-file .env \
  -v "$PWD/catalog.yaml:/data/catalog.yaml:ro" \
  icikowski/account-center:<tag>
```

## Docker Compose

Minimal Compose example with Redis-backed sessions:

```yaml
services:
  account-center:
    image: icikowski/account-center:<tag>
    restart: unless-stopped
    env_file:
      - .env
    environment:
      AC_INSTANCE_NAME: Example Inc.
      AC_INSTANCE_BASE_URL: https://account.example.com
      AC_CATALOG_PATH: /data/catalog.yaml
      AC_KB_ENABLED: "true"
      AC_KB_PATH: /data/kb
      AC_REDIS_ENABLED: "true"
      AC_REDIS_ADDRESS: redis:6379
      AC_REDIS_KEY_PREFIX: account-center
    ports:
      - "8080:8080"
    volumes:
      - ./catalog.yaml:/data/catalog.yaml:ro
      - ./kb:/data/kb:ro
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: ["redis-server", "--save", "", "--appendonly", "no"]
```

### Example `.env`

```dotenv
AC_OIDC_PROVIDER_URL="https://sso.example.com"
AC_OIDC_CLIENT_ID="account-center"
AC_OIDC_CLIENT_SECRET="replace-me"
AC_OIDC_REFRESH_BEFORE=1m
```

### Compose notes

- The application image is distroless.
- The app exposes `/health/live` and `/health/ready`.
- Remove Redis settings if you do not use Redis.
- Set `AC_SERVER_TRUSTED_PROXIES` only when a trusted proxy is in front of the app.
- In this repository, `.env` is a documented workflow file loaded by the committed `.envrc`.

## Nginx reverse proxy

### Recommended environment variables

```dotenv
AC_INSTANCE_BASE_URL="https://account.example.com"
AC_SERVER_TRUSTED_PROXIES="127.0.0.1/32"
```

### Example nginx site

```nginx
server {
    listen 80;
    server_name account.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### OIDC note

The callback must match the public URL exactly:

```text
https://account.example.com/oidc-callback
```

If the app does not know the public URL, it can generate the wrong callback origin or scheme and the provider will reject the login flow.

#### Recommended production setup

1. set `AC_INSTANCE_BASE_URL` to the public URL;
2. trust only the proxy IP/CIDR with `AC_SERVER_TRUSTED_PROXIES`;
3. register the same callback URL in the OIDC provider.

## Scaling

For multiple instances:

- use Redis-backed sessions;
- keep the same `AC_REDIS_KEY_PREFIX` across the deployment;
- ensure every instance uses the same OIDC client and public base URL.

## See also

- [Environment variables](environment-variables.md)
- [OIDC configuration](oidc.md)
- [Services catalog](services-catalog.md)
- [Knowledge base](knowledge-base.md)
