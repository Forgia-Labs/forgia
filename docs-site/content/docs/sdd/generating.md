---
title: "Generating SDDs"
description: "Use /fd-sdd to generate agent-ready execution specs from an approved Feature Design."
navigation:
  title: "Generating"
---

# Generating SDDs

```
/fd-sdd FD-001
```

`/fd-sdd` reads the approved FD and produces one SDD file per planned component.

## Requirements

The FD must have `status: approved` and `reviewed: true`. If not, `/fd-sdd` refuses and directs you to run `/fd-review` first.

## What gets generated

For an FD with 3 planned components, `/fd-sdd` creates:

```
.forgia/sdd/FD-001/
├── SDD-001-component-a.md
├── SDD-002-component-b.md
└── SDD-003-integration-wiring.md   ← always added automatically
```

Each SDD is populated with:

- **Scope** derived from the FD component description
- **Interfaces** from the FD's interface definitions
- **Constraints** from the FD's architecture section (language, framework, versions)
- **Best practices** from the dev-guide and constitution
- **Test requirements** specific to that component
- **Acceptance criteria** traceable to the FD's verification section
- **Context** — pointers to existing files the agent must read first
- **Work Log** — empty, ready to be filled by the executing agent

## The integration wiring SDD

`/fd-sdd` always appends an integration wiring SDD, even if the FD doesn't list one. This SDD covers:

- Wiring: `pub mod`, `use`, dependency injection registration
- Startup path: who calls what, in what order
- E2E test: full flow from the public entry point to the innermost component

This is not optional. History shows that components passing unit tests fail to work together when integration wiring is left implicit.

## After generation

Review each generated SDD for completeness:

- Scope should be narrow enough for one agent session
- Interfaces should match between adjacent SDDs
- Acceptance criteria should be objectively verifiable

Then execute: see [Executing SDDs](/docs/sdd/executing).
