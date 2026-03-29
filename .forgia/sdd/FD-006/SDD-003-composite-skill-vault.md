---
id: "SDD-003"
fd: "FD-006"
title: "CompositeSkill vault access — VaultReader interface + registration factory"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-29"
started: "2026-03-29"
completed: "2026-03-29"
tags: ["enhancement", "phase:2-core"]
---

# SDD-003: CompositeSkill vault access — VaultReader interface + registration factory

> Parent FD: [[FD-006]]

## Scope

Enable composite skills to read vault data (guardrails, architecture, contexts) and create the registration factory.

### 1. `VaultReader` interface in `internal/skill/skill.go`

Add a narrow interface with only the methods composite skills need:

```go
type VaultReader interface {
    Constitution(ctx context.Context) (string, error)
    GuardrailsRaw(ctx context.Context) ([]byte, error)
    GetArchitecture(ctx context.Context) (*vault.Architecture, error)
    ListContexts(ctx context.Context) ([]*vault.BoundedContext, error)
}
```

Add `vault VaultReader` field to `CompositeSkill` struct. Update `Execute()` to pass vault to PreProcess/PostProcess hooks if needed.

### 2. `internal/skill/composite.go` — Registration factory

```go
func RegisterCompositeSkills(reg *Registry, providerReg *mcp.ProviderRegistry, vault VaultReader) error
```

This function instantiates all 7 composite skills (SDD-004 primitives + SDD-005 higher-level) and registers them in both the skill registry and as providers in the MCP registry. Called from `forgia mcp serve` startup (SDD-008).

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `VaultReader` | interface | 4 methods: Constitution, GuardrailsRaw, GetArchitecture, ListContexts |
| `CompositeSkill.vault` | field | VaultReader instance, available to PreProcess/PostProcess hooks |
| `RegisterCompositeSkills()` | function | Factory that creates and registers all 7 composite skills |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: `internal/skill`, `internal/mcp`, `internal/vault`
- Dependencies / Dipendenze: `vault.Vault` satisfies `VaultReader` (no adapter needed)
- Patterns / Pattern: Interface Segregation — narrow `VaultReader` over full 17-method `Vault`

### Security (from threat model)

- `VaultReader` is intentionally narrow (4 methods) — if widened later, re-assess what data composite skills can access (source: threat model)

## Best Practices

- Error handling: `RegisterCompositeSkills` should fail fast if providerReg is nil (mandatory dependency)
- Naming: `VaultReader` — clear purpose, defined at consumer (skill package), not producer (vault package)
- Style: `vault.Vault` already implements `VaultReader` — no adapter code needed

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `VaultReader` satisfied by `*FileVault` (compile-time check) | Interface compliance |
| Unit | `RegisterCompositeSkills` registers 7 skills | Factory completeness |
| Unit | `CompositeSkill.Execute()` with vault-aware PreProcess | Vault access |

## Acceptance Criteria / Criteri di Accettazione

- [x] `VaultReader` interface defined in `internal/skill/skill.go` with 4 methods
- [x] `CompositeSkill` struct has `vault VaultReader` field
- [x] `vault.FileVault` satisfies `VaultReader` (compile-time `var _ VaultReader = (*vault.FileVault)(nil)`)
- [x] `RegisterCompositeSkills()` exists in `internal/skill/composite.go`
- [x] Factory registers all 7 composite skills (stubs OK — SDD-004/005 fill implementations)
- [x] All new functions have unit tests with `t.Parallel()`

## Context / Contesto

- [ ] `internal/skill/skill.go` — `CompositeSkill` struct to modify, `Skill` interface, `Registry`
- [ ] `internal/vault/vault.go` — `Vault` interface (VaultReader is a subset of this)
- [ ] `internal/vault/file_vault.go` — `FileVault` (concrete implementation)
- [ ] `internal/mcp/mcp.go` — `ProviderRegistry` for skill registration

## Constitution Check

- [ ] Respects code standards — Go conventions, interfaces at consumer
- [ ] Respects commit conventions — `feat(FD-006): description`
- [ ] No hardcoded secrets — no credential handling
- [ ] Tests defined and sufficient — compile-time check + unit tests

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. VaultReader interface defined in `skill.go` alongside CompositeSkill (consumer-side, per Interface Segregation) — 4 methods matching the Vault interface subset that composite skills need.
2. Compile-time check (`var _ VaultReader = (*vault.FileVault)(nil)`) placed in `composite.go` alongside the factory, since that's where vault wiring happens.
3. 7 stub composite skills split into 4 primitives (SDD-004) + 3 higher-level (SDD-005), with placeholder provider tools. PreProcess/PostProcess hooks left nil — SDD-004/005 will fill implementations.
4. `RegisterCompositeSkills` fails fast on nil `providerReg` (mandatory dependency) but accepts nil vault (some skills may not need it during early wiring).

### Output

- **Commit(s)**: f597873
- **PR**: (pending)
- **Files created/modified**:
  - `internal/skill/skill.go` — added VaultReader interface, vault field on CompositeSkill
  - `internal/skill/composite.go` — new file: compile-time check + RegisterCompositeSkills factory
  - `internal/skill/composite_test.go` — new file: 6 unit tests with t.Parallel()

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: Clean separation — VaultReader is 4 methods, no adapter needed since FileVault already implements them. Compile-time check caught compliance immediately.
- **What didn't / Cosa non ha funzionato**: Nulla di rilevante — straightforward implementation.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: SDD-004/005 should define the PreProcess/PostProcess hook signatures explicitly to avoid ambiguity about what vault data each skill injects.
