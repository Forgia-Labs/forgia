---
id: "SDD-004"
fd: "FD-001"
title: "forgia validate — SDD validation with guardrails"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, cli, validate, guardrails, cobra]
---

# SDD-004: forgia validate — SDD validation with guardrails

> Parent FD: [[FD-001]]

## Scope

Create `cmd/forgia/cmd/validate.go` — Cobra command that validates SDD files before execution.

Two modes:
1. **Single file**: `forgia validate path/to/SDD-001.md`
2. **FD directory**: `forgia validate FD-001` (validates all SDDs in `.forgia/sdd/FD-001/`)

Validation checks:
1. **Frontmatter**: required fields present (id, fd, title, status)
2. **Sections**: required sections exist (Scope, Interfaces, Constraints, Test Requirements, Acceptance Criteria, Context, Constitution Check, Work Log)
3. **Guardrails pre-check**: SDD scope/context files don't reference paths denied in deny.toml
4. **Boundaries pre-check**: if SDD has `boundaries.write_dirs`, verify they exist
5. **Parent FD exists**: the referenced FD file is present in the vault

Output: clear pass/fail per check, exit 0 if all pass, exit 1 if any fail.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `validateCmd` | `*cobra.Command` | `forgia validate <sdd-file|FD-NNN>` |
| `vault.Open(dir)` | existing | Open vault |
| `vault.GetSDD(ctx, fd, sdd)` | existing | Parse SDD frontmatter |
| `guardrails.Parse(data)` | existing | Load deny.toml |
| `guardrails.CheckFilePaths(ctx, paths)` | existing | Check context files against deny list |

## Constraints / Vincoli

- Language / Linguaggio: Go 1.25+
- Framework: Cobra
- Dependencies / Dipendenze: `internal/vault`, `internal/guardrails`
- Patterns / Pattern: Read-only (never modify SDD files during validation)
- Must handle both `.md` frontmatter and `.yaml` SDD formats
- Replaces `modules/runners/validate-sdd.sh`

## Best Practices

- Error handling: collect ALL validation errors (don't stop at first), report them all at the end
- Naming: `validateSDD(ctx, v, sddPath) []ValidationError`
- Style: each error has a category (frontmatter, section, guardrail, boundary) and clear message

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | Valid SDD passes all checks | happy path |
| Unit | Missing frontmatter field fails | each required field |
| Unit | Missing section fails | each required section |
| Unit | SDD referencing denied path fails guardrail check | deny.toml integration |
| Unit | FD directory mode validates all SDDs | batch mode |
| Unit | Nonexistent file returns error | error path |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `forgia validate SDD-001.md` validates single file
- [ ] `forgia validate FD-001` validates all SDDs in directory
- [ ] Missing frontmatter fields reported as errors
- [ ] Missing sections reported as errors
- [ ] Guardrail violations detected in SDD context paths
- [ ] Exit code 0 on success, 1 on failure
- [ ] Output lists all errors (not just first)
- [ ] `go build ./...` succeeds
- [ ] 6+ tests pass

## Context / Contesto

- [ ] `modules/runners/validate-sdd.sh` — Bash validation to replicate
- [ ] `internal/vault/sdd.go` — SDD type and frontmatter fields
- [ ] `internal/guardrails/guardrails.go` — Parse, CheckFilePaths
- [ ] `cmd/forgia/cmd/sync.go` — reference for vault opening pattern

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
