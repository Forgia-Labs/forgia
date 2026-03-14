```
  ███████╗ ██████╗ ██████╗  ██████╗ ██╗ █████╗
  ██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██║██╔══██╗
  █████╗  ██║   ██║██████╔╝██║  ███╗██║███████║
  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██║██╔══██║
  ██║     ╚██████╔╝██║  ██║╚██████╔╝██║██║  ██║
  ╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝╚═╝  ╚═╝
```

> Forgia le spec in codice. Framework spec-driven per agenti AI.

Forgia e' un framework SDD (Spec-Driven Development) che trasforma decisioni di design in contratti di esecuzione per agenti autonomi.

## Il Modello

```
Tu ──→ FD (cosa e perche') ──→ SDD (come, per agenti) ──→ Agente ──→ Codice
        design funzionale         contratto di esecuzione     OpenHands
        1 per feature             N per FD                    Claude Code
```

| Layer | Chi | Cosa |
|-------|-----|------|
| **FD** (Feature Design) | Umano + AI | Architettura, trade-off, interfacce |
| **SDD** (Spec-Driven Dev) | AI genera, umano approva | Scope, vincoli, test, acceptance criteria |
| **Constitution** | Umano | Regole immutabili del progetto |
| **Codice** | Agente | Implementazione guidata dall'SDD |

## Quick Start

```bash
# Prerequisiti: mise, docker
git clone git@github.com:Deepzima/forgia.git
cd forgia

# Installa i comandi Claude Code
mise run claude:install

# Installa il runtime OpenHands
mise run openhands:install

# Inizializza un progetto
cd /path/to/your/project
forgia init
```

## Struttura Vault

Dopo `forgia init`, il tuo progetto ottiene:

```
tuo-progetto/
  .forgia/
    constitution.md         # regole immutabili
    _dashboard.md           # panoramica Obsidian dataview
    fd/                     # Feature Design
    sdd/                    # Execution Spec (agent-ready)
    ops/                    # Task operative
    dev-guide/              # Convenzioni
```

## Incantesimi

> I comandi di Forgia — slash commands per Claude Code.

| Incantesimo | Cosa fa |
|-------------|---------|
| `/project-init` | Prepara la fucina — scaffolda il vault `.forgia/` |
| `/fd-new` | Forgia un nuovo design — crea Feature Design |
| `/fd-review` | Prova del fuoco — review obbligatoria prima di procedere |
| `/fd-sdd` | Tempra le spec — genera N SDD dal FD approvato |
| `/fd-deep` | Analisi profonda — 4 agenti esplorano il problema in parallelo |
| `/fd-explore` | Studia il pezzo — carica contesto di un FD |
| `/fd-verify` | Collaudo — verifica implementazione vs spec |
| `/fd-close` | Sigilla l'opera — archivia FD completato, retrospettiva |
| `/fd-status` | Stato della fucina — dashboard FD + SDD |
| `/sdd-assign` | Assegna il lavoro — manda un SDD a un agente |
| `/sdd-status` | Stato degli agenti — progresso esecuzione SDD |

## Attrezzi della Fucina

> Task mise per gestire l'infrastruttura.

| Attrezzo | Cosa fa |
|----------|---------|
| `mise run init` | Scaffolda il vault nel progetto corrente |
| `mise run openhands:install` | Scarica il container OpenHands |
| `mise run openhands:up` | Accendi la fucina — avvia OpenHands su :3000 |
| `mise run openhands:down` | Spegni la fucina |
| `mise run sdd <file>` | Esegui un SDD con OpenHands headless |
| `mise run sdd:batch FD-001` | Esegui tutti gli SDD di un FD in parallelo |
| `mise run status` | Dashboard di tutti FD + SDD |
| `mise run doctor` | Controllo salute (docker, openhands, vault) |
| `mise run claude:install` | Installa i comandi in Claude Code |

## Diario di Lavoro (Work Log)

Ogni SDD include un Diario di Lavoro obbligatorio:

```markdown
## Work Log
### Agente
- chi ha eseguito, quando, durata
### Decisioni
- deviazioni dal piano, problemi incontrati
### Output
- commit hash, PR link, file creati/modificati
### Retrospettiva
- cosa ha funzionato, cosa no, suggerimenti per FD futuri
```

## Agenti

| Agente | Quando |
|--------|--------|
| **Claude Code** | Interattivo, sei alla tastiera |
| **OpenHands** | Autonomo, N agenti in container paralleli |
| **Manuale** | Leggi l'SDD e implementi tu |

## Ispirato da

- [GitHub Spec Kit](https://github.com/github/spec-kit) — separazione spec/plan, constitution
- [BMAD Method](https://github.com/bmad-code-org/BMAD-METHOD) — ruoli multi-agente
- [OpenHands](https://github.com/OpenHands/OpenHands) — runtime agente sandboxato

## Licenza

MIT
