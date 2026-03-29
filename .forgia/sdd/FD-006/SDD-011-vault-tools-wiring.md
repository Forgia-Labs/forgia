---
id: "SDD-011"
fd: "FD-006"
title: "Vault tools integration tests + server wiring"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-29"
started: "2026-03-29"
completed: "2026-03-29"
tags: ["enhancement", "phase:2-core"]
---

# SDD-011: Vault tools integration tests + server wiring

> Parent FD: [[FD-006]]

## Scope

Wire `VaultProvider` into the MCP server and verify the complete tool surface works end-to-end.

### 1. Server wiring update

Update `cmd/forgia/cmd/mcp.go` startup path to register `VaultProvider` alongside existing providers:

```go
// After WireProviders and RegisterCompositeSkills:
vaultProvider := mcp.NewVaultProvider(v, guardrails)
providerReg.Register(vaultProvider)
```

The `tools/list` response should now include all vault tools alongside composite skills.

### 2. Integration tests

**Vault CRUD cycle via MCP** — spawn `forgia mcp serve`, send JSON-RPC requests:
1. `tools/call: forgia_fd_create` → creates FD, returns ID
2. `tools/call: forgia_fd_get` → reads created FD, verifies content
3. `tools/call: forgia_fd_list` → lists FDs, verifies the new one appears
4. `tools/call: forgia_sdd_create` → creates SDD under the FD
5. `tools/call: forgia_sdd_list` → lists SDDs, verifies the new one
6. `tools/call: forgia_validate` → validates the SDD
7. `tools/call: forgia_fd_update` → update FD status to approved
8. `tools/call: forgia_fd_get` → verify status changed

**Tool listing** — verify `tools/list` returns both vault tools (11) and composite skills (7) = 18 total tools.

**Vault tools without codebase-memory-mcp** — verify MCP server starts and vault tools work even when no knowledge graph provider is configured.

### 3. E2E test update

Update `tests/e2e_go_test.go` to verify `forgia mcp serve` exposes vault tools in tools/list.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `VaultProvider` registration | `ProviderRegistry.Register()` | VaultProvider added at startup |
| `tools/list` | JSON-RPC | Returns 18+ tools (7 composite + 11 vault) |
| CRUD cycle test | integration | Full FD+SDD lifecycle via MCP tools |

## Constraints / Vincoli

- Language / Linguaggio: Go (tests), Go (wiring)
- Framework: `testing`, `os/exec` for E2E
- Dependencies / Dipendenze: SDD-009 + SDD-010 (VaultProvider must be complete)
- Patterns / Pattern: Follow existing integration test patterns from SDD-007

## Best Practices

- Error handling: test helpers use `t.Helper()` for clean stack traces
- Naming: `TestIntegration_VaultTools_CRUDCycle`, `TestIntegration_VaultTools_WithoutKnowledgeGraph`
- Style: reuse `setupMinimalVault()` helper from existing E2E tests

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Integration | Full CRUD cycle: create FD → get → list → create SDD → validate → update | End-to-end |
| Integration | `tools/list` returns vault tools + composite skills | Tool surface |
| Integration | Vault tools work without codebase-memory-mcp | Independence |
| E2E | `forgia mcp serve` exposes vault tools | Binary-level |

## Acceptance Criteria / Criteri di Accettazione

- [x] `VaultProvider` registered in `forgia mcp serve` startup path
- [x] `tools/list` returns vault tools alongside composite skills
- [x] Full CRUD cycle works via MCP: create FD → get → list → create SDD → validate → update
- [x] MCP server with vault tools works without codebase-memory-mcp
- [x] E2E test verifies vault tools in `tools/list` output
- [x] `go test ./...` passes with all new tests

## Context / Contesto

- [x] `cmd/forgia/cmd/mcp.go` — startup path to update
- [x] `internal/mcp/vault_provider.go` — VaultProvider (SDD-009 + SDD-010)
- [x] `tests/integration_mcp_test.go` — existing MCP integration tests
- [x] `tests/e2e_go_test.go` — E2E tests to extend

## Constitution Check

- [x] Respects code standards — Go test conventions, `t.Parallel()`
- [x] Respects commit conventions — `feat(FD-006): description`
- [x] No hardcoded secrets — test fixtures use placeholder data
- [x] Tests defined and sufficient — CRUD cycle + tool listing + independence

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. Registered VaultProvider after composite skills (step 6) so provider registry is available for KG enrichment pass-through. Guardrails loaded from vault at registration time.
2. Extended `setupMinimalVault()` to create `fd/` and `sdd/` subdirectories — required for vault CRUD operations during integration tests.
3. Created separate `integration_vault_mcp_test.go` file for vault-specific MCP tests to keep test organization clean alongside existing `integration_mcp_test.go`.
4. Updated existing `TestIntegration_ServerE2E_ToolsList` to expect 18 tools (was 7) since vault provider now always registers.

### Output

- **Commit(s)**: pending
- **PR**: pending
- **Files created/modified**:
  - `cmd/forgia/cmd/mcp.go` — added VaultProvider registration with guardrails
  - `tests/integration_vault_mcp_test.go` — new: CRUD cycle, tools/list, no-KG tests
  - `tests/integration_mcp_test.go` — updated tool count from 7 to 18, added vault tool assertions
  - `tests/e2e_go_test.go` — added `TestE2E_MCPServe_VaultToolsExposed`, extended `setupMinimalVault`

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: VaultProvider plugged in cleanly — the existing ProviderRegistry + namespace routing required no changes. Integration tests using subprocess pattern proved reliable.
- **What didn't / Cosa non ha funzionato**: `setupMinimalVault` was too minimal — missing `fd/` and `sdd/` subdirs caused `CreateFD` to fail. Composite skill names don't use `forgia_` prefix (unlike provider tools), needed to adjust counting logic.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider documenting the naming convention difference between provider tools (`forgia_<ns>_<tool>`) and composite skills (plain names) to avoid confusion in future integration work.
