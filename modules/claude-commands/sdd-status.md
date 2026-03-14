Show execution status of all SDDs.

## Instructions

1. Read all SDD files from `.forgia/sdd/*/SDD-*.md`
2. Parse frontmatter and Work Log section
3. Display status for each SDD:

### Format

```
=== SDD Execution Status ===

FD-001: My Feature
  SDD-001  [done]         component-a     openhands   2h 15m   3 commits
  SDD-002  [in-progress]  component-b     claude-code started 10m ago
  SDD-003  [planned]      component-c     (unassigned)

FD-002: Another Feature
  SDD-001  [assigned]     setup           openhands   (not started)

Summary: 4 SDDs — 1 done, 1 in-progress, 1 assigned, 1 planned
```

4. For completed SDDs, show Work Log summary:
   - Duration
   - Number of commits
   - Any issues from Retrospective

5. Highlight problems:
   - SDDs "in-progress" for more than 4 hours without Work Log updates
   - SDDs "done" with empty Work Log
   - SDDs "assigned" but never started

## Important

- This is a READ-ONLY operation
- If checking OpenHands container status is possible (docker ps), include it
