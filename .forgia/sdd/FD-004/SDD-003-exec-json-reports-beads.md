---
id: "SDD-003"
fd: "FD-004"
title: "JSON execution reports + Beads task closure"
status: done
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: ["enhancement"]
---

# SDD-003: JSON execution reports + Beads task closure

> Parent FD: [[FD-004]]

## Scope

Ensure `forgia exec` (and by extension `forgia batch`) produces structured JSON execution reports and closes matching Beads tasks after successful SDD completion.

### Part A: JSON Execution Reports

After each SDD execution, write a JSON report to `.forgia/logs/exec-{SDD_ID}-{ISO8601}.json` with this structure:

```json
{
  "sdd": "SDD-001",
  "fd": "FD-004",
  "file": ".forgia/sdd/FD-004/SDD-001-status-enrichment.md",
  "runner": "claude",
  "started": "2026-03-28T15:30:00Z",
  "completed": "2026-03-28T15:35:00Z",
  "duration_seconds": 300,
  "exit_code": 0,
  "status": "success"
}
```

This may already be partially implemented in `internal/runner/claude.go` (check `ExecResult` → JSON marshaling). Verify the JSON format matches what bash produces and what `status` (SDD-001) will consume.

### Part B: Beads Task Closure

After successful SDD execution (exit code 0), search for a matching Beads task and close it:

1. Call `bd search {SDD_ID}` via `beads.Client.Call()`
2. If a matching task is found, call `bd update --status=done {task_id}`
3. If Beads is unavailable (`!client.Available()`), skip silently

This must be wired into `cmd/forgia/cmd/exec.go` after the runner completes successfully.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `ExecResult` → JSON | serialization | Marshal `runner.ExecResult` to JSON, write to `.forgia/logs/` |
| Log file writer | filesystem | Create `.forgia/logs/` if missing, write `exec-{SDD}-{timestamp}.json` |
| Beads search | subprocess | `bd search {SDD_ID}` → parse task ID from output |
| Beads update | subprocess | `bd update --status=done {task_id}` |
| SDD-001 contract | JSON schema | Status command reads `logs/exec-*.json` — JSON format must match |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: `encoding/json`, `internal/beads`
- Dependencies / Dipendenze: `internal/runner` (ExecResult), `internal/beads` (Client)
- Patterns / Pattern: Beads is optional — use `client.Available()` guard

### Guardrails (from deny.toml)

- `.forgia/logs/` is NOT in the deny write list — safe to write
- Do not write to `.forgia/constitution.md`, `.forgia/config.toml`, `.forgia/guardrails/deny.toml`

## Best Practices

- Error handling: JSON write failure should warn but NOT fail the exec command — the SDD execution itself succeeded
- Error handling: Beads failure should warn but NOT fail — Beads is optional
- Naming: `writeExecReport(result *runner.ExecResult, logsDir string) error`
- Style: Use `json.MarshalIndent` for human-readable JSON reports

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `writeExecReport()` — verify JSON structure and file naming | Report generation |
| Unit | Beads task closure with mock client (available, unavailable, search miss) | Beads integration |
| Integration | `forgia exec` produces JSON file in `logs/` | End-to-end report |
| Integration | JSON report parseable by status command (SDD-001) | Cross-SDD contract |

## Acceptance Criteria / Criteri di Accettazione

- [x] `forgia exec` writes JSON report to `.forgia/logs/exec-{SDD}-{timestamp}.json`
- [x] JSON contains: sdd, fd, file, runner, started, completed, duration_seconds, exit_code, status
- [x] `forgia batch` writes one JSON report per SDD executed
- [x] Beads task closed (`bd update --status=done`) on successful execution
- [x] Beads unavailable → skipped silently (no error, logged as info)
- [x] JSON write failure → warning logged, exec command still returns success
- [x] `.forgia/logs/` directory created automatically if missing
- [x] All new functions have unit tests with `t.Parallel()`

## Context / Contesto

- [x] `cmd/forgia/cmd/exec.go` — exec command to modify (add report writing + Beads closure)
- [x] `cmd/forgia/cmd/batch.go` — batch command (delegates to exec, verify reports generated per SDD)
- [x] `internal/runner/runner.go` — `ExecResult` struct definition
- [x] `internal/runner/claude.go` — existing runner, check if it already writes logs
- [x] `internal/beads/beads.go` — `Client.Call()`, `Client.Available()`
- [x] `bin/forgia` lines 963-1033 — bash `cmd_exec()` as reference for report format and Beads closure

## Constitution Check

- [x] Respects code standards — Go conventions, error wrapping, graceful degradation
- [x] Respects commit conventions — `feat(FD-004): description`
- [x] No hardcoded secrets — no secret handling
- [x] Tests defined and sufficient — unit + integration

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-28T23:00:00Z
- **Completed**: 2026-03-28T23:10:00Z
- **Duration / Durata**: ~10 min

### Decisions / Decisioni

1. Moved JSON report writing from `internal/runner/claude.go` to a new `cmd/forgia/cmd/report.go` — this way all runners (claude, dry-run, future openhands) get reports, and the runner layer stays focused on execution.
2. Created `writeExecReport()` as a standalone function callable from both `exec` and `batch` commands, avoiding code duplication.
3. Added `File` field to `ExecResult` struct to match the SDD-specified JSON schema (was missing).
4. Used `json.MarshalIndent` for human-readable reports as specified in Best Practices.
5. Beads client is instantiated per-execution in exec, and once per batch run in batch (avoids repeated `LookPath` calls).
6. `closeBeadsTask` uses structured JSON parsing of `bd search` output (same approach as bash reference) rather than regex.
7. Timestamp in filename uses `2006-01-02T15-04-05Z` format (hyphens instead of colons) to be filesystem-safe across all platforms.

### Output

- **Commit(s)**: pending
- **PR**: pending
- **Files created/modified**:
  - `internal/runner/runner.go` — added `File` field to `ExecResult`
  - `internal/runner/claude.go` — set `File` field, removed duplicate JSON report writing
  - `internal/runner/runner_test.go` — new: ExecResult JSON structure tests
  - `cmd/forgia/cmd/report.go` — new: `writeExecReport()`, `closeBeadsTask()`, `parseBeadsTaskID()`
  - `cmd/forgia/cmd/report_test.go` — new: unit tests for report writing and Beads closure
  - `cmd/forgia/cmd/exec.go` — wired in report writing + Beads task closure after execution
  - `cmd/forgia/cmd/batch.go` — wired in per-SDD report writing + Beads task closure

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: ExecResult already had most fields needed; the Claude runner already had partial JSON report logic that served as a good reference. Extracting report writing to a shared function was clean.
- **What didn't / Cosa non ha funzionato**: Nothing blocked execution. The batch command's normal execution path was quite minimal (no validation, no config) — this SDD added report/beads but the batch path still skips validation and guardrails.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider an SDD to unify batch's normal execution path with `execSDD()` to get full validation, guardrails, and config loading per SDD. Currently batch takes a shortcut.
