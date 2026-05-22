# Deployment

Account Center can be deployed:

1. as a standalone binary,
2. as a container,
3. as a Docker Compose stack,
4. behind a reverse proxy such as Nginx.

Published container repositories:

- `icikowski/account-center`
- `ghcr.io/icikowski/account-center`

## Common requirements

Every deployment needs:

- a valid OIDC provider,
- a non-empty catalog file,
- environment variables for OIDC,
- and, if enabled, a knowledge base directory.

Recommended for production:

- set `AC_INSTANCE_BASE_URL` to the public URL,
- set `AC_SERVER_TRUSTED_PROXIES` to your reverse proxy CIDR(s),
- use Redis if sessions must survive restarts,
- and store configuration outside the container image.

## Standalone binary

Build from source:

```bash
task generate
task build-static
```

Run it:

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

## Docker

The published image bakes in:

- `AC_CATALOG_PATH=/data/catalog.yaml`
- `AC_KB_PATH=/data/kb`

So the simplest pattern is to mount your content under `/data`.

```bash
docker run -d \
  --name account-center \
  -p 8080:8080 \
  --env-file .env \
  -v "$PWD/catalog.yaml:/data/catalog.yaml:ro" \
  -v "$PWD/kb:/data/kb:ro" \
  ghcr.io/icikowski/account-center:<tag>
```

If the knowledge base is disabled, you can omit the `kb` mount.

To use Docker Hub instead:

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
    image: ghcr.io/icikowski/account-center:<tag>
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

Example `.env`:

```dotenv
AC_OIDC_PROVIDER_URL="https://sso.example.com"
AC_OIDC_CLIENT_ID="account-center"
AC_OIDC_CLIENT_SECRET="replace-me"
AC_OIDC_REFRESH_BEFORE=1m
```

### Notes for Compose users

- The application image is distroless, so do not rely on shell-based troubleshooting inside the container.
- The application does **not** expose a dedicated health endpoint; use TCP checks or an external HTTP probe against `/`.
- If you disable Redis, remove the Redis service and the `AC_REDIS_*` settings.
- Add `AC_SERVER_TRUSTED_PROXIES` only when a trusted reverse proxy is actually in front of the app.
- In this repository, `.env` is a documented workflow file and is loaded by the committed `.envrc`.

## Nginx reverse proxy

Recommended environment variables when proxying through Nginx:

```dotenv
AC_INSTANCE_BASE_URL="https://account.example.com"
AC_SERVER_TRUSTED_PROXIES="127.0.0.1/32"
```

Recommended Nginx site:

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

### Why this matters for OIDC

The OIDC callback must match the public URL exactly:

```text
https://account.example.com/oidc-callback
```

If the app sits behind a reverse proxy and does not know the public URL, it can generate the wrong callback origin or scheme and the provider will reject the login flow.

The safest production approach is:

1. set `AC_INSTANCE_BASE_URL` to the public URL,
2. trust only the proxy IP/CIDR with `AC_SERVER_TRUSTED_PROXIES`,
3. and register the same callback URL in the OIDC provider.

## Scaling considerations

For multiple application instances:

- use Redis-backed sessions,
- keep the same `AC_REDIS_KEY_PREFIX` across the shared deployment,
- and ensure all instances use the same OIDC client and public base URL.

## Related guides

- [Environment variables](environment-variables.md)
- [OIDC configuration](oidc.md)
- [Services catalog](services-catalog.md)
- [Knowledge base](knowledge-base.md)
