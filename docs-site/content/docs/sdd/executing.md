---
title: "Executing SDDs"
description: "Assign SDDs to agents and run them interactively, autonomously, or in parallel."
navigation:
  title: "Executing"
---

# Executing SDDs

## Assign an SDD

```
/sdd-assign SDD-001
```

You'll be asked to choose an agent backend:

| Backend | Best for |
|---------|---------|
| `claude-code` | Complex SDDs needing human-in-the-loop |
| `openhands` | Isolated, well-defined SDDs in Docker |
| `manual` | Developer implements; SDD serves as spec |

`/sdd-assign` updates the SDD frontmatter (`agent`, `assigned_to`, `status: assigned`) and launches the appropriate workflow.

## Interactive execution (Claude Code)

```
/sdd-assign SDD-001 claude
```

Claude loads the SDD into the current session and starts implementing. You can guide, review, and adjust in real time. The agent fills the Work Log when done.

## Autonomous execution (OpenHands)

```
/sdd-assign SDD-001 openhands
```

OpenHands must be running first:

```bash
mise run openhands:up
```

Then trigger execution:

```bash
mise run sdd .forgia/sdd/FD-001/SDD-001.md
```

OpenHands runs headlessly in Docker, reads the SDD, implements the component, and fills the Work Log.

## Batch execution

Run all pending SDDs in an FD sequentially:

```bash
forgia batch FD-001
```

Forgia runs each pending SDD in order, one at a time. It stops immediately on the first failure — fix the failing SDD and re-run before continuing with the rest.

## Direct execution

```bash
forgia exec .forgia/sdd/FD-001/SDD-001.md
```

Runs a single SDD through the configured agent runtime without the `/sdd-assign` flow.

## Work Log

Every agent must fill the Work Log before marking the SDD done:

- **Executor**: who or what ran the SDD
- **Decisions**: choices made during implementation (with reasons)
- **Output**: files created, commands run, test results
- **Retrospective**: what worked, what didn't, what the next agent should know

An SDD with an empty Work Log fails `/fd-verify`.
