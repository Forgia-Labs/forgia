---
id: "SDD-007"
fd: "FD-004"
title: "bin/forgia removal and integration wiring"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-28"
started: "2026-03-28"
completed: "2026-03-28"
tags: ["enhancement"]
---

# SDD-007: bin/forgia removal and integration wiring

> Parent FD: [[FD-004]]

## Scope

This is the **final SDD** and the **integration wiring SDD**. It must only execute after all previous SDDs (001-006) are completed and E2E tests pass.

### Part A: Remove `bin/forgia`

1. `git rm bin/forgia` — remove the deprecated bash CLI
2. `git rm modules/runners/claude.sh` — bash Claude runner (replaced by `internal/runner/claude.go`)
3. `git rm modules/runners/openhands.sh` — bash OpenHands runner (replaced by SDD-004)
4. `git rm modules/runners/validate-sdd.sh` — bash validation (replaced by `cmd/forgia/cmd/validate.go`)
5. Keep `modules/vault-template/` and `modules/claude-commands/` — these are used by the Go binary via `go:embed`

### Part B: Update references

Update all files that reference `bin/forgia` or the bash CLI:

1. **README.md** — update installation instructions to Go binary only (`go install` or `mise run go:build`)
2. **CLAUDE.md** — update CLI reference from `bin/forgia (bash)` to `cmd/forgia/ (Go)`
3. **mise.toml** — remove any tasks that reference `bin/forgia`, update `run` task to use Go binary
4. **`.github/workflows/ci.yml`** — ensure CI builds and tests the Go binary only
5. **`.forgia/config.toml` template** — if any comments reference bash, update
6. **Any SDD or FD** that references `bin/forgia` — update path references

### Part C: Integration verification

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

- [x] `bin/forgia` removed from repository
- [x] `modules/runners/claude.sh` removed
- [x] `modules/runners/openhands.sh` removed
- [x] `modules/runners/validate-sdd.sh` removed
- [x] README.md updated — installation references Go binary only
- [x] CLAUDE.md updated — CLI section references `cmd/forgia/` only
- [x] mise.toml updated — no tasks reference `bin/forgia`
- [x] CI workflows build and test Go binary only
- [x] `go build ./cmd/forgia/` succeeds after removal
- [x] `go test ./...` passes after removal
- [x] E2E tests from SDD-006 pass after removal
- [x] No file in the repository references `bin/forgia` as an executable path

## Context / Contesto

- [x] `bin/forgia` — file to remove (read first to understand what's being removed)
- [x] `modules/runners/` — runner scripts to remove
- [x] `README.md` — installation instructions to update
- [x] `CLAUDE.md` — project description to update
- [x] `mise.toml` — task runner configuration
- [x] `.github/workflows/ci.yml` — CI pipeline
- [x] All SDD-001 through SDD-006 Work Logs — verify all completed

## Constitution Check

- [x] Respects code standards — clean removal, no backwards-compatibility hacks
- [x] Respects commit conventions — `refactor(FD-004): description`
- [x] No hardcoded secrets — no secret handling
- [x] Tests defined and sufficient — full E2E + regression

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-28
- **Completed**: 2026-03-28
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. **Removed bash E2E tests alongside `bin/forgia`**: The 6 bash E2E test files (`tests/e2e*.sh`) tested the bash CLI directly — some even sourced internal functions from `bin/forgia`. Since Go E2E parity tests (SDD-006, `tests/e2e_go_test.go`) cover the same command surface, the bash tests were removed rather than rewritten.
2. **mise.toml core tasks now use `go run ./cmd/forgia`**: Instead of pointing to a compiled binary path, tasks use `go run` for convenience during development. The `go:build` task still produces a binary in `build/`.
3. **install.sh rewritten to build Go binary**: The installer now runs `go build` and symlinks `build/forgia` instead of the removed `bin/forgia`. Requires Go to be installed.
4. **FD/SDD historical references preserved**: References to `bin/forgia` in design documents (FD-001, FD-004, SDD context sections) were kept as historical design context. Only executable path references (scripts, config, CI) were updated.

### Output

- **Commit(s)**: pending
- **PR**: pending
- **Files removed**:
  - `bin/forgia` — deprecated bash CLI
  - `modules/runners/claude.sh` — bash Claude runner
  - `modules/runners/openhands.sh` — bash OpenHands runner
  - `modules/runners/validate-sdd.sh` — bash SDD validator
  - `tests/e2e.sh` — bash E2E tests (replaced by Go)
  - `tests/e2e-multi-eng.sh` — bash multi-engineer E2E tests
  - `tests/e2e-knowledge-config.sh` — bash knowledge config tests
  - `tests/e2e-codebase-memory.sh` — bash codebase memory tests
  - `tests/e2e-beads-autospec.sh` — bash beads/autospec tests
  - `tests/e2e-guardrails.sh` — bash guardrails tests
- **Files modified**:
  - `CLAUDE.md` — CLI reference updated to `cmd/forgia/ (Go, Cobra)`
  - `README.md` — Quick Start updated to Go binary, added Go to prerequisites
  - `mise.toml` — core tasks rewired to `go run`, test task runs `go test ./...`
  - `.github/workflows/ci.yml` — removed bash E2E test step
  - `install.sh` — rewritten: builds Go binary, symlinks `build/forgia`
  - `CONTRIBUTING.md` — removed bash CLI references, updated E2E test docs
  - `docs/getting-started.md` — updated install and CLI usage to Go binary

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: Pre-verification strategy (build + test before removal) made it safe to remove files confidently. SDD-006 Go E2E parity tests provided full coverage, making the bash test removal clean.
- **What didn't / Cosa non ha funzionato**: Nothing significant — all verification passed on first attempt.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: When planning a "remove legacy" SDD, explicitly enumerate which test files will be removed and which replacement tests must exist. This SDD did it well by requiring SDD-006 completion first.
