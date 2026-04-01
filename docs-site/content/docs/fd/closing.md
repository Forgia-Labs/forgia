---
title: "Closing an FD"
description: "Verify all SDDs are done, then archive the FD with /fd-close."
navigation:
  title: "Closing"
---

# Closing an FD

## Verify first

```
/fd-verify FD-001
```

Before closing, `/fd-verify` checks that:
- All SDDs under `FD-001/` have status `done`
- Every Work Log is filled (Agent, Decisions, Output, Retrospective sections)
- Acceptance criteria from the FD are met

If any SDD is still in progress or its Work Log is empty, verification fails. Fix it before closing.

## Close

```
/fd-close FD-001
```

`/fd-close` performs:
1. Archives the FD (`status: closed`)
2. Updates the project changelog
3. Aggregates retrospectives from all Work Logs into a summary
4. Optionally syncs the board card to "Closed"

## Rejecting an FD

If a competitive review selects a different approach:

```
/fd-close FD-001 --reject
```

The FD is archived with `status: rejected` and `superseded_by: FD-NNN`. It serves as a negative ADR — a record of the path not taken and why.

## After closing

The Work Log retrospectives feed back into future FDs. Patterns that worked well, failure modes discovered, and interface corrections are stored in `.forgia/learnings/` for the intelligent gate to reference when reviewing future proposals.
