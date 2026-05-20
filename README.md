# Account Center

**Self-hosted, OIDC-authenticated portal for internal services and knowledge base articles.**

Account Center is a self-hosted account portal for organizations that already use OpenID Connect and want a clean place to send users after login. It combines a role-aware services catalog with an optional knowledge base, so users land on one page that shows **what they can access** and **how to use it**.

It is intentionally small and operator-friendly: configuration lives in environment variables, a YAML catalog file, and an optional directory of Markdown articles. The application validates that input on startup, reloads catalog and knowledge base changes automatically, and can keep sessions in memory or Redis.

| Login                                             | Services catalog                                      |
| ------------------------------------------------- | ----------------------------------------------------- |
| ![Login screen](docs/assets/screenshot-login.png) | ![Catalog screen](docs/assets/screenshot-catalog.png) |

| Knowledge base listing                                        | Knowledge base article                                           |
| ------------------------------------------------------------- | ---------------------------------------------------------------- |
| ![Knowledge base listing](docs/assets/screenshot-kb-list.png) | ![Knowledge base article](docs/assets/screenshot-kb-article.png) |

> **⚠️ Warning**
>
> This repository is hosted on _git.sr.ht_ and mirrored to GitHub.
> You should always refer to _git.sr.ht_ version as the primary instance.

## Why it exists

Most self-hosted environments already have:

- an identity provider,
- several internal tools,
- different user groups,
- and scattered documentation.

Account Center started from a practical gap: I wanted something with the convenience of an Okta-style application dashboard for a self-hosted stack, but couldn't find a good fit. The result is a portal that puts those pieces behind one consistent entry point:

- **OIDC sign-in** with refresh-token support.
- **Role-aware catalog** based on groups from the `groups` claim.
- **Optional knowledge base** rendered from Markdown with front matter and referenced assets.
- **Live reload** for catalog and knowledge base content.
- **Session storage choice** between in-memory and Redis-backed persistence.
- **Reverse-proxy aware deployment** with explicit trusted-proxy handling.
- **Small operational surface**: one binary, one container image, one catalog file, one optional KB directory.

## How it is structured

```text
cmd/account-center/        Main application entrypoint
internal/auth/             OIDC flow, session handling, trusted proxy logic
internal/catalog/          Catalog loading, validation, live reload
internal/knowledgebase/    Markdown loading, validation, link/asset rewriting
internal/evaluator/        Effective role calculation for visible services
internal/web/              HTTP routes, UI, assets, templates
internal/config/           Environment variable parsing and validation
schemas/                   JSON schemas for catalog and KB front matter
```

## Documentation

| Topic                                                  | Purpose                                                         |
| ------------------------------------------------------ | --------------------------------------------------------------- |
| [Configuration overview](docs/configuration.md)        | Runtime model, routes, live reload, storage, and proxy behavior |
| [Environment variables](docs/environment-variables.md) | Complete reference for every `AC_*` variable                    |
| [Deployment](docs/deployment.md)                       | Binary, Docker, Docker Compose, and Nginx deployment guidance   |
| [OIDC configuration](docs/oidc.md)                     | Provider requirements and a sanitized Authelia example          |
| [Services catalog](docs/services-catalog.md)           | YAML schema, roles, and access evaluation rules                 |
| [Knowledge base](docs/knowledge-base.md)               | Enablement, content layout, front matter, links, and assets     |

## Development workflow

### Prerequisites

- Go 1.26+
- [Task](https://taskfile.dev/)
- `golangci-lint`
- Tailwind CLI is used automatically by the task definitions

### Common commands

```bash
task generate
task build
task run
task test
task lint
```

For the live development loop:

```bash
task dev
```

That runs Tailwind in watch mode and starts the Templ development flow defined in `Taskfile.yml`.

### Setup notes

1. Start from the variable reference in `.env.example` or the docs in [`docs/environment-variables.md`](docs/environment-variables.md).
2. For local development, you can place the required `AC_*` variables in `.env`; the committed `.envrc` loads it automatically when using direnv.
3. Provide a valid catalog file and point `AC_CATALOG_PATH` at it.
4. Optionally enable the knowledge base with `AC_KB_ENABLED=true` and point `AC_KB_PATH` at your content directory.
5. If you are configuring Authelia, use the sanitized guide in [`docs/oidc.md`](docs/oidc.md).

## Contributing

This project's Go module path is `git.sr.ht/~icikowski/account-center`, and contributor-facing links in the documentation should treat **SourceHut** as the primary forge.

When contributing:

1. Keep generated assets in sync by running `task generate` after template, asset, or UI source changes.
2. Run `task fmt`, `task lint`, and `task test` before sending changes.
3. Update the docs when configuration, deployment behavior, or schemas change.
4. Prefer focused changes that preserve the operator-facing simplicity of the project.

## Container images

Published images are available as:

- `icikowski/account-center`
- `ghcr.io/Icikowski/account-center`

See [`docs/deployment.md`](docs/deployment.md) for complete examples.

## License

Licensed under the GNU GPL v3. See [`LICENSE`](LICENSE).

## Disclaimer

Code was written in spare time, without AI assistance as I find vibed crap unmaintainable. Documentation, on the other hand, was written with the help of AI, and may contain inaccuracies or inconsistencies - I have made an effort to review and correct it, but please report any issues you find.

AI-assisted contributions are welcome, but every contributor is expected to review and test the code they submit, and to be responsive to feedback on it. I will not merge contributions that I have not personally reviewed and tested, regardless of whether they were AI-assisted or not.

Use at your own risk, and expect a personal project with limited support and no SLA. Contributions are welcome, but I can't promise any specific turnaround time or feature roadmap.
