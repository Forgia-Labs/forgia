---
id: "SDD-007"
fd: "FD-001"
title: "forgia watch — fsnotify watcher with auto-exec"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, cli, watch, fsnotify, cobra]
---

# SDD-007: forgia watch — fsnotify watcher with auto-exec

> Parent FD: [[FD-001]]

## Scope

Create `cmd/forgia/cmd/watch.go` — Cobra command that watches `.forgia/sdd/FD-NNN/` for new or modified SDD files and auto-executes them.

Behavior:
1. Open vault, load config
2. Set up `fsnotify` watcher on `.forgia/sdd/<FD-NNN>/` directory
3. On file create/modify events: debounce for N seconds (configurable, default 5s)
4. After debounce: check if the SDD has `status: planned` — if yes, execute via `runSDD()` (shared logic from SDD-006)
5. After execution: update SDD status to done/failed
6. Continue watching for more events
7. Ctrl+C (SIGINT/SIGTERM) gracefully stops the watcher

Flags: `--runner`, `--debounce=5s`, `--mode`

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `watchCmd` | `*cobra.Command` | `forgia watch <FD-NNN> [flags]` |
| `--debounce` | `time.Duration` | Debounce period (default 5s) |
| `fsnotify.Watcher` | external | File system watcher |
| `runSDD()` | from SDD-006 | Shared exec logic |
| `process.ShutdownManager` | existing | Graceful shutdown |

## Constraints / Vincoli

- Language / Linguaggio: Go 1.25+
- Framework: Cobra, `github.com/fsnotify/fsnotify`
- Dependencies / Dipendenze: `internal/vault`, `internal/config`, `internal/runner` (SDD-005), `internal/process`, fsnotify
- Patterns / Pattern:
  - Debounce with `time.AfterFunc` reset on each event
  - Context cancellation for graceful shutdown
  - Only trigger on `.md` and `.yaml` files (ignore temp files, .swp, etc.)
  - Filter: only process SDDs with `status: planned`

## Best Practices

- Error handling: log watcher errors but continue watching (don't exit on transient errors)
- Naming: `watchCmd`, `debounceExec`
- Style: clear terminal output showing "Watching...", "Found SDD...", "Executing..."

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | Debounce logic: multiple events within window produce single exec | timer reset |
| Unit | File filter: .md triggers, .swp ignored | filter logic |
| Unit | Graceful shutdown on context cancel | shutdown path |
| Unit | Only planned SDDs trigger execution | status filter |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `forgia watch FD-001` starts watching the SDD directory
- [ ] New SDD file with `status: planned` triggers execution after debounce
- [ ] Modified SDD with `status: planned` triggers execution
- [ ] SDD with `status: done` is ignored
- [ ] `--debounce` flag configures wait time
- [ ] Ctrl+C stops gracefully (no zombie processes)
- [ ] Temporary files (.swp, ~, .tmp) are ignored
- [ ] `go build ./...` succeeds
- [ ] 4+ tests pass

## Context / Contesto

- [ ] `bin/forgia` `cmd_watch()` — Bash watch implementation
- [ ] `internal/process/signals.go` — ShutdownManager
- [ ] `internal/process/spawn.go` — Spawn with context cancellation
- [ ] `cmd/forgia/cmd/exec.go` — runSDD shared logic (from SDD-006)
- [ ] fsnotify docs: https://pkg.go.dev/github.com/fsnotify/fsnotify

## Constitution Check

- [x] Respects code standards
- [x] Respects commit conventions
- [x] No hardcoded secrets
- [x] Tests defined and sufficient

---

## Work Log / Diario di Lavoro

### Agent / Agente

- **Executor**: <!-- openhands | claude-code | manual | name -->
- **Started**: <!-- timestamp -->
- **Completed**: <!-- timestamp -->
- **Duration / Durata**: <!-- total time -->

### Decisions / Decisioni

1. <!-- decision 1: what and why -->

### Output

- **Commit(s)**: <!-- hash -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `path/to/file`

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**:
- **What didn't / Cosa non ha funzionato**:
- **Suggestions for future FDs / Suggerimenti per FD futuri**:
