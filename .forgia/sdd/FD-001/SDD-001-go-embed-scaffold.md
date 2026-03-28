---
id: "SDD-001"
fd: "FD-001"
title: "go:embed scaffold — vault templates + slash commands"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, embed, templates]
---

# SDD-001: go:embed scaffold — vault templates + slash commands

> Parent FD: [[FD-001]]

## Scope

Create `internal/embedded/` package that uses `go:embed` to bundle:
1. All vault template files from `modules/vault-template/` (config.toml, constitution.md, fd templates, sdd templates, dev-guide, guardrails, ops)
2. All slash command markdown files from `modules/claude-commands/` (fd-new, fd-review, fd-sdd, fd-close, fd-explore, fd-verify, fd-status, fd-deep, sdd-status, sdd-assign, project-init)

The package must export two functions:
- `VaultTemplateFS() fs.FS` — returns the embedded vault template filesystem
- `SlashCommandFS() fs.FS` — returns the embedded slash command filesystem

This eliminates the runtime dependency on the `modules/` directory. The Go binary becomes fully self-contained.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `VaultTemplateFS()` | `func() fs.FS` | Returns embedded vault templates |
| `SlashCommandFS()` | `func() fs.FS` | Returns embedded slash commands |
| `ListSlashCommands()` | `func() []string` | Returns list of available command names |
| `ReadSlashCommand(name)` | `func(string) ([]byte, error)` | Returns content of a specific command |

## Constraints / Vincoli

- Language / Linguaggio: Go 1.25+
- Framework: stdlib `embed`, `io/fs`
- Dependencies / Dipendenze: none (stdlib only)
- Patterns / Pattern: `//go:embed` directives, `fs.FS` interface
- The `go:embed` directives MUST be in the same package as the embedded files or use relative paths from the Go file location. If `internal/embedded/` is the package, the `modules/` dir must be reachable. Alternative: place the `go:embed` file at repo root and re-export.

## Best Practices

- Error handling: `ReadSlashCommand` returns error if command not found
- Naming: package name `embedded`, functions are self-documenting
- Style: no exported types, just functions returning `fs.FS`

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | VaultTemplateFS contains expected files (config.toml, constitution.md, fd template, sdd template) | 100% of expected files |
| Unit | SlashCommandFS contains all 11+ commands | all known commands |
| Unit | ListSlashCommands returns correct names | all names |
| Unit | ReadSlashCommand returns content for valid name, error for invalid | both paths |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `internal/embedded/` package exists with `go:embed` directives
- [ ] `VaultTemplateFS()` returns fs.FS containing all files from `modules/vault-template/`
- [ ] `SlashCommandFS()` returns fs.FS containing all files from `modules/claude-commands/`
- [ ] `ListSlashCommands()` returns names of all embedded commands (without .md extension)
- [ ] `ReadSlashCommand("fd-review")` returns the markdown content
- [ ] `ReadSlashCommand("nonexistent")` returns error
- [ ] `go build ./...` succeeds — embedded files compile into binary
- [ ] Binary size increase is reasonable (< 500KB for templates + commands)
- [ ] All existing tests still pass

## Context / Contesto

- [ ] `modules/vault-template/` — directory tree to embed
- [ ] `modules/claude-commands/` — markdown files to embed
- [ ] Go `embed` docs: https://pkg.go.dev/embed
- [ ] `internal/skill/skill.go` — existing Skill interface (for future integration)

## Constitution Check

- [x] Respects code standards (stdlib only, no external deps)
- [x] Respects commit conventions (`feat(FD-001): SDD-001`)
- [x] No hardcoded secrets
- [x] Tests defined and sufficient

## Guardrails

From `.forgia/guardrails/deny.toml`: no deny patterns are relevant to this SDD. The embedded files are templates, not secrets.

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
