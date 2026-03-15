Load context for working on a Feature Design.

## Instructions

Given FD identifier: $ARGUMENTS

1. Read the specified FD file from `.forgia/fd/FD-NNN*.md`
2. Read the dev-guide files:
   - `.forgia/dev-guide/coding-conventions.md`
   - `.forgia/dev-guide/commit-conventions.md`
   - `.forgia/dev-guide/review-process.md`
3. Read principles from `.forgia/dev-guide/principles/` (clean-code, SOLID, design-patterns)
4. Read language-specific conventions from `.forgia/dev-guide/lang/` — load all files present
5. Read the constitution: `.forgia/constitution.md`
4. Read any related SDDs in `.forgia/sdd/FD-NNN/` if they exist
5. Summarize:
   - The FD's current status and what needs to be done
   - Key constraints from the dev-guide and constitution
   - Related FDs that might conflict or depend on this one
   - SDDs already generated (if any) and their status

## Important

- This is a READ-ONLY operation — do not modify any files
- If no FD identifier is provided, show the dashboard (all FDs by status)
- Use `.forgia/` as the vault path (project-local)
