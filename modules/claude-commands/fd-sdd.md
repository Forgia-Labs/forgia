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
9. **Read threat model** (if present):
   - Check if `.forgia/fd/FD-NNN-threat-model.md` exists (where `FD-NNN` matches the FD identifier from step 1)
   - If the file exists, read it and extract the `## 4. Recommendations for SDDs` section
   - Parse each recommendation line (format: `N. **SDD-NNN**: <mitigation description>`)
   - Store the mapping of SDD numbers to their security mitigations for use in step 10
   - If the file does not exist, skip silently — no error, no warning. Proceed with SDD generation normally.
10. For each component listed in "SDD Previsti" in the FD, create an SDD:
   - Create directory `.forgia/sdd/FD-NNN/`
   - Create `SDD-001-kebab-name.md`, `SDD-002-kebab-name.md`, etc.
   - Populate each SDD with:
     - **Scope**: derived from the FD's component description
     - **Interfaces**: from the FD's interface definitions
     - **Constraints**: language, framework, versions from the FD's architecture
       - If threat model mitigations were found in step 9 for this SDD number, add a `### Security (from threat model)` subsection at the end of the Constraints section. Each injected constraint is a bullet point ending with `(source: threat model)`. Example:
         ```
         ### Security (from threat model)
         - Add input sanitization for title/description (source: threat model)
         ```
       - If no mitigations match this SDD number, do not add the subsection
       - Mitigations are additive — never remove or modify existing constraints
     - **Best Practices**: from dev-guide + constitution + language conventions from `lang/`
     - **Test Requirements**: specific to this component
     - **Acceptance Criteria**: verifiable, derived from FD verification criteria
     - **Context**: files to read, existing code to understand
     - **Constitution Check**: pre-filled from constitution.md
     - **Work Log**: empty, ready to be filled by the agent

   **MANDATORY: The last SDD must always be an Integration Wiring SDD.**
   This SDD is automatically added even if the FD does not list it. Its scope:
   - **Wiring**: `pub mod`, `use`, component registration, dependency injection setup
   - **Startup path**: who calls what, in what order — trace from entry point to leaf
   - **E2E test**: the complete flow from the public entry point (IPC, API, CLI) to the innermost component
   - **Acceptance criteria**: NOT "component X works" but "the user calls Y and Z happens"

   Why: without this SDD, each component is tested in isolation but never verified together.
   Pattern observed: code compiles, unit tests pass, but nothing works end-to-end because
   modules are never wired, functions are never called from the startup path, and integration
   is left as an implicit assumption that nobody verifies.

11. Update the FD status to "in-progress"
12. **Beads integration** (if `bd` is available and `.beads/` exists):
   - Create a bd epic for the FD: `bd create "FD-NNN: <title>" --type=epic`
   - For each SDD, create a bd subtask: `bd create "SDD-NNN: <title>" --parent=<epic-id>`
   - Add dependencies between SDDs if they have interface contracts
   - Tag each bd task with the SDD id for cross-reference
13. Show the user:
   - List of generated SDDs with file paths
   - Beads epic and task IDs (if created)
   - Suggested next steps:
     - Review each SDD for completeness
     - Execute directly: `forgia exec .forgia/sdd/FD-NNN/SDD-001.md`
     - Batch execute: `forgia batch FD-NNN`
     - Check ready tasks: `bd ready`

## Important

- NEVER generate SDDs from an unapproved FD
- Each SDD must be self-contained — an agent should be able to execute it without reading the FD
- The SDD is an agent-ready prompt, not traditional documentation
- Include enough context that the agent can work autonomously
- Cross-reference interfaces between SDDs so they produce compatible code
- Beads tasks mirror SDDs — they are NOT a replacement, they add dependency tracking
