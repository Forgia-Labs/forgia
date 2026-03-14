Launch parallel deep exploration for a complex Feature Design problem.

## Instructions

Given FD identifier: $ARGUMENTS

1. Read the specified FD file from `.forgia/fd/`
2. Launch 4 parallel Agent (subagent_type=Explore) searches, each exploring a different angle:
   - **Agent 1 — Algorithmic**: Find existing implementations, libraries, or patterns that solve similar problems
   - **Agent 2 — Structural**: Analyze the codebase architecture to find the best integration points
   - **Agent 3 — Incremental**: Identify the smallest possible change that delivers value
   - **Agent 4 — Risk**: Find potential issues, edge cases, security concerns, and conflicts with other FDs
3. Collect results from all 4 agents
4. Write a synthesis section in the FD file under "## Deep Analysis" with:
   - Key findings from each angle
   - Recommended approach based on all findings
   - Updated SDD breakdown if the exploration revealed new components
5. Update the FD status to "design" if it was "planned"

## Important

- This command is for COMPLEX problems that benefit from multi-angle analysis
- For simple FDs, use /fd-explore instead
- Always preserve existing FD content — append the analysis, don't overwrite
