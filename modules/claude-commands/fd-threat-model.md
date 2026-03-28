Perform STRIDE-based threat modeling on a Feature Design and write the result to a persistent file.

## Instructions

Given FD identifier: $ARGUMENTS

1. **Pre-flight check**:

   a. Parse `$ARGUMENTS` to extract the FD identifier (e.g., `FD-005`).
   b. Find the matching FD file in `.forgia/fd/` — look for a file matching `FD-NNN-*.md` (e.g., `.forgia/fd/FD-005-*.md`).
   c. If no matching FD file exists, output: `"FD non trovato: $ARGUMENTS. Verifica l'identificatore."` — then **stop**. Do NOT create any file.

2. **Read vault context** — load all of these:

   - The FD file (full content: architecture, interfaces, components, constraints, SDD plan)
   - `.forgia/constitution.md` — especially the Security section
   - `.forgia/guardrails/deny.toml` — existing deny patterns (read, execute, write sections)
   - `.forgia/dev-guide/review-process.md` (if exists)
   - `.forgia/dev-guide/principles/clean-code.md` — error handling, validation at boundaries
   - `.forgia/dev-guide/lang/*.md` — all language-specific conventions (security patterns)
   - Actually **read the content** of every file — do not just list them.

3. **Scan codebase for existing security patterns**:

   - Search for authentication, authorization, input validation, encryption, and access control patterns in the project source code.
   - **CRITICAL**: Before reading any file, check its path against the `[read]` deny patterns in `.forgia/guardrails/deny.toml`. If a file matches a deny pattern, **skip it** and do not include it in context. Specifically, never read: `.env`, `*.pem`, `*.key`, `*.p12`, `*.pfx`, `*.jks`, `*.keystore`, `.aws/credentials`, `.ssh/id_*`, `.gnupg/**`, `credentials.json`, `*.tfstate`, or any other pattern listed in the `[read]` section.
   - Note any existing security mechanisms that are already in place (middleware, validators, sanitizers, encryption helpers).

4. **Identify Assets**:

   Based on the FD's architecture, interfaces, and data flow, identify:
   - Data assets (user data, configuration, tokens, API keys, session data)
   - Access credentials and authentication mechanisms
   - Infrastructure components (databases, message queues, file storage, APIs)
   - Trust boundaries (where data crosses from trusted to untrusted zones)

5. **Enumerate Threat Actors**:

   Based on the feature's attack surface, enumerate relevant threat actors:
   - Unauthenticated external users
   - Authenticated users with limited privileges
   - Compromised plugins or extensions
   - Malicious input (injection, crafted payloads)
   - Insider threats (compromised accounts, rogue agents)
   - Supply chain threats (compromised dependencies)
   - Only include actors relevant to the specific feature — do not list generic actors that have no bearing on this FD.

6. **Perform STRIDE Analysis**:

   For each relevant component identified in the FD's architecture, analyze all 6 STRIDE categories:

   - **S**poofing — Can an attacker impersonate another identity or component?
   - **T**ampering — Can data or code be modified without detection?
   - **R**epudiation — Can actions be performed without accountability/audit trail?
   - **I**nformation Disclosure — Can sensitive data leak through this component?
   - **D**enial of Service — Can availability be degraded or destroyed?
   - **E**levation of Privilege — Can an attacker gain unauthorized access levels?

   For each component, produce a table row for each identified threat. If a STRIDE category has no threats for a given component, explicitly state "No threats identified" for that category — all 6 categories must be covered per relevant component.

   Each threat row must include: Threat description, STRIDE Category (S/T/R/I/D/E), Component, Risk level (High/Medium/Low), and Mitigation.

7. **Generate SDD Recommendations**:

   Map identified mitigations to specific SDDs listed in the FD's "Planned SDDs / SDD Previsti" section. Each recommendation must:
   - Reference a specific SDD by number (e.g., "SDD-002")
   - Describe a concrete mitigation action (e.g., "add input sanitization for title field", "enforce rate limiting on API endpoint")
   - Be actionable — an agent reading the SDD should know exactly what security control to implement.
   - If a mitigation applies to a component not covered by any planned SDD, note it as a gap.

8. **Suggest Guardrails Additions**:

   Based on identified threats, propose concrete additions to `.forgia/guardrails/deny.toml`. Each suggestion must:
   - Specify the section (`[read]`, `[execute]`, or `[write]`)
   - Provide the exact glob/command pattern
   - Include a comment explaining why
   - Not duplicate patterns already in the current `deny.toml`
   - Example: `[execute] "kubectl exec*"  # Prevent direct container shell access`

9. **Write the threat model file** at `.forgia/fd/FD-NNN-threat-model.md` using this exact format:

```markdown
---
fd: "FD-NNN"
generated: "YYYY-MM-DD"
generator: "/fd-threat-model"
---

# Threat Model: FD-NNN — <FD title>

> Generated by `/fd-threat-model` on YYYY-MM-DD.
> Based on: <FD file path>

## 1. Assets

| Asset | Type | Location | Sensitivity |
|-------|------|----------|-------------|
| ... | Data / Credential / Infrastructure | Component or boundary | High / Medium / Low |

## 2. Threat Actors

| Actor | Motivation | Access Level | Relevance to FD |
|-------|-----------|--------------|-----------------|
| ... | ... | ... | Why this actor matters for this feature |

## 3. STRIDE Analysis

### <Component Name>

| Threat | Category | Risk | Mitigation |
|--------|----------|------|------------|
| ... | S / T / R / I / D / E | High / Medium / Low | ... |

<!-- Repeat for each relevant component -->

## 4. Recommendations for SDDs

1. **SDD-NNN**: <concrete mitigation action>
2. **SDD-NNN**: <concrete mitigation action>
...

## 5. Guardrails Additions

Proposed additions to `.forgia/guardrails/deny.toml`:

```toml
[section]
# Reason for addition
"pattern"
```

If no additions are needed, state: "No additional guardrails patterns identified — existing deny.toml coverage is sufficient."
```

10. **Report to the user**:

    - Confirm the file was created: `"Threat model creato: .forgia/fd/FD-NNN-threat-model.md"`
    - Summarize: number of assets, threat actors, STRIDE threats found, SDD recommendations, and guardrails suggestions
    - Suggest next steps:
      - "Rivedi il threat model e applica le raccomandazioni ai vincoli degli SDD"
      - "Usa `/fd-review $ARGUMENTS` per la revisione completa dell'FD"
      - "Usa `/fd-sdd $ARGUMENTS` per generare gli SDD (le mitigazioni verranno iniettate automaticamente)"

## Important

- This command **writes exactly one file**: `.forgia/fd/FD-NNN-threat-model.md` — it must NEVER modify the FD itself, constitution, config.toml, guardrails/deny.toml, or any codebase file
- Guardrails suggestions in the output are **advisory** — the user decides whether to apply them
- STRIDE analysis must cover **all 6 categories** per relevant component — no category may be silently skipped
- Each STRIDE category with no threats must explicitly state "No threats identified"
- SDD recommendations must reference **specific SDD numbers** from the FD's planned SDDs
- Guardrails suggestions must be **concrete `deny.toml` patterns**, not generic advice
- The threat model file is a **persistent artifact** consumed by `/fd-review` (SDD-002) and `/fd-sdd` (SDD-003)
- Do NOT read files matching `[read]` deny patterns in `.forgia/guardrails/deny.toml`
- Do NOT write to: `.forgia/constitution.md`, `.forgia/config.toml`, `.forgia/guardrails/deny.toml`
- If the FD has no architecture diagrams or component breakdown, note this as a limitation but still perform analysis based on available information
