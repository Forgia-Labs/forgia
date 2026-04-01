---
title: "Your First Feature"
description: "End-to-end walkthrough: from idea to merged code using the Forgia workflow."
navigation:
  title: "First Feature"
---

# Your First Feature

This walkthrough covers the full Forgia workflow for a single feature.

## 1. Create a Feature Design

```
/fd-new "Add user authentication with JWT"
```

Or from a GitHub issue:

```
/fd-new #42
```

Claude creates `.forgia/fd/FD-001-user-authentication.md` pre-filled with problem statement, solutions, architecture diagrams, and planned SDDs.

**Fill in the gaps**: edit the FD to add your chosen architecture, define component interfaces, and list the SDDs you want generated.

## 2. Review the FD

```
/fd-review FD-001
```

This is a mandatory gate. The review checks:

- Problem is clearly defined (what + why, not how)
- At least 2 solutions with pros/cons
- Architecture diagrams are present and non-trivial
- Interfaces between components are defined
- Verification criteria are testable

If anything fails, the FD stays in `planned` state. Fix the issues and re-run.

On pass: status changes to `approved`.

## 3. Generate SDDs

```
/fd-sdd FD-001
```

Creates `N` files in `.forgia/sdd/FD-001/`, one per component. Each SDD is a self-contained execution contract: scope, interfaces, constraints, best practices, test requirements, acceptance criteria.

## 4. Execute

Assign each SDD to an agent:

```bash
# Interactive (Claude Code)
/sdd-assign SDD-001 claude

# Autonomous (OpenHands in Docker)
mise run sdd .forgia/sdd/FD-001/SDD-001.md

# Parallel batch
forgia batch FD-001
```

The agent reads the SDD, implements the code, and fills in the Work Log.

## 5. Verify

```
/fd-verify FD-001
```

Checks that:
- All SDDs have status `done`
- All Work Logs are filled
- Acceptance criteria are met

## 6. Close

```
/fd-close FD-001
```

Archives the FD, updates the changelog, and aggregates retrospectives from all Work Logs.

---

Next: read the [FD Guide](/docs/fd) for a deeper look at Feature Designs.
