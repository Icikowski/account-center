# OIDC configuration

**Account Center** uses the OIDC authorization-code flow with refresh tokens.

## What the application expects

### Redirect URI

The callback path is fixed:

```text
/oidc-callback
```

### Production redirect URI

```text
https://account.example.com/oidc-callback
```

### Local development

```text
http://localhost:8080/oidc-callback
```

### Required scopes

**Account Center** requests following scopes:

- `openid`
- `profile`
- `email`
- `groups`
- `offline_access`

All five matter:

- `profile` and `email` provide user identity data for the UI;
- `groups` drives catalog access evaluation;
- `offline_access` enables refresh tokens.

### Required claims

| Claim    | Why it matters                                  |
| -------- | ----------------------------------------------- |
| `sub`    | Stable subject identifier                       |
| `name`   | Display name in the UI                          |
| `email`  | Displayed in the UI                             |
| `groups` | Determines visible services and effective roles |

The `groups` claim name is fixed and not configurable. It should be a JSON array of strings.

## Required environment variables

```dotenv
AC_OIDC_PROVIDER_URL="https://sso.example.com"
AC_OIDC_CLIENT_ID="account-center"
AC_OIDC_CLIENT_SECRET="replace-me"
AC_OIDC_REFRESH_BEFORE=1m
```

For reverse-proxy deployments, also set:

```dotenv
AC_INSTANCE_BASE_URL="https://account.example.com"
AC_SERVER_TRUSTED_PROXIES="<your proxy CIRD>,127.0.0.1/32"
```

Use a reverse proxy to terminate TLS/SSL before requests reach **Account Center**.

## Sanitized Authelia example

This example uses the callback path expected by **Account Center**.

```yaml
identity_providers:
  oidc:
    clients:
      - client_id: "account-center"
        client_name: "Account Center"
        client_secret: "superSecretSecret"
        public: false
        authorization_policy: one_factor
        redirect_uris:
          - https://account.example.com/oidc-callback
          - http://localhost:8080/oidc-callback
        scopes:
          - openid
          - profile
          - email
          - groups
          - offline_access
        grant_types:
          - authorization_code
          - refresh_token
        response_types:
          - code
        token_endpoint_auth_method: client_secret_post
```

## Validation checklist

Before looking elsewhere, verify if:

1. the provider URL is correct and reachable from Account Center;
2. the client ID and secret match the registered OIDC client;
3. the redirect URI uses `/oidc-callback`;
4. the provider returns a `groups` claim;
5. `offline_access` is enabled if you expect refresh-token behavior;
6. the public URL seen by the provider matches `AC_INSTANCE_BASE_URL`.

## Common failure modes

| Symptom                                   | Likely cause                                                       |
| ----------------------------------------- | ------------------------------------------------------------------ |
| Redirect mismatch at login                | Wrong callback URI, wrong scheme or missing `AC_INSTANCE_BASE_URL` |
| User logs in but sees no services         | Missing or empty `groups` claim or no matching catalog groups      |
| Sessions do not refresh                   | `offline_access` missing or provider not issuing refresh tokens    |
| Works on localhost but fails behind Nginx | Missing trusted proxy configuration or wrong public base URL       |

## See also

- [Configuration overview](configuration.md)
- [Environment variables](environment-variables.md)
- [Deployment](deployment.md)
- [Services catalog](services-catalog.md)
