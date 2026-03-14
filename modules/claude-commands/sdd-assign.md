Assign an SDD to an agent for execution.

## Instructions

Given SDD identifier and optional agent: $ARGUMENTS
Format: `SDD-001` or `FD-001/SDD-001` or `SDD-001 openhands`

1. Find the SDD file in `.forgia/sdd/`
2. If no agent is specified, ask the user which agent backend to use:
   - **openhands** — autonomous execution in Docker container (best for isolated, well-defined SDDs)
   - **claude-code** — interactive execution with Claude Code (best for complex SDDs needing human-in-the-loop)
   - **manual** — developer implements manually (SDD serves as spec)
3. Update the SDD frontmatter:
   - Set `agent: <chosen-agent>`
   - Set `assigned_to: <agent-name or user>`
   - Set `status: assigned`
4. If agent is `openhands`:
   - Show the command to execute: `mise run sdd .forgia/sdd/FD-NNN/SDD-NNN.md`
   - Remind: OpenHands must be running (`mise run openhands:up`)
5. If agent is `claude-code`:
   - Load the SDD context into the current session
   - Suggest starting implementation immediately
6. If agent is `manual`:
   - Show the SDD as a reference spec
   - Remind to fill the Work Log when done

## Important

- An SDD must have status "planned" or "assigned" to be assignable
- Never re-assign an "in-progress" SDD without user confirmation
- The Work Log must be filled regardless of which agent executes
