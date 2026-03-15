Create a Feature Design (FD) from a GitHub issue.

## Instructions

1. Parse arguments from: $ARGUMENTS
   Format: `<issue-number> [--repo owner/repo] [--context path]`

   - `issue-number` (required): GitHub issue number
   - `--repo owner/repo` (optional): target repository. If omitted, infer from `git remote get-url origin`
   - `--context path` (optional): additional context file or directory to read

2. **Pre-flight checks**:
   - Verify `.forgia/fd/` exists. If not, tell the user: "Vault non inizializzato. Esegui `/project-init` prima."
   - Verify `gh` CLI is installed by running `command -v gh`. If not found, tell the user: "gh CLI non trovato. Installa GitHub CLI: https://cli.github.com/"

3. **Fetch the issue data**:
   - Try running `forgia fd-from-issue <args> --dry-run` first. If `forgia` CLI is available, parse the draft FD output.
   - If `forgia` CLI is not available, fall back to fetching directly:
     ```
     gh api repos/{owner}/{repo}/issues/{number}
     ```
   - Extract: `title`, `body`, `labels` (array of name strings), `assignee.login`, `milestone.title`
   - Fetch comments: `gh api repos/{owner}/{repo}/issues/{number}/comments`

4. **Discover local context**:
   - Try running `modules/context-discovery.sh --format=markdown` if available
   - If not available, manually scan for these files and read any that exist:
     - `docs/decisions/*.md` (ADRs)
     - `CONTRIBUTING.md`
     - `ARCHITECTURE.md`, `docs/ARCHITECTURE.md`
     - `.forgia/constitution.md`
     - `.forgia/dev-guide/**/*.md`
     - `CODE_OF_CONDUCT.md`
     - `docs/*.md`
   - If `--context` argument was provided, also read files at that path
   - **Actually read the content** of all discovered files — do not just list them

5. **Read the FD template** from `modules/vault-template/fd/_templates/fd-template.md` to know the exact section structure. The generated FD MUST have ALL sections from this template.

6. **Determine the next FD ID**:
   - List existing FD files in `.forgia/fd/` (excluding `_templates/`)
   - Find the highest FD number and increment by 1
   - Format: `FD-NNN` (zero-padded to 3 digits)

7. **Create the FD** at `.forgia/fd/FD-NNN-kebab-title.md` with agent-refined content:

   ### Frontmatter
   - `id`: the new FD ID
   - `title`: cleaned issue title
   - `status`: "planned"
   - `priority`: "high" if labels contain "bug", "medium" if "enhancement", else "medium"
   - `effort`: "medium" (default)
   - `impact`: "medium" (default)
   - `assignee`: from issue assignee, or empty
   - `created`: today's date
   - `reviewed`: false
   - `reviewer`: ""
   - `tags`: from issue labels (map label names to tag array)
   - `upstream_issue`: `"owner/repo#number"` (e.g. `"Deepzima/forgia#7"`)

   ### Problem / Problema
   - Restructure the issue body into a clear problem statement
   - Extract the "what" and "why" — remove implementation details, code snippets, workarounds
   - Write in a clean, concise style that describes the problem from the user's perspective
   - Do NOT dump the raw issue body — rewrite it as a proper problem definition

   ### Solutions Considered / Soluzioni Considerate
   - Identify solutions mentioned in the issue body and comments
   - If only one solution is apparent, propose a second alternative with trade-offs
   - Each option MUST have **Pro:** and **Con / Contro:** items
   - Mark the recommended option as "(chosen) / (scelta)"

   ### Architecture / Architettura
   - Generate a **Integration Context** mermaid flowchart showing where the feature integrates in the existing system
     - Existing components in grey (`fill:#f0f0f0,stroke:#999`)
     - New components in green (`fill:#d4edda,stroke:#28a745`)
   - Generate a **Data Flow** mermaid sequence diagram showing component interactions
   - Base diagrams on information from ARCHITECTURE.md, ADRs, and codebase structure
   - If the architecture cannot be determined from available context, use `<!-- TODO: fill architecture diagram based on codebase analysis -->` and tell the user

   ### Interfaces / Interfacce
   - Define component interfaces based on the proposed solution
   - Fill the interface table: Component, Input, Output, Protocol

   ### Planned SDDs / SDD Previsti
   - Suggest a component breakdown for SDD generation
   - At least 1 SDD must be listed
   - Each entry: `SDD-NNN: description of component scope`

   ### Constraints / Vincoli
   - Include constraints from: CONTRIBUTING.md rules, milestone deadlines, label-implied constraints
   - Add technical constraints from discovered context

   ### Verification / Verifica
   - Create testable verification criteria based on the issue requirements
   - Each criterion should be a checkbox item

   ### Notes / Note
   - Link to the upstream issue: `Upstream: owner/repo#number`
   - List all auto-discovered context files as links
   - Include relevant maintainer comments from the issue (summarized, not raw dumps)

8. **Verify the FD was NOT overwritten**: confirm no existing file was modified — only a new file was created.

9. **Show the user** the created file path and suggest next steps:
   - "FD creato: `.forgia/fd/FD-NNN-kebab-title.md`"
   - "Prossimo passo: `/fd-review FD-NNN`"

## Important

- Never modify existing FD files when creating a new one
- Use `.forgia/fd/` as the vault path (project-local)
- Follow the template exactly — do not skip sections. `/fd-review` will reject the FD if sections are missing
- The Problem section must be a clean rewrite, NOT a raw dump of the issue body
- Always include at least 2 solutions with pros/cons — this is a `/fd-review` requirement
- Architecture mermaid diagrams are MANDATORY — `/fd-review` will reject without them
- Actually read discovered context files to inform the FD content — don't just link them
- If any section cannot be determined from available information, use `<!-- TODO: ... -->` markers and tell the user what needs manual input
- The command must work without the `forgia` CLI — fall back to `gh api` directly
- The command must fail gracefully when `gh` is not installed
