---
id: "SDD-007"
fd: "FD-004"
title: "bin/forgia removal and integration wiring"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-28"
started: "2026-03-29"
completed: ""
tags: ["enhancement"]
---

# SDD-007: bin/forgia removal and integration wiring

> Parent FD: [[FD-004]]

## Scope

This is the **final SDD** and the **integration wiring SDD**. It must only execute after all previous SDDs (001-006) are completed and E2E tests pass.

### Part A: Verify E2E parity before deletion

**Before removing any file**, verify that every test scenario from the bash E2E tests (`tests/e2e*.sh`, ~215 assertions across 6 files) has a Go equivalent in `tests/e2e_go_test.go` (SDD-006). Run a side-by-side comparison:

1. For each bash test file, list all scenarios and assertions
2. For each scenario, confirm a matching Go test exists
3. If any scenario is missing from Go tests, **stop and port it first** — do NOT delete bash tests with uncovered scenarios

### Part B: Remove `bin/forgia` and bash-only files

Only after Part A confirms full coverage:

1. `git rm bin/forgia` — remove the deprecated bash CLI
2. `git rm modules/runners/claude.sh` — bash Claude runner (replaced by `internal/runner/claude.go`)
3. `git rm modules/runners/openhands.sh` — bash OpenHands runner (replaced by SDD-004)
4. `git rm modules/runners/validate-sdd.sh` — bash validation (replaced by `cmd/forgia/cmd/validate.go`)
5. `git rm tests/e2e*.sh` — bash E2E tests (replaced by `tests/e2e_go_test.go`)
6. Keep `modules/vault-template/` and `modules/claude-commands/` — these are used by the Go binary via `go:embed`

### Part C: Update references

Update all files that reference `bin/forgia` or the bash CLI:

1. **README.md** — update installation instructions to Go binary only (`go install` or `mise run go:build`)
2. **CLAUDE.md** — update CLI reference from `bin/forgia (bash)` to `cmd/forgia/ (Go)`
3. **mise.toml** — remove any tasks that reference `bin/forgia`, update `run` task to use Go binary
4. **`.github/workflows/ci.yml`** — ensure CI builds and tests the Go binary only
5. **`.forgia/config.toml` template** — if any comments reference bash, update
6. **Any SDD or FD** that references `bin/forgia` — update path references

### Part D: Integration verification

Verify the complete system works end-to-end:

1. `go build ./cmd/forgia/` — binary compiles
2. `forgia init` in a fresh directory — vault scaffolded correctly
3. `forgia status` — displays all dashboard sections
4. `forgia doctor` — all checks run
5. `forgia validate` — SDD validation works
6. `forgia skills` — lists all slash commands
7. `go test ./...` — all tests pass
8. E2E tests from SDD-006 pass

### Prerequisite

**ALL** previous SDDs (001-006) must be completed and their E2E tests passing before this SDD executes. This is the point of no return — after `bin/forgia` is removed, only the Go binary exists.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| Git rm | version control | Remove bash CLI files |
| Documentation update | markdown | README, CLAUDE.md, mise.toml, CI workflows |
| Integration test | end-to-end | Full command suite verification |

## Constraints / Vincoli

- Language / Linguaggio: N/A (file removal + documentation)
- Framework: Git, mise
- Dependencies / Dipendenze: SDD-001 through SDD-006 must ALL be done
- Patterns / Pattern: This is a destructive operation — verify everything works before removing

### Guardrails (from deny.toml)

- Do not modify `.forgia/constitution.md` or `.forgia/guardrails/deny.toml`
- `.forgia/config.toml` is write-denied — update the template in `modules/vault-template/` instead

## Best Practices

- Error handling: Run all verification checks BEFORE removing files — if any fail, stop and report
- Naming: Commit message: `refactor(FD-004): remove bin/forgia bash CLI — Go binary is now the sole CLI`
- Style: Clean removal — no backwards-compatibility shims, no "deprecated" stubs, no re-exports

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| E2E | Full command suite after removal | Regression |
| E2E | `go test ./...` passes | Unit + integration |
| E2E | SDD-006 tests still pass | Parity |
| Manual | Fresh clone → `go build` → `forgia init` → `forgia status` | Clean install |

## Acceptance Criteria / Criteri di Accettazione

- [ ] E2E parity verified — every bash E2E scenario has a Go equivalent (Part A)
- [ ] `bin/forgia` removed from repository
- [ ] `modules/runners/claude.sh` removed
- [ ] `modules/runners/openhands.sh` removed
- [ ] `modules/runners/validate-sdd.sh` removed
- [ ] `tests/e2e*.sh` removed (only after Go equivalents confirmed)
- [ ] README.md updated — installation references Go binary only
- [ ] CLAUDE.md updated — CLI section references `cmd/forgia/` only
- [ ] mise.toml updated — no tasks reference `bin/forgia`
- [ ] CI workflows build and test Go binary only
- [ ] `go build ./cmd/forgia/` succeeds after removal
- [ ] `go test ./...` passes after removal
- [ ] E2E tests from SDD-006 pass after removal
- [ ] No file in the repository references `bin/forgia` as an executable path

## Context / Contesto

- [ ] `bin/forgia` — file to remove (read first to understand what's being removed)
- [ ] `modules/runners/` — runner scripts to remove
- [ ] `tests/e2e*.sh` — bash E2E tests (verify Go equivalents exist before removing)
- [ ] `README.md` — installation instructions to update
- [ ] `CLAUDE.md` — project description to update
- [ ] `mise.toml` — task runner configuration
- [ ] `.github/workflows/ci.yml` — CI pipeline
- [ ] All SDD-001 through SDD-006 Work Logs — verify all completed

## Constitution Check

- [ ] Respects code standards — clean removal, no backwards-compatibility hacks
- [ ] Respects commit conventions — `refactor(FD-004): description`
- [ ] No hardcoded secrets — no secret handling
- [ ] Tests defined and sufficient — full E2E + regression

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~10 min

### Decisions / Decisioni

1. The earlier SDD-007 execution (commit `6ca0155`) had already removed `bin/forgia`, `modules/runners/`, and `tests/e2e*.sh` and updated CLAUDE.md, README, mise.toml, CI. This re-execution verified the removal was correct and complete.
2. Verified all FD/SDD references to `bin/forgia` are historical design context (in `.forgia/fd/` and `.forgia/sdd/`) — these were intentionally preserved, not operational paths.
3. Fixed 3 parity gaps in Go init before verifying removal: .gitignore content, CODEOWNERS conditional creation, CODEOWNERS granular entries. These fixes were part of the SDD-006 re-execution.
4. Full integration verification passed: `go build`, `go test ./...` (16 packages), E2E tests (37 tests, 35 pass, 2 skip).

### Output

- **Commit(s)**: pending (parity fixes in init.go + E2E test expansion)
- **PR**: #70
- **Files verified removed** (by earlier commit):
  - `bin/forgia`, `modules/runners/claude.sh`, `modules/runners/openhands.sh`, `modules/runners/validate-sdd.sh`
  - `tests/e2e.sh`, `tests/e2e-multi-eng.sh`, `tests/e2e-knowledge-config.sh`, `tests/e2e-codebase-memory.sh`, `tests/e2e-beads-autospec.sh`, `tests/e2e-guardrails.sh`
- **Files modified in this execution**:
  - `cmd/forgia/cmd/init.go` — fixed .gitignore, CODEOWNERS parity
  - `tests/e2e_go_test.go` — expanded to 37 tests with full parity coverage
  - `.forgia/sdd/FD-004/SDD-006-e2e-parity-tests.md` — updated scope, work log
  - `.forgia/sdd/FD-004/SDD-007-bash-removal-wiring.md` — updated work log

### Retrospective / Retrospettiva

- **What worked**: The Part A parity verification step caught that the earlier SDD-007 execution had deleted bash tests without full Go E2E coverage. Re-running SDD-006 first with expanded scope (37 tests) and then fixing the 3 parity gaps in init.go made this execution clean.
- **What didn't**: The original SDD-007 execution was premature — it deleted bash tests before Go equivalents existed. The two-step approach (SDD-006 first, SDD-007 after) was the correct sequencing.
- **Suggestions for future FDs**: Always gate destructive SDDs on a verification step that runs BEFORE deletion, not after. The Part A addition to this SDD was the right pattern.
