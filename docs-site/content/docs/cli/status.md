---
title: "forgia status"
description: "Show FD and SDD status across the vault."
navigation:
  title: "status"
---

# forgia status

```
forgia status [FD-NNN]
```

Displays the current state of FDs and SDDs in the vault.

## Project overview

```bash
forgia status
```

Output:

```
FD-001  user-authentication        approved     3 SDDs (1 done, 1 in-progress, 1 planned)
FD-002  api-rate-limiting          planned      —
FD-003  dashboard-redesign         closed       4 SDDs (4 done)
```

## FD detail

```bash
forgia status FD-001
```

Output:

```
FD-001: User Authentication with JWT
  Status:   approved
  Reviewed: true (claude)
  Created:  2026-03-15

  SDDs:
    SDD-001  auth-middleware       done         agent: claude-code
    SDD-002  token-storage         in-progress  agent: openhands
    SDD-003  integration-wiring    planned      —
```

## Filtering

```bash
# Only show in-progress work
forgia status --filter in-progress

# Only show planned FDs
forgia status --filter planned
```

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | All FDs complete or closed |
| `1` | One or more FDs have SDDs in `planned` or `in-progress` state |

Use exit code 1 in CI to gate releases on complete work.
