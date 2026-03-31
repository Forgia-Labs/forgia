---
id: "SDD-004"
fd: "FD-007"
title: "Integration Wiring & E2E Verification"
status: planned
agent: ""
assigned_to: ""
created: "2026-04-01"
started: ""
completed: ""
tags: ["integration", "e2e", "verification"]
---

# SDD-004: Integration Wiring & E2E Verification

> Parent FD: [[FD-007]]

## Scope

Verificare che tutti i componenti prodotti da SDD-001, SDD-002, SDD-003 funzionino correttamente insieme come sistema integrato. Questo SDD non produce nuovo codice applicativo — produce un **smoke test script** e completa le verifiche di integrazione che confermano il sistema end-to-end.

**Dipendenza forte**: SDD-001, SDD-002, SDD-003 devono essere completati prima.

### Flusso E2E da verificare

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
      → Mobile layout: sidebar collassa correttamente ✓
```

### Output di questo SDD

1. **Script di smoke test locale** (`docs-site/scripts/smoke-test.sh`) che:
   - Esegue `pnpm run generate`
   - Verifica che i file HTML critici esistano in `.output/public/`
   - Verifica che nessun file HTML contenga link interni rotti (pattern `/forgia/docs/` non trovato)
   - Exit code 0 se tutto OK, 1 con lista di errori altrimenti

2. **Aggiornamento `package.json`** con script `"smoke": "bash scripts/smoke-test.sh"`

3. **Step di smoke test nel workflow** (modifica minima a `docs.yml` — aggiungere uno step tra `generate` e `upload artifact` che esegue `pnpm run smoke`)

4. **Verifica finale manuale** documentata nel Work Log: push di tag `v0.1.0` e conferma che il sito sia live su GitHub Pages.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `docs-site/scripts/smoke-test.sh` | Bash script | Input: `.output/public/`; Output: exit 0 (OK) o exit 1 + lista errori |
| `package.json` `smoke` script | npm script | Chiama `smoke-test.sh`; usato localmente e in CI |
| `docs.yml` step `Smoke test` | GitHub Actions step | Aggiunto dopo `generate`, prima di `upload-pages-artifact` |
| Path critici da verificare | Filesystem | `/forgia/`, `/forgia/docs/`, `/forgia/docs/getting-started/`, `/forgia/docs/fd/`, `/forgia/docs/sdd/`, `/forgia/docs/cli/`, `/forgia/docs/constitution/` |

## Constraints / Vincoli

- Language / Linguaggio: Bash per lo script (`set -euo pipefail`)
- Lo script deve essere idempotente e non avere side effects oltre alla lettura di `.output/public/`
- Nessuna dipendenza esterna nello script (solo `grep`, `find`, `test` — tool Unix standard)
- Lo step nel workflow deve usare `working-directory: docs-site` e il medesimo environment del build job
- Non modificare la logica di deploy in SDD-003 — aggiungere solo uno step di verifica prima dell'upload
- Rispetta `deny.toml`: nessun `cat *.pem`, nessun accesso a credenziali
- Lo script deve fallire esplicitamente con messaggio chiaro se `.output/public/` non esiste (gen non è stato eseguito)

## Best Practices

- Error handling: `set -euo pipefail` nello script Bash; ogni check fallito stampa il path mancante prima di exit 1
- Naming: `smoke-test.sh` (kebab-case), variabili locali con `local`
- Style: script breve (< 50 righe), commenti dove non ovvio

### Checklist pagine da verificare nello script

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

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| E2E locale | `pnpm run smoke` dopo `pnpm run generate` — exit 0 | Tutti i path critici |
| E2E CI | Step `Smoke test` in `docs.yml` — workflow verde su push tag `v*.*.*` | Pipeline completa |
| Negative | Rimuovere un file content e verificare che smoke-test rilevi il path mancante | Smoke test reliability |
| Manual | Push tag `v0.1.0` — sito live e navigabile su GitHub Pages | Deploy end-to-end |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `docs-site/scripts/smoke-test.sh` esiste, è eseguibile (`chmod +x`), e passa con `set -euo pipefail`
- [ ] `pnpm run smoke` eseguito localmente dopo `pnpm run generate` — exit 0
- [ ] `docs.yml` include step `Smoke test` tra `generate` e `upload-pages-artifact`
- [ ] Il workflow completo (build → smoke → deploy) è verde su push tag `v*.*.*` in CI
- [ ] Il sito è accessibile su `forgia-labs.github.io/forgia/` dopo il primo deploy
- [ ] Navigazione sidebar mostra tutte e 5 le sezioni nel sito live
- [ ] Viewport mobile 375px: sidebar collassa correttamente (verificato su sito live o `pnpm dev`)
- [ ] `baseURL: /forgia/` funziona — nessun asset 404 nei DevTools del browser
- [ ] Commit: `test(FD-007): add smoke test and E2E verification`

## Context / Contesto

- [ ] `.forgia/sdd/FD-007/SDD-001-nuxt-content-setup.md` — output: struttura `content/` e path `.output/public/`
- [ ] `.forgia/sdd/FD-007/SDD-002-initial-content.md` — output: 17 file `.md` da verificare
- [ ] `.forgia/sdd/FD-007/SDD-003-github-actions-workflow.md` — workflow da modificare aggiungendo lo step smoke
- [ ] `docs-site/.output/public/` — directory prodotta da `nuxt generate` (esiste dopo SDD-001 completato)
- [ ] `.forgia/dev-guide/lang/shell.md` — convenzioni Bash per lo script

## Constitution Check

- [ ] Rispetta code standards: Bash con `set -euo pipefail`, variabili locali
- [ ] Rispetta commit conventions: `test(FD-007): add smoke test and E2E verification`
- [ ] No hardcoded secrets: lo script non accede a credenziali
- [ ] Tests definiti: lo script stesso È il test — verifica il sistema integrato end-to-end

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
  - `docs-site/scripts/smoke-test.sh`
  - `docs-site/package.json` (aggiunto script `smoke`)
  - `.github/workflows/docs.yml` (aggiunto step smoke test)

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**:
- **What didn't / Cosa non ha funzionato**:
- **Suggestions for future FDs / Suggerimenti per FD futuri**:
