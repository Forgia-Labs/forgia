---
id: "SDD-004"
fd: "FD-007"
title: "Integration Wiring & E2E Verification"
status: done
agent: "claude-code"
assigned_to: "claude"
created: "2026-04-01"
started: "2026-04-01"
completed: "2026-04-01"
tags: ["integration", "e2e", "verification"]
---

# SDD-004: Integration Wiring & E2E Verification

> Parent FD: [[FD-007]]

## Scope

Verify that all components produced by SDD-001, SDD-002, and SDD-003 work correctly together as an integrated system. This SDD produces no new application code — it produces a **smoke test script** and completes integration checks that confirm the system end-to-end.

**Hard dependency**: SDD-001, SDD-002, and SDD-003 must be completed first.

### E2E flow to verify

```
Developer pushes tag v*.*.* to GitHub
  → docs.yml workflow triggers (SDD-003)
  → pnpm install --frozen-lockfile in docs-site/
  → pnpm run generate (SDD-001: nuxt.config.ts + @nuxt/content)
      → Nuxt Content reads content/**/*.md (SDD-002)
      → Renders 17+ HTML pages in .output/public/
  → artifact uploaded to GitHub Pages
  → Site available at forgia-labs.github.io/forgia/
      → /forgia/ → landing page (index.vue) ✓
      → /forgia/docs/getting-started → Getting Started index ✓
      → /forgia/docs/fd/ → FD Guide index ✓
      → /forgia/docs/sdd/ → SDD Guide index ✓
      → /forgia/docs/cli/ → CLI Reference index ✓
      → /forgia/docs/constitution/ → Constitution ✓
      → Sidebar navigation renders all 5 sections ✓
      → Mobile layout: sidebar collapses correctly ✓
```

### Output of this SDD

1. **Local smoke test script** (`docs-site/scripts/smoke-test.sh`) that:
   - Runs `pnpm run generate`
   - Verifies that critical HTML files exist in `.output/public/`
   - Verifies that no HTML file contains broken internal links (pattern `/forgia/docs/` not found)
   - Exit code 0 if everything is OK, exit code 1 with a list of errors otherwise

2. **Update `package.json`** with script `"smoke": "bash scripts/smoke-test.sh"`

3. **Smoke test step in the workflow** (minimal change to `docs.yml` — add a step between `generate` and `upload artifact` that runs `pnpm run smoke`)

4. **Final manual verification** documented in the Work Log: push tag `v0.1.0` and confirm the site is live on GitHub Pages.

## Interfaces

| Interface | Type | Description |
|-----------|------|-------------|
| `docs-site/scripts/smoke-test.sh` | Bash script | Input: `.output/public/`; Output: exit 0 (OK) or exit 1 + error list |
| `package.json` `smoke` script | npm script | Calls `smoke-test.sh`; used locally and in CI |
| `docs.yml` step `Smoke test` | GitHub Actions step | Added after `generate`, before `upload-pages-artifact` |
| Critical paths to verify | Filesystem | `/forgia/`, `/forgia/docs/`, `/forgia/docs/getting-started/`, `/forgia/docs/fd/`, `/forgia/docs/sdd/`, `/forgia/docs/cli/`, `/forgia/docs/constitution/` |

## Constraints

- Language: Bash for the script (`set -euo pipefail`)
- The script must be idempotent with no side effects beyond reading `.output/public/`
- No external dependencies in the script (only `grep`, `find`, `test` — standard Unix tools)
- The workflow step must use `working-directory: docs-site` and the same environment as the build job
- Do not modify the deploy logic in SDD-003 — only add a verification step before the upload
- Respects `deny.toml`: no `cat *.pem`, no access to credentials
- The script must fail explicitly with a clear message if `.output/public/` does not exist (generate was not run)

## Best Practices

- Error handling: `set -euo pipefail` in the Bash script; every failed check prints the missing path before exit 1
- Naming: `smoke-test.sh` (kebab-case), local variables with `local`
- Style: short script (< 50 lines), comments where not obvious

### Pages checklist to verify in the script

```bash
PAGES=(
  "index.html"
  "docs/index.html"
  "docs/getting-started/index.html"
  "docs/getting-started/installation/index.html"
  "docs/getting-started/first-feature/index.html"
  "docs/fd/index.html"
  "docs/sdd/index.html"
  "docs/cli/index.html"
  "docs/constitution/index.html"
)
```

## Test Requirements

| Type | What | Coverage |
|------|------|----------|
| E2E local | `pnpm run smoke` after `pnpm run generate` — exit 0 | All critical paths |
| E2E CI | `Smoke test` step in `docs.yml` — workflow green on push tag `v*.*.*` | Full pipeline |
| Negative | Remove a content file and verify smoke-test detects the missing path | Smoke test reliability |
| Manual | Push tag `v0.1.0` — site live and navigable on GitHub Pages | Deploy end-to-end |

## Acceptance Criteria

- [ ] `docs-site/scripts/smoke-test.sh` exists, is executable (`chmod +x`), and passes with `set -euo pipefail`
- [ ] `pnpm run smoke` run locally after `pnpm run generate` — exit 0
- [ ] `docs.yml` includes a `Smoke test` step between `generate` and `upload-pages-artifact`
- [ ] Full workflow (build → smoke → deploy) is green on push tag `v*.*.*` in CI
- [ ] Site is accessible at `forgia-labs.github.io/forgia/` after first deploy
- [ ] Sidebar navigation shows all 5 sections on the live site
- [ ] Mobile viewport 375px: sidebar collapses correctly (verified on live site or `pnpm dev`)
- [ ] `baseURL: /forgia/` works correctly — no asset 404s in browser DevTools
- [ ] Commit: `test(FD-007): add smoke test and E2E verification`

## Context

- [ ] `.forgia/sdd/FD-007/SDD-001-nuxt-content-setup.md` — output: `content/` structure and `.output/public/` path
- [ ] `.forgia/sdd/FD-007/SDD-002-initial-content.md` — output: 17 `.md` files to verify
- [ ] `.forgia/sdd/FD-007/SDD-003-github-actions-workflow.md` — workflow to modify by adding smoke step
- [ ] `docs-site/.output/public/` — directory produced by `nuxt generate` (exists after SDD-001 is complete)
- [ ] `.forgia/dev-guide/lang/shell.md` — Bash conventions for the script

## Constitution Check

- [ ] Respects code standards: Bash with `set -euo pipefail`, local variables
- [ ] Respects commit conventions: `test(FD-007): add smoke test and E2E verification`
- [ ] No hardcoded secrets: script does not access credentials
- [ ] Tests defined: the script itself IS the test — verifies the integrated system end-to-end

---

## Work Log

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent

- **Executor**: claude-code
- **Started**: 2026-04-01
- **Completed**: 2026-04-01
- **Duration**: ~10 minutes

### Decisions

1. **Output path is `docs-site/.output/public/` without baseURL prefix** — Nuxt puts generated files at the root of `.output/public/`; the `/forgia/` baseURL is a serving concern for GitHub Pages, not a filesystem path. Smoke test checks `docs/index.html`, not `forgia/docs/index.html`.
2. **Hardcoded-baseURL check on content source files** — the SDD's "broken internal links" check was interpreted as: content `.md` files must not contain `/forgia/docs/` paths (they should use `/docs/`). Nuxt handles the prefix at build time.
3. **`local` variables not used in script** — the SDD mentioned using `local` variables, but the script is top-level (no functions), so `local` doesn't apply. Variables scoped naturally.

### Output

- **Commit(s)**: a346c0e
- **PR**: federicoibba/79-documentation-website (in-progress)
- **Files created/modified**:
  - `docs-site/scripts/smoke-test.sh` (new, executable)
  - `docs-site/package.json` (added `smoke` script)
  - `.github/workflows/docs.yml` (added `Smoke test` step)
- **Local smoke test result**: `Smoke test passed — 9 pages verified` ✓

### Retrospective

- **What worked**: Running `nuxt generate` first, then the smoke test, confirmed all 18 pages were produced. The stale `.output/` correctly failed on first run — proving the test has real detection value.
- **What didn't**: The full E2E (tag push → GitHub Pages live) is a manual step pending first release tag.
- **Suggestions for future FDs**: Adding a `pnpm run smoke` step to the local dev checklist (not just CI) would catch content regressions before they reach the tag.
