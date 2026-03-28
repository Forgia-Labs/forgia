---
id: "SDD-001"
fd: "FD-004"
title: "status enrichment — OPS, logs, Beads, knowledge stats"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-28"
started: "2026-03-28"
completed: "2026-03-28"
tags: ["enhancement"]
---

# SDD-001: status enrichment — OPS, logs, Beads, knowledge stats

> Parent FD: [[FD-004]]

## Scope

Enrich `cmd/forgia/cmd/status.go` to display the full dashboard matching bash `forgia status`. Currently the Go version only shows FDs with nested SDDs. Add four new sections:

1. **OPS tasks** — read `.forgia/ops/active/OPS-*.md`, parse frontmatter (id, title, priority), display as a table
2. **Last executions** — read `.forgia/logs/exec-*.json`, parse JSON (sdd, duration_seconds, exit_code, status), display the most recent N entries
3. **Beads ready tasks** — call `bd ready` via `beads.Client`, display output if available
4. **Knowledge layer stats** — call `codebase-memory-mcp cli list_projects`, extract node_count, edge_count, last_indexed, display with relative time

### Dependency

SDD-003 (JSON execution reports) should ideally complete first so that `logs/exec-*.json` files exist for testing. However, this SDD can be implemented independently — the logs section simply shows "No executions found" when no JSON files exist.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| OPS file reader | filesystem | Read `.forgia/ops/active/OPS-*.md`, parse YAML frontmatter for id, title, priority |
| Exec log reader | filesystem | Read `.forgia/logs/exec-*.json`, parse JSON for sdd, duration, status, timestamp |
| Beads integration | subprocess | Call `bd ready` via `beads.Client.Call()`, display raw output |
| Knowledge stats | subprocess | Call `codebase-memory-mcp cli list_projects`, parse JSON for node_count, edge_count, last_indexed |
| Dashboard output | stdout | Formatted tabwriter output with section headers |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: Cobra CLI, text/tabwriter for formatting
- Dependencies / Dipendenze: `internal/beads` (existing), `internal/vault` (existing)
- Patterns / Pattern: Graceful degradation — Beads and knowledge sections skipped silently when tools unavailable

### Guardrails (from deny.toml)

- Do not read files matching `[read]` deny patterns
- OPS files and exec logs in `.forgia/` are NOT denied — safe to read

## Best Practices

- Error handling: Never fail the entire status command because one section errors — log warning via `slog`, skip section, continue
- Naming: Follow existing `status.go` patterns — keep functions prefixed with `print*` (e.g., `printOPSTasks`, `printExecLogs`)
- Style: Use `text/tabwriter` for aligned output, match existing FD/SDD table formatting
- Relative time: Convert `last_indexed` ISO timestamp to relative ("2h ago") — use `time.Since()` with formatted output

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `parseOPSFile()` — parse frontmatter from OPS markdown | Frontmatter extraction |
| Unit | `parseExecLog()` — parse JSON execution report | JSON parsing |
| Unit | Relative time formatting | Time formatting |
| Integration | `forgia status` with vault containing OPS + logs | Full dashboard |
| Integration | `forgia status` with no OPS, no logs, no Beads, no knowledge | Graceful degradation |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `forgia status` displays OPS tasks section with id, title, priority from `.forgia/ops/active/`
- [ ] `forgia status` displays last execution summaries from `.forgia/logs/exec-*.json`
- [ ] `forgia status` displays Beads ready tasks when `bd` is available
- [ ] `forgia status` displays knowledge layer stats when codebase-memory-mcp is available
- [ ] Each section degrades gracefully — missing tools or empty directories produce no errors
- [ ] Existing FD/SDD dashboard sections unchanged
- [ ] All new functions have unit tests with `t.Parallel()`

## Context / Contesto

- [ ] `cmd/forgia/cmd/status.go` — existing status command to extend
- [ ] `cmd/forgia/cmd/status_test.go` — existing tests
- [ ] `internal/beads/beads.go` — `Client.Call()` for Beads integration
- [ ] `internal/vault/file_vault.go` — vault file reading patterns
- [ ] `bin/forgia` lines 591-713 — bash `cmd_status()` as reference for output format
- [ ] `.forgia/ops/active/` — OPS task directory structure
- [ ] `.forgia/logs/exec-*.json` — execution log format (created by SDD-003)

## Constitution Check

- [ ] Respects code standards — Go conventions, error wrapping, `t.Parallel()`
- [ ] Respects commit conventions — `feat(FD-004): description`
- [ ] No hardcoded secrets — no secret handling
- [ ] Tests defined and sufficient — unit + integration tests

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-28T22:52:00+01:00
- **Completed**: 2026-03-28T22:56:00+01:00
- **Duration / Durata**: ~4 min

### Decisions / Decisioni

1. Kept all new sections in `status.go` rather than splitting into separate files — the file is small (~270 lines) and cohesive around the dashboard concept.
2. Used `gopkg.in/yaml.v3` (already in go.mod) for OPS frontmatter parsing, matching vault's existing approach.
3. Capped exec logs display to 10 most recent entries, sorted by `started` timestamp descending, to avoid overwhelming output.
4. Knowledge layer calls `codebase-memory-mcp cli list_projects` directly (subprocess with 5s timeout) rather than going through the MCP subsystem — matches the bash implementation and avoids coupling to the MCP provider chain.
5. Each section prints nothing (not even a header) when data is absent, matching the bash behavior where empty sections are fully omitted.
6. Tests for Beads and Knowledge graceful degradation run against the real system — if tools are unavailable, the functions return silently, which is the correct behavior.

### Output

- **Commit(s)**: <!-- hash — will be filled after commit -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `cmd/forgia/cmd/status.go` — enriched with 4 new sections (OPS, exec logs, Beads, knowledge)
  - `cmd/forgia/cmd/status_enrichment_test.go` — 14 new tests covering parsing, formatting, and graceful degradation

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: The bash reference implementation provided a clear target. Existing `beads.Client` and `vault.Vault` patterns made integration straightforward. All tests passed on first run.
- **What didn't / Cosa non ha funzionato**: Nothing significant — SDD-003 (JSON reports) hasn't run yet so exec log files have empty `sdd` fields, but the code handles this gracefully with `(unknown)` label.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider adding an `io.Writer` parameter to `printDashboard` and section functions to enable output capture in integration tests without stdout redirection.
