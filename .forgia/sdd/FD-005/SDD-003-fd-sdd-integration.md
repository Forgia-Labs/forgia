---
id: "SDD-003"
fd: "FD-005"
title: "/fd-sdd integration — inject threat mitigations into SDDs"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-29"
started: "2026-03-29"
completed: "2026-03-29"
tags: ["phase:2-core"]
---

# SDD-003: /fd-sdd integration — inject threat mitigations into SDDs

> Parent FD: [[FD-005]]

## Scope

Update `modules/claude-commands/fd-sdd.md` to read the threat model file (if present) and inject relevant security mitigations into each SDD's Constraints section during generation.

### Changes

Add a new step after reading the FD and before generating SDDs:

1. Check if `.forgia/fd/FD-NNN-threat-model.md` exists
2. If present, read the "Recommendations for SDDs" section
3. For each SDD being generated, find matching recommendations (by SDD number or component name)
4. Inject matched mitigations into the SDD's "Constraints / Vincoli" section under a `### Security (from threat model)` subsection
5. If no threat model exists, skip silently — no error, no warning (threat model is optional)

### Example

If the threat model contains:
```
## Recommendations for SDDs
- SDD-002: add input sanitization for title/description
- SDD-003: add CSP headers, escape user content in React
```

Then SDD-002's Constraints section would include:
```
### Security (from threat model)
- Add input sanitization for title/description (source: threat model)
```

### Dependency

SDD-001 must be completed first (threat model file format defined).

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| Threat model file | filesystem | `.forgia/fd/FD-NNN-threat-model.md` — read "Recommendations for SDDs" section |
| SDD Constraints section | markdown | Inject `### Security (from threat model)` subsection with matched mitigations |
| SDD-001 contract | file format | Threat model has "## Recommendations for SDDs" section with `- SDD-NNN: description` entries |

## Constraints / Vincoli

- Language / Linguaggio: Markdown (slash command prompt update)
- Framework: Existing `/fd-sdd` structure
- Dependencies / Dipendenze: SDD-001 (defines threat model format and recommendations section)
- Patterns / Pattern: Graceful skip when threat model absent — no error, no warning

### Guardrails

- Only read the threat model file — never modify it
- Injected constraints are additive — never remove existing SDD constraints
- Do not modify `.forgia/constitution.md` or `.forgia/guardrails/deny.toml`

## Best Practices

- Error handling: Missing threat model → skip silently. Malformed threat model → warn but continue SDD generation
- Naming: Security constraints subsection: `### Security (from threat model)` — clearly sourced
- Style: Each injected constraint ends with `(source: threat model)` for traceability

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Manual | Run `/fd-sdd` on FD with threat model — verify SDDs have `### Security (from threat model)` section | Injection present |
| Manual | Run `/fd-sdd` on FD without threat model — verify SDDs generated normally, no errors | Skip path |
| Manual | Verify mitigations map to correct SDDs by number | Mapping accuracy |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `/fd-sdd` reads `.forgia/fd/FD-NNN-threat-model.md` when present
- [ ] Matching mitigations injected into SDD Constraints under `### Security (from threat model)`
- [ ] Each injected constraint includes `(source: threat model)` for traceability
- [ ] Missing threat model → SDDs generated normally, no error or warning
- [ ] Existing fd-sdd behavior unchanged when no threat model exists
- [ ] Mitigations correctly match SDD numbers from the recommendations section

## Context / Contesto

- [ ] `modules/claude-commands/fd-sdd.md` — file to modify
- [ ] `.forgia/sdd/FD-005/SDD-001-fd-threat-model-command.md` — defines threat model file format
- [ ] `.forgia/fd/FD-005-fd-threat-model.md` — parent FD

## Constitution Check

- [ ] Respects code standards — minimal change to existing command
- [ ] Respects commit conventions — `feat(FD-005): description`
- [ ] No hardcoded secrets — no secret handling
- [ ] Tests defined and sufficient — 3 manual test scenarios

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-29
- **Completed**: 2026-03-29
- **Duration / Durata**: ~10 min

### Decisions / Decisioni

1. Inserted the threat model reading as step 9 (between guardrails reading and SDD generation) to maintain the natural data-flow: all context is loaded before generation begins.
2. Used the exact `## 4. Recommendations for SDDs` heading and `N. **SDD-NNN**: <description>` line format defined by SDD-001's output template for parsing, ensuring tight contract coupling between producer and consumer.
3. Placed the injection instructions inside the existing Constraints bullet of step 10 rather than as a separate step, keeping the SDD generation logic cohesive — the security subsection is just another part of the Constraints section.
4. Fixed pre-existing step numbering errors (steps 8, 9, 10 after the SDD generation block were misnumbered) to maintain a clean 1–13 sequence.
5. Added explicit "mitigations are additive" guardrail to prevent accidental removal of existing constraints during injection.

### Output

- **Commit(s)**: pending
- **PR**: pending
- **Files created/modified**:
  - `modules/claude-commands/fd-sdd.md` (modified — added threat model integration)
  - `.forgia/sdd/FD-005/SDD-003-fd-sdd-integration.md` (modified — Work Log)

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: The threat model output format (defined in SDD-001) has a clear, parseable structure (`## 4. Recommendations for SDDs` with numbered `**SDD-NNN**:` lines) that made the integration straightforward. The fd-sdd.md prompt was well-structured with clear step numbering, making insertion clean.
- **What didn't / Cosa non ha funzionato**: The original fd-sdd.md had step numbering errors (steps 8, 9, 10 after the big step 9 block restarted at 8 instead of continuing). Fixed as part of this SDD.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: When an SDD produces an output file format that other SDDs consume (producer/consumer pattern), include the exact parsing format in the consumer SDD's Interfaces section — not just a reference to the producer SDD. This avoids requiring the agent to read the producer SDD during execution.
