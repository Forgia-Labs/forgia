# Review Process

## Gates

Forgia enforces 3 gates before code reaches production:

```
FD Review ──→ Arch Review ──→ Threat Model ──→ SDD Generation ──→ Implementation ──→ Verification ──→ Close
 /fd-review   /fd-arch-review  /fd-threat-model  /fd-sdd           agent executes     /fd-verify       /fd-close
 GATE 1        (optional)       (optional)                                               GATE 2           GATE 3
```

### Gate 1: FD Review (`/fd-review`)

Before any SDD can be generated:
- Problem clearly defined (what + why, not how)
- At least 2 solutions considered
- Architecture diagram present
- Interfaces between components defined
- Constitution compliance checked

### Architecture Review (`/fd-arch-review`) — Optional

After FD approval, before SDD generation:
- Run `/fd-arch-review FD-NNN` to analyze the proposed architecture against the actual codebase
- Produces a structured report: pattern analysis, anti-pattern detection, dependency graph, SOLID compliance, recommendations
- **Advisory, not a gate** — the report informs but does not block `/fd-sdd`
- Recommended for FDs that introduce new components, change interfaces, or touch multiple packages
- Can be run repeatedly during FD authoring for iterative improvement

### Optional: Threat Model (`/fd-threat-model`)

Between FD Review and SDD Generation, run `/fd-threat-model FD-NNN` for security-sensitive FDs.

- **When to run**: recommended for FDs that involve authentication, authorization, data storage, external integrations, or user-facing APIs
- **What it produces**: `.forgia/fd/FD-NNN-threat-model.md` — a STRIDE-based security analysis with identified threats, risk ratings, and mitigation recommendations
- **Advisory, not blocking**: the FD can proceed to SDD generation without a threat model. `/fd-review` will flag its absence as a suggestion, not a failure
- **Downstream effects**: if a threat model exists, `/fd-sdd` automatically injects relevant mitigations into each SDD's constraints section

### Gate 2: Verification (`/fd-verify`)

After all SDDs are executed:
- All acceptance criteria met
- All tests pass
- Work Log completed in every SDD
- Constitution compliance re-checked
- Commits follow conventions

### Gate 3: Close (`/fd-close`)

Before archiving:
- FD status is "complete" (Gate 2 passed)
- Retrospective insights aggregated
- Changelog updated

## Who Reviews

| Reviewer | When |
|----------|------|
| Claude (AI) | Default reviewer for `/fd-review` |
| Human | Complex architectural decisions, security-sensitive FDs |
| Both | Recommended for high-priority FDs |

## Review Strictness

- **Be strict**: vague problems, missing diagrams, undefined interfaces = FAIL
- **Be specific**: every failed check includes actionable feedback
- **Be consistent**: always check against the constitution
- **No exceptions**: the gates cannot be bypassed

## Work Log Review

The Work Log in each SDD is reviewed during `/fd-verify`:
- Agent section: who, when, how long
- Decisions: any deviations from the plan
- Output: commits, PRs, files
- Retrospective: learnings for future FDs

The retrospective is the most valuable artifact — it feeds back into better FDs.

## MCP Server — Enhanced Analysis

When `forgia mcp serve` is running, slash commands (`/fd-review`, `/fd-arch-review`, etc.) gain access to the codebase knowledge graph via composite skills. This enables:

- **Blast radius analysis**: automatic risk labeling (High/Medium/Low) based on dependency count
- **Architecture coherence**: drift detection between actual call paths and documented architecture
- **Security scanning**: cross-references code patterns against guardrails deny.toml
- **Context mapping**: maps symbols and changes to bounded contexts

The MCP server exposes 7 composite skills alongside the 19 embedded slash commands. Run `forgia skills` to see the full list with mode indicators (Slash Command vs MCP Tool).
