---
id: "SDD-005"
fd: "FD-004"
title: "Knowledge layer integration — codebase-memory-mcp"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-28"
started: "2026-03-28"
completed: "2026-03-28"
tags: ["enhancement"]
---

# SDD-005: Knowledge layer integration — codebase-memory-mcp

> Parent FD: [[FD-004]]

## Scope

Wire codebase-memory-mcp into three Go CLI commands, matching the bash integration:

### Part A: `forgia init` — auto-index + `.mcp.json`

After vault scaffolding, if `config.Knowledge.AutoIndex` is true and codebase-memory-mcp is installed:

1. Run `codebase-memory-mcp cli index_repository path=./` — display symbols indexed count
2. Create/update `.mcp.json` with codebase-memory-mcp server config:
   ```json
   {"mcpServers": {"codebase-memory-mcp": {"type": "stdio", "command": "codebase-memory-mcp"}}}
   ```
   If `.mcp.json` exists, merge the entry (don't overwrite other servers).

Modify `cmd/forgia/cmd/init.go` to add this as a post-scaffolding step.

### Part B: `forgia status` — stats display

Already covered by SDD-001 (knowledge stats section). This SDD provides the underlying helper functions that SDD-001 consumes.

Create `internal/knowledge/knowledge.go`:
- `Available() bool` — check if `codebase-memory-mcp` is in PATH
- `GetStats(ctx) (*Stats, error)` — call `list_projects`, parse JSON
- `Index(ctx) (int, error)` — call `index_repository`, return symbol count
- `EnsureMCPJSON(ctx, dir) error` — create/merge `.mcp.json`

### Part C: `forgia doctor` — health check

Already covered by SDD-002. This SDD provides the `Available()` and `GetStats()` functions.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `knowledge.Available()` | function | Returns true if `codebase-memory-mcp` in PATH |
| `knowledge.GetStats(ctx)` | function | Returns `*Stats{Symbols, Edges, LastIndexed}` |
| `knowledge.Index(ctx)` | function | Triggers indexing, returns symbols count |
| `knowledge.EnsureMCPJSON(ctx, dir)` | function | Creates/merges `.mcp.json` |
| codebase-memory-mcp CLI | subprocess | `cli list_projects`, `cli index_repository` |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: `os/exec` for subprocess, `encoding/json`
- Dependencies / Dipendenze: codebase-memory-mcp CLI (optional — graceful degradation)
- Patterns / Pattern: Follow `internal/beads/` pattern — client with `Available()` guard

### Guardrails (from deny.toml)

- `.mcp.json` is NOT in the deny write list — safe to create/modify
- Do not overwrite existing `.mcp.json` content — merge only

## Best Practices

- Error handling: If codebase-memory-mcp is not installed, skip silently — log info via `slog`
- Error handling: If indexing fails, warn but don't fail `init`
- Naming: Package `internal/knowledge/`, types `Client`, `Stats`
- Style: Mirror `internal/beads/beads.go` structure — `NewClient()`, `Available()`, `Call()`

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `Available()` with mock PATH | Tool detection |
| Unit | JSON parsing of `list_projects` response | Stats extraction |
| Unit | `EnsureMCPJSON()` — create new, merge into existing | MCP config management |
| Integration | `forgia init` with codebase-memory-mcp installed | Auto-indexing (CI-optional) |

## Acceptance Criteria / Criteri di Accettazione

- [x] `internal/knowledge/` package created with `Client`, `Stats`, `Available()`, `GetStats()`, `Index()`, `EnsureMCPJSON()`
- [x] `forgia init` auto-indexes codebase when codebase-memory-mcp is installed and `auto_index=true`
- [x] `forgia init` creates/merges `.mcp.json` with codebase-memory-mcp server config
- [x] Knowledge layer unavailable → skipped silently, no errors
- [x] `.mcp.json` merge preserves existing server entries
- [x] Stats include symbols (node_count), edges (edge_count), last indexed timestamp
- [x] All new functions have unit tests with `t.Parallel()`

## Context / Contesto

- [ ] `cmd/forgia/cmd/init.go` — init command to extend with knowledge layer
- [ ] `internal/beads/beads.go` — pattern reference for optional CLI client
- [ ] `internal/config/config.go` — `KnowledgeConfig` struct (provider, auto_index, auto_sync)
- [ ] `bin/forgia` lines 205-357 — bash knowledge layer functions as reference
- [ ] `.mcp.json` — MCP server configuration file format

## Constitution Check

- [x] Respects code standards — Go conventions, error wrapping, graceful degradation
- [x] Respects commit conventions — `feat(FD-004): description`
- [x] No hardcoded secrets — no secret handling
- [x] Tests defined and sufficient — unit + integration (tool-dependent tests CI-optional)

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-28
- **Completed**: 2026-03-28
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. Mirrored `internal/beads/` pattern exactly: `NewClient()` checks `LookPath`, `Available()` guard, `call()` with timeout and context — consistent API across optional integrations.
2. Used `json.RawMessage` in `mcpJSON.MCPServers` map to preserve arbitrary fields in existing server entries during merge, avoiding data loss.
3. `EnsureMCPJSON` is a package-level function (not a method on Client) since it doesn't require the binary to be installed — it only writes config.
4. `initKnowledge` in init.go loads config to check `AutoIndex` before indexing, matching the bash implementation's behavior. Indexing failure warns but does not fail init.
5. Longer timeout (30s vs beads' 5s) since repository indexing can take significant time on large codebases.

### Output

- **Commit(s)**: (pending)
- **PR**: (pending)
- **Files created/modified**:
  - `internal/knowledge/knowledge.go` — new: Client, Stats, Available(), GetStats(), Index(), EnsureMCPJSON()
  - `internal/knowledge/knowledge_test.go` — new: 12 unit tests covering all functions
  - `cmd/forgia/cmd/init.go` — modified: added initKnowledge post-scaffolding step

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: The beads package provided an excellent pattern to follow. Mirroring it made the implementation fast and consistent. Using `json.RawMessage` for the mcpServers map cleanly solved the merge-without-data-loss requirement.
- **What didn't / Cosa non ha funzionato**: Nothing significant — the SDD was well-specified and the reference code was clear.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider defining a shared interface/trait for optional CLI tool clients (beads, knowledge) to reduce duplication of the availability-check + timeout + fallback pattern.
