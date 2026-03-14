---
id: "{{SDD_ID}}"
fd: "{{FD_ID}}"
title: "{{TITLE}}"
status: planned
agent: ""
assigned_to: ""
created: "{{DATE}}"
started: ""
completed: ""
tags: []
---

# {{SDD_ID}}: {{TITLE}}

> Parent FD: [[{{FD_ID}}]]

## Scope

<!-- Cosa costruire. Derivato dal FD. Preciso, non ambiguo. -->

## Interfacce

<!-- Input/output, tipi, trait, API endpoint, schema -->

| Interfaccia | Tipo | Descrizione |
|-------------|------|-------------|
| | | |

## Vincoli

<!-- Linguaggio, framework, versioni, pattern obbligatori, dipendenze -->

- Linguaggio:
- Framework:
- Dipendenze:
- Pattern:

## Best Practices

<!-- Error handling, naming, style specifici per questo componente -->

- Error handling:
- Naming:
- Style:

## Test Requirements

<!-- Cosa testare, coverage atteso, tipi di test -->

| Tipo | Cosa | Coverage |
|------|------|----------|
| Unit | | |
| Integration | | |
| E2E | | |

## Acceptance Criteria

<!-- Quando il lavoro e' "done". Ogni criterio deve essere verificabile. -->

- [ ] <!-- criterio 1 -->
- [ ] <!-- criterio 2 -->
- [ ] <!-- criterio 3 -->

## Contesto

<!-- File da leggere, doc da consultare, codice esistente da capire -->

- [ ] `path/to/file`
- [ ] `docs/reference`

## Constitution Check

<!-- Verifica che questo SDD rispetti la constitution del progetto -->

- [ ] Rispetta le code standards
- [ ] Rispetta le commit conventions
- [ ] Nessun secret hardcoded
- [ ] Test definiti e sufficienti

---

## Work Log

> Questa sezione e' **obbligatoria**. Deve essere compilata dall'agent o dallo sviluppatore durante e dopo l'esecuzione.

### Agent

- **Executor**: <!-- openhands | claude-code | manual | nome -->
- **Started**: <!-- timestamp -->
- **Completed**: <!-- timestamp -->
- **Duration**: <!-- tempo totale -->

### Decisioni Prese

<!-- Deviazioni dal plan, problemi incontrati, scelte fatte durante l'implementazione -->

1. <!-- decisione 1: cosa e perche -->

### Output

- **Commit(s)**: <!-- hash -->
- **PR**: <!-- link -->
- **File creati/modificati**:
  - `path/to/file`

### Retrospettiva

- **Cosa ha funzionato**:
- **Cosa non ha funzionato**:
- **Suggerimenti per FD futuri**:
