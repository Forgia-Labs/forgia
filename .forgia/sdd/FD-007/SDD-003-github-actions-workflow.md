---
id: "SDD-003"
fd: "FD-007"
title: "GitHub Actions Workflow — Build & Deploy to GitHub Pages"
status: done
agent: "claude-code"
assigned_to: "claude"
created: "2026-04-01"
started: "2026-04-01"
completed: "2026-04-01"
tags: ["ci-cd", "github-actions", "github-pages"]
---

# SDD-003: GitHub Actions Workflow — Build & Deploy to GitHub Pages

> Parent FD: [[FD-007]]

## Scope

Create the GitHub Actions workflow `.github/workflows/docs.yml` that:

1. Triggers **only on push of a release tag** matching `v*.*.*`
2. Installs Node dependencies with `pnpm` (using cache)
3. Runs `pnpm run generate` in the `docs-site/` directory
4. Uploads the artifact with `actions/upload-pages-artifact`
5. Deploys to GitHub Pages with `actions/deploy-pages`

**Dependency**: SDD-001 must be completed — `pnpm run generate` must exist and work.

The workflow must NOT trigger on push to `main`, push to branches, or pull requests — only on `v*.*.*` tags.

### File to create

```
.github/
└── workflows/
    └── docs.yml
```

### Expected workflow structure

```yaml
name: Deploy Documentation

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 10          # verify pnpm version from pnpm-workspace.yaml or package.json
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
          cache-dependency-path: docs-site/pnpm-lock.yaml
      - name: Install dependencies
        run: pnpm install --frozen-lockfile
        working-directory: docs-site
      - name: Generate static site
        run: pnpm run generate
        working-directory: docs-site
      - name: Upload Pages artifact
        uses: actions/upload-pages-artifact@v3
        with:
          path: docs-site/.output/public

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

**Implementation notes**:
- Verify the exact `pnpm` version from `docs-site/package.json` or `pnpm-workspace.yaml` before hardcoding
- `cache-dependency-path` must point to `docs-site/pnpm-lock.yaml` (not the repo root)
- The `deploy` job requires `environment: github-pages` for protected deployment
- `concurrency: cancel-in-progress: false` prevents cancelling an in-progress deploy if a second tag is pushed in quick succession

## Interfaces

| Interface | Type | Description |
|-----------|------|-------------|
| Trigger | GitHub webhook `push.tags` | Fires only on `v*.*.*` — NOT on branch pushes or main |
| `docs-site/.output/public/` | Directory | Output of `nuxt generate` (SDD-001) — artifact to upload |
| `actions/upload-pages-artifact@v3` | GitHub Action | Packages `.output/public/` for Pages |
| `actions/deploy-pages@v4` | GitHub Action | Deploys the artifact to `forgia-labs.github.io/forgia/` |
| `permissions: pages: write, id-token: write` | OIDC | Required for GitHub Pages deploy without a PAT |

**Contract with SDD-001**: the path `docs-site/.output/public/` must exist after `nuxt generate`. If SDD-001 changes the output path, this workflow must be updated accordingly.

## Constraints

- Language: YAML (GitHub Actions syntax)
- The trigger must be **exclusively** `on: push: tags: ['v*.*.*']` — no `push: branches`, no `pull_request`, no `workflow_dispatch` (unless explicitly requested in future)
- `permissions` at workflow level (not job level): `contents: read`, `pages: write`, `id-token: write`
- `concurrency.cancel-in-progress: false` — do not cancel in-progress deploys
- Node version: 22 LTS (compatible with Nuxt 4)
- pnpm: use `pnpm/action-setup@v4` with version from lockfile, do not install globally via npm
- `--frozen-lockfile` required in CI — fails if lockfile is not up to date
- The workflow must not read secrets (no `${{ secrets.XXX }}`) — GitHub Pages OIDC does not require a PAT
- Respects `deny.toml`: no `echo $GITHUB_TOKEN`, no credential exposure in logs

## Best Practices

- Error handling: every step must fail explicitly if the previous one fails (default GitHub Actions behaviour — do not use `continue-on-error: true`)
- Naming: descriptive job names (`build`, `deploy`), step names in English
- Style: YAML indented with 2 spaces, no tabs
- Action versions pinned: use major version tags (e.g. `@v4`) — not `@latest`, not SHAs for now

## Test Requirements

| Type | What | Coverage |
|------|------|----------|
| Manual | Push tag `v0.0.1-docs-test` on a test branch — verify workflow triggers | Pre-merge smoke test |
| Negative | Push a commit to `main` without a tag — verify workflow does NOT trigger | Trigger isolation |
| CI | `pnpm install --frozen-lockfile` does not fail (lockfile up to date) | Build job |
| CI | `pnpm run generate` produces `docs-site/.output/public/` | Build job |

## Acceptance Criteria

- [ ] File `.github/workflows/docs.yml` created and valid (no YAML syntax errors)
- [ ] Workflow triggers ONLY on push of tag `v*.*.*` — verified with `on:` block
- [ ] Workflow does NOT include triggers for `push: branches` or `pull_request`
- [ ] `permissions: pages: write, id-token: write` present
- [ ] `build` job uses pnpm with cache and `--frozen-lockfile`
- [ ] `deploy` job uses `actions/deploy-pages@v4` with `environment: github-pages`
- [ ] Commit: `chore(FD-007): add GitHub Actions workflow for docs deployment`

## Context

- [ ] `docs-site/package.json` — verify `generate` script and pnpm version
- [ ] `docs-site/pnpm-lock.yaml` — path required for `cache-dependency-path`
- [ ] `docs-site/pnpm-workspace.yaml` — verify declared pnpm version
- [ ] `.forgia/guardrails/deny.toml` — no `echo $GITHUB_TOKEN` or similar patterns
- [ ] GitHub Actions OIDC docs: https://docs.github.com/en/actions/security-for-github-actions/security-hardening-your-deployments/about-security-hardening-with-openid-connect
- [ ] `actions/deploy-pages` docs: https://github.com/actions/deploy-pages

## Constitution Check

- [ ] Respects code standards: valid YAML, no hardcoded secrets
- [ ] Respects commit conventions: `chore(FD-007): add GitHub Actions workflow for docs deployment`
- [ ] No hardcoded secrets: uses OIDC, no `secrets.XXX` needed
- [ ] Tests defined: manual smoke test with test tag

---

## Work Log

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent

- **Executor**: claude-code
- **Started**: 2026-04-01
- **Completed**: 2026-04-01
- **Duration**: ~5 minutes

### Decisions

1. **pnpm version 10** — confirmed from `pnpm-lock.yaml` lockfileVersion `9.0` which corresponds to pnpm 10. The SDD suggested verifying before hardcoding; inspection of the lockfile confirmed the correct version.
2. **No `workflow_dispatch`** — kept trigger exclusively on `v*.*.*` tags per the SDD constraint. Adding `workflow_dispatch` would require a future FD change.
3. **Workflow matches SDD template exactly** — no deviations needed; the spec was precise and complete.

### Output

- **Commit(s)**: 71128c7
- **PR**: federicoibba/79-documentation-website (in-progress)
- **Files created/modified**:
  - `.github/workflows/docs.yml`

### Retrospective

- **What worked**: The SDD provided a near-complete YAML template — implementation was verification + copy with one confirmed value (pnpm version).
- **What didn't**: Nothing. This was a well-scoped SDD.
- **Suggestions for future FDs**: Workflow SDDs benefit from the "expected YAML structure" pattern used here — it removes ambiguity and speeds execution significantly.
