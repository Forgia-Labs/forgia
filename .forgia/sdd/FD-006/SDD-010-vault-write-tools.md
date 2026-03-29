---
id: "SDD-010"
fd: "FD-006"
title: "Vault write tools — fd_create, fd_update, sdd_create, sdd_update"
status: done
agent: ""
assigned_to: ""
created: "2026-03-29"
started: ""
completed: ""
tags: ["enhancement", "phase:2-core"]
---

# SDD-010: Vault write tools — MCP tools for vault write operations

> Parent FD: [[FD-006]]

## Scope

Extend `VaultProvider` (from SDD-009) with write operation tools. These tools create and update FDs and SDDs through the MCP server.

### Tools

| Tool | Vault Method | Input | Output |
|------|-------------|-------|--------|
| `forgia_fd_create` | `CreateFD(ctx, fd)` | `{title, author, status, priority, tags, upstream_issue, body}` | `{id: "FD-NNN", path: ".forgia/fd/FD-NNN-kebab.md"}` |
| `forgia_fd_update` | `UpdateFD(ctx, fd)` | `{id: "FD-001", status: "approved", reviewed: true, reviewer: "claude"}` | `{updated: true}` |
| `forgia_sdd_create` | `CreateSDD(ctx, sdd)` | `{fd, title, scope, interfaces, constraints, acceptance_criteria}` | `{id: "SDD-NNN", path: ".forgia/sdd/FD-NNN/SDD-NNN-kebab.md"}` |
| `forgia_sdd_update` | `UpdateSDD(ctx, sdd)` | `{fd: "FD-001", id: "SDD-001", status: "done", agent: "claude-code"}` | `{updated: true}` |

### Knowledge graph enrichment

Write tools leverage codebase-memory-mcp when available to produce richer output:

- **`forgia_fd_create`**: if registry has `"code"` provider, calls `forgia_arch_init` to auto-detect architecture for Mermaid diagrams, `forgia_search_code` to identify components for the Interfaces table, and detects language stack for Constraints. Falls back to user-provided content only.
- **`forgia_sdd_create`**: if registry has `"code"` provider, calls `forgia_blast_radius` to assess scope impact, `forgia_search_code` to populate Context section with relevant file paths. Falls back to FD-derived content only.
- **`forgia_fd_update`** and **`forgia_sdd_update`**: pure writes, no knowledge graph needed.

The pattern: check `registry != nil` → attempt composite skill call with 3s timeout → merge results into the created/updated document → if error/timeout, proceed with vault-only data.

### Input validation

Each write tool validates its input before calling vault methods:
- **fd_create**: title required (non-empty), status must be valid (planned/approved/in-progress/complete/closed/rejected)
- **fd_update**: id required, at least one field to update
- **sdd_create**: fd required (must exist), title required
- **sdd_update**: fd + id required, at least one field to update

### ID generation

- `fd_create`: uses `vault.NewFDID(title, author)` — hash-based, collision-proof (e.g., `FD-a3f2`). This is the canonical Go implementation. Slash commands that use the MCP tool MUST use the hash-based ID, not sequential numbering.
- `sdd_create`: auto-generates next SDD-NNN sequentially within the FD directory (SDDs are always sequential within an FD)

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `VaultProvider` (extended) | `ToolProvider` | SDD-009 provider extended with 4 write tools |
| `vault.Vault` | dependency | `CreateFD`, `UpdateFD`, `CreateSDD`, `UpdateSDD` methods |
| `guardrails.Guardrails` | dependency | Write patterns checked before creating/updating files |
| `ProviderRegistry` | optional dependency | For KG enrichment — calls `forgia_arch_init`, `forgia_blast_radius`, `forgia_search_code` when available |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: `internal/mcp` (VaultProvider from SDD-009), `internal/vault`
- Dependencies / Dipendenze: SDD-009 (VaultProvider must exist)
- Patterns / Pattern: Extend VaultProvider's `Tools()` and `Call()` dispatch

### Security

- Write tools MUST respect deny.toml `[write]` patterns — never write to constitution.md, config.toml, deny.toml
- Validate all input fields — reject empty titles, invalid status values, malformed IDs
- ID generation must be collision-safe — check existing files before creating
- Log every write operation via `slog.InfoContext` (tool name + target path)

## Best Practices

- Error handling: return MCP error for validation failures (invalid status, empty title, missing parent FD)
- Error handling: return MCP error if write would violate deny.toml patterns
- Naming: tool names match read tools convention (`forgia_fd_create`, `forgia_sdd_update`)
- Style: share `VaultProvider` struct from SDD-009, just add tools and dispatch cases

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `forgia_fd_create` with valid input → creates FD file | Create path |
| Unit | `forgia_fd_create` with empty title → returns error | Validation |
| Unit | `forgia_fd_create` auto-generates next ID | ID generation |
| Unit | `forgia_fd_update` changes status → verifies on disk | Update path |
| Unit | `forgia_sdd_create` under existing FD → creates SDD | Create path |
| Unit | `forgia_sdd_create` under missing FD → returns error | Missing parent |
| Unit | `forgia_sdd_update` changes agent + status | Update path |
| Unit | Write to deny.toml-protected path → rejected | Security |
| Unit | `forgia_fd_create` with KG → architecture auto-detected, interfaces populated | KG enrichment |
| Unit | `forgia_sdd_create` with KG → context paths populated, scope enriched | KG enrichment |
| Unit | `forgia_fd_create` without KG → user content only, no error | Fallback |

## Acceptance Criteria / Criteri di Accettazione

- [x] `forgia_fd_create` creates FD in vault, returns file path + auto-generated ID
- [x] `forgia_fd_update` updates FD frontmatter fields
- [x] `forgia_sdd_create` creates SDD under correct FD directory with auto-generated ID
- [x] `forgia_sdd_update` updates SDD frontmatter fields
- [x] Input validation: empty titles rejected, invalid statuses rejected, missing parent FDs rejected
- [x] Write tools respect deny.toml `[write]` patterns
- [x] Every write operation logged (tool name + target path)
- [x] `forgia_fd_create` enriches with architecture/interfaces from knowledge graph when available
- [x] `forgia_sdd_create` enriches with context paths/scope from knowledge graph when available
- [x] All write tools work without codebase-memory-mcp (fallback to user-provided content)
- [x] All functions have unit tests with `t.Parallel()`

## Context / Contesto

- [ ] `.forgia/sdd/FD-006/SDD-009-vault-read-tools.md` — VaultProvider to extend
- [ ] `internal/vault/vault.go` — `CreateFD`, `UpdateFD`, `CreateSDD`, `UpdateSDD`
- [ ] `internal/vault/file_vault.go` — concrete write implementations
- [ ] `internal/guardrails/guardrails.go` — `CheckFilePaths` for write validation
- [ ] `.forgia/guardrails/deny.toml` — write deny patterns

## Constitution Check

- [ ] Respects code standards — Go conventions, error wrapping, input validation
- [ ] Respects commit conventions — `feat(FD-006): description`
- [ ] No hardcoded secrets — vault tools handle specs, not secrets
- [ ] Tests defined and sufficient — CRUD + validation + security

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~30 min

### Decisions / Decisioni

1. Write handlers in separate file (`vault_provider_write.go`) to keep code organized while sharing the same `VaultProvider` struct from SDD-009.
2. FD ID generation uses `vault.NewFDID(title, author)` (hash-based) as specified. SDD IDs are sequential (`SDD-NNN`) within each FD directory, parsing existing IDs to find the next number.
3. KG enrichment returns data in the result's `enrichment` field rather than modifying the created file — keeps the vault abstraction clean and lets the caller decide how to use it.
4. Guardrails check uses relative paths (`.forgia/fd/FD-xxx.md`) consistent with deny.toml patterns. CheckWritePaths is called before vault.CreateFD/UpdateFD to fail-close.
5. Status validation uses exhaustive switch on vault.FDStatus/SDDStatus constants — any new status added to the type requires updating the validator.
6. Write tests in separate file (`vault_provider_write_test.go`) reusing the existing mock from `vault_provider_test.go`, with capture fields added to the mock struct.

### Output

- **Commit(s)**: <!-- hash — to be filled after commit -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `internal/mcp/vault_provider_write.go` (new — 4 write handlers, validation, KG enrichment, ID generation helpers)
  - `internal/mcp/vault_provider_write_test.go` (new — 27 unit tests covering CRUD, validation, guardrails, KG enrichment, fallback)
  - `internal/mcp/vault_provider.go` (modified — 4 tool definitions in Tools(), 4 dispatch cases in Call(), updated comment)
  - `internal/mcp/vault_provider_test.go` (modified — mock capture fields, tool count 7→11, expected tool names updated)
  - `.forgia/sdd/FD-006/SDD-010-vault-write-tools.md` (modified — status→done, Work Log filled)

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: The existing VaultProvider pattern (ToolProvider interface, namespace routing, KG enrichment with timeout) was straightforward to extend. The mock vault with capture fields made write tests clean and fast.
- **What didn't / Cosa non ha funzionato**: The vault.CreateFD/CreateSDD methods write a default body; passing user-provided body content through the MCP tool would require vault interface changes (out of scope). Accepted as a limitation.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider adding a `body` parameter to vault.CreateFD/CreateSDD for richer file creation from MCP tools. The `enrichSDDWithKG` fdID parameter is currently unused — when blast_radius supports FD-scoped analysis, wire it through.
