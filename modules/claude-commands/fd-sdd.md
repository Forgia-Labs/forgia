Generate execution specs (SDDs) from an approved Feature Design.

## Instructions

Given FD identifier: $ARGUMENTS

1. Read the specified FD file from `.forgia/fd/`
2. Verify the FD has `reviewed: true` and `status: approved`. If not, refuse and suggest `/fd-review` first
3. Read the SDD template from `.forgia/sdd/_templates/sdd-template.md`
4. Read the constitution from `.forgia/constitution.md`
5. Read the dev-guide from `.forgia/dev-guide/` (general files)
6. Read principles from `.forgia/dev-guide/principles/` (clean-code, SOLID, design-patterns)
7. Read language-specific conventions from `.forgia/dev-guide/lang/` — load only the files relevant to each SDD's language/stack
8. Read guardrails from `.forgia/guardrails/deny.toml` — include the deny list in each SDD's constraints
9. For each component listed in "SDD Previsti" in the FD, create an SDD:
   - Create directory `.forgia/sdd/FD-NNN/`
   - Create `SDD-001-kebab-name.md`, `SDD-002-kebab-name.md`, etc.
   - Populate each SDD with:
     - **Scope**: derived from the FD's component description
     - **Interfaces**: from the FD's interface definitions
     - **Constraints**: language, framework, versions from the FD's architecture
     - **Best Practices**: from dev-guide + constitution + language conventions from `lang/`
     - **Test Requirements**: specific to this component
     - **Acceptance Criteria**: verifiable, derived from FD verification criteria
     - **Context**: files to read, existing code to understand
     - **Constitution Check**: pre-filled from constitution.md
     - **Work Log**: empty, ready to be filled by the agent
7. Update the FD status to "in-progress"
8. Show the user:
   - List of generated SDDs with file paths
   - Suggested next steps:
     - Review each SDD for completeness
     - Assign agent: `/sdd-assign SDD-001`
     - Or execute directly: `mise run sdd .forgia/sdd/FD-NNN/SDD-001.md`
     - Or batch execute: `mise run sdd:batch FD-NNN`

## Important

- NEVER generate SDDs from an unapproved FD
- Each SDD must be self-contained — an agent should be able to execute it without reading the FD
- The SDD is an agent-ready prompt, not traditional documentation
- Include enough context that the agent can work autonomously
- Cross-reference interfaces between SDDs so they produce compatible code
