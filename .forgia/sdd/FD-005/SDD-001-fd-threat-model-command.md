---
id: "SDD-001"
fd: "FD-005"
title: "/fd-threat-model slash command"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-29"
started: "2026-03-29"
completed: "2026-03-29"
tags: ["phase:2-core"]
---

# SDD-001: /fd-threat-model slash command

> Parent FD: [[FD-005]]

## Scope

Create `modules/claude-commands/fd-threat-model.md` — a Claude Code slash command that performs STRIDE-based threat modeling on a Feature Design and writes the result to a persistent file.

The command accepts an FD identifier (`$ARGUMENTS` = `FD-NNN`) and produces `.forgia/fd/FD-NNN-threat-model.md` with these sections:

1. **Assets** — identify data, access credentials, tokens, infrastructure components at risk based on the FD's architecture and interfaces
2. **Threat Actors** — enumerate relevant threat actors (unauthenticated users, compromised plugins, malicious input, insider threats) based on the feature's attack surface
3. **STRIDE Analysis** — per-component table covering all 6 categories:
   - **S**poofing — identity impersonation
   - **T**ampering — data/code modification
   - **R**epudiation — deniable actions
   - **I**nformation Disclosure — data leaks
   - **D**enial of Service — availability attacks
   - **E**levation of Privilege — unauthorized access escalation
   Each row: Threat, Category, Component, Risk (High/Medium/Low), Mitigation
4. **Recommendations for SDDs** — numbered list mapping mitigations to specific SDD numbers with concrete actions (e.g., "SDD-002: add input sanitization for title field")
5. **Guardrails Additions** — concrete `deny.toml` patterns to suggest adding (e.g., `[execute] "kubectl exec"`)

### Output

The command **writes a file** at `.forgia/fd/FD-NNN-threat-model.md`. This is a persistent artifact that `/fd-review` and `/fd-sdd` will consume (SDD-002 and SDD-003).

### What this SDD does NOT cover

- Updating `/fd-review` to check for threat model (SDD-002)
- Updating `/fd-sdd` to inject mitigations (SDD-003)
- Updating review-process.md or verifying skill loading (SDD-004)

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `$ARGUMENTS` input | string | FD identifier (e.g., `FD-005`). Must match existing file in `.forgia/fd/` |
| Threat model output | file | `.forgia/fd/FD-NNN-threat-model.md` — persistent markdown artifact |
| Vault reads | filesystem | FD file, constitution.md, dev-guide/*, guardrails/deny.toml |
| Codebase reads | filesystem | Scan for existing security patterns (auth, validation, encryption). Must respect deny.toml |

## Constraints / Vincoli

- Language / Linguaggio: Markdown (Claude Code slash command prompt)
- Framework: Claude Code slash command system (`$ARGUMENTS` substitution)
- Dependencies / Dipendenze: None — standalone markdown file
- Patterns / Pattern: Follow structure of `fd-arch-review.md` for multi-step analysis

### Guardrails (from deny.toml)

- Before scanning codebase files, check paths against `[read]` deny patterns — skip denied files
- Never modify the FD itself, constitution, config.toml, or deny.toml
- The only file written is `.forgia/fd/FD-NNN-threat-model.md`

## Best Practices

- Error handling: If FD file doesn't exist, output: `"FD non trovato: FD-NNN. Verifica l'identificatore."` — do not create partial threat model
- Naming: Output file must be `FD-NNN-threat-model.md` (kebab, matches FD naming convention)
- Style: Follow the output format shown in issue #21 (STRIDE table with Risk/Mitigation columns, numbered SDD recommendations)
- STRIDE must cover all 6 categories per relevant component — if a category has no threats, explicitly state "No threats identified"

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Manual | Run `/fd-threat-model FD-003` — verify threat model file created with all 5 sections | File creation + sections |
| Manual | Run `/fd-threat-model FD-999` — verify error message | Error path |
| Manual | Verify STRIDE table has all 6 categories | STRIDE completeness |
| Manual | Verify SDD recommendations reference specific SDD numbers | Recommendation quality |
| Manual | Verify guardrails suggestions are concrete deny.toml patterns | Guardrails quality |
| Manual | Verify no files matching deny.toml patterns are read | Guardrails compliance |

## Acceptance Criteria / Criteri di Accettazione

- [ ] File `modules/claude-commands/fd-threat-model.md` exists and is a valid Claude Code slash command
- [ ] Command creates `.forgia/fd/FD-NNN-threat-model.md` with all 5 sections: Assets, Threat Actors, STRIDE Analysis, SDD Recommendations, Guardrails Suggestions
- [ ] STRIDE table covers all 6 categories (S, T, R, I, D, E) per relevant component
- [ ] SDD recommendations reference specific SDD numbers with concrete mitigations
- [ ] Guardrails suggestions are concrete `deny.toml` patterns
- [ ] Command reads constitution, dev-guide, and existing deny.toml patterns
- [ ] Command only writes the threat model file — does not modify FD, guardrails, or codebase
- [ ] Command respects deny.toml read patterns when scanning codebase
- [ ] Error handling: non-existent FD produces Italian error message, no partial file
- [ ] Threat model file format is consumable by `/fd-review` (SDD-002) and `/fd-sdd` (SDD-003)

## Context / Contesto

- [ ] `.forgia/fd/FD-005-fd-threat-model.md` — parent FD with full requirements
- [ ] `modules/claude-commands/fd-arch-review.md` — reference for multi-step analysis command structure
- [ ] `modules/claude-commands/fd-review.md` — will be updated in SDD-002 to consume threat model
- [ ] `modules/claude-commands/fd-sdd.md` — will be updated in SDD-003 to consume threat model
- [ ] `.forgia/constitution.md` — security section loaded by the command
- [ ] `.forgia/guardrails/deny.toml` — deny patterns loaded and checked against
- [ ] GitHub issue #21 — original requirements and example output format

## Constitution Check

- [ ] Respects code standards — markdown follows existing slash command conventions
- [ ] Respects commit conventions — `feat(FD-005): description`
- [ ] No hardcoded secrets — command handles no secrets
- [ ] Tests defined and sufficient — 6 manual test scenarios

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. Followed `arch-review.md` and `fd-review.md` as structural references for the multi-step analysis command. The command uses the same pattern: numbered steps, vault reads, guardrail compliance, and a structured output format.
2. Used a tabular format for Assets, Threat Actors, and STRIDE Analysis sections to ensure machine-parseable output that SDD-002 (`/fd-review` integration) and SDD-003 (`/fd-sdd` integration) can consume.
3. Made STRIDE analysis per-component with sub-headings rather than a single flat table, so each component's threats are clearly scoped and the output scales with FD complexity.
4. Included explicit "No threats identified" requirement for empty STRIDE categories to ensure completeness is verifiable.
5. Output file uses frontmatter (`fd`, `generated`, `generator` fields) so downstream commands can programmatically identify and parse threat models.
6. No Go code changes needed — the `go:embed all:modules/claude-commands` directive and `LoadEmbedded` walker automatically pick up new `.md` files.

### Output

- **Commit(s)**: pending
- **PR**: pending
- **Files created/modified**:
  - `modules/claude-commands/fd-threat-model.md` (created — the slash command)
  - `.forgia/sdd/FD-005/SDD-001-fd-threat-model-command.md` (modified — Work Log)

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: The embedded skill system made this a zero-config addition — just drop a `.md` file and it auto-registers with correct category, description, and `$ARGUMENTS` substitution. The existing commands (`fd-review.md`, `arch-review.md`) provided clear structural patterns to follow.
- **What didn't / Cosa non ha funzionato**: The SDD referenced `fd-arch-review.md` as the structural reference, but the actual file is `arch-review.md` — minor naming mismatch in the Context section.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: When referencing existing files in SDD Context sections, use the exact filename (e.g., `arch-review.md` not `fd-arch-review.md`) to avoid confusion during execution.
