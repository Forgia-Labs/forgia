---
id: "SDD-009"
fd: "FD-001"
title: "Bash deprecation — banner, install.sh, README, E2E migration"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, bash, deprecation, docs]
---

# SDD-009: Bash deprecation — banner, install.sh, README, E2E migration

> Parent FD: [[FD-001]]

## Scope

Final step: deprecate the Bash CLI and point everything to the Go binary.

### 1. Deprecation banner in `bin/forgia`
Add at the top of `bin/forgia` (after shebang):
```bash
echo "⚠  bin/forgia is DEPRECATED. Use the Go binary: go install github.com/Deepzima/forgia/cmd/forgia@latest" >&2
echo "   Or: mise run go:build && build/forgia <command>" >&2
echo "" >&2
```
Keep the Bash CLI functional for 1 release cycle.

### 2. Update `install.sh`
- Primary installation: download Go binary from GitHub Releases
- Remove `bin/forgia` from PATH setup
- Keep `modules/claude-commands/` install for backward compatibility (users who haven't updated)

### 3. Update `README.md`
- Quick Start: use Go binary commands
- Replace `bin/forgia` references with `forgia` (Go binary)
- Add "Migration from Bash CLI" section

### 4. Update `mise.toml`
- `mise run forgia:*` tasks point to Go binary instead of `bin/forgia`
- Keep `mise run test` for E2E (but note it tests Go binary now)

### 5. E2E test migration
- Update `tests/e2e-beads-autospec.sh` to test Go binary instead of `bin/forgia`
- `FORGIA` variable points to `build/forgia` instead of `$ROOT_DIR/bin/forgia`
- Add new E2E tests for Go-only features (skill, validate improvements)

### 6. Remove Bash runner
- Delete `modules/runners/validate-sdd.sh` (replaced by Go `forgia validate`)
- Keep `modules/runners/claude.sh` for now (Go runner is SDD-005, but may need fallback)

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `bin/forgia` | Bash script | Add deprecation banner |
| `install.sh` | Bash script | Switch to Go binary download |
| `README.md` | Markdown | Update Quick Start and references |
| `mise.toml` | TOML | Point tasks to Go binary |
| `tests/e2e*.sh` | Bash tests | Use Go binary |

## Constraints / Vincoli

- Language / Linguaggio: Bash (modifications), Markdown (docs)
- Dependencies / Dipendenze: all SDD-001 through SDD-008 must be complete
- Patterns / Pattern: deprecation, not removal — keep bin/forgia functional
- DO NOT delete `modules/claude-commands/` — they're still the source for `go:embed`
- DO NOT delete `modules/vault-template/` — they're still the source for `go:embed`

## Best Practices

- Error handling: deprecation banner goes to stderr (not stdout) to not break piped output
- Naming: n/a
- Style: clear, actionable deprecation message with migration instructions

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| E2E | All existing e2e tests pass with Go binary | regression |
| E2E | `forgia init` + `forgia status` + `forgia validate` round-trip | integration |
| Unit | Deprecation banner appears when running bin/forgia | banner present |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `bin/forgia` shows deprecation banner on every invocation
- [ ] `install.sh` downloads Go binary from GitHub Releases
- [ ] `README.md` Quick Start uses Go binary
- [ ] `mise.toml` tasks use Go binary
- [ ] E2E tests pass against Go binary
- [ ] `modules/runners/validate-sdd.sh` removed
- [ ] No functionality regression
- [ ] `go build ./...` still succeeds

## Context / Contesto

- [ ] `bin/forgia` — current Bash CLI (add banner)
- [ ] `install.sh` — current installer
- [ ] `README.md` — current docs
- [ ] `mise.toml` — task definitions
- [ ] `tests/e2e*.sh` — E2E test files
- [ ] `modules/runners/validate-sdd.sh` — to delete

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
