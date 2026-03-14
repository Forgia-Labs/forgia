# Review Process

## Gates

Forgia enforces 3 gates before code reaches production:

```
FD Review ──→ SDD Generation ──→ Implementation ──→ Verification ──→ Close
 /fd-review    /fd-sdd            agent executes     /fd-verify       /fd-close
 GATE 1                                              GATE 2           GATE 3
```

### Gate 1: FD Review (`/fd-review`)

Before any SDD can be generated:
- Problem clearly defined (what + why, not how)
- At least 2 solutions considered
- Architecture diagram present
- Interfaces between components defined
- Constitution compliance checked

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
