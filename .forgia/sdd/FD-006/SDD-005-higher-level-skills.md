---
id: "SDD-005"
fd: "FD-006"
title: "Higher-level composite skills — security_scan, arch_coherence, context_map"
status: done
agent: ""
assigned_to: ""
created: "2026-03-29"
started: ""
completed: ""
tags: ["enhancement", "phase:2-core"]
---

# SDD-005: Higher-level composite skills — security_scan, arch_coherence, context_map

> Parent FD: [[FD-006]]

## Scope

Implement 3 higher-level composite skills that compose primitive skills (SDD-004) with vault data reads.

### 1. `forgia_security_scan`

- **Composes**: `forgia_search_code` (primitive) + vault guardrails
- **PreProcess**: build search queries for security patterns (auth, validation, crypto, secrets handling, input sanitization)
- **PostProcess**: cross-reference findings with `deny.toml` patterns — identify gaps where code handles secrets but deny.toml doesn't protect the files. Return file paths and pattern names ONLY, never actual secret values.

### 2. `forgia_arch_coherence`

- **Composes**: `forgia_trace_calls` + `forgia_arch_init` (primitives) + vault architecture
- **PreProcess**: load documented architecture from `.forgia/architecture/` via `VaultReader.GetArchitecture()`
- **PostProcess**: compare actual call paths (from trace_calls) against documented component dependencies (from architecture). Flag drift: components that call each other but aren't documented as connected.

### 3. `forgia_context_map`

- **Composes**: `forgia_search_code` (primitive) + vault contexts
- **PreProcess**: load bounded contexts from `.forgia/contexts/` via `VaultReader.ListContexts()`
- **PostProcess**: map code symbols/changes to their bounded context. For each change, identify which context it belongs to and whether it crosses context boundaries.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `forgia_security_scan` | CompositeSkill | Security pattern search + guardrails gap analysis |
| `forgia_arch_coherence` | CompositeSkill | Call path vs documented architecture drift detection |
| `forgia_context_map` | CompositeSkill | Symbol-to-context mapping |
| `VaultReader` | dependency | Reads architecture, contexts, guardrails |
| `ProviderRegistry` | dependency | Calls primitive skills via `registry.Call()` |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: `internal/skill`, `internal/mcp`, `internal/vault`, `internal/guardrails`
- Dependencies / Dipendenze: SDD-003 (VaultReader), SDD-004 (primitive skills must be registered)
- Patterns / Pattern: Higher-level skills call primitives via `registry.Call()`, not directly

### Security (from threat model)

- `forgia_security_scan` must NOT return actual secret values found in code — only file paths and pattern names. Apply the same approach as `guardrails.ScanForSecrets()` (report location without content) (source: threat model)

## Best Practices

- Error handling: if a primitive skill call fails, return partial results with error annotation — don't fail entirely
- Error handling: if vault reads fail (architecture or contexts not initialized), skip enrichment and return raw results
- Naming: `security_scan`, `arch_coherence`, `context_map` — clear purpose
- Style: each higher-level skill is a separate function in `composite.go` for readability

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `forgia_security_scan` returns paths only, no secret content | Security |
| Unit | `forgia_arch_coherence` detects drift between actual and documented architecture | Drift detection |
| Unit | `forgia_context_map` maps symbols to correct bounded contexts | Mapping accuracy |
| Unit | Each higher-level skill degrades gracefully when primitives fail | Graceful degradation |
| Unit | Each higher-level skill degrades when vault data missing | Vault-less fallback |

## Acceptance Criteria / Criteri di Accettazione

- [x] `forgia_security_scan` searches for auth/validation/crypto patterns and cross-references with deny.toml gaps
- [x] `forgia_security_scan` returns file locations only, never actual secret content
- [x] `forgia_arch_coherence` compares actual call paths against documented `.forgia/architecture/`
- [x] `forgia_arch_coherence` flags undocumented component dependencies (drift)
- [x] `forgia_context_map` maps code changes to bounded contexts from `.forgia/contexts/`
- [x] All 3 skills call primitives via `registry.Call()`, not direct provider access
- [x] All 3 skills degrade gracefully when vault data or primitives unavailable
- [x] All 3 skills registered in `RegisterCompositeSkills()` from SDD-003
- [x] All new functions have unit tests with `t.Parallel()`

## Context / Contesto

- [x] `internal/skill/composite.go` — registration factory (SDD-003), primitive skill implementations (SDD-004)
- [x] `internal/guardrails/guardrails.go` — `ScanForSecrets()` pattern (lines 256-280) as reference for security_scan
- [x] `internal/vault/architecture.go` — `Architecture` type for arch_coherence
- [x] `internal/vault/vault.go` — `BoundedContext` for context_map
- [x] `.forgia/fd/FD-006-threat-model.md` — security_scan output constraint

## Constitution Check

- [x] Respects code standards — Go conventions, error wrapping, graceful degradation
- [x] Respects commit conventions — `feat(FD-006): description`
- [x] No hardcoded secrets — security_scan reports locations only
- [x] Tests defined and sufficient — security + drift + mapping + degradation

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~30 min

### Decisions / Decisioni

1. Higher-level skills use the same `CompositeSkill` struct as primitives, with `graceful: true` enabling error annotation instead of propagation when provider calls fail. This avoids adding a new type while supporting partial-result degradation.
2. Each higher-level skill is a constructor function (`securityScanSkill`, `archCoherenceSkill`, `contextMapSkill`) that returns a `*CompositeSkill` — keeps `RegisterCompositeSkills` clean and each skill self-contained.
3. Skills call provider tools via `registry.Call()` (through the `CompositeSkill.Execute` flow) and use `PostProcess` closures to compose with vault data. This keeps the architecture consistent with primitives.
4. `security_scan` PostProcess builds a sanitized output with only `path` and `name` fields — any `content` or value fields from the provider are explicitly excluded (threat model constraint).
5. `arch_coherence` uses `findContextForPath` (existing helper) to map call path files to bounded contexts, then checks against `buildDocumentedDeps` for drift. Requires both architecture and contexts to be available.
6. `context_map` detects boundary crossings by checking adjacent results for context changes — simple and effective for the search_graph result format.
7. Replaced 3 SDD-005 placeholder skills (impact-analysis, arch-validate, design-assist) with the specified implementations. Categories shifted: Design went from 3 to 1, Knowledge from 2 to 4 (total still 7).

### Output

- **Commit(s)**: <!-- hash — to be filled after commit -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `internal/skill/skill.go` — added `graceful` field to CompositeSkill, updated Execute for graceful degradation
  - `internal/skill/composite.go` — replaced 3 placeholder skills with security_scan, arch_coherence, context_map; added 8 new functions + securityPatternQueries variable
  - `internal/skill/composite_test.go` — updated registration/category tests, added 8 new test functions covering security, drift detection, mapping, graceful degradation, and vault-less fallback

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: The existing `CompositeSkill` pattern (PreProcess/PostProcess hooks) was flexible enough to implement all three higher-level skills without structural changes beyond the `graceful` flag. Reusing `findContextForPath` across both `arch_coherence` and `context_map` reduced duplication.
- **What didn't / Cosa non ha funzionato**: Initial test for `context_map` crossings expected 1 crossing but there were 2 (bidirectional context switches). Minor oversight, quickly fixed.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: The `graceful` degradation mechanism could be formalized as a skill "tier" (primitive vs higher-level) instead of a boolean flag. Also, `buildDocumentedDeps` currently treats DependsOn and DependedBy symmetrically — consider whether directional semantics matter for drift detection accuracy.
