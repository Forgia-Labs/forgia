Scaffold a Forgia vault in the current project.

## Instructions

1. Check if `.forgia/` already exists in the current working directory
2. If it exists, ask the user if they want to update templates (non-destructive: never overwrite existing FDs/SDDs)
3. Create the directory structure:
   ```
   .forgia/
     constitution.md
     _dashboard.md
     fd/_templates/fd-template.md
     sdd/_templates/sdd-template.md
     ops/_templates/ops-task.md
     ops/active/
     ops/done/
     dev-guide/
       coding-conventions.md
       commit-conventions.md
       review-process.md
   ```
4. Copy templates from the forgia repo's `modules/vault-template/` directory
5. If the forgia repo is not found, use the built-in templates embedded in this command
6. Ask the user to customize `constitution.md` with their project-specific rules
7. Show next steps:
   - `/fd-new "feature name"` to create the first Feature Design
   - `forgia doctor` to verify the setup

## Important

- Never overwrite existing FDs, SDDs, or OPS tasks
- Always create the full directory structure even if some dirs are empty
- The `.forgia/` directory should be committed to version control
- Add `.forgia/ops/` to `.gitignore` if the user prefers private task tracking
