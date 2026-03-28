---
id: "SDD-002"
fd: "FD-001"
title: "forgia init — Cobra command with embedded templates"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, cli, init, cobra]
---

# SDD-002: forgia init — Cobra command with embedded templates

> Parent FD: [[FD-001]]

## Scope

Create `cmd/forgia/cmd/init.go` — Cobra command that scaffolds `.forgia/` vault using embedded templates from SDD-001.

Behavior must match current Bash `cmd_init()` in `bin/forgia`:
1. Create `.forgia/` directory structure
2. Copy templates from `embedded.VaultTemplateFS()` (not from filesystem `modules/`)
3. Auto-detect project stack (Go, Rust, Python, Node, Shell) by checking for `go.mod`, `Cargo.toml`, `pyproject.toml`, `package.json`
4. Copy matching language conventions from embedded templates
5. Initialize Beads if `bd` is available (via `beads.Client`)
6. Initialize Knowledge layer if codebase-memory-mcp is available
7. Create architecture templates (DDD)
8. Create `.gitignore` for `.forgia/` (exclude logs, run, beads)
9. Create `.github/CODEOWNERS` if not exists
10. Print summary of created files and next steps

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `initCmd` | `*cobra.Command` | Registered as subcommand of rootCmd |
| `--dir` flag | `string` | Target directory (default: current dir) |
| `embedded.VaultTemplateFS()` | `fs.FS` | Source of template files (from SDD-001) |
| `vault.InitVault(dir)` | existing | Creates directory structure |
| `beads.Client.Available()` | existing | Check if bd is installed |
| `config.LoadConfig()` | existing | Not needed for init (config doesn't exist yet) |

## Constraints / Vincoli

- Language / Linguaggio: Go 1.25+
- Framework: Cobra, `io/fs` for walking embedded templates
- Dependencies / Dipendenze: `internal/embedded` (SDD-001), `internal/vault`, `internal/beads`
- Patterns / Pattern: `fs.WalkDir` to copy from embed.FS to disk, `os.MkdirAll` for directories
- Must be idempotent: running init twice should skip existing files (not overwrite)
- Must NOT import `internal/guardrails` or `internal/board` (not needed at init time)

## Best Practices

- Error handling: wrap all errors with context, fail on first critical error, warn on optional failures (beads, knowledge)
- Naming: `initCmd` (unexported, registered in `init()` — WAIT: no init(). Register in package-level var or in root.go)
- Style: use `slog.InfoContext` for all progress messages, structured output

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | Init creates all expected directories | all dirs |
| Unit | Init copies templates from embedded FS | config.toml, constitution, fd template, sdd template |
| Unit | Init detects Go stack (go.mod present) | go.md copied |
| Unit | Init is idempotent (second run skips existing) | no overwrite |
| Unit | Init with --dir flag works in custom directory | alternate path |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `forgia init` scaffolds `.forgia/` identical to Bash version
- [ ] Templates come from embedded FS, not filesystem `modules/`
- [ ] Stack detection works for Go, Rust, Python, Node
- [ ] Idempotent: existing files not overwritten
- [ ] Beads init runs if `bd` available (graceful skip if not)
- [ ] `--dir` flag changes target directory
- [ ] Progress output via slog (not bare fmt.Println)
- [ ] `go build ./...` succeeds
- [ ] 5+ tests pass

## Context / Contesto

- [ ] `bin/forgia` lines 88-340 — current Bash `cmd_init()` implementation
- [ ] `internal/vault/file_vault.go` — `InitVault()` function
- [ ] `internal/embedded/` — from SDD-001
- [ ] `modules/vault-template/` — reference for expected output structure
- [ ] `cmd/forgia/cmd/root.go` — where to register the command
- [ ] `cmd/forgia/cmd/sync.go` — reference for Cobra command pattern in this project

## Constitution Check

- [x] Respects code standards
- [x] Respects commit conventions
- [x] No hardcoded secrets
- [x] Tests defined and sufficient

## Guardrails

Init creates `.forgia/guardrails/deny.toml` from embedded template. It does NOT read/write any denied files. The `constitution.md` and `config.toml` it creates are NEW files (not modifying existing protected ones).

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
