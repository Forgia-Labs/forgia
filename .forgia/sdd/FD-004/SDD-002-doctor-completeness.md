---
id: "SDD-002"
fd: "FD-004"
title: "doctor completeness — OpenHands, API keys, Beads, knowledge"
status: done
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: ["enhancement"]
---

# SDD-002: doctor completeness — OpenHands, API keys, Beads, knowledge

> Parent FD: [[FD-004]]

## Scope

Extend `cmd/forgia/cmd/doctor.go` to add the missing health checks that bash `forgia doctor` performs. The Go version currently checks: vault, config, guardrails, git, claude, docker, bd, fswatch, yq. Add:

1. **OpenHands image pulled** — `docker image inspect ghcr.io/openhands/openhands:latest`
2. **OpenHands container running** — `docker ps --filter name=openhands`
3. **Claude Code slash commands installed** — check `$HOME/.claude/commands/fd-*.md` and `sdd-*.md` exist
4. **Beads circuit breaker state** — read `/tmp/beads-dolt-circuit-*.json`, check for `"state":"open"`
5. **LLM API key presence** — check `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` environment variables are set (check presence only, never log values)
6. **Knowledge layer health** — check `codebase-memory-mcp` installed, call `list_projects` for stats (symbols, edges, last sync)

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| Docker checks | subprocess | `docker image inspect`, `docker ps` via `exec.Command` |
| Claude commands check | filesystem | Glob `$HOME/.claude/commands/{fd,sdd}-*.md` |
| Beads circuit breaker | filesystem | Read `/tmp/beads-dolt-circuit-*.json`, parse JSON |
| API key check | environment | `os.Getenv("ANTHROPIC_API_KEY")`, `os.Getenv("OPENAI_API_KEY")` — presence only |
| Knowledge health | subprocess | `codebase-memory-mcp cli list_projects` → JSON with node_count, edge_count, last_indexed |
| Doctor output | stdout | Formatted lines with status icons (pass/fail/skip) |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: Cobra CLI
- Dependencies / Dipendenze: None new — uses `os/exec`, `os`, `filepath`, `encoding/json`
- Patterns / Pattern: Follow existing check pattern in `doctor.go` — each check prints a line with status icon

### Guardrails (from deny.toml)

- NEVER log or print API key values — only check `len(os.Getenv(...)) > 0`
- Do not read files matching deny patterns (not applicable here — `/tmp/` and `$HOME/.claude/` are not in deny list)

### Security

- API key check must ONLY verify presence (non-empty string), never print, log, or store the value
- Circuit breaker JSON may contain internal state — only extract `state` field

## Best Practices

- Error handling: Each check is independent — one failing check must not prevent others from running. Use the existing pass/fail/skip pattern.
- Naming: `check*` function naming (e.g., `checkOpenHands`, `checkAPIKeys`, `checkKnowledgeLayer`)
- Style: Match existing doctor output format — aligned columns with status icon, component name, details

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `checkAPIKeys()` with env vars set/unset | Key presence detection |
| Unit | Circuit breaker JSON parsing (open, closed, missing file) | Beads health |
| Unit | Claude commands glob with mock HOME | Command detection |
| Integration | `forgia doctor` runs without crash on dev machine | Smoke test |

## Acceptance Criteria / Criteri di Accettazione

- [x] `forgia doctor` checks OpenHands image pulled and container running
- [x] `forgia doctor` checks Claude Code slash commands installed
- [x] `forgia doctor` checks Beads circuit breaker state
- [x] `forgia doctor` checks LLM API key presence (ANTHROPIC_API_KEY or OPENAI_API_KEY)
- [x] `forgia doctor` checks codebase-memory-mcp installed, displays symbols/edges/sync time
- [x] API key values are NEVER logged or printed — only presence is checked
- [x] Each check is independent — one failure does not block others
- [x] All new functions have unit tests with `t.Parallel()` (except env-dependent tests — Go constraint)

## Context / Contesto

- [x] `cmd/forgia/cmd/doctor.go` — existing doctor command to extend
- [x] `bin/forgia` lines 717-921 — bash `cmd_doctor()` as reference
- [x] `internal/beads/beads.go` — Beads client patterns
- [x] `internal/config/config.go` — `OpenHandsConfig` struct for image name
- [x] `.forgia/guardrails/deny.toml` — confirm no path conflicts

## Constitution Check

- [x] Respects code standards — Go conventions, error wrapping
- [x] Respects commit conventions — `feat(FD-004): description`
- [x] No hardcoded secrets — API keys checked for presence only, never logged
- [x] Tests defined and sufficient — unit + integration

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-28T22:55:00+01:00
- **Completed**: 2026-03-28T23:05:00+01:00
- **Duration / Durata**: ~10 min

### Decisions / Decisioni

1. Split `checkClaudeCommands()` into a public wrapper and a `checkClaudeCommandsIn(home)` helper — allows unit testing with mock HOME directory without `t.Setenv` (which is incompatible with `t.Parallel()` in Go).
2. Extracted `parseCircuitBreakerState()` as a standalone function for testability — only reads the `state` field from Beads circuit breaker JSON, ignoring all other internal state per security constraint.
3. API key tests use `t.Setenv` without `t.Parallel()` — this is a Go language constraint (`t.Setenv` mutates process-level env, cannot run concurrently). All other tests use `t.Parallel()`.
4. Reused `KnowledgeProject` struct and `formatRelativeTime` from `status.go` — no duplication, same package.
5. OpenHands checks skip gracefully when Docker is unavailable (LookPath gate), matching the bash reference behavior.

### Output

- **Commit(s)**: b26e63b
- **PR**: <!-- link -->
- **Files created/modified**:
  - `cmd/forgia/cmd/doctor.go` — extended with 6 new check functions + circuit breaker types
  - `cmd/forgia/cmd/doctor_test.go` — new file, 18 unit tests covering all new checks

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: Existing `checkResult` pattern made adding new checks straightforward. The bash reference `cmd_doctor()` was a clear blueprint. Splitting testable logic (parseCircuitBreakerState, checkClaudeCommandsIn) from system-dependent wrappers enabled good test coverage.
- **What didn't / Cosa non ha funzionato**: `t.Setenv` + `t.Parallel()` incompatibility required dropping parallelism for env-dependent tests (5 tests). Minor trade-off.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider defining a `Check` interface (`Name() string`, `Run(ctx) checkResult`) to enable plugin-style check registration, making future extensions even simpler.
