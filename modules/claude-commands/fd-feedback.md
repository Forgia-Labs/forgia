Close the feedback loop for a completed Feature Design: read SDD Work Logs, route feedback to contexts/architecture/learnings, and trigger architecture review.

## Instructions

Given FD identifier: $ARGUMENTS

1. **Parse the FD identifier**:
   - Extract the FD number from `$ARGUMENTS` (e.g., `FD-010`)
   - If no argument is provided, refuse: "Specificare un identificativo FD (es. `/fd-feedback FD-010`)."

2. **Pre-flight checks**:

   a. Read the FD file from `.forgia/fd/FD-NNN*.md` (match the FD number from the argument).
      If the file does not exist, refuse: "FD non trovato: FD-NNN."

   b. Read the FD frontmatter and check the `status` field.
      If the status is NOT "closed", refuse:
      "FD-NNN non e' chiuso. Esegui `/fd-close FD-NNN` prima."
      — then stop. Do NOT continue.

   c. Check if `.forgia/architecture/` exists AND contains at least one `.yaml` file.
      If not, set `SKIP_ARCHITECTURE=true` and warn:
      "Nessuna architecture/ trovata — aggiornamenti architettura/contesti saltati. Learning record creato. Esegui `/arch-init` per scaffoldare l'architettura."
      — do NOT stop. Continue to create learnings file.

3. **Read SDD Work Logs**:

   a. Find all SDD files in `.forgia/sdd/FD-NNN/SDD-*.md`
   b. For each SDD file, read the full content and extract:
      - The **Decisions / Decisioni** section (numbered list of decisions)
      - The **Retrospective / Retrospettiva** section (what worked, what didn't, suggestions)
      - The **Output** section (files created/modified)
   c. If ALL SDD Work Logs still contain only template placeholders (e.g., `<!-- decision 1: what and why -->` or `<!-- timestamp -->`), set `EMPTY_WORKLOGS=true` and warn:
      "Nessun dato trovato nei Work Log — creo learning record minimo."
      — do NOT stop. Continue to create a minimal learnings file.
   d. Collect all filled Work Log entries into a structured list for analysis.

4. **Read current architecture and context files** (skip if `SKIP_ARCHITECTURE=true`):

   - Read all `.forgia/architecture/*.yaml` files:
     - `system-context.yaml`
     - `containers.yaml`
     - `technology-decisions.yaml`
     - `quality-attributes.yaml`
     - `constraints.yaml`
     - `glossary.yaml`
   - Read all `.forgia/contexts/*.yaml` files (excluding `_template.yaml`)
   - Do NOT read files matching guardrail deny patterns: `**/.env`, `**/*.pem`, `**/*.key`, `**/.ssh/id_*`, `**/.aws/credentials`

5. **Read existing learnings** (if any):

   - Check if `.forgia/learnings/FD-NNN.yaml` already exists
   - If it exists, set `MERGE_MODE=true` and warn:
     "Learning record gia' esistente — unisco nuove voci senza sovrascrivere."
   - Read the existing file content for merging

6. **Analyze Work Logs and route feedback**:

   If `EMPTY_WORKLOGS=true`, skip to step 8 (create minimal learnings file).

   For each filled Work Log entry, determine routing based on this table:

   | Work Log content | Routes to | Action |
   |-----------------|-----------|--------|
   | Interface was different than planned | `contexts/<name>.yaml` interfaces | Update contract, protocol, direction with actual values |
   | New dependency discovered | `contexts/<name>.yaml` dependencies | Add to `depends_on` / `depended_by` |
   | New container/service introduced | `architecture/containers.yaml` | Add container entry |
   | Technology decision changed | `architecture/technology-decisions.yaml` | Update or add decision record |
   | Latency/performance measured | `architecture/quality-attributes.yaml` | Update with actual measured values |
   | New term introduced | `architecture/glossary.yaml` | Add term + meaning |
   | Failure mode discovered | `learnings/FD-NNN.yaml` failure_modes | Add pattern, context, outcome, recommendation |
   | Pattern that worked well | `learnings/FD-NNN.yaml` successful_patterns | Add pattern, context, outcome |
   | Suggestion for future FDs | `learnings/FD-NNN.yaml` suggestions | Add suggestion string |
   | Interface correction (planned vs actual) | `learnings/FD-NNN.yaml` interface_corrections | Add context, interface, planned, actual |
   | New term from Work Log | `learnings/FD-NNN.yaml` new_terms | Add term + meaning |

   For each potential change, record:
   - The **source**: which SDD and which Work Log excerpt motivated the change
   - The **target**: which file and field will be updated
   - The **action**: add, update, or remove
   - The **details**: the exact values to write

   If a Work Log entry is ambiguous, prefer NOT updating over making a wrong update — route it as a suggestion in the learnings file instead.

7. **Apply architecture and context updates** (skip if `SKIP_ARCHITECTURE=true`):

   Follow the SAME update rules as `/arch-update`:

   - **YAML modification rules**:
     - Read the current file content before modifying
     - Preserve existing structure — add/update fields, never rewrite the entire file
     - Preserve comments in YAML files
     - Use 2-space indentation (no tabs)
     - Quote strings that contain special characters (`:`, `#`, `{`, `}`, `[`, `]`, `,`, `&`, `*`, `?`, `|`, `-`, `<`, `>`, `=`, `!`, `%`, `@`, `\`)
     - Empty arrays: `[]` on the same line
     - No trailing whitespace

   - **Adding a container** to `containers.yaml`:
     ```yaml
     - name: "<container-name>"
       technology: "<tech>"
       description: "<description from Work Log>"
       bounded_context: "<context-name>"
       ports: []
       depends_on: []
     ```

   - **Adding a technology decision** to `technology-decisions.yaml`:
     ```yaml
     - id: "TD-NNN"           # next sequential ID
       title: "<title>"
       status: "accepted"     # it was actually implemented
       date: "<date from Work Log>"
       context: "<context from Work Log>"
       decision: "<what was decided>"
       alternatives: []
       consequences: []
     ```

   - **Adding a glossary term** to `glossary.yaml`:
     ```yaml
     - term: "<term>"
       meaning: "<meaning from Work Log>"
     ```

   - **Updating a context** in `contexts/<name>.yaml`:
     - For interfaces: update or add the interface entry in the `interfaces` list
     - For dependencies: add to `depends_on` or `depended_by` as appropriate
     - For responsibilities: add or update items in the `responsibilities` list
     - For ubiquitous_language: add new terms if introduced

   - **Updating quality attributes** in `quality-attributes.yaml`:
     - If an attribute was measured, update its `target` or add a new field `actual` with the measured value

   - **Safety check**: If more than 3 files will be modified, present the planned changes to the user BEFORE applying them and wait for confirmation.

8. **Create or merge learnings file**:

   Create `.forgia/learnings/FD-NNN.yaml` conforming to `modules/vault-template/learnings/_template.yaml` schema.

   Read the FD title and closed date from the FD frontmatter.

   **If `EMPTY_WORKLOGS=true`** — create a minimal learnings file:
   ```yaml
   fd: "FD-NNN"
   title: "<title from FD>"
   closed: "<date from FD frontmatter or today>"

   failure_modes: []

   successful_patterns: []

   suggestions: []

   interface_corrections: []

   new_terms: []
   ```

   **If `MERGE_MODE=true`** — read the existing file and APPEND new entries to each array. Do NOT overwrite existing entries. Do NOT duplicate entries that already exist.

   **Otherwise** — create the full learnings file with all extracted data:
   ```yaml
   fd: "FD-NNN"
   title: "<title from FD>"
   closed: "<date from FD frontmatter or today>"

   failure_modes:
     - pattern: "<what went wrong>"
       context: "<in which situation>"
       outcome: "<what happened as a result>"
       recommendation: "<how to avoid it next time>"

   successful_patterns:
     - pattern: "<what worked well>"
       context: "<in which situation>"
       outcome: "<the positive result>"

   suggestions:
     - "<free-form improvement idea>"

   interface_corrections:
     - context: "<which bounded context>"
       interface: "<which interface>"
       planned: "<what was originally planned>"
       actual: "<what was actually implemented>"

   new_terms:
     - term: "<new term>"
       meaning: "<its definition>"
   ```

   YAML rules:
   - 2-space indentation (no tabs)
   - Quote strings that contain special characters
   - Empty arrays: `[]` on the same line
   - No trailing whitespace

9. **Report results**:

   Produce a detailed report clearly separating the three routing destinations:

   ```
   === Feedback Loop — FD-NNN ===

   CONTESTI aggiornati:
     1. [contexts/<name>.yaml] <description of change>
        Motivo (SDD-XXX Decisioni #N): "<excerpt from Work Log>"

   ARCHITETTURA aggiornata:
     1. [containers.yaml] <description of change>
        Motivo (SDD-XXX Decisioni #N): "<excerpt from Work Log>"

   LEARNING RECORD:
     File: .forgia/learnings/FD-NNN.yaml
     - N failure modes
     - N successful patterns
     - N suggestions
     - N interface corrections
     - N new terms

   Totale: N aggiornamenti contesti, N aggiornamenti architettura, 1 learning record.
   ```

   Each architecture/context change MUST include the Work Log excerpt that motivated it.

   If a section had no changes, report it as:
   ```
   CONTESTI aggiornati: nessuna modifica.
   ```

   If `SKIP_ARCHITECTURE=true`, report:
   ```
   CONTESTI aggiornati: saltati (nessuna architecture/ trovata).
   ARCHITETTURA aggiornata: saltata (nessuna architecture/ trovata).
   ```

10. **Trigger architecture review**:

    After all updates are applied and reported, automatically run `/arch-review` to validate the updated architecture:
    "Eseguo `/arch-review` per validare la coerenza dell'architettura aggiornata..."

    Then execute the `/arch-review` skill to perform the coherence checks.

    If `SKIP_ARCHITECTURE=true`, skip the review and report:
    "Nessuna architettura presente — `/arch-review` saltato."

## Difference from /arch-update

`/arch-update` focuses on architecture/context YAML updates only.
`/fd-feedback` does the same updates PLUS creates the learnings record.

In practice, run `/fd-feedback` after `/fd-close` (it is the superset).
If only architecture updates are needed without learnings, use `/arch-update`.

## Important

- This skill **MODIFIES vault files** — writes to `architecture/`, `contexts/`, and `learnings/`
- Before modifying any YAML file, always read its current content first — never blindly overwrite
- Preserve existing YAML structure and comments — add/update fields, never rewrite entire files
- Each update must be traceable: report the SDD and Work Log excerpt that motivated it
- NEVER close or modify the FD file itself — this skill only updates architecture, contexts, and learnings
- Do NOT read files matching guardrail deny patterns: `**/.env`, `**/*.pem`, `**/*.key`, `**/.ssh/id_*`, `**/.aws/credentials`
- Do NOT write to: `.forgia/constitution.md`, `.forgia/config.toml`, `.forgia/guardrails/deny.toml`
- If a Work Log mentions something ambiguous, prefer NOT updating over making a wrong update — report it as a suggestion in the learnings file instead
- Generated/updated YAML must conform to the schemas in `modules/vault-template/`
- The learnings YAML must conform to `modules/vault-template/learnings/_template.yaml`
- On non-closed FD: refuse immediately with clear message
- On empty Work Logs: create minimal learnings file (not an error — some FDs have clean executions)
- On missing architecture/: create learnings file, skip architecture/context updates, warn the user
- On existing learnings file: merge (append to arrays), never overwrite
