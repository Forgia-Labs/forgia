---
id: "SDD-003"
fd: "FD-007"
title: "GitHub Actions Workflow — Build & Deploy to GitHub Pages"
status: planned
agent: ""
assigned_to: ""
created: "2026-04-01"
started: ""
completed: ""
tags: ["ci-cd", "github-actions", "github-pages"]
---

# SDD-003: GitHub Actions Workflow — Build & Deploy to GitHub Pages

> Parent FD: [[FD-007]]

## Scope

Creare il workflow GitHub Actions `.github/workflows/docs.yml` che:

1. Si attiva **solo al push di un tag di release** con pattern `v*.*.*`
2. Installa le dipendenze Node con `pnpm` (usando cache)
3. Esegue `pnpm run generate` nella directory `docs-site/`
4. Carica l'artifact con `actions/upload-pages-artifact`
5. Deploya su GitHub Pages con `actions/deploy-pages`

**Dipendenza**: SDD-001 deve essere completato — `pnpm run generate` deve esistere e funzionare.

Il workflow NON deve attivarsi su push a `main`, push a branch, o pull request — solo su tag `v*.*.*`.

### File da creare

```
.github/
└── workflows/
    └── docs.yml
```

### Struttura workflow attesa

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
          version: 10          # versione pnpm da pnpm-workspace.yaml o package.json
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

**Note implementative**:
- Verificare la versione esatta di `pnpm` da `docs-site/package.json` o `pnpm-workspace.yaml` prima di hardcodare
- `cache-dependency-path` deve puntare a `docs-site/pnpm-lock.yaml` (non alla root)
- Il job `deploy` richiede `environment: github-pages` per il deploy protetto
- `concurrency: cancel-in-progress: false` evita di cancellare un deploy in corso se arriva un secondo tag in rapida successione

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| Trigger | GitHub webhook `push.tags` | Si attiva solo su `v*.*.*` — NON su push a branch o main |
| `docs-site/.output/public/` | Directory | Output di `nuxt generate` (SDD-001) — artifact da caricare |
| `actions/upload-pages-artifact@v3` | GitHub Action | Impacchetta `.output/public/` per Pages |
| `actions/deploy-pages@v4` | GitHub Action | Deploya l'artifact su `forgia-labs.github.io/forgia/` |
| `permissions: pages: write, id-token: write` | OIDC | Obbligatori per deploy GitHub Pages senza PAT |

**Contratto con SDD-001**: il path `docs-site/.output/public/` deve esistere dopo `nuxt generate`. Se SDD-001 cambia il path di output, questo workflow va aggiornato di conseguenza.

## Constraints / Vincoli

- Language / Linguaggio: YAML (GitHub Actions syntax)
- Il trigger deve essere **esclusivamente** `on: push: tags: ['v*.*.*']` — nessun `push: branches`, nessun `pull_request`, nessun `workflow_dispatch` (a meno che non sia esplicitamente richiesto in futuro)
- `permissions` a livello di workflow (non di job): `contents: read`, `pages: write`, `id-token: write`
- `concurrency.cancel-in-progress: false` — non cancellare deploy in corso
- Node version: 22 LTS (compatibile con Nuxt 4)
- pnpm: usare `pnpm/action-setup@v4` con versione da lockfile, non installare globalmente con npm
- `--frozen-lockfile` obbligatorio in CI — fallisce se il lockfile non è aggiornato
- Il workflow non deve leggere secret (nessun `${{ secrets.XXX }}`) — GitHub Pages OIDC non richiede PAT
- Rispetta `deny.toml`: nessun `echo $GITHUB_TOKEN`, nessuna esposizione di credenziali nei log

## Best Practices

- Error handling: ogni step deve fallire esplicitamente se il precedente fallisce (comportamento default GitHub Actions — non usare `continue-on-error: true`)
- Naming: job names descrittivi (`build`, `deploy`), step names in inglese
- Style: YAML indentato con 2 spazi, nessuna tab
- Versioni action pinned: usare tag di versione major (es. `@v4`) — non `@latest`, non SHA per ora

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Manual | Push di tag `v0.0.1-docs-test` su branch di test — verifica che il workflow si attivi | Smoke test pre-merge |
| Negative | Push di commit a `main` senza tag — verifica che il workflow NON si attivi | Trigger isolation |
| CI | `pnpm install --frozen-lockfile` non fallisce (lockfile aggiornato) | Build job |
| CI | `pnpm run generate` produce `docs-site/.output/public/` | Build job |

## Acceptance Criteria / Criteri di Accettazione

- [ ] File `.github/workflows/docs.yml` creato e valido (nessun errore di syntax YAML)
- [ ] Il workflow si attiva SOLO su push di tag `v*.*.*` — verificato con `on:` block
- [ ] Il workflow NON include trigger su `push: branches` o `pull_request`
- [ ] `permissions: pages: write, id-token: write` presenti
- [ ] Job `build` usa pnpm con cache e `--frozen-lockfile`
- [ ] Job `deploy` usa `actions/deploy-pages@v4` con `environment: github-pages`
- [ ] Commit: `chore(FD-007): add GitHub Actions workflow for docs deployment`

## Context / Contesto

- [ ] `docs-site/package.json` — verificare script `generate` e versione pnpm
- [ ] `docs-site/pnpm-lock.yaml` — path necessario per `cache-dependency-path`
- [ ] `docs-site/pnpm-workspace.yaml` — verificare versione pnpm dichiarata
- [ ] `.forgia/guardrails/deny.toml` — nessun `echo $GITHUB_TOKEN` o pattern analoghi
- [ ] Docs GitHub Actions OIDC: https://docs.github.com/en/actions/security-for-github-actions/security-hardening-your-deployments/about-security-hardening-with-openid-connect
- [ ] Docs `actions/deploy-pages`: https://github.com/actions/deploy-pages

## Constitution Check

- [ ] Rispetta code standards: YAML valido, nessun secret hardcoded
- [ ] Rispetta commit conventions: `chore(FD-007): add GitHub Actions workflow for docs deployment`
- [ ] No hardcoded secrets: usa OIDC, nessun `secrets.XXX` necessario
- [ ] Tests definiti: smoke test manuale con tag di test

---

## Work Log / Diario di Lavoro

> Questa sezione è **obbligatoria**. Deve essere compilata dall'agent o dallo sviluppatore durante e dopo l'esecuzione.

### Agent / Agente

- **Executor**: <!-- openhands | claude-code | manual | name -->
- **Started**: <!-- timestamp -->
- **Completed**: <!-- timestamp -->
- **Duration / Durata**: <!-- total time -->

### Decisions / Decisioni

1. <!-- decisione 1: cosa e perché -->

### Output

- **Commit(s)**: <!-- hash -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `.github/workflows/docs.yml`

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**:
- **What didn't / Cosa non ha funzionato**:
- **Suggestions for future FDs / Suggerimenti per FD futuri**:
