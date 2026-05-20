# Knowledge base

The knowledge base is an optional module that serves Markdown articles from a directory on disk and exposes them under `/kb`.

It is designed for operator-maintained runbooks, onboarding guides, and internal service documentation that should live next to the services catalog.

## Enablement

```dotenv
AC_KB_ENABLED=true
AC_KB_PATH="./kb"
AC_KB_RELOAD_DEBOUNCE=500ms
```

If `AC_KB_ENABLED=false`, the route still exists but shows the disabled-state page instead of article content.

## Directory model

Each article is a Markdown file with YAML front matter.

Example layout:

```text
kb/
  grafana/
    index.md
    assets/
      hero.svg
  prometheus/
    index.md
  runbooks/
    reset-password.md
```

## Front matter

Reference schema:

```text
schemas/frontmatter.schema.json
```

Supported fields:

| Field            | Required | Purpose                                |
| ---------------- | -------- | -------------------------------------- |
| `title`          | Yes      | Article title                          |
| `description`    | Yes      | Short article summary used in listings |
| `featured_image` | No       | Optional image path or URL             |

Minimal example:

```markdown
---
title: Building overview dashboards in Grafana
description: Placeholder runbook for arranging panels, validating queries, and making dashboards screenshot-ready.
featured_image: ./assets/hero.svg
---

## Building overview dashboards in Grafana

Start with a compact top row for health and alert status.
```

Every article must start with YAML front matter delimited by `---`. Missing or unclosed front matter is rejected.

## Slug rules

Slugs are derived from file paths:

| File                            | URL slug                   |
| ------------------------------- | -------------------------- |
| `kb/grafana/index.md`           | `/grafana`                 |
| `kb/runbooks/reset-password.md` | `/runbooks/reset-password` |
| `kb/prometheus/index.md`        | `/prometheus`              |

### Important restriction

`kb/index.md` at the root is **not allowed**. Root-level `index.md` causes validation to fail.

Use named files or subdirectories instead.

## Internal links and assets

Account Center rewrites relative references during KB loading.

### Relative Markdown links

Links to other Markdown files are turned into internal KB routes.

Example:

```markdown
[Reviewing scrape targets in Prometheus](../prometheus/index.md)
```

Becomes a link to the matching `/kb/...` article route at runtime.

### Relative assets

Relative `src` and non-Markdown `href` values are treated as KB assets and served through `/kb/attachments/...`.

Example:

```markdown
![Hero image](./assets/hero.svg)
```

### External and absolute URLs

These are left unchanged:

- `https://docs.example.com/...`
- `mailto:...`
- `/absolute/path`

## Asset rules

The KB loader rejects:

- references to missing files,
- references to directories,
- paths that escape the KB root,
- and `featured_image` values that point to Markdown files.

Relative asset paths are resolved from the current article location.

## Writing guidelines

### Good patterns

- Keep one service or one operational task per directory.
- Use `index.md` inside a directory when you want a short, clean slug such as `/grafana`.
- Keep shared images or diagrams next to the article that uses them.
- Link related articles with relative Markdown links.

### Avoid

- root `index.md`,
- missing `title` or `description`,
- links that point outside the KB directory,
- and assuming raw file paths will be exposed directly.

## Example design pattern

A practical KB structure usually looks like:

- one directory per service,
- `index.md` as the main page for that service,
- optional `featured_image`,
- related-article links using relative Markdown paths,
- and direct links out to the service itself when useful.

## Operational behavior

- The knowledge base is loaded and validated at startup.
- When enabled, file changes are watched and reloaded automatically.
- Invalid content prevents a successful reload and is surfaced as a validation error.

## Related guides

- [Services catalog](services-catalog.md)
- [Environment variables](environment-variables.md)
- [Deployment](deployment.md)
