Compare competing Feature Designs and generate a structured decision report.

## Instructions

Given arguments: $ARGUMENTS

1. **Parse arguments**:
   - Extract 2 or more FD identifiers (e.g., `FD-001 FD-002` or `FD-a3f2 FD-b1c4`)
   - Detect `--post-to-issue` flag: if present, post the comparison as a GitHub issue comment
   - If fewer than 2 FD identifiers provided, refuse: "Serve almeno 2 FD da confrontare. Uso: /fd-compare FD-A FD-B [--post-to-issue]"

2. **Pre-flight checks**:
   - Verify `.forgia/fd/` exists
   - Read each specified FD file from `.forgia/fd/FD-NNN*.md`
   - If any FD file not found, refuse with the specific missing ID
   - Verify the FDs are related: they must share the same `upstream_issue` OR have each other in `competes_with`. If neither, warn: "Questi FD non sembrano essere in competizione (nessun upstream_issue condiviso e nessun competes_with). Procedere comunque? (Y/n)"

3. **Read each FD fully** — extract for comparison:
   - Frontmatter: id, title, status, priority, effort, impact, author, tags
   - Problem section: key points
   - Solutions: chosen solution with pros/cons
   - Architecture: mermaid diagrams
   - Interfaces: component table
   - Planned SDDs: count and breakdown
   - Constraints: key constraints
   - Verification criteria: list

4. **Generate comparison report** using this exact format:

```
=== FD Comparison: FD-A vs FD-B ===

## Dimensions

| Dimension | FD-A: <title> | FD-B: <title> |
|-----------|--------------|--------------|
| Author | <author A> | <author B> |
| Priority | <priority A> | <priority B> |
| Effort | <effort A> | <effort B> |
| Impact | <impact A> | <impact B> |
| SDDs planned | <count A> | <count B> |
| Status | <status A> | <status B> |

## Solutions Compared

### FD-A: <chosen solution title>
**Pro:** <pros from FD-A>
**Con:** <cons from FD-A>

### FD-B: <chosen solution title>
**Pro:** <pros from FD-B>
**Con:** <cons from FD-B>

## Architecture Comparison

### FD-A
<mermaid diagram from FD-A>

### FD-B
<mermaid diagram from FD-B>

## Acceptance Criteria Overlap

- N/M criteria are identical between FD-A and FD-B
- FD-A unique: <criteria only in A>
- FD-B unique: <criteria only in B>

## Trade-off Analysis

| Factor | FD-A | FD-B | Winner |
|--------|------|------|--------|
| <factor 1> | <A detail> | <B detail> | <A or B> |
| <factor 2> | <A detail> | <B detail> | <A or B> |

## Recommendation

Based on the analysis above, **FD-X** is recommended because:
1. <reason 1>
2. <reason 2>
3. <reason 3>

Risks of choosing FD-X over FD-Y:
- <risk 1>
- <risk 2>
```

5. **If `--post-to-issue` flag is present**:
   - Extract `upstream_issue` from the first FD (format: "owner/repo#number")
   - Parse owner, repo, and issue number
   - Post the comparison report as a comment using: `gh api repos/{owner}/{repo}/issues/{number}/comments -f body="<report>"`
   - Confirm: "Comparazione postata su <upstream_issue>"

6. **Show the report** to the user in the terminal.

## Important

- This is a **READ-ONLY** operation — never modify any FD file
- Be objective: the recommendation must be based on concrete trade-offs, not opinion
- If the FDs are too similar to differentiate, say so explicitly
- If one FD is clearly incomplete (missing sections, placeholder diagrams), note it as a weakness
- Support 3+ FDs (not just pairs) — the comparison table expands with more columns
- Do NOT read files matching guardrail deny patterns
