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

### Architecture
- [ ] Mermaid diagram is present and accurate
- [ ] Components and dependencies are identified
- [ ] Interfaces between components are defined
- [ ] No significant technical debt introduced

### SDD Planning
- [ ] SDD breakdown is listed (which components become separate SDDs)
- [ ] Each SDD has a clear, isolated scope
- [ ] Interfaces between SDDs are defined (contracts)

### Constitution Compliance
- [ ] Respects all principles in constitution.md
- [ ] Security considerations addressed
- [ ] No violations of code standards

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
