---
id: "SDD-002"
fd: "FD-006"
title: "Provider wiring — config → registry + config template"
status: done
agent: ""
assigned_to: ""
created: "2026-03-29"
started: ""
completed: ""
tags: ["enhancement", "phase:2-core"]
---

# SDD-002: Provider wiring — config → registry + config template

> Parent FD: [[FD-006]]

## Scope

Create the orchestrator that loads MCP provider configuration and registers providers in the registry. Two deliverables:

### 1. Provider orchestrator

Create a function (in `cmd/forgia/cmd/mcp.go` or a new `internal/mcp/wiring.go`) that:

1. Reads `config.MCP.Providers` map
2. For each provider config: creates `MCPSubprocessProvider`, registers in `ProviderRegistry`
3. Auto-registers codebase-memory-mcp from `[knowledge]` config when `provider` is set
4. Starts all providers (`provider.Start(ctx)`)
5. Returns the populated `ProviderRegistry`

```go
func WireProviders(ctx context.Context, cfg *config.Config) (*ProviderRegistry, error)
```

### 2. Config template update

Add `[mcp.providers.*]` section with commented examples to `modules/vault-template/config.toml`:

```toml
# [mcp.providers.code]
# command = "codebase-memory-mcp"
# namespace = "code"
# lazy = false
# restart_on_crash = true
```

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `WireProviders(ctx, cfg)` | function | Returns populated `*ProviderRegistry` with all providers started |
| `config.MCPConfig.Providers` | map | Config source — maps provider names to `MCPProviderConfig` |
| `config.KnowledgeConfig` | struct | Auto-register codebase-memory-mcp when configured |
| Config template | TOML | Commented `[mcp.providers.*]` examples |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: `internal/config`, `internal/mcp`
- Dependencies / Dipendenze: SDD-001 (ProviderRegistry must exist)
- Patterns / Pattern: Follow existing `LoadConfig` + `NewMCPSubprocessProvider` patterns

### Security (from threat model)

- Validate provider command paths before spawning — reject commands containing shell metacharacters (`; | & $ \``) (source: threat model)
- Use `exec.LookPath` to resolve command to absolute path before passing to `NewMCPSubprocessProvider` (source: threat model)
- Log resolved command path at startup via `slog.InfoContext` (source: threat model)
- Add shell command patterns to config template comments as security advisory: `# WARNING: do not use shell commands (sh -c, bash -c) — use direct binary paths` (source: threat model)

## Best Practices

- Error handling: if one provider fails to start, log error but continue with others — don't fail the entire server
- Error handling: if `exec.LookPath` fails (command not found), skip provider with warning
- Naming: `WireProviders` — clear action name
- Style: return `*ProviderRegistry` so caller owns lifecycle

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `WireProviders` with mock config (2 providers) | Wiring logic |
| Unit | Command validation rejects shell metacharacters | Security |
| Unit | Missing command skipped gracefully | Graceful degradation |
| Unit | Knowledge config auto-registers codebase-memory-mcp | Auto-registration |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `WireProviders(ctx, cfg)` creates and starts providers from `config.MCP.Providers`
- [ ] Codebase-memory-mcp auto-registered from `[knowledge]` config
- [ ] Provider command paths validated — shell metacharacters rejected
- [ ] `exec.LookPath` resolves commands to absolute paths
- [ ] Resolved command path logged at startup
- [ ] Failed providers logged but don't prevent other providers from starting
- [ ] `[mcp.providers.*]` section added to config template with commented examples
- [ ] Config template includes security advisory about shell commands
- [ ] All new functions have unit tests with `t.Parallel()`

## Context / Contesto

- [ ] `internal/config/config.go` — `MCPConfig`, `MCPProviderConfig`, `KnowledgeConfig`
- [ ] `internal/mcp/subprocess.go` — `NewMCPSubprocessProvider(cfg)`
- [ ] `internal/mcp/mcp.go` — `ProviderRegistry.Register()`
- [ ] `modules/vault-template/config.toml` — template to update
- [ ] `.forgia/fd/FD-006-threat-model.md` — command injection mitigation

## Constitution Check

- [ ] Respects code standards — Go conventions, error wrapping
- [ ] Respects commit conventions — `feat(FD-006): description`
- [ ] No hardcoded secrets — no credential handling
- [ ] Tests defined and sufficient — unit tests for wiring + security validation

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. Placed `WireProviders` and `ValidateCommand` in `internal/mcp/wiring.go` — keeps wiring logic close to the types it operates on (`ProviderRegistry`, `MCPSubprocessProvider`) without polluting the command layer.
2. `ValidateCommand` is exported — future SDDs (e.g., SDD-008 integration wiring) can reuse command validation without duplicating it.
3. Knowledge auto-registration uses `"code"` as the default namespace and only registers if no explicit `"code"` provider already exists — prevents accidental override of a user-configured provider.
4. Tests use `lazy: true` mode to avoid spawning real MCP subprocesses — validates wiring logic (validation, registration, namespace defaults) without external dependencies.
5. `maps.Copy` modernize suggestion from linter intentionally not applied — the loop also handles the knowledge auto-registration merge logic, and the explicit loop is clearer for this two-source merge pattern.

### Output

- **Commit(s)**: (pending)
- **PR**: —
- **Files created/modified**:
  - `internal/mcp/wiring.go` (new — `WireProviders`, `ValidateCommand`)
  - `internal/mcp/wiring_test.go` (new — 11 test cases)
  - `modules/vault-template/config.toml` (modified — `[mcp.providers.*]` section with security advisory)

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: SDD context section was accurate — all referenced types (`MCPSubprocessProvider`, `ProviderRegistry`, `config.MCPProviderConfig`, `KnowledgeConfig`) existed exactly as documented. Security constraints were clear and directly implementable.
- **What didn't / Cosa non ha funzionato**: Nothing — the SDD scope was well-bounded and all acceptance criteria were achievable.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: The `lazy: true` testing pattern works well for wiring tests. Future SDDs that test orchestration logic should document this pattern explicitly to avoid test flakiness from real subprocess spawning.
