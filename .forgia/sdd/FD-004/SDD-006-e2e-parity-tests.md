---
id: "SDD-006"
fd: "FD-004"
title: "E2E parity tests — validate Go binary against bash"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-28"
started: ""
completed: ""
tags: ["enhancement"]
---

# SDD-006: E2E parity tests — validate Go binary against bash

> Parent FD: [[FD-004]]

## Scope

Port and adapt `tests/e2e.sh` to run against the Go binary (`build/forgia`). The tests must validate that every command the bash CLI supports produces equivalent behavior in the Go binary.

### Test Coverage

1. **`forgia init`** — scaffolds `.forgia/` with all expected directories and files
2. **`forgia status`** — displays FDs, SDDs, OPS tasks, execution logs (requires SDD-001)
3. **`forgia doctor`** — runs health checks without crashing, reports correct status
4. **`forgia validate`** — validates SDD files, catches missing sections, reports errors
5. **`forgia exec --dry-run`** — runs dry-run feasibility simulation
6. **`forgia batch --dry-run`** — runs batch dry-run for all SDDs in an FD
7. **`forgia skills`** — lists all registered slash commands
8. **`forgia version`** — prints version string

### Approach

- Create `tests/e2e_go_test.go` as a Go test file using `os/exec` to invoke the compiled binary
- Use `t.TempDir()` for isolated vault scaffolding per test
- Each test creates a fresh vault, populates with fixture data, and runs the command
- Assert on exit codes and key output strings (not exact formatting — allow minor differences)

### Prerequisite

All previous SDDs (001-005) should be completed before running E2E tests, as the tests validate the enriched commands.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| Go binary | executable | `build/forgia` compiled via `go build` |
| Test fixtures | filesystem | Vault templates, sample FDs, SDDs, OPS tasks, exec logs |
| Test assertions | Go testing | Exit codes, stdout substring matching |

## Constraints / Vincoli

- Language / Linguaggio: Go (test files)
- Framework: `testing`, `os/exec`, `t.TempDir()`
- Dependencies / Dipendenze: Compiled Go binary (`go build -o build/forgia ./cmd/forgia/`)
- Patterns / Pattern: Table-driven tests where applicable, `t.Parallel()` for independent tests

### Guardrails (from deny.toml)

- Test fixtures must NOT include real secrets — use placeholder values
- Tests run in temp directories — no risk of modifying real vault

## Best Practices

- Error handling: Tests must clean up temp dirs (automatic with `t.TempDir()`)
- Naming: `TestE2E_Init`, `TestE2E_Status`, `TestE2E_Doctor`, etc.
- Style: Assert on behavior (exit code, key output strings), not exact formatting
- Skip tests that require external tools (Docker, bd, codebase-memory-mcp) with `t.Skip("requires ...")`

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| E2E | Init → validate vault structure | Scaffolding |
| E2E | Status → displays FDs, SDDs | Dashboard |
| E2E | Doctor → runs without crash | Health check |
| E2E | Validate → catches invalid SDD | Validation |
| E2E | Skills → lists fd-arch-review and others | Skill registry |
| E2E | Version → prints version string | Version |
| E2E | Exec --dry-run → produces feasibility report (requires claude) | Dry-run (CI-optional) |

## Acceptance Criteria / Criteri di Accettazione

- [x] `tests/e2e_go_test.go` exists with tests for all core commands
- [x] `forgia init` test verifies vault directory structure
- [x] `forgia status` test verifies FD/SDD listing with fixture data
- [x] `forgia doctor` test verifies command runs without crash
- [x] `forgia validate` test verifies valid SDD passes, invalid SDD fails
- [x] `forgia skills` test verifies known slash commands listed
- [x] `forgia version` test verifies output format
- [x] All tests use `t.TempDir()` for isolation
- [x] Tests that require external tools are skipped with `t.Skip()`
- [x] `go test ./tests/...` passes

## Context / Contesto

- [ ] `tests/e2e.sh` — existing bash E2E tests to port
- [ ] `cmd/forgia/cmd/*.go` — all commands being tested
- [ ] `modules/vault-template/` — template files for init testing
- [ ] `cmd/forgia/cmd/*_test.go` — existing unit/integration tests for reference

## Constitution Check

- [x] Respects code standards — Go test conventions, `t.Parallel()`, table-driven
- [x] Respects commit conventions — `test(FD-004): description`
- [x] No hardcoded secrets — fixture data uses placeholder values
- [x] Tests defined and sufficient — E2E coverage for all core commands

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-28
- **Completed**: 2026-03-28
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. Used `TestMain` to build the Go binary once before all tests, stored path in package-level var. This avoids rebuilding per-test and keeps tests fast (~2s total).
2. Set `HOME` env var to temp dir in `runForgia` helper to suppress beads/claude-commands noise that depends on the host's home directory.
3. Used `CombinedOutput` (stdout+stderr) for assertions since slog writes to stderr while fmt.Print writes to stdout — both are relevant for behavior verification.
4. Skipped `exec --dry-run` and `batch --dry-run` tests with `t.Skip("requires claude CLI")` since they need the claude binary for simulation.
5. Used table-driven subtests for language detection (Go, Rust, Python) following Go test conventions.
6. Created helper functions (`writeFD`, `writeSDD`, `writeValidSDD`, `writeInvalidSDD`) for fixture generation to keep tests focused on assertions.

### Output

- **Commit(s)**: (pending)
- **PR**: (pending)
- **Files created/modified**:
  - `tests/e2e_go_test.go` — new: 15 E2E tests (13 pass, 2 skip)
  - `.forgia/sdd/FD-004/SDD-006-e2e-parity-tests.md` — updated: acceptance criteria, work log

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: Building the binary in TestMain and running all tests in parallel with t.TempDir() isolation made tests fast and reliable. The existing bash e2e.sh was a good reference for what to assert on, and the Go command implementations were clean enough to predict output strings accurately.
- **What didn't / Cosa non ha funzionato**: Minor issue — the bash e2e tests assert on strings like "Feature Designs" and "Execution Specs" which don't match the Go binary output. The Go binary uses different phrasing ("No Feature Designs found", "Forgia Dashboard"). Had to verify actual binary output before writing assertions.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider adding a CI step that runs `go test ./tests/...` alongside the bash e2e tests. The Go E2E tests are faster (~2s vs ~10s for bash) and provide better isolation. Once bash CLI is fully removed (SDD-007), the bash e2e tests can be retired.
