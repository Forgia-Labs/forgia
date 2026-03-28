Review architecture coherence across the vault and produce a structured report.

## Instructions

This is a READ-ONLY operation. Never modify any vault file.

1. **Pre-flight check**:

   a. Verify `.forgia/architecture/` exists AND contains at least one `.yaml` file. If not, refuse and tell the user:
      "Architettura non trovata. Esegui `/arch-init` per creare l'architettura di progetto."
      — then stop. Do NOT continue.

2. **Read vault files** — load all of these that exist:

   - All `.forgia/architecture/*.yaml` files (system-context, containers, technology-decisions, quality-attributes, constraints, glossary)
   - All `.forgia/contexts/*.yaml` files (bounded contexts)
   - All `.forgia/fd/FD-*.md` files (feature designs — read frontmatter and body)
   - All `.forgia/learnings/*.yaml` files (if the directory exists)
   - Do NOT read files matching guardrail deny patterns: `**/.env`, `**/*.pem`, `**/*.key`, `**/.ssh/id_*`, `**/.aws/credentials`
   - Actually **read the content** of every discovered file — do not just list them.

3. **Perform coherence checks** — evaluate ALL 8 checks from the table below. For each check, determine if it passes, fails as error, or fails as warning. Record the exact file and field that caused any issue.

### Coherence Checks

| # | Check | Source | Severity | How to verify |
|---|-------|--------|----------|---------------|
| 1 | All containers in `containers.yaml` have a corresponding `contexts/*.yaml` | `architecture/containers.yaml` field `bounded_context` vs files in `contexts/` | **Warning** | For each container, check that a file `contexts/<bounded_context>.yaml` exists. If a container references a bounded context that has no corresponding context file, it is a warning. |
| 2 | All `contexts/*.yaml` interfaces are bidirectional (A→B implies B→A) | `contexts/` field `interfaces[].with` | **Error** | For each context A that declares an interface `with: B`, verify that context B also declares an interface `with: A`. The direction fields must be compatible (A outbound → B inbound, or both bidirectional). If A declares an interface with B but B has no matching interface back to A, report an error. |
| 3 | No orphan contexts (context exists but no container references it) | `architecture/containers.yaml` vs `contexts/` | **Warning** | For each context file in `contexts/`, check that at least one container in `containers.yaml` has `bounded_context` pointing to it. If a context file exists but no container references it, it is a warning. |
| 4 | FD scopes align with bounded context boundaries | `fd/FD-*.md` vs `contexts/` | **Error** | For each FD that is in-progress or approved (not draft), check if the components or services it modifies/creates fall within a single bounded context or explicitly declare cross-context coordination. If an FD implicitly crosses bounded context boundaries without acknowledging it (no mention of the relevant contexts in its body), report an error. |
| 5 | Technology decisions in FDs match `technology-decisions.yaml` | `fd/FD-*.md` vs `architecture/technology-decisions.yaml` | **Error** | If an FD proposes using a technology (language, database, framework, protocol) that contradicts an `accepted` decision in `technology-decisions.yaml`, report an error. FDs may propose NEW technologies not yet in the decisions file — that is acceptable but should be noted as informational. Only flag contradictions with existing accepted decisions. |
| 6 | Quality attributes are still achievable given implemented FDs | `architecture/quality-attributes.yaml` vs `fd/` (closed FDs) | **Warning** | Review closed FDs and their Work Log retrospectives. If a closed FD's outcome (actual performance, actual architecture) makes a quality attribute target clearly unachievable, report a warning. If no closed FDs exist or no quality attributes are defined, this check passes. |
| 7 | No FD contradicts decisions in closed FDs' Work Logs | `fd/` retrospectives | **Error** | For each in-progress or approved FD, check if it proposes something that was explicitly rejected or marked as a failure in a closed FD's Work Log retrospective. If a contradiction is found, report an error specifying both the current FD and the closed FD whose retrospective it contradicts. If no closed FDs exist, this check passes. |
| 8 | Glossary terms are used consistently across contexts | `architecture/glossary.yaml` vs `contexts/` | **Warning** | For each term in `glossary.yaml`, check if contexts use the same term with a different meaning in their `ubiquitous_language` section. If a context redefines a glossary term with a different meaning, report a warning. If the glossary is empty or contexts have no ubiquitous_language entries, this check passes. |

4. **Calculate coherence score**:
   - Start at 100
   - Subtract 15 for each **error**
   - Subtract 5 for each **warning**
   - Minimum score is 0 (never go negative)

5. **Produce the report** using this exact format:

```
=== Architecture Review ===

  Coherence Score: NN/100

  ERRORS (must fix):
    x [Check #N] description of error — file: <path>, field: <field>
    x [Check #N] description of error — file: <path>, field: <field>

  WARNINGS (review recommended):
    ! [Check #N] description of warning — file: <path>, field: <field>
    ! [Check #N] description of warning — file: <path>, field: <field>

  PASS:
    v [Check #N] description of passing check
    v [Check #N] description of passing check
```

   - If there are no errors, omit the ERRORS section entirely
   - If there are no warnings, omit the WARNINGS section entirely
   - Every check must appear exactly once — either as error, warning, or pass
   - Each error/warning must specify the exact file path and field that caused the issue
   - Sort errors and warnings by check number

6. **Empty/missing sections handling**:
   - If `contexts/` directory is empty or missing: report as a warning ("No bounded contexts defined"), and checks 1-4, 8 that depend on contexts all pass vacuously (no data to contradict)
   - If no FD files exist: checks 4-7 pass vacuously
   - If `learnings/` directory is empty or missing: not an issue, check 7 relies on FD Work Logs not learnings
   - If `glossary.yaml` is empty or has no terms: check 8 passes
   - If `quality-attributes.yaml` is empty: check 6 passes
   - If `technology-decisions.yaml` is empty or has no decisions: check 5 passes

## Important

- This is a **READ-ONLY** operation — never modify any vault file
- Be strict and specific: vague findings like "might be inconsistent" are not acceptable — cite exact files and fields
- Every check must produce a clear PASS, ERROR, or WARNING — no ambiguous results
- The report must be actionable: a developer reading it should know exactly what to fix
- Do NOT read files matching guardrail deny patterns
- Do NOT write to: `.forgia/constitution.md`, `.forgia/config.toml`, `.forgia/guardrails/deny.toml`
