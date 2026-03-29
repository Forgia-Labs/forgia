---
id: "SDD-007"
fd: "FD-006"
title: "Integration tests — server, providers, composite skills, slash fallback"
status: done
agent: ""
assigned_to: ""
created: "2026-03-29"
started: ""
completed: ""
tags: ["enhancement", "phase:2-core"]
---

# SDD-007: Integration tests — server, providers, composite skills, slash fallback

> Parent FD: [[FD-006]]

## Scope

Comprehensive integration tests covering the full MCP stack: server transport, provider wiring, composite skill execution, and slash command fallback.

### Test suites

1. **Server E2E** — spawn `forgia mcp serve` as subprocess, send JSON-RPC on stdin, verify responses on stdout:
   - `initialize` → returns serverInfo
   - `tools/list` → returns registered tools
   - `tools/call` with known tool → returns result
   - `tools/call` with unknown tool → returns error
   - Oversized request (>10MB) → rejected

2. **Provider wiring** — create mock provider config, verify `WireProviders` creates and starts providers:
   - Valid config → provider started
   - Missing command → skipped with warning
   - Shell metacharacter in command → rejected

3. **Composite skill execution** — register mock provider, call composite skills:
   - `forgia_search_code` → results filtered through guardrails
   - `forgia_search_code` with deny.toml-matched paths in results → stripped
   - `forgia_blast_radius` → results enriched with context annotations
   - Higher-level skill calling primitive → chain works

4. **Slash command fallback** — verify 6 commands work without MCP:
   - Each command produces output when MCP unavailable (fallback path)

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| Server E2E | subprocess | Spawn binary, communicate via stdin/stdout |
| Mock provider | Go test | Implements `ToolProvider` returning fixture data |
| Fixture deny.toml | test data | Contains patterns that must be filtered |
| Fixture results | test data | Contains paths matching deny patterns |

## Constraints / Vincoli

- Language / Linguaggio: Go (test files)
- Framework: `testing`, `os/exec`, `t.TempDir()`
- Dependencies / Dipendenze: SDD-001 through SDD-006 must be complete
- Patterns / Pattern: `t.Parallel()`, table-driven where applicable

### Security (from threat model)

- Add test: `forgia_search_code` with fixture containing `.env` and `*.pem` paths → verify stripped from results (source: threat model)
- Add test: oversized JSON-RPC request → verify server rejects gracefully with error response (source: threat model)

## Best Practices

- Error handling: test helpers should use `t.Helper()` for clean stack traces
- Naming: `TestIntegration_ServerE2E_Initialize`, `TestIntegration_SearchCode_GuardrailsFilter`, etc.
- Style: mock provider as a simple struct implementing `ToolProvider`, not a subprocess

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Integration | Server initialize → returns serverInfo | Handshake |
| Integration | Server tools/list → returns tools | Tool listing |
| Integration | Server tools/call → dispatches to provider | Tool dispatch |
| Integration | Server oversized request → rejected | Security |
| Integration | WireProviders → creates providers from config | Wiring |
| Integration | WireProviders → rejects shell metacharacters | Security |
| Integration | search_code → strips deny.toml-matched paths | Guardrails |
| Integration | security_scan → returns paths only, no secret content | Security |
| Integration | Higher-level skill → calls primitive → chain works | Composition |

## Acceptance Criteria / Criteri di Accettazione

- [x] Server E2E test: initialize, tools/list, tools/call, unknown tool error
- [x] Server rejects oversized JSON-RPC requests (>10MB)
- [x] Provider wiring test: valid config, missing command, shell metacharacter rejection
- [x] `forgia_search_code` guardrails test: fixture with `.env` + `*.pem` paths stripped
- [x] `forgia_security_scan` test: returns file locations only, not secret content
- [x] Higher-level skill composition test: calls primitive, gets enriched result
- [x] All tests use `t.Parallel()` and `t.TempDir()`
- [x] `go test ./...` passes with all new tests

## Context / Contesto

- [x] `internal/mcp/server.go` — server to test (SDD-001)
- [x] `internal/mcp/mcp.go` — ProviderRegistry
- [x] `internal/skill/composite.go` — composite skills (SDD-003, SDD-004, SDD-005)
- [x] `.forgia/guardrails/deny.toml` — deny patterns for test fixtures
- [x] `tests/e2e_go_test.go` — existing E2E test pattern reference

## Constitution Check

- [x] Respects code standards — Go test conventions, `t.Parallel()`
- [x] Respects commit conventions — `test(FD-006): description`
- [x] No hardcoded secrets — test fixtures use placeholder values
- [x] Tests defined and sufficient — covers server, wiring, skills, security

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~20 min

### Decisions / Decisioni

1. Split integration tests across 3 files by test level: `tests/integration_mcp_test.go` (subprocess E2E), `internal/mcp/integration_test.go` (in-process server full-stack), `internal/skill/integration_test.go` (composite skills + slash fallback). Rationale: each file tests at a different integration level and lives in the package closest to the code under test.
2. Server E2E subprocess tests spawn `forgia mcp serve` binary via stdin/stdout pipes with timeouts. Oversized request tested both at subprocess level (11MB JSON string) and in-process (transport + server recovery).
3. Provider wiring integration tests already existed in `wiring_test.go` (valid config, missing command, shell metachar). Acceptance criteria satisfied by existing tests — no duplication needed.
4. Security scan content sanitization test injects mock results with `content` and `value` fields containing secret-like data, then verifies the serialized output contains none of it. Also verifies guardrail gap detection for unprotected files.
5. Higher-level composition chain tested via `arch_coherence`: verifies the underlying `trace_call_path` primitive is called, params are passed through, and drift detection enrichment is produced with documented vs undocumented dependency classification.
6. Slash fallback tested by reading actual `.md` command files from `modules/claude-commands/` and verifying MCP availability check, fallback path, specific tool reference, and timeout specification. Also tested via `LoadEmbedded` + registry lookup.

### Output

- **Commit(s)**: <!-- hash — to be filled after commit -->
- **PR**: —
- **Files created/modified**:
  - `tests/integration_mcp_test.go` (new — 7 Server E2E subprocess tests)
  - `internal/mcp/integration_test.go` (new — 3 in-process server integration tests)
  - `internal/skill/integration_test.go` (new — 6 composite skill + slash fallback tests)
  - `.forgia/sdd/FD-006/SDD-007-integration-tests.md` (updated — acceptance criteria + work log)

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: Existing mock patterns (mockProvider, mockVaultReader, mockSkillDispatcher) were reusable across all test files. The `TestMain` binary build in `tests/` made subprocess E2E tests straightforward. All 16 new integration tests passed on first run.
- **What didn't / Cosa non ha funzionato**: Nothing significant — the codebase was well-structured for testability.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider adding a test tag (e.g., `//go:build integration`) to separate fast unit tests from slower subprocess tests in CI. The E2E tests build the binary each run (~1s overhead).
