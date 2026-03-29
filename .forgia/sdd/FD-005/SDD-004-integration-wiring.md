---
id: "SDD-004"
fd: "FD-005"
title: "Integration wiring and documentation"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-29"
started: "2026-03-29"
completed: "2026-03-29"
tags: ["phase:2-core"]
---

# SDD-004: Integration wiring and documentation

> Parent FD: [[FD-005]]

## Scope

Wire the `/fd-threat-model` slash command and its integrations into the Forgia workflow. Verify everything works end-to-end.

### Deliverables

1. **Update `.forgia/dev-guide/review-process.md`** — add `/fd-threat-model` as an optional security step in the workflow. The updated flow:

   ```
   FD Review ──→ Arch Review ──→ Threat Model ──→ SDD Generation ──→ ...
   /fd-review    /fd-arch-review   /fd-threat-model   /fd-sdd
   GATE 1        (optional)        (optional)
   ```

   Add a subsection documenting `/fd-threat-model`: when to run, what it produces, that it's advisory but recommended for security-sensitive FDs.

2. **Verify embedded skill loading** — confirm `modules/claude-commands/fd-threat-model.md` is picked up by `go:embed` and `forgia skills` lists `fd-threat-model`.

3. **E2E validation** — run `go build ./cmd/forgia/` and verify:
   - `forgia skills` lists `fd-threat-model`
   - The full workflow works: `/fd-threat-model` produces file → `/fd-review` detects it → `/fd-sdd` injects mitigations

### Dependency

SDDs 001-003 must all be completed before this SDD executes.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| SDD-001 output | file | `modules/claude-commands/fd-threat-model.md` must exist |
| SDD-002 output | file | `modules/claude-commands/fd-review.md` must have threat model check |
| SDD-003 output | file | `modules/claude-commands/fd-sdd.md` must have threat model injection |
| `review-process.md` | markdown | Updated workflow with `/fd-threat-model` step |
| `forgia skills` CLI | stdout | Must list `fd-threat-model` |

## Constraints / Vincoli

- Language / Linguaggio: Markdown (documentation), Go (build verification only)
- Framework: Forgia dev-guide conventions
- Dependencies / Dipendenze: SDD-001, SDD-002, SDD-003 must all be completed
- Patterns / Pattern: Follow existing review-process.md structure

### Guardrails

- Do not modify `.forgia/constitution.md`, `.forgia/config.toml`, `.forgia/guardrails/deny.toml`

## Best Practices

- Error handling: If any dependency SDD output is missing, report and stop
- Naming: Keep review-process.md structure consistent with existing sections
- Style: Match existing workflow diagram format (ASCII art)

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Manual | `go build ./cmd/forgia/` succeeds | Build verification |
| Manual | `forgia skills` lists `fd-threat-model` | Skill registration |
| Manual | review-process.md has `/fd-threat-model` in workflow | Documentation |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `.forgia/dev-guide/review-process.md` updated with `/fd-threat-model` in workflow diagram
- [ ] New subsection in review-process.md documents `/fd-threat-model`: purpose, output, advisory nature
- [ ] `go build ./cmd/forgia/` succeeds
- [ ] `forgia skills` lists `fd-threat-model`
- [ ] No Go code was modified — only markdown files
- [ ] Full workflow verified: threat model file → fd-review detection → fd-sdd injection

## Context / Contesto

- [ ] `.forgia/dev-guide/review-process.md` — file to update
- [ ] `modules/claude-commands/fd-threat-model.md` — created by SDD-001
- [ ] `modules/claude-commands/fd-review.md` — updated by SDD-002
- [ ] `modules/claude-commands/fd-sdd.md` — updated by SDD-003
- [ ] `internal/skill/embedded.go` — `inferCategory()` confirms `fd-` prefix → `CategoryFD`
- [ ] `embedded.go` — `go:embed all:modules/claude-commands`

## Constitution Check

- [ ] Respects code standards — documentation follows existing conventions
- [ ] Respects commit conventions — `feat(FD-005): description`
- [ ] No hardcoded secrets — no secret handling
- [ ] Tests defined and sufficient — 3 verification steps

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~5 min

### Decisions / Decisioni

1. Placed `/fd-threat-model` as an optional step between Gate 1 (FD Review) and SDD Generation in the workflow diagram, matching the SDD spec's requested flow
2. Added a dedicated subsection documenting purpose, output, advisory nature, and downstream effects — kept it concise and consistent with existing gate descriptions
3. Confirmed no Go code changes needed — `go:embed all:modules/claude-commands` automatically picks up the new `fd-threat-model.md` file

### Output

- **Commit(s)**: pending
- **PR**: part of ferruvich/issues-21 branch
- **Files created/modified**:
  - `.forgia/dev-guide/review-process.md` — updated workflow diagram and added `/fd-threat-model` subsection
  - `.forgia/sdd/FD-005/SDD-004-integration-wiring.md` — Work Log filled

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: Clean dependency chain — all SDD-001/002/003 outputs were in place, making integration straightforward. The `go:embed` auto-discovery pattern meant zero Go changes were needed.
- **What didn't / Cosa non ha funzionato**: Nothing blocked execution.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Integration/wiring SDDs are ideal for verifying the full chain works. Consider making E2E validation a standard final SDD in multi-SDD FDs.
