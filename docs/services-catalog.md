# Services catalog

The services catalog is the core data source for Account Center. It defines:

- which services exist,
- which URL and icon each service uses,
- and how OIDC groups map to visible roles.

The catalog file is required and must contain at least one service.

## File location

By default the application expects:

- `./catalog.yaml` for bare-binary deployments,
- `/data/catalog.yaml` in the published container image.

Override this with `AC_CATALOG_PATH` if needed.

## Schema reference

The repository includes a reference schema at:

```text
schemas/catalog.schema.json
```

You can use this editor-assistance header in your catalog file:

```yaml
# yaml-language-server: $schema=https://git.sr.ht/~icikowski/account-center/blob/main/schemas/catalog.schema.json
```

## Minimal example

```yaml
# yaml-language-server: $schema=https://git.sr.ht/~icikowski/account-center/blob/main/schemas/catalog.schema.json

global_access:
  platform-admins: system_administrator

services:
  - name: Grafana
    url: https://grafana.example.com
    icon: https://assets.example.com/icons/grafana.png
    roles:
      grafana-admins: administrator
      grafana-editors: editor
      grafana-viewers: viewer

  - name: Wiki
    url: https://wiki.example.com
```

In that example:

- members of `platform-admins` receive `system_administrator` on **every** service,
- Grafana-specific groups can grant a more specific role,
- and `Wiki` is visible to every authenticated user because it has no `roles` block.

## Top-level fields

| Field           | Type    | Required | Purpose                                 |
| --------------- | ------- | -------- | --------------------------------------- |
| `global_access` | mapping | No       | Groups that get access to every service |
| `services`      | array   | Yes      | List of service definitions             |

## Service fields

| Field   | Type       | Required | Purpose                                |
| ------- | ---------- | -------- | -------------------------------------- |
| `name`  | string     | Yes      | Service display name                   |
| `url`   | URI string | Yes      | Service base URL                       |
| `icon`  | URI string | No       | Optional icon shown in the UI          |
| `roles` | mapping    | No       | Group-to-role mapping for this service |

## Role values

Operator-assignable role values are:

| Role                   | Typical meaning                     |
| ---------------------- | ----------------------------------- |
| `superuser`            | Highest level of access             |
| `system_administrator` | Platform-wide administrative access |
| `administrator`        | Service admin                       |
| `redactor`             | Content-heavy write access          |
| `editor`               | Standard editing access             |
| `viewer`               | Read-only or dashboard access       |
| `user`                 | Normal user access                  |
| `guest`                | Lowest explicit access level        |

### Important note about `general_access`

You may see `general_access` in the codebase, but it is **not** meant to be authored in the catalog. It is assigned automatically when a service has no `roles` section, meaning every authenticated user can see that service in Account Center.

## How access is evaluated

Account Center evaluates access like this:

1. Read the user's `groups` claim from OIDC.
2. Match those groups against the service-specific `roles` map.
3. Match those same groups against `global_access`.
4. If a service has no `roles`, assign automatic `general_access`.
5. If multiple roles match, keep the highest role according to the built-in hierarchy.

### Effective role order

From highest to lowest:

1. `superuser`
2. `system_administrator`
3. `administrator`
4. `redactor`
5. `editor`
6. `viewer`
7. `user`
8. `guest`

The catalog controls **visibility and displayed effective role inside Account Center**. It does not replace authorization inside the downstream service itself.

## Authoring guidance

### Use stable group names

The catalog must use the exact group names emitted by your OIDC provider in the `groups` claim.

### Keep URLs absolute

Both `url` and `icon` should be absolute URLs.

### Prefer explicit service mappings

Use `global_access` for broad platform-wide access only. Put service-specific group mappings under each service when access differs between tools.

## Common validation failures

| Problem                                      | Result                                              |
| -------------------------------------------- | --------------------------------------------------- |
| Missing `services` array                     | Startup failure                                     |
| Empty `services` array                       | Startup failure                                     |
| Invalid role name                            | Startup failure                                     |
| Invalid service URL or icon URL              | Startup failure                                     |
| Group names do not match OIDC `groups` claim | Users authenticate but do not see expected services |

## Example design pattern

A practical catalog often uses:

- one `global_access` group for broad administrative access,
- service-specific user groups,
- and different role levels for tools like Grafana.

That is a good structure to follow when you want one portal to aggregate several unrelated services without duplicating access logic in the portal itself.

## Related guides

- [OIDC configuration](oidc.md)
- [Knowledge base](knowledge-base.md)
- [Environment variables](environment-variables.md)
