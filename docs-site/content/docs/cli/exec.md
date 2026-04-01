---
title: "forgia exec"
description: "Execute a single SDD against the configured agent runtime."
navigation:
  title: "exec"
---

# forgia exec

```
forgia exec <sdd-file> [--runner <backend>] [--mode <guardrail-mode>] [--dry-run]
```

Runs a single SDD file through the configured agent runtime.

## Basic usage

```bash
forgia exec .forgia/sdd/FD-001/SDD-002-token-storage.md
```

Uses the runner configured in `.forgia/config.toml` (defaults to `claude`).

## Flags

| Flag | Description |
|------|-------------|
| `--runner` | Runner backend: `claude` (default) or `dry-run` |
| `--mode` | Guardrail enforcement mode: `off`, `careful`, `freeze`, `guard` |
| `--dry-run` | Simulate execution without modifying files |

## Guardrail modes

| Mode | Behaviour |
|------|-----------|
| `off` | No guardrail checks |
| `careful` | Warn on violations |
| `freeze` | Block writes outside allowed dirs |
| `guard` | Block all writes not explicitly allowed |

## Dry run

```bash
forgia exec .forgia/sdd/FD-001/SDD-001.md --dry-run
```

Runs a feasibility simulation and prints a GO / NO-GO report without touching any files.

## Batch execution

To run all pending SDDs for an FD sequentially:

```bash
forgia batch FD-001
```

`forgia batch` calls `forgia exec` for each pending SDD in order. It stops immediately on the first failure — fix the failing SDD before continuing.

## After execution

Check the SDD's Work Log — the agent should have filled:

- Executor and timestamp
- Decisions made during implementation
- Files created or modified
- Test results
- Retrospective

If the Work Log is empty, the execution did not complete cleanly. Re-run or assign manually.
