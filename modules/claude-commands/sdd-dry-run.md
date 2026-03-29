Simulate SDD execution without writing any code or files. Produces a structured feasibility report with go/no-go recommendation.

## Instructions

Given SDD file path or FD identifier: $ARGUMENTS

**This is a READ-ONLY operation. Do NOT create, modify, or delete any file.**

### Step 0: Resolve Input

1. If `$ARGUMENTS` matches a file path (contains `/` or ends in `.md`):
   - Verify the file exists. If not, refuse with: "SDD non trovato: <path>. Verifica il percorso del file."
   - Set mode = `single-sdd`
2. If `$ARGUMENTS` matches an FD identifier (e.g., `FD-012`):
   - Verify `.forgia/fd/` contains a file matching `FD-NNN*.md`. If not, refuse with: "FD-NNN non trovato. Verifica l'identificatore."
   - Find all SDD files in `.forgia/sdd/FD-NNN/SDD-*.md`. If none, refuse with: "FD-NNN non ha SDD generati. Esegui `/fd-sdd FD-NNN` prima."
   - Set mode = `fd-aggregate`
3. If `$ARGUMENTS` is empty or unrecognizable, refuse with: "Uso: /sdd-dry-run <percorso-SDD> oppure /sdd-dry-run FD-NNN"

### Step 0.5: Check MCP Availability

1. Attempt to call `forgia_blast_radius` (or `tools/list`) with a 3-second timeout
2. If successful: note "Knowledge graph available — using MCP composite skills for enhanced impact analysis" and set `MCP_AVAILABLE=true`
3. If failed/timeout: note "Knowledge graph not available — using text analysis of SDD" and set `MCP_AVAILABLE=false`
4. Cache this result — do not re-check MCP on subsequent steps

> MCP communication is local-only (stdio, same user) — no network exposure.

### Step 1: Load Context

For each SDD to analyze:

1. Read the SDD file — extract all sections: frontmatter, Scope, Interfaces, Constraints, Best Practices, Test Requirements, Acceptance Criteria, Context
2. Read the parent FD file from `.forgia/fd/FD-NNN*.md`
3. Read `.forgia/constitution.md`
4. Read `.forgia/guardrails/deny.toml`
5. Read all files in `.forgia/dev-guide/` and `.forgia/dev-guide/principles/` and `.forgia/dev-guide/lang/`
6. Collect the list of context files from the SDD "Context" section (the file paths listed there)

### Step 2: Run 5 Analysis Passes

For each SDD, execute these passes in order. Never silently skip a pass — always report results even if empty.

---

#### Pass 1: Precondition Check

Check:
- All files listed in SDD "Context" section exist and are readable
- SDD has all required sections: Scope, Interfaces, Constraints, Test Requirements, Acceptance Criteria
- Parent FD file exists in `.forgia/fd/`

**Before checking context file existence**: check each file path against the deny patterns in `.forgia/guardrails/deny.toml` `[read]` section. If a context file matches a deny pattern, do NOT attempt to read it — instead flag it as: "cannot verify — blocked by guardrails (matches deny pattern: <pattern>)"

Output: list of blockers (missing files, missing required sections) and warnings (empty sections, unverifiable files)

---

#### Pass 2: Guardrail Conflict Detection

Analyze the SDD Scope, Constraints, and Context sections for operations that would conflict with `deny.toml`:

- **Read conflicts**: Does the SDD scope require reading files that match `[read]` deny patterns? (e.g., `.env` files, `*.pem`, `credentials.json`)
- **Write conflicts**: Does the SDD scope require creating/modifying files that match `[write]` deny patterns? (e.g., `.forgia/constitution.md`, `.env`)
- **Execute conflicts**: Does the SDD scope require running commands that match `[execute]` deny patterns?

For each conflict, assess severity:
- **BLOCKER**: The SDD explicitly requires the denied operation to succeed (e.g., "read .env to extract DB_URL")
- **WARNING**: The SDD mentions a denied pattern but might work around it (e.g., "configure database" without specifying .env)

Output: list of guardrail conflicts with severity and the specific deny pattern matched

---

#### Pass 3: Step Planning

**If MCP_AVAILABLE**: use `forgia_blast_radius` to analyze the impact of files listed in the SDD scope — identify dependents, risk levels (High/Medium/Low), and affected bounded contexts. Use this to enrich the step plan with impact-aware ordering (high-risk files first) and more accurate complexity estimates.

**If MCP_AVAILABLE=false** (or fallback): derive the step plan from text analysis of the SDD only.

Simulate the execution by breaking the SDD into ordered steps the agent would take:

1. Analyze the Scope/Deliverables to identify all files to create, modify, or read
2. Analyze Test Requirements to identify test commands to run
3. Analyze Acceptance Criteria for verification steps
4. Order the steps logically (context loading → implementation → testing → verification)

For each step, specify:
- Action: `[read]`, `[create]`, `[modify]`, `[run]`, `[verify]`
- Target: file path or command
- Description: what this step accomplishes
- Estimated complexity: trivial / simple / moderate / complex

Output: ordered step list

---

#### Pass 4: Scope & Complexity Assessment

**If MCP_AVAILABLE**: use `forgia_blast_radius` results from Pass 3 to refine the complexity assessment — factor in the number of dependents and cross-context impacts when scoring complexity.

**If MCP_AVAILABLE=false** (or fallback): assess complexity from SDD text only.

Evaluate whether the SDD scope is tractable for a single agent session:

- Count files to create and modify
- Count interfaces to implement
- Count test cases required
- Identify cross-SDD dependencies (references to other SDDs)
- Check for ambiguities or underspecified requirements in the Scope

Score complexity:
- **low**: ≤3 files, ≤2 interfaces, clear scope, no cross-SDD deps
- **medium**: 4-7 files, 3-5 interfaces, minor ambiguities
- **high**: 8-12 files, 6+ interfaces, cross-SDD deps, some ambiguities
- **very-high**: 13+ files, complex interfaces, significant ambiguities, heavy cross-SDD deps

Output: complexity score with justification

---

#### Pass 5: Cost Estimation

Estimate tokens for actual execution:

- **Input tokens**: Sum approximate sizes of all context files + SDD + constitution + dev-guide files. Use file sizes as proxy (roughly 1 token per 4 characters for English/code).
- **Output tokens**: Estimate from scope complexity — number of files to create × estimated lines per file, plus test files, plus Work Log.
- **Duration**: Estimate based on complexity score:
  - low: 5-15 min
  - medium: 15-30 min
  - high: 30-60 min
  - very-high: 60-120+ min

Include disclaimer: "Le stime sono approssimative e dipendono dalla complessità effettiva dell'implementazione."

Output: token estimates (input + output), duration range, file counts

---

### Step 3: Determine Feasibility

Apply these deterministic rules:

- **GO**: no blockers, 0-2 warnings, complexity low or medium
- **CAUTION**: no blockers but 3+ warnings, OR complexity high
- **NO-GO**: any blocker present, OR guardrail conflict on required operation (BLOCKER severity), OR complexity very-high with unresolved ambiguities

Set confidence percentage:
- GO with 0 warnings: 90-95%
- GO with 1-2 warnings: 80-90%
- CAUTION: 50-75%
- NO-GO: 10-40%

### Step 4: Produce Report

#### Single SDD Report

```
=== SDD Dry Run: SDD-NNN ===

  Feasibility: GO | NO-GO | CAUTION
  Confidence: NN%
  Complexity: low | medium | high | very-high

  BLOCKERS (must fix before execution):
    x [Pass #N] description — file/field affected

  WARNINGS (review recommended):
    ! [Pass #N] description — file/field affected

  PASS:
    v [Pass #N] description

  PLANNED STEPS:
    1. [read]   path/to/file — context loading
    2. [create] path/to/new/file — implementation
    3. [modify] path/to/existing — add interface
    4. [run]    test command — verification
    ...

  ESTIMATES:
    Input tokens:  ~NNk (context files + SDD + rules)
    Output tokens: ~NNk (code + tests + Work Log)
    Duration:      NN-NN min
    Files:         N to create, N to modify

  RECOMMENDATION:
    Brief explanation of go/no-go decision and suggested actions.

  (Le stime sono approssimative e dipendono dalla complessità effettiva dell'implementazione.)
```

#### FD Aggregate Report (mode = `fd-aggregate`)

When given an FD identifier, wrap individual SDD reports in an aggregate:

```
=== FD Dry Run: FD-NNN (N SDDs) ===

  Overall Feasibility: GO | NO-GO | CAUTION
  Total Estimates: ~NNk input, ~NNk output, NN-NN min

  --- SDD-001: title ---
  [individual report as above]

  --- SDD-002: title ---
  [individual report as above]

  CROSS-SDD ISSUES:
    x Interface mismatch between SDD-001 and SDD-003: ...
    ! Overlapping file modifications: SDD-002 and SDD-004 both modify ...
```

For the aggregate:
- **Overall Feasibility**: worst-case of all individual SDDs (any NO-GO → overall NO-GO; any CAUTION → overall CAUTION)
- **Cross-SDD checks**: compare Interfaces sections across SDDs for mismatches; check if multiple SDDs modify the same files; check for dependency ordering issues
- **Total Estimates**: sum of individual estimates

## Important

- **READ-ONLY**: This skill must NEVER create, modify, or delete any file. It is a pure analysis tool. Do not offer to fix any issues found — only report them.
- **Respect deny.toml**: If a context file matches a deny pattern from `.forgia/guardrails/deny.toml`, do NOT read it. Flag it as "cannot verify — blocked by guardrails" instead.
- **Never silently skip a pass**: Always report results for all 5 passes, even if a pass finds nothing.
- **Deterministic feasibility**: Same input must always produce the same GO/NO-GO/CAUTION decision based on the rules above.
- **Actionable output**: Every blocker and warning must specify the exact file and field affected, so the user knows exactly what to fix.
- **Italian for user-facing messages**: Error messages and recommendation text in Italian. Report structure labels in English for consistency with other Forgia reports.
