---
id: "SDD-008"
fd: "FD-001"
title: "forgia skill — skill registry + embedded command loader"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, cli, skill, embed, cobra]
---

# SDD-008: forgia skill — skill registry + embedded command loader

> Parent FD: [[FD-001]]

## Scope

Create `cmd/forgia/cmd/skill.go` and wire `internal/skill/` to the embedded slash commands from SDD-001.

### `forgia skill <name> [args]`
Execute an embedded slash command via Claude CLI:
1. Look up command by name in `embedded.SlashCommandFS()`
2. Read the markdown content
3. Build system context (constitution + guardrails + dev-guide) using `runner.BuildSystemContext()` from SDD-005
4. Replace `$ARGUMENTS` placeholder in the markdown with the provided args
5. Invoke Claude CLI with the command as task prompt + system context
6. Stream output to terminal

### `forgia skills`
List all available slash commands with descriptions (first line of each markdown file).

### Skill Registry (`internal/skill/registry.go`)
Wire the existing `Skill` interface to embedded commands:
- `EmbeddedSkill` struct: loads from `embed.FS`, implements `Skill` interface
- `Registry.LoadEmbedded(fs.FS)` — populate registry from embedded filesystem
- `Registry.Get(name)` — lookup by name
- `Registry.List()` — all registered skills

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `skillCmd` | `*cobra.Command` | `forgia skill <name> [args]` |
| `skillsCmd` | `*cobra.Command` | `forgia skills` (list) |
| `EmbeddedSkill` | `struct` implementing `Skill` | Wraps a markdown file |
| `Registry.LoadEmbedded(fs)` | `func(fs.FS) error` | Populates registry |
| `Registry.Get(name)` | `func(string) (Skill, error)` | Lookup |
| `Registry.List()` | `func() []Skill` | All skills |
| `embedded.SlashCommandFS()` | from SDD-001 | Source of commands |
| `runner.BuildSystemContext()` | from SDD-005 | Context builder |

## Constraints / Vincoli

- Language / Linguaggio: Go 1.25+
- Framework: Cobra, `io/fs`
- Dependencies / Dipendenze: `internal/embedded` (SDD-001), `internal/skill`, `internal/runner` (SDD-005), `internal/vault`
- Patterns / Pattern:
  - Skills are invoked via Claude CLI (same as current slash commands)
  - `$ARGUMENTS` placeholder in markdown is replaced before invocation
  - Some skills may become native Go in the future (fd-status, sdd-status) — the registry should support both embedded and native skills
- Read-only: `forgia skills` never modifies anything
- `forgia skill` may modify vault files (e.g., `/fd-new` creates files) — this is expected

## Best Practices

- Error handling: unknown skill name → clear error with list of available skills
- Naming: `EmbeddedSkill`, `NativeSkill` (future), `Registry`
- Style: `forgia skills` output in table format (name, first-line description)

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | Registry.LoadEmbedded loads all commands from test FS | loading |
| Unit | Registry.Get returns correct skill | lookup |
| Unit | Registry.Get returns error for unknown name | error path |
| Unit | Registry.List returns all skills | list |
| Unit | EmbeddedSkill.Content() returns markdown with args substituted | substitution |
| Unit | skillsCmd lists all skills | CLI integration |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `forgia skills` lists all embedded slash commands
- [ ] `forgia skill fd-status` executes the fd-status command via Claude
- [ ] `forgia skill fd-new "some description"` passes args correctly
- [ ] Unknown skill name returns helpful error with available skills list
- [ ] System context (constitution + guardrails) automatically included
- [ ] `EmbeddedSkill` implements existing `Skill` interface
- [ ] `Registry` populated from `embedded.SlashCommandFS()`
- [ ] `go build ./...` succeeds
- [ ] 6+ tests pass

## Context / Contesto

- [ ] `internal/skill/skill.go` — existing Skill interface and Registry
- [ ] `modules/claude-commands/` — current markdown slash commands
- [ ] `internal/embedded/` — from SDD-001
- [ ] `internal/runner/claude.go` — from SDD-005 (BuildSystemContext)
- [ ] Current usage: commands are installed to `~/.claude/commands/` — this replaces that with embedded execution

## Constitution Check

- [x] Respects code standards
- [x] Respects commit conventions
- [x] No hardcoded secrets
- [x] Tests defined and sufficient

---

## Work Log / Diario di Lavoro

### Agent / Agente

- **Executor**: <!-- openhands | claude-code | manual | name -->
- **Started**: <!-- timestamp -->
- **Completed**: <!-- timestamp -->
- **Duration / Durata**: <!-- total time -->

### Decisions / Decisioni

1. <!-- decision 1: what and why -->

### Output

- **Commit(s)**: <!-- hash -->
- **PR**: <!-- link -->
- **Files created/modified**:
  - `path/to/file`

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**:
- **What didn't / Cosa non ha funzionato**:
- **Suggestions for future FDs / Suggerimenti per FD futuri**:
