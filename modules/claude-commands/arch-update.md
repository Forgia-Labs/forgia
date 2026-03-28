Update architecture and context files with what was actually implemented, based on SDD Work Logs from a closed FD.

## Instructions

Given FD identifier: $ARGUMENTS

1. **Parse the FD identifier**:
   - Extract the FD number from `$ARGUMENTS` (e.g., `FD-010`)
   - If no argument is provided, refuse: "Specificare un identificativo FD (es. `/arch-update FD-010`)."

2. **Pre-flight checks**:

   a. Verify `.forgia/architecture/` exists AND contains at least one `.yaml` file. If not, refuse and tell the user:
      "Nessuna architettura trovata. Esegui `/arch-init` per creare l'architettura di progetto."
      — then stop. Do NOT continue.

   b. Read the FD file from `.forgia/fd/FD-NNN*.md` (match the FD number from the argument).
      If the file does not exist, refuse: "FD non trovato: FD-NNN."

   c. Read the FD frontmatter and check the `status` field.
      If the status is NOT "closed", refuse:
      "FD-NNN non e' chiuso. Esegui `/fd-close FD-NNN` prima."
      — then stop. Do NOT continue.

3. **Read SDD Work Logs**:

   a. Find all SDD files in `.forgia/sdd/FD-NNN/SDD-*.md`
   b. For each SDD file, read the full content and extract:
      - The **Decisions / Decisioni** section (numbered list of decisions)
      - The **Retrospective / Retrospettiva** section (what worked, what didn't, suggestions)
      - The **Output** section (files created/modified)
   c. If ALL SDD Work Logs still contain only template placeholders (e.g., `<!-- decision 1: what and why -->` or `<!-- timestamp -->`), warn:
      "Nessun dato trovato nei Work Log — nulla da aggiornare."
      — then stop. Do NOT continue.
   d. Collect all filled Work Log entries into a structured list for analysis.

4. **Read current architecture and context files**:

   - Read all `.forgia/architecture/*.yaml` files:
     - `system-context.yaml`
     - `containers.yaml`
     - `technology-decisions.yaml`
     - `quality-attributes.yaml`
     - `constraints.yaml`
     - `glossary.yaml`
   - Read all `.forgia/contexts/*.yaml` files (excluding `_template.yaml`)
   - Do NOT read files matching guardrail deny patterns: `**/.env`, `**/*.pem`, `**/*.key`, `**/.ssh/id_*`, `**/.aws/credentials`

5. **Analyze Work Logs and compute changes**:

   For each filled Work Log entry, determine what architecture updates are needed based on these rules:

   | Work Log content | Target file | Update action |
   |-----------------|-------------|---------------|
   | New container/service introduced | `architecture/containers.yaml` | Add container entry following the schema: name, technology, description, bounded_context, ports, depends_on |
   | Container removed or renamed | `architecture/containers.yaml` | Update or remove the entry |
   | Interface was different than planned | `contexts/<name>.yaml` interfaces | Update interface contract, protocol, direction |
   | New dependency between contexts | `contexts/<name>.yaml` dependencies | Add to `depends_on` / `depended_by` |
   | Technology decision changed | `architecture/technology-decisions.yaml` | Update existing decision or add new one (next sequential ID: TD-NNN) |
   | Quality attribute measured | `architecture/quality-attributes.yaml` | Update with actual measured values |
   | New term introduced | `architecture/glossary.yaml` | Add term + meaning to the `terms` list |
   | Bounded context scope changed | `contexts/<name>.yaml` responsibilities | Update responsibilities list |

   For each potential change, record:
   - The **source**: which SDD and which Work Log excerpt motivated the change
   - The **target**: which file and field will be updated
   - The **action**: add, update, or remove
   - The **details**: the exact values to write

6. **Safety check — preview significant changes**:

   If more than 3 files will be modified, present the planned changes to the user BEFORE applying them:
   ```
   === Modifiche pianificate da FD-NNN ===

   File: .forgia/architecture/containers.yaml
     + Aggiunta container: <name> (da SDD-XXX: "<excerpt>")

   File: .forgia/contexts/<name>.yaml
     ~ Aggiornata interfaccia con <other>: protocol http → grpc (da SDD-YYY: "<excerpt>")

   Procedere con le modifiche? (Y/n)
   ```
   Wait for user confirmation before proceeding.

   If 3 or fewer files are affected, proceed directly.

7. **Apply updates**:

   For each change identified in step 5, modify the target YAML file:

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

8. **Report changes**:

   After all updates are applied, produce a detailed report:

   ```
   === Aggiornamento Architettura da FD-NNN ===

   Modifiche applicate:

   1. [containers.yaml] Aggiunto container "<name>"
      Motivo (SDD-XXX Decisioni #N): "<excerpt from Work Log>"

   2. [contexts/<name>.yaml] Aggiornata interfaccia con <other>
      Motivo (SDD-YYY Decisioni #M): "<excerpt from Work Log>"

   3. [glossary.yaml] Aggiunto termine "<term>"
      Motivo (SDD-ZZZ Decisioni #K): "<excerpt from Work Log>"

   Totale: N file modificati, M modifiche applicate.
   ```

   Each change MUST include the Work Log excerpt that motivated it — the user must see exactly what changed and why.

   If no changes were detected after analyzing all Work Logs:
   "Architettura aggiornata — nessuna modifica necessaria dai Work Log di FD-NNN."

9. **Trigger architecture review**:

   After all updates are applied and reported, automatically run `/arch-review` to validate the updated architecture:
   "Eseguo `/arch-review` per validare la coerenza dell'architettura aggiornata..."

   Then execute the `/arch-review` skill to perform the coherence checks.

## Important

- This skill **MODIFIES vault files** — unlike `/arch-review` (read-only), this skill writes to `architecture/` and `contexts/`
- Before modifying any YAML file, always read its current content first — never blindly overwrite
- Preserve existing YAML structure and comments — add/update fields, never rewrite entire files
- Each update must be traceable: report the SDD and Work Log excerpt that motivated it
- NEVER close or modify the FD file itself — this skill only updates architecture and context files
- Do NOT read files matching guardrail deny patterns: `**/.env`, `**/*.pem`, `**/*.key`, `**/.ssh/id_*`, `**/.aws/credentials`
- Do NOT write to: `.forgia/constitution.md`, `.forgia/config.toml`, `.forgia/guardrails/deny.toml`
- If a Work Log mentions something ambiguous, prefer NOT updating over making a wrong update — report it as a suggestion instead
- Generated/updated YAML must conform to the schemas in `modules/vault-template/`
