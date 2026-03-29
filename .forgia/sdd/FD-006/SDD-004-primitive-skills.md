---
id: "SDD-004"
fd: "FD-006"
title: "Primitive composite skills — arch_init, blast_radius, trace_calls, search_code"
status: done
agent: ""
assigned_to: ""
created: "2026-03-29"
started: ""
completed: ""
tags: ["enhancement", "phase:2-core"]
---

# SDD-004: Primitive composite skills — arch_init, blast_radius, trace_calls, search_code

> Parent FD: [[FD-006]]

## Scope

Implement 4 primitive composite skills that each wrap a single `codebase-memory-mcp` tool with Forgia-specific PreProcess/PostProcess logic. All skills are registered in `internal/skill/composite.go` via `RegisterCompositeSkills()` (from SDD-003).

### 1. `forgia_arch_init`

- **Wraps**: `get_architecture()` from codebase-memory-mcp
- **PreProcess**: pass-through (no modification needed)
- **PostProcess**: map knowledge graph architecture output to `.forgia/architecture/` YAML structure (system-context, containers, technology-decisions)
- **Note**: this is the ONLY composite skill that writes files — merge with existing YAML, don't overwrite

### 2. `forgia_blast_radius`

- **Wraps**: `detect_changes()` from codebase-memory-mcp
- **PreProcess**: pass-through
- **PostProcess**: add risk labels (High/Medium/Low based on number of dependents), annotate with bounded context names from `.forgia/contexts/`

### 3. `forgia_trace_calls`

- **Wraps**: `trace_call_path()` from codebase-memory-mcp
- **PreProcess**: pass-through
- **PostProcess**: annotate call path nodes with bounded context from `.forgia/contexts/`

### 4. `forgia_search_code`

- **Wraps**: `search_graph()` from codebase-memory-mcp
- **PreProcess**: pass-through
- **PostProcess**: filter results through `guardrails.CheckFilePaths()` — strip any file paths matching `[read]` deny patterns from results

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| Each skill | `CompositeSkill` | Registered with name, ProviderName="code", ProviderTool=<tool> |
| `forgia_search_code` PostProcess | function | Calls `guardrails.Parse()` + `CheckFilePaths()` on result paths |
| `forgia_arch_init` PostProcess | function | Writes YAML to `.forgia/architecture/` (merge, not overwrite) |
| `forgia_blast_radius` / `forgia_trace_calls` PostProcess | function | Reads `.forgia/contexts/` via `VaultReader.ListContexts()` |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: `internal/skill`, `internal/mcp`, `internal/guardrails`, `internal/vault`
- Dependencies / Dipendenze: SDD-003 (VaultReader, RegisterCompositeSkills factory)
- Patterns / Pattern: `CompositeSkill` with PreProcess/PostProcess hooks

### Security (from threat model)

- `forgia_search_code` PostProcess MUST apply `guardrails.CheckFilePaths()` to filter results matching `[read]` deny patterns — test with fixture containing `.env` and `*.pem` paths, verify they are stripped (source: threat model)
- `forgia_arch_init` must merge with existing architecture files, not overwrite — check file existence before writing (source: threat model)

## Best Practices

- Error handling: if codebase-memory-mcp tool returns error, propagate as-is — don't mask provider errors
- Error handling: guardrails parsing failure in search_code → log warning, return unfiltered results (fail-open for usability, but log the gap)
- Naming: skill names match FD spec exactly: `arch_init`, `blast_radius`, `trace_calls`, `search_code`
- Style: each skill's hooks are defined as closures in `RegisterCompositeSkills()` for locality

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `forgia_search_code` PostProcess strips deny.toml-matched paths | Guardrails filtering |
| Unit | `forgia_search_code` PostProcess with fixture containing `.env`, `*.pem` | Specific deny patterns |
| Unit | `forgia_arch_init` PostProcess merges, not overwrites | Merge safety |
| Unit | `forgia_blast_radius` PostProcess adds context annotations | Context enrichment |
| Unit | `forgia_trace_calls` PostProcess adds context annotations | Context enrichment |
| Unit | Each skill calls correct provider tool name | Routing |

## Acceptance Criteria / Criteri di Accettazione

- [x] `forgia_arch_init` wraps `get_architecture()` with YAML mapping PostProcess
- [x] `forgia_blast_radius` wraps `detect_changes()` with risk labels + context annotations
- [x] `forgia_trace_calls` wraps `trace_call_path()` with context annotations
- [x] `forgia_search_code` wraps `search_graph()` with guardrails filtering PostProcess
- [x] `forgia_search_code` strips `.env`, `*.pem`, and other deny.toml-matched paths from results
- [x] `forgia_arch_init` merges with existing architecture YAML, does not overwrite
- [x] All 4 skills registered in `RegisterCompositeSkills()` from SDD-003
- [x] All new functions have unit tests with `t.Parallel()`

## Context / Contesto

- [ ] `internal/skill/composite.go` — registration factory (from SDD-003)
- [ ] `internal/skill/skill.go` — `CompositeSkill` struct, `VaultReader`
- [ ] `internal/guardrails/guardrails.go` — `Parse()`, `CheckFilePaths()` for search_code filtering
- [ ] `internal/vault/architecture.go` — `Architecture` type for arch_init mapping
- [ ] `internal/vault/vault.go` — `BoundedContext` type for context annotations
- [ ] `.forgia/guardrails/deny.toml` — deny patterns for search_code test fixtures
- [ ] `.forgia/fd/FD-006-threat-model.md` — guardrails filtering requirement

## Constitution Check

- [x] Respects code standards — Go conventions, error wrapping
- [x] Respects commit conventions — `feat(FD-006): description`
- [x] No hardcoded secrets — no credential handling
- [x] Tests defined and sufficient — guardrails filtering + merge safety + routing

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~20 min

### Decisions / Decisioni

1. **Skill names use underscores**: `arch_init`, `blast_radius`, `trace_calls`, `search_code` — as specified in FD spec ("skill names match FD spec exactly").
2. **arch_init writes via Dir() interface check**: Instead of adding `vaultDir` parameter to `RegisterCompositeSkills` (which would break SDD-003 signature), the closure checks if VaultReader implements `Dir() string` via interface assertion. `FileVault` already has `Dir()`, so it works in production; tests use `mockVaultReaderWithDir`.
3. **Risk label thresholds**: High ≥ 5 dependents, Medium ≥ 2, Low < 2 — reasonable defaults derived from typical blast-radius analysis.
4. **Context matching by responsibility substring**: `findContextForPath` uses `strings.Contains(path, responsibility)` to match files to bounded contexts. Simple but effective for directory-based context boundaries.
5. **Guardrails fail-open for search_code**: On parse failure, logs warning and returns unfiltered results (as specified in Best Practices). Tested with invalid TOML input.
6. **Preserved RegisterCompositeSkills signature**: No change to `func RegisterCompositeSkills(reg, providerReg, v)` — backward compatible with SDD-003.

### Output

- **Commit(s)**: c9dd96b
- **PR**: <!-- link -->
- **Files created/modified**:
  - `internal/skill/composite.go` — replaced 4 stub primitives with full implementations including PostProcess closures + helper functions
  - `internal/skill/composite_test.go` — updated existing tests for new skill names/categories + added 16 new tests covering all SDD-004 requirements

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: La struttura `CompositeSkill` con `PostProcess` closures è risultata molto flessibile. Il pattern di interface assertion per `Dir()` ha evitato di modificare la firma di `RegisterCompositeSkills`. I test con `mockProvider.callFn` rendono facile simulare qualsiasi output del provider MCP.
- **What didn't / Cosa non ha funzionato**: L'output del provider MCP (`get_architecture`, `detect_changes`, etc.) è `any` — il mapping da `map[string]any` a tipi strutturati richiede codice difensivo. Quando il provider reale sarà disponibile, potrebbe servire adattare i nomi dei campi.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Definire un contratto formale (schema JSON) per l'output di ciascun tool codebase-memory-mcp, così i PostProcess possono fare parsing tipizzato invece di asserzioni su `map[string]any`. Considerare anche un `VaultWriter` interface per skill che scrivono (attualmente solo `arch_init`).
