---
id: "SDD-003"
fd: "FD-001"
title: "forgia status + forgia doctor — read-only vault commands"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, cli, status, doctor, cobra]
---

# SDD-003: forgia status + forgia doctor — read-only vault commands

> Parent FD: [[FD-001]]

## Scope

Create two Cobra commands:

### `forgia status` (`cmd/forgia/cmd/status.go`)
Read-only dashboard showing:
- All FDs with status, priority, author (table format)
- All SDDs per FD with status (nested under FD)
- Board connection status (if configured in config.toml)
- Beads status (if available)
- Knowledge layer status (if codebase-memory-mcp configured)

### `forgia doctor` (`cmd/forgia/cmd/doctor.go`)
Health check showing:
- Vault: `.forgia/` exists and is valid (config.toml parseable, constitution exists)
- Git: `git` available, repo initialized
- Docker: `docker` available (optional)
- Claude: `claude` CLI available
- Beads: `bd` available, database connected
- fswatch: available (optional, for watch command)
- yq: available (optional, for YAML validation)
- MCP providers: healthy (if configured)
- Guardrails: deny.toml parseable

Each check outputs pass/fail/skip with clear messages.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `statusCmd` | `*cobra.Command` | `forgia status` |
| `doctorCmd` | `*cobra.Command` | `forgia doctor` |
| `vault.Open(dir)` | existing | Open vault for reading |
| `vault.ListFDs(ctx)` | existing | Get all FDs |
| `vault.ListSDDs(ctx, fdID)` | existing | Get SDDs for an FD |
| `config.LoadConfig(ctx, dir)` | existing | Load config for board/beads settings |
| `board.Resolve(ctx, provider, owner, number)` | existing | Check board connectivity |
| `beads.Client.Available()` | existing | Check beads |
| `exec.LookPath(name)` | stdlib | Check tool availability |

## Constraints / Vincoli

- Language / Linguaggio: Go 1.25+
- Framework: Cobra, `text/tabwriter` for aligned output
- Dependencies / Dipendenze: `internal/vault`, `internal/config`, `internal/board`, `internal/beads`
- Patterns / Pattern: both commands are read-only — NEVER modify vault files
- `status` must work even if config.toml is missing (just skip board/beads sections)
- `doctor` must not panic on missing tools — report and continue

## Best Practices

- Error handling: `status` returns error only if vault cannot be opened. Missing board/beads are warnings, not errors. `doctor` always returns exit 0 (it reports issues, doesn't fail on them).
- Naming: clear section headers in output
- Style: `slog.InfoContext` for structured output, `tabwriter` for tables

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | Status with empty vault shows "no FDs" | empty case |
| Unit | Status with FDs shows correct table | table format |
| Unit | Doctor detects git (always present in CI) | pass case |
| Unit | Doctor reports missing optional tools as skip | skip case |
| Unit | Status works without config.toml | graceful degradation |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `forgia status` lists all FDs with status in table format
- [ ] `forgia status` shows SDDs nested under each FD
- [ ] `forgia status` shows board status when configured (skip when not)
- [ ] `forgia status` works with empty vault (no FDs)
- [ ] `forgia doctor` checks: vault, git, docker, claude, bd, fswatch, yq
- [ ] `forgia doctor` reports pass/fail/skip per tool
- [ ] `forgia doctor` validates config.toml is parseable
- [ ] `forgia doctor` validates deny.toml is parseable
- [ ] Both commands are strictly read-only
- [ ] `go build ./...` succeeds
- [ ] 5+ tests pass

## Context / Contesto

- [ ] `bin/forgia` `cmd_status()` — Bash implementation
- [ ] `bin/forgia` `cmd_doctor()` — Bash implementation
- [ ] `internal/vault/file_vault.go` — ListFDs, ListSDDs
- [ ] `internal/config/config.go` — LoadConfig
- [ ] `cmd/forgia/cmd/sync.go` — reference for vault opening pattern

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
