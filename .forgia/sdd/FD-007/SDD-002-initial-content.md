---
id: "SDD-002"
fd: "FD-007"
title: "Initial Documentation Content"
status: planned
agent: ""
assigned_to: ""
created: "2026-04-01"
started: ""
completed: ""
tags: ["docs", "content", "markdown"]
---

# SDD-002: Initial Documentation Content

> Parent FD: [[FD-007]]

## Scope

Write the initial Markdown content for the documentation site, structured across the 5 site sections. Content uses existing files in `docs/` and `CLAUDE.md` as reference, but is **not** copied verbatim — it must be rewritten/adapted for the web format (short pages, front matter, internal links).

**Dependency**: SDD-001 must be completed first — the `content/` structure and docs layout must exist.

### Sections to produce

```
content/
├── index.md                          ← redirect to /docs/getting-started
├── getting-started/
│   ├── index.md                      ← overview + prerequisites
│   ├── installation.md               ← install forgia CLI, mise, Claude Code
│   └── first-feature.md              ← full workflow: fd-new → fd-review → fd-sdd → exec → verify → close
├── fd/
│   ├── index.md                      ← what is an FD, when to use it
│   ├── creating.md                   ← /fd-new, from issue or free-text
│   ├── reviewing.md                  ← /fd-review, what it checks
│   └── closing.md                    ← /fd-close, /fd-verify
├── sdd/
│   ├── index.md                      ← what is an SDD, structure
│   ├── generating.md                 ← /fd-sdd
│   └── executing.md                  ← forgia exec, forgia batch, /sdd-assign
├── cli/
│   ├── index.md                      ← commands overview
│   ├── init.md                       ← forgia init
│   ├── status.md                     ← forgia status
│   ├── doctor.md                     ← forgia doctor
│   └── exec.md                       ← forgia exec / batch / watch
└── constitution/
    └── index.md                      ← rules, commit conventions, code standards
```

### Required front matter for every file

```yaml
---
title: "Page title"
description: "Short description (used for SEO and social cards)"
navigation:
  title: "Sidebar title"  # if different from title
---
```

### Tone and language

- All content: English
- Code blocks, command names, variable names: English
- Style: concise, technical, action-oriented — no marketing language

## Interfaces

| Interface | Type | Description |
|-----------|------|-------------|
| `content/**/*.md` | Markdown + YAML front matter | Contract with SDD-001: every file must have `title`, `description`, `navigation.title` in front matter |
| Sidebar navigation | Front matter `navigation` | SDD-001 uses `queryCollection()` to build the menu — files must respect the directory structure |
| Internal links | Markdown `[text](/docs/section/page)` | All internal links use absolute paths with `/docs/` prefix (consistent with `baseURL: /forgia/`) |

**Contract with SDD-001**: the folders written here must exactly match the structure created by SDD-001. Do not create additional folders without coordination.

**Contract with SDD-004**: every page produced here must be reachable via `nuxt generate` — broken internal links will fail the build.

## Constraints

- Language: Markdown (CommonMark) with YAML front matter
- Framework: Nuxt Content v3 — use MDC syntax if needed for inline Vue components
- Every `.md` file must have front matter with at least: `title`, `description`
- Do not copy verbatim blocks from `docs/functional-architecture.md` — complex Mermaid diagrams should be simplified or removed for the site (too verbose for a user discovering Forgia)
- No unverified external links
- No `.env` files, secrets or credentials in content
- `constitution/index.md` is a summary — NOT a copy of `.forgia/constitution.md` (that is the project's authoritative source, not the site)

## Best Practices

- Error handling: every page must be self-contained — a user landing directly on the page must understand the context without having read the previous ones
- Naming: files in kebab-case, folders in kebab-case
- Style: short sections (max ~300 words per page), many code examples, frequent internal links

## Test Requirements

| Type | What | Coverage |
|------|------|----------|
| Build | `pnpm run generate` with the produced content files — no broken internal links (no 404) | All `.md` files |
| Manual | Every page is reachable from the sidebar navigation | Local smoke test |
| Manual | Front matter `title` and `description` present on every page | 100% of files |

## Acceptance Criteria

- [ ] All 5 sections present with at least 1 content page each (not just `.gitkeep`)
- [ ] Every `.md` file has front matter with `title` and `description`
- [ ] `pnpm run generate` completes without errors with this content (no broken internal links)
- [ ] The docs layout sidebar shows structured navigation across the 5 sections
- [ ] Getting Started / `first-feature.md` covers the full workflow: `fd-new → fd-review → fd-sdd → exec → verify → close`
- [ ] CLI Reference documents at least: `init`, `status`, `doctor`, `exec`
- [ ] Commit: `docs(FD-007): add initial documentation content`

## Context

- [ ] `docs/getting-started.md` — installation and first-use guide (primary source for getting-started/)
- [ ] `docs/concepts.md` — FD/SDD concepts (source for fd/ and sdd/)
- [ ] `docs/functional-architecture.md` — architecture (reference for CLI and advanced concepts)
- [ ] `docs/go-architecture.md` — Go architecture (reference for CLI Reference)
- [ ] `CLAUDE.md` — project rules (source for constitution/)
- [ ] `.forgia/constitution.md` — immutable rules (source for constitution/index.md)
- [ ] `docs-site/app/layouts/docs.vue` — layout already created by SDD-001: understand how the sidebar is built to write front matter correctly

## Constitution Check

- [ ] Respects code standards: valid Markdown, well-formed YAML front matter
- [ ] Respects commit conventions: `docs(FD-007): add initial documentation content`
- [ ] No hardcoded secrets: no tokens, no API keys in content
- [ ] Tests defined: build test (generate) + manual smoke test

---

## Work Log

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent

- **Executor**: <!-- openhands | claude-code | manual | name -->
- **Started**: <!-- timestamp -->
- **Completed**: <!-- timestamp -->
- **Duration**: <!-- total time -->

### Decisions

1. <!-- decision 1: what and why -->

### Output

- **Commit(s)**: <!-- hash -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `docs-site/content/index.md`
  - `docs-site/content/getting-started/index.md`
  - `docs-site/content/getting-started/installation.md`
  - `docs-site/content/getting-started/first-feature.md`
  - `docs-site/content/fd/index.md`
  - `docs-site/content/fd/creating.md`
  - `docs-site/content/fd/reviewing.md`
  - `docs-site/content/fd/closing.md`
  - `docs-site/content/sdd/index.md`
  - `docs-site/content/sdd/generating.md`
  - `docs-site/content/sdd/executing.md`
  - `docs-site/content/cli/index.md`
  - `docs-site/content/cli/init.md`
  - `docs-site/content/cli/status.md`
  - `docs-site/content/cli/doctor.md`
  - `docs-site/content/cli/exec.md`
  - `docs-site/content/constitution/index.md`

### Retrospective

- **What worked**:
- **What didn't**:
- **Suggestions for future FDs**:
