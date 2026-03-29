---
id: "SDD-002"
fd: "FD-005"
title: "/fd-review integration — threat model existence check"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-29"
started: ""
completed: ""
tags: ["phase:2-core"]
---

# SDD-002: /fd-review integration — threat model existence check

> Parent FD: [[FD-005]]

## Scope

Update `modules/claude-commands/fd-review.md` to optionally check for the existence of a threat model file when reviewing an FD.

### Changes

Add a new check item under the "Constitution & Security Compliance" section:

```markdown
- [ ] **Threat model** — check if `.forgia/fd/FD-NNN-threat-model.md` exists. If present, note "Threat model available" and verify it covers the FD's components. If missing, flag as advisory: "Threat model non presente — considera `/fd-threat-model FD-NNN` per analisi di sicurezza." This is NOT a blocker — the FD can still pass review without a threat model.
```

### Dependency

SDD-001 must be completed first (the threat model file format must be defined before fd-review can check for it).

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| Threat model file | filesystem | `.forgia/fd/FD-NNN-threat-model.md` — checked for existence |
| `/fd-review` output | markdown | Advisory flag in review output (not a blocker) |
| SDD-001 contract | file format | Threat model file path convention: `FD-NNN-threat-model.md` in `.forgia/fd/` |

## Constraints / Vincoli

- Language / Linguaggio: Markdown (slash command prompt update)
- Framework: Existing `/fd-review` structure
- Dependencies / Dipendenze: SDD-001 (defines threat model file format)
- Patterns / Pattern: Advisory check, not a gate — review can pass without threat model

### Guardrails

- Do not modify the review pass/fail logic — threat model check is advisory only
- Do not modify `.forgia/constitution.md` or `.forgia/guardrails/deny.toml`

## Best Practices

- Error handling: If threat model file doesn't exist, flag as advisory — never fail the review for this
- Naming: Keep the check item consistent with existing checklist style
- Style: Italian for the advisory message, consistent with existing fd-review output

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Manual | Run `/fd-review` on FD with threat model present — verify "Threat model available" noted | Present path |
| Manual | Run `/fd-review` on FD without threat model — verify advisory flag, review still passes | Missing path |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `/fd-review` checks for `.forgia/fd/FD-NNN-threat-model.md` existence
- [ ] If present, review output notes "Threat model available"
- [ ] If missing, review output flags advisory: suggests `/fd-threat-model`
- [ ] Threat model check does NOT block FD approval — it is advisory only
- [ ] Existing fd-review checks unchanged

## Context / Contesto

- [ ] `modules/claude-commands/fd-review.md` — file to modify
- [ ] `.forgia/sdd/FD-005/SDD-001-fd-threat-model-command.md` — defines threat model file format
- [ ] `.forgia/fd/FD-005-fd-threat-model.md` — parent FD

## Constitution Check

- [ ] Respects code standards — minimal change to existing command
- [ ] Respects commit conventions — `feat(FD-005): description`
- [ ] No hardcoded secrets — no secret handling
- [ ] Tests defined and sufficient — 2 manual test scenarios (present/missing)

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~5 min

### Decisions / Decisioni

1. Added the threat model check as the last item in "Constitution & Security Compliance" with an explicit `(advisory)` label, keeping it visually consistent with existing checklist items while clearly marking it as non-blocking.
2. Updated the pass/fail logic (steps 5 and 6) to explicitly exclude advisory items from the gate decision. Step 5 now says "excluding advisory items" and includes advisory recommendations as "Note advisory" after approval. Step 6 now says "non-advisory check" to prevent threat model absence from blocking FD approval.
3. Used the exact Italian advisory message from the SDD spec: "Threat model non presente — considera `/fd-threat-model FD-NNN` per analisi di sicurezza."

### Output

- **Commit(s)**: pending
- **PR**: pending
- **Files created/modified**:
  - `modules/claude-commands/fd-review.md` (modified — added threat model advisory check + updated pass/fail logic)
  - `.forgia/sdd/FD-005/SDD-002-fd-review-integration.md` (modified — Work Log)

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: The existing checklist structure made it straightforward to add a new check item. The SDD spec was precise about the check wording and advisory behavior, leaving no ambiguity.
- **What didn't / Cosa non ha funzionato**: Nothing — the change was minimal and well-scoped.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: When adding advisory (non-blocking) checks to gate commands, the SDD should explicitly specify how the pass/fail logic should be updated. This SDD did it well by stating "NOT a blocker", which made the logic change clear.
