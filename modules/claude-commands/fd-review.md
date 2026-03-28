Perform mandatory review of a Feature Design before SDD generation.

## Instructions

Given FD identifier: $ARGUMENTS

1. Read the specified FD file from `.forgia/fd/`
2. Read the project constitution from `.forgia/constitution.md`
3. Read the dev-guide conventions from `.forgia/dev-guide/` (general files)
4. Read principles from `.forgia/dev-guide/principles/` (clean-code, SOLID, design-patterns)
5. Read language-specific conventions from `.forgia/dev-guide/lang/` — load ALL files present (they were auto-detected during `forgia init`)
6. Perform the review by checking EVERY item:

### Problem Definition
- [ ] Problem is clearly defined (not vague or too broad)
- [ ] Problem describes WHAT and WHY (not HOW)
- [ ] At least 2 solutions were considered with pros/cons
- [ ] The chosen solution is justified

### Architecture & Mermaid Diagrams (MANDATORY)
- [ ] **Integration Context diagram** is present — shows where the feature sits in the existing system (existing components in grey, new in green)
- [ ] **Data Flow diagram** is present — sequence diagram showing the interaction between components
- [ ] Diagrams are NOT the default template placeholders — they must reflect the actual feature
- [ ] Diagrams use correct Mermaid syntax (flowchart/sequenceDiagram)
- [ ] Components and dependencies are identified in the diagrams
- [ ] Interfaces between components are defined and match the diagrams
- [ ] No significant technical debt introduced

### SDD Planning
- [ ] SDD breakdown is listed (which components become separate SDDs)
- [ ] Each SDD has a clear, isolated scope
- [ ] Interfaces between SDDs are defined (contracts)
- [ ] **Integration Wiring SDD is present or acknowledged** — the last planned SDD must cover: wiring (mod/use/registration), startup path, E2E test from public entry point to inner component. If missing, the FD FAILS this check. If the FD has only 1 SDD (single component), integration wiring can be part of that SDD's scope — but it must be explicitly mentioned.

### Constitution & Security Compliance
- [ ] Respects all principles in constitution.md
- [ ] No violations of code standards
- [ ] Security guardrails reviewed (`.forgia/guardrails/deny.toml`)
- [ ] No proposed interfaces expose or require secrets in plaintext
- [ ] No file patterns in the FD would violate deny.toml read/write/execute rules

### Verification
- [ ] Verification criteria are defined and sufficient
- [ ] Each criterion is objectively testable

5. If ALL checks pass:
   - Set `reviewed: true` and `reviewer: "claude"` in frontmatter
   - Update status to "approved"
   - Report: "APPROVATO — FD pronto per generazione SDD. Usa /fd-sdd FD-NNN"

6. If ANY check fails:
   - Keep `reviewed: false`
   - List all failed checks with specific feedback
   - Report: "REVISIONE RICHIESTA" with action items

## Important

- This is a GATE — the FD CANNOT proceed to SDD generation without passing review
- Be strict: vague problems, missing diagrams, or undefined interfaces are failures
- Always check against the constitution
