---
title: "Reviewing an FD"
description: "What /fd-review checks, why it's a mandatory gate, and how to fix failures."
navigation:
  title: "Reviewing"
---

# Reviewing an FD

```
/fd-review FD-001
```

`/fd-review` is a **mandatory gate**. An FD cannot proceed to SDD generation without passing it.

## What gets checked

### Problem definition
- Problem is clearly defined — not vague or too broad
- Describes WHAT and WHY, not HOW
- At least 2 solutions considered with pros/cons
- Chosen solution is justified

### Architecture diagrams (mandatory)
- **Integration Context** diagram present — shows where the feature sits in the existing system (existing components grey, new ones green)
- **Data Flow** diagram present — sequence diagram showing component interactions
- Diagrams are not template placeholders — they reflect the actual feature
- Mermaid syntax is correct

### SDD planning
- SDD breakdown listed (one per component/service)
- Each SDD has a clear, isolated scope
- Interfaces between SDDs are defined
- Last SDD covers integration wiring and E2E verification

### Constitution & security
- Respects all rules in `constitution.md`
- No proposed interfaces expose secrets
- No file patterns would violate `guardrails/deny.toml`

### Verification
- Acceptance criteria are present and objectively testable

## On pass

FD status changes to `approved`. Work can proceed to `/fd-sdd`.

## On failure

The review lists every failed check with specific feedback. The FD stays in `planned`. Fix the issues and re-run.

Common failures:
- **Vague problem**: add concrete user impact, not just "the system is slow"
- **Missing diagrams**: draw the actual component topology, not the template placeholder
- **Missing integration wiring SDD**: add a final SDD that wires everything together and includes an E2E test
