# Forgia Docs

Documentation website for [Forgia](https://github.com/forgia-labs/forgia) — built with Nuxt 4, Nuxt UI v4, and Nuxt Content v3. Deployed to GitHub Pages on every `v*.*.*` release tag.

Live site: **https://forgia-labs.github.io/forgia/**

## Stack

- **Nuxt 4** — Vue framework with static generation (`nuxt generate`)
- **Nuxt UI v4** — component library (navigation, TOC, layout)
- **Nuxt Content v3** — Markdown-based content with SQLite index at build time
- **pnpm 10** — package manager

## Local development

```bash
pnpm install
pnpm dev        # http://localhost:3000/forgia/
```

## Build & verify

```bash
pnpm generate   # builds to .output/public/
pnpm smoke      # verifies critical pages exist in .output/public/
pnpm typecheck  # TypeScript check
pnpm lint       # ESLint
```

---

## Adding documentation pages

All content lives in `content/docs/`. Nuxt Content reads every `.md` file there and assigns it a URL path matching the directory structure.

### 1. Create the Markdown file

```
content/docs/<section>/<page-name>.md
```

Every file needs this frontmatter:

```md
---
title: "Page Title"
description: "One-sentence description — used for SEO and link previews."
navigation:
  title: "Sidebar label"   # optional: shorter label for the left nav
---

# Page Title

Content here...
```

### 2. URL mapping

| File path | URL |
|-----------|-----|
| `content/docs/getting-started/installation.md` | `/docs/getting-started/installation` |
| `content/docs/cli/exec.md` | `/docs/cli/exec` |
| `content/docs/my-section/index.md` | `/docs/my-section` |

### 3. Add a new section

Create an `index.md` for the section landing page:

```
content/docs/my-section/index.md
```

The left sidebar automatically picks up the section and its pages from the frontmatter — no manual registration needed.

### 4. Internal links

Use root-relative paths without the `/forgia/` baseURL prefix — Nuxt rewrites them at build time:

```md
See [Creating an FD](/docs/fd/creating) for details.   ✅
See [Creating an FD](/forgia/docs/fd/creating) ...     ❌ don't hardcode the baseURL
```

### 5. Table of contents

The right sidebar TOC is built automatically from the `##` and `###` headings in the page. No extra configuration needed — just use proper heading hierarchy.

### 6. Verify before pushing

```bash
pnpm generate && pnpm smoke
```

The smoke test checks that all critical pages rendered correctly. If you added a new required page, add it to `scripts/smoke-test.sh`.

---

## Deployment

The site deploys automatically when a release tag matching `v*.*.*` is pushed:

```bash
git tag v1.2.0
git push origin v1.2.0
```

The GitHub Actions workflow (`.github/workflows/docs.yml`) runs `pnpm generate`, uploads `.output/public/`, and deploys to GitHub Pages via OIDC — no PAT required.

Pushing to `main` or opening a PR does **not** trigger a deploy.
