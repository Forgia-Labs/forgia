Create a new Feature Design (FD) in the Forgia vault.

## Instructions

1. Ask the user for a brief description of the feature/task if not provided as argument: $ARGUMENTS
2. Verify `.forgia/fd/` exists. If not, suggest running `/project-init` first
3. Determine the next FD ID by counting existing FD files in `.forgia/fd/`
4. Read the FD template from `.forgia/fd/_templates/fd-template.md`
5. Create the new FD file at `.forgia/fd/FD-NNN-kebab-title.md` with:
   - Auto-incremented ID (FD-001, FD-002, etc.)
   - Title from the user's description
   - Today's date
   - Status: "planned"
   - All template sections with placeholder guidance
6. Pre-populate the "SDD Previsti" section based on the feature scope (suggest component breakdown)
7. Show the user the created file path and suggest next steps:
   - Fill in the problem, solutions, and architecture sections
   - When ready: `/fd-review FD-NNN`

## Important

- Never modify existing FD files when creating a new one
- Use `.forgia/fd/` as the vault path (project-local, not ~/Obsidian)
- Follow the template exactly — do not skip sections
- The FD must be technology-agnostic in the Problem section (what, not how)
