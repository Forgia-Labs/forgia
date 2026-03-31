---
id: "SDD-001"
fd: "FD-007"
title: "Nuxt Content Setup & Docs Layout"
status: planned
agent: ""
assigned_to: ""
created: "2026-04-01"
started: ""
completed: ""
tags: ["nuxt", "docs", "frontend"]
---

# SDD-001: Nuxt Content Setup & Docs Layout

> Parent FD: [[FD-007]]

## Scope

Lo scaffold `docs-site/` esiste già (Nuxt 4 + Nuxt UI v4, untracked). Questo SDD ha il compito di:

1. Aggiungere `@nuxt/content` v3 alle dipendenze
2. Configurare `nuxt.config.ts` per: modulo Content, `baseURL: '/forgia/'`, prerender di tutte le pagine content
3. Aggiungere il script `"generate": "nuxt generate"` a `package.json`
4. Creare il layout `docs` con: sidebar navigazione, table of contents, area contenuto principale
5. Creare la pagina catch-all `app/pages/[...slug].vue` che serve il contenuto Nuxt Content
6. Creare la struttura directory `content/` con index e cartelle placeholder
7. Committare tutto su `main` come primo commit tracciato di `docs-site/`

**Non** toccare la homepage (`app/pages/index.vue`) né i componenti esistenti (`HeroSection`, `FeaturesSection`, `CtaSection`).

### Stato attuale del repo

```
docs-site/
├── app/
│   ├── app.vue              ✅ header/footer UHeader/UFooter
│   ├── app.config.ts        ✅ colori custom: forge, anvil
│   ├── assets/css/main.css  ✅
│   ├── assets/images/       ✅
│   ├── pages/index.vue      ✅ landing page
│   └── components/          ✅ HeroSection, FeaturesSection, CtaSection
├── nuxt.config.ts           ✅ (manca: @nuxt/content, baseURL, prerender)
├── package.json             ✅ (mancano: @nuxt/content, script generate)
├── pnpm-lock.yaml           ✅
├── pnpm-workspace.yaml      ✅
├── eslint.config.mjs        ✅
├── tsconfig.json            ✅
└── public/                  ✅
```

### Output atteso

```
docs-site/
├── app/
│   ├── layouts/
│   │   └── docs.vue         ← NUOVO: sidebar + TOC + content area
│   └── pages/
│       └── docs/
│           └── [...slug].vue ← NUOVO: catch-all per Nuxt Content
├── content/
│   ├── index.md             ← NUOVO: redirect o intro breve
│   ├── getting-started/
│   │   └── .gitkeep         ← placeholder (contenuto in SDD-002)
│   ├── fd/
│   │   └── .gitkeep
│   ├── sdd/
│   │   └── .gitkeep
│   ├── cli/
│   │   └── .gitkeep
│   └── constitution/
│       └── .gitkeep
├── nuxt.config.ts           ← MODIFICATO
└── package.json             ← MODIFICATO
```

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `nuxt.config.ts` | Config | Esporta config con `@nuxt/content`, `baseURL: '/forgia/'`, prerender `/**` |
| `app/layouts/docs.vue` | Vue SFC | Layout con slot `default` per il contenuto, sidebar sinistra, TOC destra |
| `app/pages/docs/[...slug].vue` | Vue SFC | Pagina catch-all che usa `<ContentDoc />` o `queryCollection()` |
| `content/` directory | Filesystem | Struttura cartelle per SDD-002 — contratto: `getting-started/`, `fd/`, `sdd/`, `cli/`, `constitution/` |
| `package.json` | JSON | Script `generate: nuxt generate`, dipendenza `@nuxt/content` v3 |

**Contratto con SDD-002**: la struttura `content/` creata qui deve corrispondere esattamente alle cartelle in cui SDD-002 scriverà i file Markdown.

**Contratto con SDD-003**: `nuxt generate` (aggiunto qui) deve completare con exit code 0 affinché il workflow CI possa deployare.

## Constraints / Vincoli

- Language / Linguaggio: TypeScript, Vue 3 Composition API con `<script setup>`
- Framework: Nuxt 4, Nuxt UI v4 (`@nuxt/ui: ^4.6.0`), Nuxt Content v3 (`@nuxt/content: ^3.x`)
- Package manager: `pnpm` (non usare npm o yarn)
- `baseURL` obbligatorio: `/forgia/` — senza trailing slash nelle route, con trailing slash nel baseURL
- Prerender: configurare `nitro.prerender.routes` o `routeRules` per prerenderizzare `/docs/**`
- Il layout `docs.vue` deve usare **solo componenti Nuxt UI v4** per sidebar e TOC (es. `UNavigationMenu`, `UContentToc` o equivalenti disponibili)
- Strict TypeScript: nessun `any` esplicito nelle nuove pagine/layout
- Non modificare `app/pages/index.vue` né i componenti landing page
- Rispettare il pattern ESLint esistente (commaDangle: never, braceStyle: 1tbs)
- I file in `deny.toml` non devono essere toccati: nessun `.env`, nessun secret, nessun file di configurazione Forgia

## Best Practices

- Error handling: se `queryCollection()` restituisce null, mostrare una pagina 404 esplicita (non silently fail)
- Naming: componenti Vue in PascalCase, file in kebab-case
- Style: Composition API `<script setup>` su tutti i nuovi componenti/pagine; no Options API
- Il layout docs deve essere mobile-responsive (sidebar collassabile su mobile usando `USlideover` o pattern equivalente)
- Non aggiungere CSS custom: usare solo Tailwind utility classes e componenti Nuxt UI

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Build | `pnpm run generate` completa senza errori | 100% — build deve passare |
| Build | Asset path: nessun 404 su `/forgia/` come baseURL | Verificare output `.output/public/` |
| Manual | Layout docs visualizzato correttamente su `/docs/` con sidebar | Smoke test locale `pnpm dev` |
| TypeScript | `pnpm run typecheck` senza errori | 100% |
| Lint | `pnpm run lint` senza errori | 100% |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `@nuxt/content` v3 aggiunto a `package.json` e installato via `pnpm install`
- [ ] `nuxt.config.ts` include `@nuxt/content` in `modules` e `baseURL: '/forgia/'`
- [ ] Script `"generate": "nuxt generate"` presente in `package.json`
- [ ] `pnpm run generate` completa senza errori e produce `.output/public/`
- [ ] `app/layouts/docs.vue` esiste con sidebar navigazione e area contenuto principale
- [ ] `app/pages/docs/[...slug].vue` esiste e serve contenuto Nuxt Content
- [ ] Struttura `content/` creata con le 5 cartelle: `getting-started/`, `fd/`, `sdd/`, `cli/`, `constitution/`
- [ ] `pnpm run typecheck` e `pnpm run lint` passano senza errori
- [ ] Tutto committato su `main` con messaggio `feat(FD-007): add Nuxt Content and docs layout`

## Context / Contesto

- [ ] `docs-site/nuxt.config.ts` — config attuale (no Content, no baseURL)
- [ ] `docs-site/package.json` — dipendenze attuali (no @nuxt/content)
- [ ] `docs-site/app/app.vue` — layout root con UHeader/UFooter; il layout docs si inserisce tra questi
- [ ] `docs-site/app/app.config.ts` — colori custom `forge` e `anvil`
- [ ] `docs/getting-started.md` — contenuto di riferimento (non copiare, usare come guida struttura)
- [ ] Docs Nuxt Content v3: https://content.nuxt.com/
- [ ] Docs Nuxt UI v4: https://ui.nuxt.com/

## Constitution Check

- [ ] Rispetta code standards: TypeScript strict, no any, Composition API
- [ ] Rispetta commit conventions: `feat(FD-007): add Nuxt Content and docs layout`
- [ ] No hardcoded secrets: nessun token, nessuna API key
- [ ] Tests definiti: build test + typecheck + lint

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
  - `docs-site/nuxt.config.ts`
  - `docs-site/package.json`
  - `docs-site/app/layouts/docs.vue`
  - `docs-site/app/pages/docs/[...slug].vue`
  - `docs-site/content/index.md`
  - `docs-site/content/getting-started/.gitkeep`
  - `docs-site/content/fd/.gitkeep`
  - `docs-site/content/sdd/.gitkeep`
  - `docs-site/content/cli/.gitkeep`
  - `docs-site/content/constitution/.gitkeep`

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**:
- **What didn't / Cosa non ha funzionato**:
- **Suggestions for future FDs / Suggerimenti per FD futuri**:
