Create a new Feature Design (FD) in the Forgia vault.

## Instructions

1. Parse arguments from: $ARGUMENTS
   Format: `[description or GitHub/GitLab issue URL/reference] [--repo owner/repo] [--gitlab-host hostname] [--context path]`
   - First argument can be:
     - A **GitHub issue reference**: URL (`https://github.com/owner/repo/issues/42`), `#42`, or just `42`
     - A **GitLab issue reference**: URL (`https://gitlab.com/owner/repo/-/issues/42`, `https://custom.host/owner/repo/-/issues/42`), or `gl#42`
     - A **free-text description** of the feature/task
     - Omitted — ask the user for a brief description
   - `--repo owner/repo` (optional): target repository for issue fetch. If omitted, infer from `git remote get-url origin`
   - `--gitlab-host hostname` (optional): GitLab host for self-hosted instances when using `gl#N` or bare number with GitLab target. Default: `gitlab.com`. Not needed when a full URL is provided (host is extracted from the URL).
   - `--context path` (optional): additional context file or directory to read

   **Platform detection rules:**
   - Argument matches `https://github.com/...` → GitHub
   - Argument matches `https?://[host]/[path]/-/issues/[0-9]+` → GitLab; extract host, owner/repo path, issue number
   - Argument matches `gl#[0-9]+` → GitLab; use `--gitlab-host` (default `gitlab.com`) and `--repo`
   - Argument matches `#[0-9]+` or bare `[0-9]+` → GitHub (default)
   - Anything else → free-text
   - Reject malformed references with: `"Riferimento issue non riconosciuto: <argument>. Usa un URL GitHub/GitLab, #N, gl#N, o una descrizione testuale."`

2. **Pre-flight checks**:
   - Verify `.forgia/fd/` exists. If not, tell the user: "Vault non inizializzato. Esegui `/project-init` prima."
   - If the argument is a **GitHub issue reference**, verify `gh` CLI is installed by running `command -v gh`. If not found, tell the user: "gh CLI non trovato. Installa GitHub CLI: https://cli.github.com/"
   - If the argument is a **GitLab issue reference**, verify that at least one of the following is available:
     - `GITLAB_TOKEN` environment variable is set (non-empty)
     - `glab` CLI is installed (`command -v glab`)
       If neither is available, tell the user: "GitLab source richiede GITLAB_TOKEN o glab CLI. Installa glab: https://gitlab.com/gitlab-org/cli oppure esporta GITLAB_TOKEN." — then stop execution and create no FD file.

3. **Determine the source** — GitHub issue, GitLab issue, or free-text:

   ### Path A: GitHub issue source

   Fetch the issue data using `gh api`:

   ```
   gh api repos/{owner}/{repo}/issues/{number}
   ```

   - Extract: `title`, `body`, `labels` (array of name strings), `assignee.login`, `milestone.title`
   - Fetch comments: `gh api repos/{owner}/{repo}/issues/{number}/comments`

   ### Path A-bis: GitLab issue source

   URL-encode the project path: replace every `/` in `owner/repo` with `%2F` (e.g. `group/subgroup/repo` → `group%2Fsubgroup%2Frepo`). Use `{host}` extracted from the URL or from `--gitlab-host` (default `gitlab.com`).

   **Primary path — `GITLAB_TOKEN` is set:**

   Fetch the issue:

   ```
   curl -sf \
     -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
     "https://{host}/api/v4/projects/{owner%2Frepo}/issues/{number}"
   ```

   Fetch the notes (comments):

   ```
   curl -sf \
     -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
     "https://{host}/api/v4/projects/{owner%2Frepo}/issues/{number}/notes?per_page=20"
   ```

   **Fallback path — `glab` is available (no `GITLAB_TOKEN`):**

   Fetch the issue:

   ```
   glab api projects/{owner%2Frepo}/issues/{number}
   ```

   Fetch the notes:

   ```
   glab api "projects/{owner%2Frepo}/issues/{number}/notes?per_page=20"
   ```

   Extract from the issue response:

   | GitLab field                     | Maps to                                         |
   | -------------------------------- | ----------------------------------------------- |
   | `title`                          | FD `title`                                      |
   | `description`                    | Problem section body (rewrite, do not dump raw) |
   | `labels[]` (array of strings)    | `tags` frontmatter array                        |
   | `assignees[0].username`          | `assignee` frontmatter                          |
   | `milestone.title`                | Constraint note                                 |
   | notes `body` + `author.username` | Notes section (summarized)                      |

   Never echo or log `$GITLAB_TOKEN` in any output visible to the user.

   ### Path B: Free-text description

   Use the user's description as the basis for the FD. Ask clarifying questions if the description is too vague.

4. **Discover local context**:
   - Scan for these files and read any that exist:
     - `docs/decisions/*.md` (ADRs)
     - `CONTRIBUTING.md`
     - `ARCHITECTURE.md`, `docs/ARCHITECTURE.md`
     - `.forgia/constitution.md`
     - `.forgia/dev-guide/**/*.md`
     - `CODE_OF_CONDUCT.md`
     - `docs/*.md`
   - If `--context` argument was provided, also read files at that path
   - **Actually read the content** of all discovered files — do not just list them

5. **Read the FD template** from `.forgia/fd/_templates/fd-template.md` (fallback: `modules/vault-template/fd/_templates/fd-template.md`) to know the exact section structure. The generated FD MUST have ALL sections from this template.

6. **Determine the next FD ID**:
   - List existing FD files in `.forgia/fd/` (excluding `_templates/`)
   - Find the highest FD number and increment by 1
   - Format: `FD-NNN` (zero-padded to 3 digits)

7. **Create the FD** at `.forgia/fd/FD-NNN-kebab-title.md`:

   ### Frontmatter
   - `id`: the new FD ID
   - `title`: cleaned title (from issue or description)
   - `status`: "planned"
   - `priority`: if from issue — "high" if labels contain "bug", else "medium"; if from description — "medium"
   - `effort`: "medium" (default)
   - `impact`: "medium" (default)
   - `author`: detect automatically — try `git config user.name`, fallback to `gh api user --jq '.login'`, fallback to empty
   - `assignee`: from issue assignee if available, or empty
   - `created`: today's date
   - `reviewed`: false
   - `reviewer`: ""
   - `tags`: from issue labels if available, else `[]`
   - `upstream_issue`: if from an issue source, use format `"owner/repo#number"`:
     - GitHub: `"owner/repo#42"`
     - GitLab on `gitlab.com`: `"owner/repo#42"` (same convention)
     - GitLab on a self-hosted instance: prefix with the hostname — `"git.example.com/owner/repo#42"`
     - Omit entirely if from free-text description

   ### Problem / Problema
   - If from issue: restructure the issue body into a clear problem statement. Extract the "what" and "why" — remove implementation details, code snippets, workarounds. Do NOT dump the raw issue body — rewrite it as a proper problem definition.
   - If from description: write a clean problem statement from the user's input.
   - Always write from the user's perspective, technology-agnostic (what, not how).

   ### Solutions Considered / Soluzioni Considerate
   - Identify solutions mentioned in the issue body/comments or user description
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
   - Create testable verification criteria based on the requirements
   - Each criterion should be a checkbox item

   ### Notes / Note
   - If from issue: link to the upstream issue (`Upstream: owner/repo#number`) and include relevant maintainer comments (summarized, not raw dumps)
   - List all auto-discovered context files as links

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
- The command must fail gracefully when `gh` is not installed (only affects issue path)
