---
id: "SDD-008"
fd: "FD-006"
title: "Integration wiring — startup, shutdown, skill listing, documentation"
status: done
agent: ""
assigned_to: ""
created: "2026-03-29"
started: ""
completed: ""
tags: ["enhancement", "phase:2-core"]
---

# SDD-008: Integration wiring — startup, shutdown, skill listing, documentation

> Parent FD: [[FD-006]]

## Scope

This is the **integration wiring SDD**. Wire all components together and verify the complete system works end-to-end.

### 1. Startup initialization path

Verify `forgia mcp serve` executes this sequence:
1. `vault.Open(".")` — open vault
2. `config.LoadConfig(ctx, vault.Dir())` — load config
3. `WireProviders(ctx, cfg)` — create + start providers (SDD-002)
4. `skill.NewRegistry()` → `LoadEmbedded()` — load slash commands
5. `skill.RegisterCompositeSkills(reg, providerReg, vault)` — register composite skills (SDD-003)
6. `mcp.NewMCPServer(transport, providerReg, skillReg)` — create server (SDD-001)
7. `server.Serve(ctx)` — block until signal or EOF
8. `provider.Stop()` for all providers — graceful shutdown

### 2. Graceful shutdown

Verify:
- SIGINT triggers shutdown
- SIGTERM triggers shutdown
- All providers stopped before process exits
- No goroutine leaks after shutdown

### 3. `forgia skills` lists composite skills

Verify that `forgia skills` output includes all 7 composite skills alongside the existing 19 embedded slash commands. Composite skills should show their `Mode: MCP Tool` designation.

### 4. Documentation update

Update `.forgia/dev-guide/review-process.md` to note that MCP server enables enhanced analysis in slash commands. Update README if needed.

### 5. E2E verification

The user calls `forgia mcp serve` → agent calls `tools/list` → sees all tools → calls `forgia_blast_radius` → gets enriched result. Full chain from CLI entry point to knowledge graph and back.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| Startup path | sequential | vault → config → providers → skills → server.Serve() |
| Shutdown | signals | SIGINT/SIGTERM → provider.Stop() → clean exit |
| `forgia skills` | CLI | Lists all 26 skills (19 embedded + 7 composite) |
| E2E chain | end-to-end | CLI → server → registry → provider → codebase-memory-mcp → result |

## Constraints / Vincoli

- Language / Linguaggio: Go + Markdown
- Framework: All previous SDDs (001-007) must be complete
- Dependencies / Dipendenze: Everything — this SDD verifies the integration of all components
- Patterns / Pattern: Integration wiring SDD per Forgia convention

## Best Practices

- Error handling: if any startup step fails, log the error and exit cleanly — don't leave orphan processes
- Naming: startup function should be readable top-to-bottom in `runMCPServe()`
- Style: the startup sequence should read like a recipe — each step clearly depends on the previous

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| E2E | `forgia mcp serve` starts, responds to tools/list, shuts down on signal | Full lifecycle |
| E2E | `forgia skills` lists composite skills | Skill visibility |
| Manual | Send tools/call for a composite skill → get result | Full chain |
| Manual | SIGINT during serve → clean shutdown, no goroutine leaks | Shutdown |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `forgia mcp serve` executes the full startup path: vault → config → providers → skills → serve
- [ ] SIGINT/SIGTERM triggers graceful shutdown — all providers stopped
- [ ] `forgia skills` lists all 7 composite skills alongside 19 embedded skills
- [ ] Full E2E chain works: CLI → server → composite skill → provider → knowledge graph → enriched result
- [ ] `review-process.md` updated to note MCP server enables enhanced analysis
- [ ] No goroutine leaks after shutdown
- [ ] `go build ./cmd/forgia/` succeeds
- [ ] `go test ./...` passes

## Context / Contesto

- [ ] `cmd/forgia/cmd/mcp.go` — serve command (SDD-001)
- [ ] `internal/mcp/server.go` — MCP server (SDD-001)
- [ ] `internal/mcp/mcp.go` — ProviderRegistry
- [ ] `internal/skill/composite.go` — RegisterCompositeSkills (SDD-003)
- [ ] `internal/process/signals.go` — ShutdownManager
- [ ] `.forgia/dev-guide/review-process.md` — to update
- [ ] All SDD-001 through SDD-007 Work Logs — verify all completed

## Constitution Check

- [ ] Respects code standards — Go conventions
- [ ] Respects commit conventions — `feat(FD-006): description`
- [ ] No hardcoded secrets — no credential handling
- [ ] Tests defined and sufficient — E2E lifecycle + skill listing + shutdown

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~30 min

### Decisions / Decisioni

1. **Added `StopAll()` to `ProviderRegistry`** — iterates all providers and calls `Stop()` with best-effort error logging. Used by `runMCPServe()` via `defer` for graceful shutdown.
2. **Added `MetadataProvider`** — lightweight `ToolProvider` implementation that only exposes metadata (name, namespace, healthy=true). Used by `forgia skills` to register composite skills without spawning actual provider subprocesses.
3. **Introduced `loadFullRegistry()` in skill.go** — loads embedded skills + composite skills (when vault + config available). `forgia skills` uses this to list all 26 skills with Mode column.
4. **Updated E2E MCP tests to create minimal vault** — `forgia mcp serve` now requires vault → config → providers → skills wiring. E2E tests that spawn the binary create `.forgia/` directory so startup succeeds.
5. **Updated `tools/list` E2E test** — expects 7 composite skill tools (registered even without external providers) instead of 0.

### Output

- **Commit(s)**: <!-- hash filled after commit -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `cmd/forgia/cmd/mcp.go` — full startup path: vault → config → WireProviders → LoadEmbedded → RegisterCompositeSkills → Serve; graceful shutdown via defer StopAll()
  - `cmd/forgia/cmd/skill.go` — `loadFullRegistry()`, Mode column in `forgia skills` output
  - `internal/mcp/mcp.go` — `StopAll()`, `MetadataProvider` type
  - `internal/mcp/wiring_integration_test.go` — E2E lifecycle, shutdown, MetadataProvider tests (SDD-008)
  - `internal/skill/integration_test.go` — `TestIntegration_SkillListing_EmbeddedAndComposite` (19+7=26 skills)
  - `tests/e2e_go_test.go` — `setupMinimalVault()` helper, `TestE2E_Skills_WithVault`
  - `tests/integration_mcp_test.go` — updated to create minimal vault, updated tools/list expectation
  - `.forgia/dev-guide/review-process.md` — MCP server enhanced analysis section

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: All prior SDDs (001-007) were well-structured — wiring was straightforward. The `ProviderRegistry` and `skill.Registry` designs made integration clean. Test infrastructure (mockProvider, mockVaultReader, server test helpers) was reusable.
- **What didn't / Cosa non ha funzionato**: E2E subprocess tests needed updating because the startup path change (vault-required) broke tests that ran in bare temp dirs. This was expected but required careful attention to all test sites.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Integration wiring SDDs should explicitly list which E2E tests will break due to interface changes. A "migration impact" section would help agents plan test updates upfront.
