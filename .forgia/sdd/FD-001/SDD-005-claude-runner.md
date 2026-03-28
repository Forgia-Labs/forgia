---
id: "SDD-005"
fd: "FD-001"
title: "Claude runner — concrete Runner implementation"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, runner, claude, subprocess]
---

# SDD-005: Claude runner — concrete Runner implementation

> Parent FD: [[FD-001]]

## Scope

Create `internal/runner/claude.go` — concrete implementation of the `Runner` interface that executes SDDs via the `claude` CLI.

Responsibilities:
1. **System context builder**: assemble constitution + guardrails + dev-guide principles + lang conventions into a single context string (replaces `_build_system_context()` from Bash)
2. **Claude CLI invocation**: call `claude --dangerously-skip-permissions --max-turns N --append-system-prompt <context> -p <task>` via `exec.CommandContext`
3. **File monitor**: background goroutine that watches `git status --short` every 5s and logs new/modified files (replaces Bash background monitor)
4. **Execution logging**: write log file to `.forgia/logs/exec-SDD-NNN-<timestamp>.log` and JSON report to `.forgia/logs/exec-SDD-NNN-<timestamp>.json`
5. **ExecResult**: populate with duration, exit code, status, files changed

Does NOT implement OpenHands runner (out of scope, follow-up).

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `ClaudeRunner` | `struct` implementing `Runner` | `Name() string`, `Execute(ctx, sdd, opts) (*ExecResult, error)` |
| `var _ Runner = (*ClaudeRunner)(nil)` | compile-time check | Interface satisfaction |
| `BuildSystemContext(v vault.Vault)` | `func(vault.Vault) (string, error)` | Exported helper to build context from vault |
| `ExecOpts.AutoApprove` | `bool` | Maps to `--dangerously-skip-permissions` |
| `ExecOpts.MaxTurns` | `int` | Maps to `--max-turns` |
| `ExecOpts.DryRun` | `bool` | If true, delegate to DryRunRunner |
| `ExecOpts.GuardrailMode` | `guardrails.Mode` | Passed to Enforce before exec |

## Constraints / Vincoli

- Language / Linguaggio: Go 1.25+
- Framework: `os/exec`, `context`
- Dependencies / Dipendenze: `internal/vault`, `internal/guardrails`, `internal/process`, `internal/embedded` (for dev-guide if needed)
- Patterns / Pattern:
  - Use `exec.CommandContext(ctx, ...)` — NOT `exec.Command`
  - SIGTERM → 5s → SIGKILL on cancellation (use `internal/process.Spawn` pattern)
  - File monitor uses separate goroutine with `done` channel, stopped on exec complete
  - Log file and JSON report written even on failure
- `claude` CLI must be checked with `exec.LookPath` before execution

## Best Practices

- Error handling: wrap `exec.Command` errors with context, capture stderr separately
- Naming: `ClaudeRunner`, `BuildSystemContext`, `fileMonitor`
- Style: `slog.InfoContext` for progress, structured JSON report via `encoding/json`

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | BuildSystemContext reads constitution + guardrails + dev-guide | context assembly |
| Unit | ClaudeRunner.Name() returns "claude" | trivial |
| Unit | ExecResult populated correctly on mock execution | result fields |
| Unit | File monitor detects new files in git status | monitor goroutine |
| Unit | ClaudeRunner fails gracefully when claude CLI not found | LookPath error |
| Integration | Execute with real claude CLI (skip in CI) | optional, manual |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `ClaudeRunner` implements `Runner` interface (compile-time check)
- [ ] `BuildSystemContext` assembles constitution + guardrails + dev-guide
- [ ] Claude CLI invoked with `--append-system-prompt` and `-p` flags
- [ ] `--dangerously-skip-permissions` used when `AutoApprove=true`
- [ ] `--max-turns` passed from config
- [ ] Log file written to `.forgia/logs/`
- [ ] JSON report written with duration, exit code, status
- [ ] File monitor shows created/modified files during execution
- [ ] Graceful failure when `claude` not installed (clear error message)
- [ ] `go build ./...` succeeds
- [ ] 5+ tests pass

## Context / Contesto

- [ ] `modules/runners/claude.sh` — Bash runner to replicate (198 lines)
- [ ] `internal/runner/runner.go` — Runner interface definition
- [ ] `internal/runner/dryrun.go` — DryRunRunner (reference implementation)
- [ ] `internal/process/spawn.go` — Spawn pattern with SIGTERM/SIGKILL
- [ ] `internal/vault/file_vault.go` — Constitution(), GuardrailsRaw()
- [ ] `internal/guardrails/guardrails.go` — Parse, Enforce

## Constitution Check

- [x] Respects code standards
- [x] Respects commit conventions
- [x] No hardcoded secrets (system context built from vault files, not hardcoded)
- [x] Tests defined and sufficient

## Guardrails

The runner itself enforces guardrails before execution via `guardrails.Enforce()`. The system context includes deny.toml content so Claude is aware of restrictions.

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
