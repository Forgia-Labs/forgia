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

Scrivere il contenuto Markdown iniziale del sito di documentazione, strutturato nelle 5 sezioni del sito. Il contenuto usa come fonte di riferimento i file esistenti in `docs/` e `CLAUDE.md`, ma **non** viene copiato direttamente — va riscritto/adattato per il formato web (pagine corte, front matter, link interni).

**Dipendenza**: SDD-001 deve essere completato prima — la struttura `content/` e il layout docs devono esistere.

### Sezioni da produrre

```
content/
├── index.md                          ← redirect a /docs/getting-started
├── getting-started/
│   ├── index.md                      ← panoramica + prerequisiti
│   ├── installation.md               ← install forgia CLI, mise, Claude Code
│   └── first-feature.md              ← flusso completo: fd-new → fd-review → fd-sdd → exec → verify → close
├── fd/
│   ├── index.md                      ← cos'è un FD, quando usarlo
│   ├── creating.md                   ← /fd-new, da issue o free-text
│   ├── reviewing.md                  ← /fd-review, cosa controlla
│   └── closing.md                    ← /fd-close, /fd-verify
├── sdd/
│   ├── index.md                      ← cos'è un SDD, struttura
│   ├── generating.md                 ← /fd-sdd
│   └── executing.md                  ← forgia exec, forgia batch, /sdd-assign
├── cli/
│   ├── index.md                      ← panoramica comandi
│   ├── init.md                       ← forgia init
│   ├── status.md                     ← forgia status
│   ├── doctor.md                     ← forgia doctor
│   └── exec.md                       ← forgia exec / batch / watch
└── constitution/
    └── index.md                      ← regole, commit conventions, code standards
```

### Front matter obbligatorio per ogni file

```yaml
---
title: "Titolo pagina"
description: "Breve descrizione (usata per SEO e card social)"
navigation:
  title: "Titolo sidebar"  # se diverso da title
---
```

### Tono e lingua

- Sezioni narrative: italiano (coerente con le convention del progetto)
- Blocchi di codice, nomi di comandi, variabili: inglese
- Stile: conciso, tecnico, orientato all'azione — no marketing language

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `content/**/*.md` | Markdown + YAML front matter | Contratto con SDD-001: ogni file deve avere `title`, `description`, `navigation.title` nel front matter |
| Navigazione sidebar | Front matter `navigation` | SDD-001 usa `queryCollection()` per costruire il menu — i file devono rispettare la struttura di directory |
| Link interni | Markdown `[text](/docs/section/page)` | Tutti i link interni usano path assoluti con `/docs/` prefix (coerente con `baseURL: /forgia/`) |

**Contratto con SDD-001**: le cartelle scritte qui devono corrispondere esattamente alla struttura creata da SDD-001. Non creare cartelle aggiuntive senza coordinamento.

**Contratto con SDD-004**: ogni pagina prodotta qui deve essere raggiungibile via `nuxt generate` — i link interni rotti bloccano il build.

## Constraints / Vincoli

- Language / Linguaggio: Markdown (CommonMark) con YAML front matter
- Framework: Nuxt Content v3 — usare MDC syntax se necessario per componenti Vue inline
- Ogni file `.md` deve avere front matter con almeno: `title`, `description`
- Non copiare blocchi di testo verbatim da `docs/functional-architecture.md` — i diagrammi Mermaid complessi vanno semplificati o rimossi per il sito (troppo verbosi per un utente che scopre Forgia)
- Nessun link esterno non verificato
- Nessun file `.env`, secret o credenziale nei contenuti
- Il contenuto di `constitution/index.md` è una sintesi — NON una copia di `.forgia/constitution.md` (quello è la fonte autoritativa del progetto, non del sito)

## Best Practices

- Error handling: ogni pagina deve essere auto-contenuta — un utente che arriva direttamente alla pagina deve capire il contesto senza aver letto le precedenti
- Naming: file in kebab-case, cartelle in kebab-case
- Style: sezioni corte (max ~300 parole per pagina), molti esempi di codice, link interni frequenti

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Build | `pnpm run generate` con i file di contenuto prodotti — nessun link interno 404 | Tutti i file `.md` |
| Manual | Ogni pagina è raggiungibile dalla sidebar navigazione | Verifica smoke test locale |
| Manual | Front matter `title` e `description` presenti su ogni pagina | 100% dei file |

## Acceptance Criteria / Criteri di Accettazione

- [ ] Tutte le 5 sezioni presenti con almeno 1 pagina di contenuto ciascuna (non solo `.gitkeep`)
- [ ] Ogni file `.md` ha front matter con `title` e `description`
- [ ] `pnpm run generate` completa senza errori con questo contenuto (nessun link interno rotto)
- [ ] La sidebar del layout docs mostra la navigazione strutturata nelle 5 sezioni
- [ ] Getting Started / `first-feature.md` copre l'intero flusso: `fd-new → fd-review → fd-sdd → exec → verify → close`
- [ ] CLI Reference documenta almeno: `init`, `status`, `doctor`, `exec`
- [ ] Commit: `docs(FD-007): add initial documentation content`

## Context / Contesto

- [ ] `docs/getting-started.md` — guida all'installazione e primo utilizzo (fonte principale per getting-started/)
- [ ] `docs/concepts.md` — concetti FD/SDD (fonte per fd/ e sdd/)
- [ ] `docs/functional-architecture.md` — architettura (riferimento per CLI e concetti avanzati)
- [ ] `docs/go-architecture.md` — architettura Go (riferimento per CLI Reference)
- [ ] `CLAUDE.md` — regole progetto (fonte per constitution/)
- [ ] `.forgia/constitution.md` — regole immutabili (fonte per constitution/index.md)
- [ ] `docs-site/app/layouts/docs.vue` — layout già creato da SDD-001: capire come viene costruita la sidebar per scrivere il front matter correttamente

## Constitution Check

- [ ] Rispetta code standards: Markdown valido, front matter YAML ben formato
- [ ] Rispetta commit conventions: `docs(FD-007): add initial documentation content`
- [ ] No hardcoded secrets: nessun token, nessuna chiave API nei contenuti
- [ ] Tests definiti: build test (generate) + smoke test manuale

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

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**:
- **What didn't / Cosa non ha funzionato**:
- **Suggestions for future FDs / Suggerimenti per FD futuri**:
