---
id: "SDD-006"
fd: "FD-006"
title: "Slash command MCP integration — 6 commands with MCP fallback"
status: done
agent: ""
assigned_to: ""
created: "2026-03-29"
started: ""
completed: ""
tags: ["enhancement", "phase:2-core"]
---

# SDD-006: Slash command MCP integration — 6 commands with MCP fallback

> Parent FD: [[FD-006]]

## Scope

Update 6 slash commands to check MCP reachability and prefer composite skills (knowledge graph) when available, with automatic fallback to direct scanning.

### MCP Reachability Pattern

Each command adds a "Step 0: Check MCP availability" at the beginning:

1. Attempt a `tools/list` call to the local MCP server with a 3-second timeout
2. If successful: set `MCP_AVAILABLE=true`, cache for the duration of this command invocation
3. If timeout or error: set `MCP_AVAILABLE=false`, proceed with existing direct scanning approach
4. Do NOT re-check MCP on every step — one check per command invocation

### Commands to update

| Slash Command | File | MCP path (when available) | Fallback (when unavailable) |
|---|---|---|---|
| `/fd-arch-review` | `fd-arch-review.md` | `forgia_trace_calls` + `forgia_search_code` for pattern/dependency analysis | Glob/Grep/Read (current behavior) |
| `/fd-threat-model` | `fd-threat-model.md` | `forgia_security_scan` for security pattern detection | Glob/Grep for auth/validation/crypto |
| `/arch-init` | `arch-init.md` | `forgia_arch_init` for knowledge graph architecture extraction | Directory scanning |
| `/arch-review` | `arch-review.md` | `forgia_arch_coherence` for call-path drift detection | Vault coherence check |
| `/arch-update` | `arch-update.md` | `forgia_arch_coherence` + `forgia_context_map` | Work Log reading |
| `/sdd-dry-run` | `sdd-dry-run.md` | `forgia_blast_radius` for impact analysis | Text analysis of SDD |

### Change pattern per command

For each command, insert after "Step 0: Resolve Input" and before the main analysis steps:

```markdown
### Step 0.5: Check MCP Availability

1. Attempt: call `forgia_code_search_graph` (or `tools/list`) with a 3-second timeout
2. If successful: note "Knowledge graph available — using MCP composite skills for enhanced analysis"
3. If failed/timeout: note "Knowledge graph not available — using direct codebase scanning (Glob/Grep/Read)"
4. Cache result — do not re-check on subsequent steps
```

Then wrap each analysis step with: "If MCP available, use `forgia_<skill>`. Otherwise, fall back to <existing approach>."

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| MCP reachability check | MCP call | `tools/list` with 3s timeout → boolean |
| 6 slash command files | markdown | Each updated with Step 0.5 + conditional MCP/fallback paths |
| Composite skills | MCP tools | Called when MCP available |
| Direct scanning | Glob/Grep/Read | Called when MCP unavailable (current behavior preserved) |

## Constraints / Vincoli

- Language / Linguaggio: Markdown (slash command prompt updates)
- Framework: Claude Code slash command system
- Dependencies / Dipendenze: SDD-004 + SDD-005 (composite skills must be registered)
- Patterns / Pattern: Conditional MCP check + fallback

### Security (from threat model)

- Cache MCP reachability result for the duration of a single slash command invocation — do not re-check per step (source: threat model)
- Document that MCP communication is local-only (stdio, same user) — no network exposure (source: threat model)

## Best Practices

- Error handling: MCP check failure → fallback silently, no error to user. Just note "using direct scanning"
- Naming: "Step 0.5" to avoid renumbering all existing steps
- Style: keep existing command structure intact — only add the MCP check and conditional wrappers

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Manual | Run `/fd-arch-review` with MCP available → verify uses composite skills | MCP path |
| Manual | Run `/fd-arch-review` without MCP → verify falls back to Glob/Grep/Read | Fallback path |
| Manual | Run each of the 6 commands without MCP → verify unchanged behavior | Regression |

## Acceptance Criteria / Criteri di Accettazione

- [x] `/fd-arch-review` uses `forgia_trace_calls` + `forgia_search_code` when MCP available
- [x] `/fd-threat-model` uses `forgia_security_scan` when MCP available
- [x] `/arch-init` uses `forgia_arch_init` when MCP available
- [x] `/arch-review` uses `forgia_arch_coherence` when MCP available
- [x] `/arch-update` uses `forgia_arch_coherence` + `forgia_context_map` when MCP available
- [x] `/sdd-dry-run` uses `forgia_blast_radius` when MCP available
- [x] All 6 commands work without MCP server (fallback to direct scanning)
- [x] MCP reachability checked once per command invocation (cached)

## Context / Contesto

- [x] `modules/claude-commands/fd-arch-review.md` — updated
- [x] `modules/claude-commands/fd-threat-model.md` — updated
- [x] `modules/claude-commands/arch-init.md` — updated
- [x] `modules/claude-commands/arch-review.md` — updated
- [x] `modules/claude-commands/arch-update.md` — updated
- [x] `modules/claude-commands/sdd-dry-run.md` — updated

## Constitution Check

- [x] Respects code standards — slash commands follow existing conventions
- [x] Respects commit conventions — `feat(FD-006): description`
- [x] No hardcoded secrets — no credential handling
- [x] Tests defined and sufficient — manual tests for MCP + fallback paths

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. Used "Step 0.5" naming convention as specified by SDD to avoid renumbering existing steps in users' mental model. However, for files using numbered list format (fd-threat-model, arch-init, arch-review, arch-update), the new step was inserted as a numbered item and subsequent steps were renumbered for consistency.
2. Each MCP conditional block uses consistent phrasing: "If MCP_AVAILABLE" for the MCP path and "If MCP_AVAILABLE=false (or fallback)" for the direct scanning path, making it easy to grep/search across all 6 files.
3. Added the security note "MCP communication is local-only (stdio, same user) — no network exposure" as a blockquote in each Step 0.5, per the threat model constraint.
4. Kept existing command structure intact — MCP conditionals wrap analysis steps, never replace the existing instructions which serve as the fallback path.

### Output

- **Commit(s)**: <!-- to be filled after commit -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `modules/claude-commands/fd-arch-review.md` — added Step 0.5, MCP conditional on Steps 2-3
  - `modules/claude-commands/fd-threat-model.md` — added step 2 (MCP check), MCP conditional on step 4 (security scan)
  - `modules/claude-commands/arch-init.md` — added step 3 (MCP check), MCP conditional on step 7 (analysis)
  - `modules/claude-commands/arch-review.md` — added step 2 (MCP check), MCP conditional on step 4 (coherence checks)
  - `modules/claude-commands/arch-update.md` — added step 3 (MCP check), MCP conditional on step 6 (analysis)
  - `modules/claude-commands/sdd-dry-run.md` — added Step 0.5, MCP conditional on Pass 3-4

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: The SDD was very precise about which composite skill maps to which command — the mapping table made implementation straightforward. The "Step 0.5" naming convention avoids confusing users who know the existing step numbers.
- **What didn't / Cosa non ha funzionato**: Files using numbered lists (1, 2, 3...) required renumbering when inserting the MCP check step, which creates a diff noise. The "Step 0.5" approach only works cleanly for files using named steps (### Step N).
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Consider using named steps (### Step N: Name) consistently across all slash commands instead of numbered lists — this makes insertions cleaner and avoids renumbering.
