---
title: "forgia status"
description: "Show FD and SDD status across the vault."
navigation:
  title: "status"
---

# forgia status

```
forgia status
```

Displays the full Forgia dashboard: all Feature Designs, their SDDs, active ops tasks, recent executions, and knowledge layer stats.

## Output

```
=== Forgia Dashboard ===

ID       TITLE                    STATUS      PRIORITY  AUTHOR
──       ─────                    ──────      ────────  ──────
FD-001   user-authentication      approved    medium    alice
  └ SDD-001  auth-middleware       done
  └ SDD-002  token-storage         in-progress
  └ SDD-003  integration-wiring    ready
FD-002   api-rate-limiting        planned     low       bob
FD-003   dashboard-redesign       closed      high      alice
  └ SDD-001  layout                done
```

Additional sections appear when data is present:

- **Active Tasks** — ops tasks from `.forgia/ops/active/`
- **Last Executions** — recent `forgia exec` runs from `.forgia/logs/`
- **Beads** — ready tasks from the Beads circuit (if `bd` is installed)
- **Knowledge Layer** — codebase index stats (if `codebase-memory-mcp` is installed)

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Command completed (regardless of FD states) |

Use `/fd-status` in Claude Code for the same dashboard inside a session.
