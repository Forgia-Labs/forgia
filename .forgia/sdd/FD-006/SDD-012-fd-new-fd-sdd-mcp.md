---
id: "SDD-012"
fd: "FD-006"
title: "/fd-new and /fd-sdd MCP integration — vault tools + hash-based IDs"
status: complete
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-29"
started: "2026-03-29"
completed: "2026-03-29"
tags: ["enhancement", "phase:2-core"]
---

# SDD-012: /fd-new and /fd-sdd MCP integration — vault tools + hash-based IDs

> Parent FD: [[FD-006]]

## Scope

Update `/fd-new` and `/fd-sdd` slash commands to use vault MCP tools when available. This changes how FDs and SDDs are created — the MCP tool path uses `vault.NewFDID()` for hash-based IDs.

### `/fd-new` changes

Add MCP reachability check (same pattern as SDD-006). When MCP is available:

1. Agent builds the FD content (Problem, Solutions, Architecture, etc.) as it does today
2. Instead of writing the file directly via Write tool, call `forgia_fd_create` with the content
3. `forgia_fd_create` uses `vault.NewFDID(title, author)` → hash-based ID (e.g., `FD-a3f2`)
4. If KG is available, `forgia_fd_create` enriches with architecture detection and component identification (SDD-010)
5. Report the hash-based ID to the user: "FD creato: `.forgia/fd/FD-a3f2-kebab-title.md`"

When MCP is unavailable (fallback):
1. Current behavior preserved — agent scans existing FDs, increments highest number → sequential ID (`FD-007`)
2. Writes file directly via Write tool

**ID convention change**: this is the key behavioral difference. When MCP is available, IDs are hash-based (`FD-a3f2`). When unavailable, IDs are sequential (`FD-007`). Both are valid — the vault accepts any `FD-*` format.

### `/fd-sdd` changes

When MCP is available:
1. Agent generates SDD content from the FD as today
2. Call `forgia_sdd_create` for each SDD instead of writing directly
3. `forgia_sdd_create` auto-generates sequential SDD-NNN within the FD directory
4. If KG available, enriches with context paths and scope assessment

When MCP unavailable: current behavior preserved.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `/fd-new` MCP path | `forgia_fd_create` tool call | Creates FD with hash-based ID via vault |
| `/fd-new` fallback path | Write tool | Creates FD with sequential ID (existing behavior) |
| `/fd-sdd` MCP path | `forgia_sdd_create` tool call | Creates SDDs via vault |
| `/fd-sdd` fallback path | Write tool | Creates SDDs directly (existing behavior) |
| `vault.NewFDID()` | Go function | Hash-based ID generation — canonical source |

## Constraints / Vincoli

- Language / Linguaggio: Markdown (slash command prompt updates)
- Framework: Claude Code slash command system
- Dependencies / Dipendenze: SDD-009 + SDD-010 (vault tools must be registered), SDD-011 (wiring complete)
- Patterns / Pattern: MCP check + fallback (same as SDD-006)

### ID Convention

- **MCP available** → hash-based ID from `vault.NewFDID()` (e.g., `FD-a3f2`, `FD-7bc1`)
- **MCP unavailable** → sequential ID from file scan (e.g., `FD-007`, `FD-008`)
- Both formats are valid — vault accepts any `FD-*` pattern
- This is intentional — hash-based IDs are collision-proof and don't require scanning existing files

## Best Practices

- Error handling: if `forgia_fd_create` fails, fall back to direct Write — don't fail the command
- Naming: keep existing command structure, add MCP conditional wrapping
- Style: "Step 0.5: Check MCP" pattern from SDD-006

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Manual | `/fd-new` with MCP available → creates FD with hash-based ID | MCP path |
| Manual | `/fd-new` without MCP → creates FD with sequential ID | Fallback path |
| Manual | `/fd-sdd` with MCP available → creates SDDs via vault tools | MCP path |
| Manual | `/fd-sdd` without MCP → creates SDDs directly | Fallback path |
| Manual | Hash-based FD ID accepted by `/fd-review`, `/fd-sdd`, `/fd-verify` | ID compatibility |

## Acceptance Criteria / Criteri di Accettazione

- [x] `/fd-new` checks MCP reachability and calls `forgia_fd_create` when available
- [x] `/fd-new` with MCP produces hash-based ID (e.g., `FD-a3f2`) from `vault.NewFDID()`
- [x] `/fd-new` without MCP falls back to sequential ID (e.g., `FD-007`) — existing behavior
- [x] `/fd-sdd` checks MCP reachability and calls `forgia_sdd_create` when available
- [x] `/fd-sdd` without MCP falls back to direct file creation — existing behavior
- [x] Hash-based IDs work across all downstream commands (`/fd-review`, `/fd-sdd`, `/fd-verify`, `/fd-close`)
- [x] MCP reachability checked once per command invocation (cached)

## Context / Contesto

- [x] `modules/claude-commands/fd-new.md` — updated with MCP check + vault tool path
- [x] `modules/claude-commands/fd-sdd.md` — updated with MCP check + vault tool path
- [x] `internal/vault/vault.go` — `NewFDID()` implementation (hash-based) — read, no changes needed
- [x] `.forgia/sdd/FD-006/SDD-010-vault-write-tools.md` — `forgia_fd_create` tool spec — read for reference
- [x] `.forgia/sdd/FD-006/SDD-006-slash-command-mcp.md` — MCP check pattern reference — read for reference

## Constitution Check

- [x] Respects code standards — slash commands follow existing conventions
- [x] Respects commit conventions — `feat(FD-006): description`
- [x] No hardcoded secrets — no credential handling
- [x] Tests defined and sufficient — manual tests for MCP + fallback + ID compatibility

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. Used "Step 2.5" naming convention (consistent with SDD-006 pattern which uses "Step 0.5") — placed MCP check after pre-flight checks but before source determination, since MCP availability doesn't affect argument parsing or issue fetching.
2. `/fd-new` probes `forgia_fd_list` (not `tools/list`) for MCP reachability — this tests the actual vault tools rather than just the transport layer. Same pattern: `/fd-sdd` probes `forgia_sdd_list`.
3. Error handling: if `forgia_fd_create` or `forgia_sdd_create` fails, fall back to direct Write — graceful degradation, never fail the command due to MCP errors.
4. Content preparation is MCP-agnostic — the agent builds the same frontmatter and body sections regardless. Only the file-write step differs (MCP tool vs Write tool).
5. Verified all 10 downstream commands use glob-based file matching (`FD-NNN*.md`) — hash-based IDs like `FD-a3f2` are fully compatible with no changes needed to downstream commands.
6. `/fd-sdd` passes `body` content assembled from all sections as parameters to `forgia_sdd_create` — the `scope`, `constraints`, `acceptance_criteria` map directly to tool parameters, while remaining SDD sections go in the body written by the tool.

### Output

- **Commit(s)**: b5af4db
- **PR**: <!-- link -->
- **Files created/modified**:
  - `modules/claude-commands/fd-new.md` — added Step 2.5 MCP check, conditional Step 6 ID generation, Step 7b MCP/fallback write path
  - `modules/claude-commands/fd-sdd.md` — added Step 2.5 MCP check, conditional "Write each SDD file" section with MCP/fallback paths
  - `.forgia/sdd/FD-006/SDD-012-fd-new-fd-sdd-mcp.md` — updated status, acceptance criteria, work log

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: The MCP check pattern from SDD-006 was well-established — consistent "Step N.5" convention, 3-second timeout, cache-once-per-invocation made it straightforward to apply to new commands. The `forgia_fd_create` and `forgia_sdd_create` tool interfaces were well-documented by SDD-010.
- **What didn't / Cosa non ha funzionato**: Nothing blocked — the scope was clean and well-defined.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: The SDD body content (all markdown sections beyond what maps to tool parameters) needs to be passed as a single `body` string to `forgia_fd_create`. Future SDDs could consider a richer parameter schema that maps individual sections (Problem, Solutions, Architecture) to separate parameters for better structured creation.
