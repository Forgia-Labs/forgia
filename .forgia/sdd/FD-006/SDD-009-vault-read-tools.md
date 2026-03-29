---
id: "SDD-009"
fd: "FD-006"
title: "Vault read tools — fd_get, fd_list, sdd_get, sdd_list, status, validate, threat_model_get"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-29"
started: "2026-03-29"
completed: "2026-03-29"
tags: ["enhancement", "phase:2-core"]
---

# SDD-009: Vault read tools — MCP tools for vault read operations

> Parent FD: [[FD-006]]

## Scope

Implement a `VaultProvider` that exposes vault read operations as MCP tools. This provider implements the `ToolProvider` interface and calls `vault.Vault` methods directly — no knowledge graph dependency.

### Tools

| Tool | Vault Method | Input | Output |
|------|-------------|-------|--------|
| `forgia_fd_get` | `GetFD(ctx, id)` | `{id: "FD-001"}` | FD frontmatter + body as JSON |
| `forgia_fd_list` | `ListFDs(ctx)` | `{status: "planned"}` (optional filter) | Array of FD summaries (id, title, status, priority, author) |
| `forgia_sdd_get` | `GetSDD(ctx, fdID, sddID)` | `{fd: "FD-001", id: "SDD-001"}` | SDD frontmatter + body as JSON |
| `forgia_sdd_list` | `ListSDDs(ctx, fdID)` | `{fd: "FD-001", status: "planned"}` (optional filter) | Array of SDD summaries |
| `forgia_status` | `ListFDs` + `ListSDDs` + OPS/logs | `{}` | Full dashboard data as JSON |
| `forgia_validate` | Validate SDD against template | `{path: "..."}` or `{fd: "FD-001"}` | Validation result with errors/warnings |
| `forgia_threat_model_get` | Read file | `{fd: "FD-001"}` | Threat model content or `null` |

### Implementation

Create `internal/mcp/vault_provider.go`:

```go
type VaultProvider struct {
    vault    vault.Vault
    guard    *guardrails.Guardrails  // for validate tool
    registry *ProviderRegistry       // for KG-enriched tools (optional)
}

func NewVaultProvider(v vault.Vault, g *guardrails.Guardrails, reg *ProviderRegistry) *VaultProvider
```

Implements `ToolProvider`:
- `Name()` → `"vault"`
- `Tools()` → 7 tool definitions with JSON Schema parameters
- `Call(ctx, tool, params)` → dispatch to the appropriate method
- `Start(ctx)` → no-op (in-process)
- `Stop()` → no-op
- `Healthy()` → true if vault is open

### Knowledge graph enrichment

Tools that benefit from codebase-memory-mcp use the `registry` to call composite skills when available:

- **`forgia_status`**: if registry has `"code"` provider, includes knowledge layer stats (symbols, edges, sync time) alongside vault data. Falls back to vault-only dashboard.
- **`forgia_validate`**: if registry has `"code"` provider, calls `forgia_search_code` to verify context file paths actually exist in the codebase. Falls back to filesystem-only checks.

The pattern: check `registry != nil` → attempt composite skill call with 3s timeout → if error/timeout, fall back to vault-only result. Same pattern as slash commands (SDD-006).

### Tool parameter schemas

Each tool has a JSON Schema for its `inputSchema` so Claude and other agents know what parameters to pass. Example for `fd_get`:

```json
{
  "type": "object",
  "properties": {
    "id": {"type": "string", "description": "FD identifier (e.g., FD-001)"}
  },
  "required": ["id"]
}
```

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `VaultProvider` | `ToolProvider` implementation | Exposes vault reads as MCP tools |
| `vault.Vault` | dependency | Existing vault interface — all reads go through this |
| `guardrails.Guardrails` | dependency | For `forgia_validate` tool |
| Tool definitions | `[]ToolDefinition` | 7 tools with JSON Schema parameters |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: `internal/mcp` (ToolProvider), `internal/vault` (Vault), `internal/guardrails`
- Dependencies / Dipendenze: None new — uses existing vault and guardrails packages
- Patterns / Pattern: Follow `BoardProvider` as reference `ToolProvider` implementation

### Guardrails

- Read tools respect deny.toml `[read]` patterns (vault files are not in deny list, so this is inherently safe)
- `forgia_validate` applies guardrails checking on SDD context paths

## Best Practices

- Error handling: return MCP error responses for invalid IDs, missing FDs/SDDs — don't panic
- Naming: tool names prefixed with `forgia_` for namespace consistency
- Style: each tool handler is a separate method on `VaultProvider` for readability

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `forgia_fd_get` with existing FD → returns data | Read path |
| Unit | `forgia_fd_get` with missing FD → returns error | Error path |
| Unit | `forgia_fd_list` with filter → returns filtered results | Filtering |
| Unit | `forgia_sdd_list` for FD → returns SDDs | SDD listing |
| Unit | `forgia_validate` with valid SDD → no errors | Validation |
| Unit | `forgia_validate` with invalid SDD → returns errors | Validation errors |
| Unit | `forgia_threat_model_get` when exists → returns content | Read path |
| Unit | `forgia_threat_model_get` when missing → returns null | Missing path |
| Unit | `forgia_status` → returns dashboard JSON | Dashboard |
| Unit | `forgia_status` with KG available → includes knowledge stats | KG enrichment |
| Unit | `forgia_validate` with KG available → verifies context paths in codebase | KG enrichment |
| Unit | `forgia_status` without KG → vault-only dashboard (no error) | Fallback |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `VaultProvider` implements `ToolProvider` interface
- [ ] `forgia_fd_get` returns FD frontmatter + body as JSON
- [ ] `forgia_fd_list` returns all FDs, supports optional status filter
- [ ] `forgia_sdd_get` returns SDD by FD + SDD ID
- [ ] `forgia_sdd_list` returns SDDs for an FD, supports optional status filter
- [ ] `forgia_status` returns full dashboard data as JSON
- [ ] `forgia_validate` validates SDD and returns errors/warnings
- [ ] `forgia_threat_model_get` returns content or null
- [ ] All tools have JSON Schema parameter definitions
- [ ] Invalid IDs/missing files return MCP error responses (don't crash)
- [ ] `forgia_status` includes knowledge layer stats when codebase-memory-mcp available
- [ ] `forgia_validate` verifies context paths via knowledge graph when available
- [ ] All read tools work without codebase-memory-mcp (fallback to vault-only)
- [ ] All functions have unit tests with `t.Parallel()`

## Context / Contesto

- [ ] `internal/mcp/board_provider.go` — reference `ToolProvider` implementation
- [ ] `internal/vault/vault.go` — `Vault` interface methods
- [ ] `internal/vault/file_vault.go` — concrete implementation
- [ ] `internal/guardrails/guardrails.go` — for validate tool
- [ ] `cmd/forgia/cmd/status.go` — status dashboard logic (reuse for `forgia_status`)
- [ ] `cmd/forgia/cmd/validate.go` — validation logic (reuse for `forgia_validate`)

## Constitution Check

- [ ] Respects code standards — Go conventions, error wrapping
- [ ] Respects commit conventions — `feat(FD-006): description`
- [ ] No hardcoded secrets — vault tools read specs, not secrets
- [ ] Tests defined and sufficient — all 7 tools tested

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~30m

### Decisions / Decisioni

1. Tool namespace set to "vault" following BoardProvider pattern → tools exposed as `forgia_vault_fd_get`, `forgia_vault_sdd_get`, etc. SDD table used shorter names (`forgia_fd_get`) but the codebase ToolName convention adds the namespace prefix. Consistent with existing routing logic.
2. Validation logic ported from `cmd/forgia/cmd/validate.go` into VaultProvider — the cmd package functions aren't importable from internal/mcp. Core checks preserved: frontmatter fields, required sections, parent FD reference, guardrails on context paths.
3. KG enrichment uses 3s timeout with fallback: status tool calls `forgia_code_list_projects` via registry; validate tool calls `forgia_code_search` for context path verification, falls back to `os.Stat` on filesystem.
4. Threat model path follows existing convention: `.forgia/fd/{fdID}-threat-model.md` (from `/fd-threat-model` and `/fd-sdd` slash commands).
5. FD/SDD get tools read the file from `FilePath` to extract the markdown body alongside structured frontmatter data.

### Output

- **Commit(s)**: <!-- hash — to be filled after commit -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `internal/mcp/vault_provider.go` — VaultProvider implementation (7 tools, ~730 LOC)
  - `internal/mcp/vault_provider_test.go` — 32 test functions covering all tools, error paths, KG enrichment, guardrails

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: BoardProvider as reference pattern made the ToolProvider implementation straightforward. The existing vault.Vault interface covered all read needs without modification. Mock-based testing with temp dirs gave good coverage including filesystem interactions (validate, threat_model_get, exec logs).
- **What didn't / Cosa non ha funzionato**: Tool naming — the SDD specified `forgia_fd_get` but the ToolName convention produces `forgia_vault_fd_get`. The implementation follows the codebase pattern for consistency but the SDD should be updated.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider extracting the SDD validation logic from `cmd/forgia/cmd/validate.go` into an `internal/` package so it can be shared between CLI and MCP without duplication.
