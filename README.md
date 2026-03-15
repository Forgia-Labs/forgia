```
  ███████╗ ██████╗ ██████╗  ██████╗ ██╗ █████╗
  ██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██║██╔══██╗
  █████╗  ██║   ██║██████╔╝██║  ███╗██║███████║
  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██║██╔══██║
  ██║     ╚██████╔╝██║  ██║╚██████╔╝██║██║  ██║
  ╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝╚═╝  ╚═╝

  Forge specs into code. FD → SDD → Agent execution framework.
```

> Spec-driven development framework for AI agents.
> Forgia turns human design decisions into execution contracts for autonomous agents.

## The Model

```
Tu ──→ FD (cosa e perche') ──→ SDD (come, per agent) ──→ Agent ──→ Codice
        design funzionale        contratto esecutivo       OpenHands
        1 per feature            N per FD                  Claude Code
        decisioni umane          prompt strutturato        parallelo
```

| Layer | Chi | Cosa |
|-------|-----|------|
| **FD** (Feature Design) | Umano + AI | Architettura, trade-off, interfacce |
| **SDD** (Spec-Driven Dev) | AI genera, umano approva | Scope, vincoli, test, criteri accettazione |
| **Constitution** | Umano | Regole immutabili del progetto |
| **Code** | Agent | Implementazione guidata dall'SDD |

## Quick Start

```bash
# Prerequisiti: mise, docker (opzionale per OpenHands)
git clone git@github.com:Deepzima/forgia.git
cd forgia

# Installa i comandi Claude Code
mise run claude:install

# Installa Beads task tracker (opzionale)
mise run bd:install

# Inizializza un nuovo progetto
cd /path/to/your/project
forgia init
```

## Struttura del Vault

Dopo `forgia init`, il progetto ottiene:

```
your-project/
  .forgia/
    config.toml               # configurazione runner e beads
    constitution.md            # regole immutabili
    _dashboard.md              # panoramica Obsidian dataview
    fd/                        # Feature Design (cosa e perche')
    sdd/                       # Execution Specs (come, per agent)
      _templates/
        sdd-template.md        # formato markdown
        sdd-template.yaml      # formato YAML (machine-parseable)
    ops/                       # task operativi manuali
    dev-guide/
      principles/              # clean-code, SOLID, design-patterns
      lang/                    # convenzioni per linguaggio (auto-detect)
      coding-conventions.md
      commit-conventions.md
      review-process.md
```

## Incantesimi (Spells)

> I comandi di Forgia — slash commands per Claude Code.

| Incantesimo | Cosa fa |
|-------------|---------|
| `/fd-new` | **Forgia un nuovo design** — crea un Feature Design |
| `/fd-review` | **Prova del fuoco** — review gate obbligatorio |
| `/fd-sdd` | **Tempra le spec** — genera N SDD da un FD approvato |
| `/fd-deep` | **Analisi profonda** — 4 agenti esplorano in parallelo |
| `/fd-explore` | **Studia il pezzo** — carica il contesto di un FD |
| `/fd-verify` | **Controllo qualita'** — verifica implementazione vs spec |
| `/fd-close` | **Sigilla il lavoro** — archivia FD completato + retrospettiva |
| `/fd-status` | **Stato della fucina** — dashboard FD + SDD |

## Attrezzi della Fucina (CLI)

> Comandi `forgia` e task `mise` per gestire il workflow.

| Attrezzo | Cosa fa |
|----------|---------|
| `forgia init` | Forgia il vault nel progetto corrente |
| `forgia status` | Dashboard FD + SDD + Beads |
| `forgia doctor` | Verifica salute (docker, bd, vault, mise) |
| `forgia validate <sdd>` | Valida un SDD prima dell'esecuzione |
| `forgia exec <sdd>` | Esegui un SDD (auto-detect runner) |
| `forgia batch <FD-NNN>` | Esegui tutti gli SDD di un FD |
| `forgia watch <FD-NNN>` | Osserva e esegui nuovi SDD automaticamente |

### Runner

```bash
# Claude Code — usa la tua subscription Max (gratis)
forgia exec .forgia/sdd/FD-001/SDD-001.md --runner=claude

# OpenHands — usa API key, container Docker isolato
forgia exec .forgia/sdd/FD-001/SDD-001.md --runner=openhands

# Default dal config.toml
forgia exec .forgia/sdd/FD-001/SDD-001.md
```

| Runner | Quando usarlo |
|--------|---------------|
| **Claude Code** | Interattivo, sub Max, worktree isolato |
| **OpenHands** | Autonomo, N container paralleli, API key |
| **Manual** | Leggi l'SDD e implementa tu |

## Beads Integration

[Beads (bd)](https://github.com/steveyegge/beads) gestisce il grafo delle dipendenze tra SDD:

```bash
# Dopo /fd-sdd: crea epic + subtask con dipendenze
bd ready              # mostra task sbloccati
forgia batch FD-001   # esegue solo quelli pronti (dependency-aware)
```

## Knowledge Stack

Ogni agent carica automaticamente:

```
.forgia/dev-guide/
  principles/          ← SEMPRE (clean-code, SOLID, design-patterns)
  lang/                ← auto-detect per progetto (rust, python, ts, go, shell)
  coding-conventions   ← regole operative
  commit-conventions   ← formato commit
```

## Work Log

Ogni SDD include un Work Log obbligatorio — compilato dall'agent o dallo sviluppatore:

```yaml
work_log:
  executor: claude-code
  started: 2026-03-15T10:00:00
  completed: 2026-03-15T11:30:00
  decisions:
    - what: "Used tower middleware instead of manual auth"
      why: "Composable, testable, idiomatic axum"
  output:
    commits: ["abc123"]
    files_changed: ["src/auth.rs", "tests/auth_test.rs"]
  retrospective:
    worked: "Builder pattern for config was clean"
    suggestions: "Add integration test template to SDD"
```

## Inspired By

- [AutoSpec](https://github.com/ariel-frischer/autospec) — YAML specs, auto-validation, session isolation
- [Beads](https://github.com/steveyegge/beads) — distributed graph issue tracker for AI agents
- [GitHub Spec Kit](https://github.com/github/spec-kit) — spec/plan separation, constitution
- [OpenHands](https://github.com/all-hands-ai/OpenHands) — sandboxed agent runtime

## License

MIT
