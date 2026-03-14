Archive a completed Feature Design and update project records.

## Instructions

Given FD identifier: $ARGUMENTS

1. Read the specified FD file from `.forgia/fd/`
2. Verify the FD status is "complete" — if not, refuse and suggest `/fd-verify` first
3. Archive:
   - Update FD status to "closed" in frontmatter
   - Move related OPS tasks from `ops/active/` to `ops/done/` if they reference this FD
4. Update changelog:
   - If `.forgia/CHANGELOG.md` exists, prepend an entry
   - If not, create it with the first entry
   - Format: `## [FD-NNN] Title — YYYY-MM-DD`
   - Include: summary of changes, SDDs completed, key decisions from Work Logs
5. Compile retrospective:
   - Aggregate "Retrospective" sections from all SDDs
   - Add a summary to the changelog entry
6. Report:
   - Confirmation of closure
   - Changelog entry created
   - Aggregate retrospective insights
   - Suggest `/fd-status` to see updated dashboard

## Important

- NEVER close a FD that hasn't been verified (status must be "complete")
- The closed FD stays in `.forgia/fd/` as historical documentation (not moved)
- SDDs stay in `.forgia/sdd/FD-NNN/` as permanent record
- The Work Log retrospectives are the most valuable output — preserve them
