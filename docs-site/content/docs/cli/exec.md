---
title: "forgia exec"
description: "Execute a single SDD against the configured agent runtime."
navigation:
  title: "exec"
---

# forgia exec

```
forgia exec <path-to-sdd> [--agent <backend>] [--dry-run]
```

Runs a single SDD file through the configured agent runtime.

## Basic usage

```bash
forgia exec .forgia/sdd/FD-001/SDD-002-token-storage.md
```

Uses the agent specified in the SDD frontmatter (`agent: openhands`). If no agent is set, prompts for one.

## Flags

| Flag | Description |
|------|-------------|
| `--agent` | Override the agent backend (`openhands`, `claude-code`, `manual`) |
| `--dry-run` | Print what would be executed without running it |
| `--timeout` | Maximum execution time (default: 30m) |
| `--worktree` | Run in an isolated git worktree |

## Worktree isolation

```bash
forgia exec .forgia/sdd/FD-001/SDD-001.md --worktree
```

Creates a temporary git worktree, runs the agent there, and merges changes back on success. Prevents partially-applied changes from polluting the working tree if the agent fails midway.

## Batch execution

To run all SDDs for an FD in parallel:

```bash
forgia batch FD-001
```

`forgia batch` calls `forgia exec` for each SDD, respecting the dependency order implied by interface contracts. SDDs with no dependencies on each other run concurrently.

## After execution

Check the SDD's Work Log — the agent should have filled:

- Executor and timestamp
- Decisions made during implementation
- Files created or modified
- Test results
- Retrospective

If the Work Log is empty, the execution did not complete cleanly. Re-run or assign manually.
