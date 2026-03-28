---
id: "SDD-006"
fd: "FD-001"
title: "forgia exec + forgia batch — execution commands"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, cli, exec, batch, cobra]
---

# SDD-006: forgia exec + forgia batch — execution commands

> Parent FD: [[FD-001]]

## Scope

Create two Cobra commands:

### `forgia exec` (`cmd/forgia/cmd/exec.go`)
Execute a single SDD:
1. Open vault, load config
2. Parse SDD file (frontmatter + body)
3. Pre-exec validation (delegate to SDD-004 validate logic)
4. Load guardrails, enforce in configured mode
5. Resolve runner from config or `--runner` flag
6. Call `runner.Execute(ctx, sdd, opts)`
7. Post-exec: update SDD work log with results
8. Post-exec: update Beads task status if available
9. Report: duration, exit code, files changed

Flags: `--runner=claude|openhands`, `--dry-run`, `--mode=off|careful|freeze|guard`

### `forgia batch` (`cmd/forgia/cmd/batch.go`)
Execute all pending SDDs for an FD:
1. Open vault, list SDDs in `.forgia/sdd/FD-NNN/`
2. Filter: only `status: planned` or `status: validated`
3. Order: by SDD number (SDD-001 before SDD-002), or by Beads dependency (`bd ready`) if available
4. Sequential execution: call exec logic for each SDD, stop on failure
5. Summary: N succeeded, M failed, K skipped

Flags: same as exec + `--continue-on-error`

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `execCmd` | `*cobra.Command` | `forgia exec <sdd-file> [flags]` |
| `batchCmd` | `*cobra.Command` | `forgia batch <FD-NNN> [flags]` |
| `runner.Runner` | existing interface | Execute(ctx, sdd, opts) |
| `runner.ClaudeRunner` | from SDD-005 | Concrete runner |
| `guardrails.Enforce()` | from SDD-004/existing | Pre-exec check |
| `vault.UpdateSDD()` | existing | Write work log |
| `beads.Client` | existing | Task status update |

## Constraints / Vincoli

- Language / Linguaggio: Go 1.25+
- Framework: Cobra
- Dependencies / Dipendenze: `internal/vault`, `internal/config`, `internal/guardrails`, `internal/runner` (SDD-005), `internal/beads`
- Patterns / Pattern:
  - exec and batch share core logic — extract `runSDD(ctx, v, cfg, sdd, opts) error` helper
  - batch uses sequential execution (parallel is #15, out of scope)
  - `--dry-run` delegates to `runner.DryRunRunner`
  - Work log update: set executor, started, completed, duration in SDD file

## Best Practices

- Error handling: exec returns the runner's exit code as process exit code. Batch aggregates results.
- Naming: `runSDD` for shared logic, `execCmd`/`batchCmd` for Cobra commands
- Style: progress output during execution (file monitor from SDD-005)

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | Exec with mock runner succeeds | happy path |
| Unit | Exec fails on invalid SDD (validation gate) | validation integration |
| Unit | Exec fails on guardrail violation | guardrails integration |
| Unit | Batch collects all pending SDDs in order | ordering |
| Unit | Batch stops on first failure (default) | stop behavior |
| Unit | Batch continues on error with --continue-on-error | flag behavior |
| Unit | --dry-run flag routes to DryRunRunner | flag routing |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `forgia exec SDD.md` executes a single SDD
- [ ] `forgia exec SDD.md --dry-run` simulates without modifying files
- [ ] `forgia exec SDD.md --mode=guard` enforces boundaries
- [ ] `forgia exec` validates SDD before execution (rejects invalid)
- [ ] `forgia exec` updates SDD work log after execution
- [ ] `forgia batch FD-001` executes all pending SDDs sequentially
- [ ] `forgia batch` respects dependency order (if Beads available)
- [ ] `forgia batch --continue-on-error` doesn't stop on first failure
- [ ] Exit codes propagated correctly
- [ ] `go build ./...` succeeds
- [ ] 7+ tests pass

## Context / Contesto

- [ ] `bin/forgia` `cmd_exec()` — Bash exec implementation
- [ ] `bin/forgia` `cmd_batch()` — Bash batch implementation
- [ ] `internal/runner/claude.go` — from SDD-005
- [ ] `internal/runner/dryrun.go` — existing DryRunRunner
- [ ] `internal/guardrails/guardrails.go` — Enforce()
- [ ] `cmd/forgia/cmd/sync.go` — reference for vault/config wiring

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
