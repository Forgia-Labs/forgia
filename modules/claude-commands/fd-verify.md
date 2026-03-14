Verify implementation of a Feature Design after all SDDs are complete.

## Instructions

Given FD identifier: $ARGUMENTS

1. Read the specified FD file from `.forgia/fd/`
2. Read all SDDs in `.forgia/sdd/FD-NNN/`
3. Read the constitution from `.forgia/constitution.md`
4. For each SDD, verify:
   - Status is "done"
   - Work Log is filled (Agent, Decisions, Output, Retrospective sections)
   - All Acceptance Criteria are checked
   - Commit references exist and are valid
5. For the FD, verify:
   - All verification criteria in the FD's "Verifica" section are met
   - All listed SDDs have been completed
   - Run actual validation where possible (tests, linting, build)
6. Check constitution compliance:
   - Commit conventions followed
   - Code standards respected
   - No security violations
7. If ALL verifications pass:
   - Update FD status to "complete"
   - Report: "VERIFICATO — FD completato. Usa /fd-close FD-NNN per archiviare"
8. If ANY verification fails:
   - Keep FD status as "in-progress"
   - List all failures with specific issues
   - Identify which SDDs need rework
   - Report: "VERIFICA FALLITA" with action items

## Important

- Run actual validation commands where possible (test suites, build, lint)
- Every SDD must have a completed Work Log — this is non-negotiable
- This is the final gate before a FD can be closed
