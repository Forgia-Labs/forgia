---
title: "Feature Design"
description: "What an FD is, when to create one, and how it fits in the Forgia workflow."
navigation:
  title: "FD Guide"
---

# Feature Design (FD)

A Feature Design is a **functional design document written for humans**. It captures decisions, trade-offs, and architecture — not implementation details.

```
FD answers: WHAT needs to be built, and WHY
SDD answers: HOW an agent should build one component
```

## What an FD contains

| Section | Purpose |
|---------|---------|
| **Problem** | What is broken or missing, and why it matters |
| **Solutions Considered** | At least 2 options with pros/cons, one chosen |
| **Architecture** | Integration context diagram + data flow diagram (Mermaid) |
| **Interfaces** | Component contracts — inputs, outputs, protocols |
| **Planned SDDs** | Which components will become separate execution specs |
| **Constraints** | Technical, time, and budget limits |
| **Verification** | Testable acceptance criteria |

## Lifecycle

```
planned → approved → in-progress → complete → closed
              ↑
         /fd-review (mandatory gate)
```

An FD can also be `rejected` (loses a competitive review) or `abandoned` (team decides to drop it).

## When to create an FD

Create an FD for any non-trivial change: new features, significant refactors, architectural decisions. Skip it for typo fixes, dependency bumps, or single-file changes with no design decisions.

## Pages

- [Creating an FD](/docs/fd/creating) — `/fd-new` from issue, free-text, or manual
- [Reviewing an FD](/docs/fd/reviewing) — what `/fd-review` checks
- [Closing an FD](/docs/fd/closing) — `/fd-verify` and `/fd-close`
