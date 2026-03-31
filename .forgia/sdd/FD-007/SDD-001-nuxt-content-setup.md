---
id: "SDD-001"
fd: "FD-007"
title: "Nuxt Content Setup & Docs Layout"
status: done
agent: "claude-code"
assigned_to: "claude"
created: "2026-04-01"
started: ""
completed: ""
tags: ["nuxt", "docs", "frontend"]
---

# SDD-001: Nuxt Content Setup & Docs Layout

> Parent FD: [[FD-007]]

## Scope

The `docs-site/` scaffold already exists (Nuxt 4 + Nuxt UI v4, untracked). This SDD must:

1. Add `@nuxt/content` v3 to the dependencies
2. Configure `nuxt.config.ts` with: Content module, `baseURL: '/forgia/'`, prerender for all content pages
3. Add the `"generate": "nuxt generate"` script to `package.json`
4. Create the `docs` layout with: navigation sidebar, table of contents, main content area
5. Create the catch-all page `app/pages/docs/[...slug].vue` that serves Nuxt Content
6. Create the `content/` directory structure with an index and placeholder folders
7. Commit everything to `main` as the first tracked commit of `docs-site/`

**Do not** touch the homepage (`app/pages/index.vue`) or existing components (`HeroSection`, `FeaturesSection`, `CtaSection`).

### Current repository state

```
docs-site/
├── app/
│   ├── app.vue              ✅ header/footer UHeader/UFooter
│   ├── app.config.ts        ✅ custom colors: forge, anvil
│   ├── assets/css/main.css  ✅
│   ├── assets/images/       ✅
│   ├── pages/index.vue      ✅ landing page
│   └── components/          ✅ HeroSection, FeaturesSection, CtaSection
├── nuxt.config.ts           ✅ (missing: @nuxt/content, baseURL, prerender)
├── package.json             ✅ (missing: @nuxt/content, generate script)
├── pnpm-lock.yaml           ✅
├── pnpm-workspace.yaml      ✅
├── eslint.config.mjs        ✅
├── tsconfig.json            ✅
└── public/                  ✅
```

### Expected output

```
docs-site/
├── app/
│   ├── layouts/
│   │   └── docs.vue         ← NEW: sidebar + TOC + content area
│   └── pages/
│       └── docs/
│           └── [...slug].vue ← NEW: catch-all for Nuxt Content
├── content/
│   ├── index.md             ← NEW: redirect or short intro
│   ├── getting-started/
│   │   └── .gitkeep         ← placeholder (content in SDD-002)
│   ├── fd/
│   │   └── .gitkeep
│   ├── sdd/
│   │   └── .gitkeep
│   ├── cli/
│   │   └── .gitkeep
│   └── constitution/
│       └── .gitkeep
├── nuxt.config.ts           ← MODIFIED
└── package.json             ← MODIFIED
```

## Interfaces

| Interface | Type | Description |
|-----------|------|-------------|
| `nuxt.config.ts` | Config | Exports config with `@nuxt/content`, `baseURL: '/forgia/'`, prerender `/**` |
| `app/layouts/docs.vue` | Vue SFC | Layout with `default` slot for content, left sidebar, right TOC |
| `app/pages/docs/[...slug].vue` | Vue SFC | Catch-all page using `<ContentDoc />` or `queryCollection()` |
| `content/` directory | Filesystem | Folder structure for SDD-002 — contract: `getting-started/`, `fd/`, `sdd/`, `cli/`, `constitution/` |
| `package.json` | JSON | `generate: nuxt generate` script, `@nuxt/content` v3 dependency |

**Contract with SDD-002**: the `content/` structure created here must exactly match the folders where SDD-002 will write Markdown files.

**Contract with SDD-003**: `nuxt generate` (added here) must complete with exit code 0 so that the CI workflow can deploy.

## Constraints

- Language: TypeScript, Vue 3 Composition API with `<script setup>`
- Framework: Nuxt 4, Nuxt UI v4 (`@nuxt/ui: ^4.6.0`), Nuxt Content v3 (`@nuxt/content: ^3.x`)
- Package manager: `pnpm` (do not use npm or yarn)
- `baseURL` required: `/forgia/` — no trailing slash in routes, trailing slash in baseURL
- Prerender: configure `nitro.prerender.routes` or `routeRules` to prerender `/docs/**`
- The `docs.vue` layout must use **only Nuxt UI v4 components** for sidebar and TOC (e.g. `UNavigationMenu`, `UContentToc` or equivalent available components)
- Strict TypeScript: no explicit `any` in new pages/layouts
- Do not modify `app/pages/index.vue` or landing page components
- Respect the existing ESLint pattern (commaDangle: never, braceStyle: 1tbs)
- Files listed in `deny.toml` must not be touched: no `.env`, no secrets, no Forgia config files

## Best Practices

- Error handling: if `queryCollection()` returns null, show an explicit 404 page (do not silently fail)
- Naming: Vue components in PascalCase, files in kebab-case
- Style: Composition API `<script setup>` on all new components/pages; no Options API
- The docs layout must be mobile-responsive (collapsible sidebar on mobile using `USlideover` or equivalent pattern)
- Do not add custom CSS: use only Tailwind utility classes and Nuxt UI components

## Test Requirements

| Type | What | Coverage |
|------|------|----------|
| Build | `pnpm run generate` completes without errors | 100% — build must pass |
| Build | Asset paths: no 404s with `/forgia/` as baseURL | Verify `.output/public/` output |
| Manual | Docs layout correctly rendered at `/docs/` with sidebar | Local smoke test `pnpm dev` |
| TypeScript | `pnpm run typecheck` with no errors | 100% |
| Lint | `pnpm run lint` with no errors | 100% |

## Acceptance Criteria

- [ ] `@nuxt/content` v3 added to `package.json` and installed via `pnpm install`
- [ ] `nuxt.config.ts` includes `@nuxt/content` in `modules` and `baseURL: '/forgia/'`
- [ ] Script `"generate": "nuxt generate"` present in `package.json`
- [ ] `pnpm run generate` completes without errors and produces `.output/public/`
- [ ] `app/layouts/docs.vue` exists with navigation sidebar and main content area
- [ ] `app/pages/docs/[...slug].vue` exists and serves Nuxt Content
- [ ] `content/` structure created with 5 folders: `getting-started/`, `fd/`, `sdd/`, `cli/`, `constitution/`
- [ ] `pnpm run typecheck` and `pnpm run lint` pass without errors
- [ ] Everything committed to `main` with message `feat(FD-007): add Nuxt Content and docs layout`

## Context

- [ ] `docs-site/nuxt.config.ts` — current config (no Content, no baseURL)
- [ ] `docs-site/package.json` — current dependencies (no @nuxt/content)
- [ ] `docs-site/app/app.vue` — root layout with UHeader/UFooter; the docs layout sits inside this
- [ ] `docs-site/app/app.config.ts` — custom colors `forge` and `anvil`
- [ ] `docs/getting-started.md` — reference content (do not copy, use as structural guide)
- [ ] Nuxt Content v3 docs: https://content.nuxt.com/
- [ ] Nuxt UI v4 docs: https://ui.nuxt.com/

## Constitution Check

- [ ] Respects code standards: TypeScript strict, no any, Composition API
- [ ] Respects commit conventions: `feat(FD-007): add Nuxt Content and docs layout`
- [ ] No hardcoded secrets: no tokens, no API keys
- [ ] Tests defined: build test + typecheck + lint

---

## Work Log

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent

- **Executor**: claude-code
- **Started**: 2026-04-01
- **Completed**: 2026-04-01
- **Duration**: ~1 session

### Decisions

1. Moved content tree from `content/` into `content/docs/` so Nuxt Content assigns paths like `/docs/getting-started`, matching the URL pattern used in `[...slug].vue` query (`/docs/${slug}`). Without this, all docs pages would return 404 at build time.
2. Added `content.config.ts` at project root — required by Nuxt Content v3 to define collections; without it the module logs a warning and falls back to defaults which broke the build.
3. Added `onlyBuiltDependencies: [better-sqlite3]` in `pnpm-workspace.yaml` — pnpm 10 ignores native build scripts by default; `better-sqlite3` requires a native `.node` binary compiled for the current Node version.
4. Added a minimal `content/docs/getting-started/index.md` stub so the prerender crawler (which follows links from `content/docs/index.md`) does not 404 during `nuxt generate`.

### Output

- **Commit(s)**: `bbe8a0c`, `43ec24a`
- **PR**: —
- **Files created/modified**:
  - `docs-site/nuxt.config.ts`
  - `docs-site/package.json`
  - `docs-site/pnpm-lock.yaml`
  - `docs-site/pnpm-workspace.yaml`
  - `docs-site/content.config.ts`
  - `docs-site/app/layouts/docs.vue`
  - `docs-site/app/pages/docs/[...slug].vue`
  - `docs-site/content/docs/index.md`
  - `docs-site/content/docs/getting-started/index.md`
  - `docs-site/content/docs/getting-started/.gitkeep`
  - `docs-site/content/docs/fd/.gitkeep`
  - `docs-site/content/docs/sdd/.gitkeep`
  - `docs-site/content/docs/cli/.gitkeep`
  - `docs-site/content/docs/constitution/.gitkeep`

### Retrospective

- **What worked**: Nuxt Content v3 + Nuxt UI v4 integration is clean; `queryCollectionNavigation` + `ContentRenderer` pattern works well for docs sites.
- **What didn't**: pnpm 10 silently ignores build scripts for native modules — `better-sqlite3` built fine once added to `onlyBuiltDependencies`, but the failure was not obvious from the initial error message. Future FDs using native Node modules with pnpm 10 need this.
- **Suggestions for future FDs**: When adding `@nuxt/content` to a new project, always create `content.config.ts` first and place content under `content/docs/` immediately to match the URL routing convention.
