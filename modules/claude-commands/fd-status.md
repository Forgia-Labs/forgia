Show the Forgia dashboard — current status of all FDs and SDDs.

## Instructions

1. Read all FD files from `.forgia/fd/FD-*.md`
2. Read all SDD files from `.forgia/sdd/*/SDD-*.md`
3. Parse frontmatter from each file
4. Display a formatted dashboard grouped by status:

### Format

```
=== Forgia Dashboard ===

--- Feature Designs ---
  FD-001  [approved]   my-feature        effort:large   (3 SDDs)
  FD-002  [planned]    another-thing     effort:medium  (0 SDDs)

--- Execution Specs ---
  FD-001/SDD-001  [done]        component-a     agent:openhands
  FD-001/SDD-002  [in-progress] component-b     agent:claude-code
  FD-001/SDD-003  [planned]     component-c     (unassigned)

--- Active Tasks ---
  OPS-001  [high]  Fix build pipeline
  OPS-002  [low]   Update docs

Summary: 2 FDs, 3 SDDs (1 done, 1 in-progress, 1 planned), 2 tasks
```

5. Highlight issues:
   - FDs approved but with no SDDs generated
   - SDDs assigned but stale (no Work Log updates)
   - SDDs with incomplete Work Logs marked as "done"
   - Multiple SDDs assigned to the same agent

## Important

- This is a READ-ONLY operation
- If no FDs exist, suggest `/fd-new` to create one
- If no vault exists, suggest `/project-init`
