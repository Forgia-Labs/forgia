# Forgia

Spec-driven development framework. Forge specs into code.

## Architecture

```
FD (Feature Design)  →  SDD (Spec-Driven Development)  →  Code
 "cosa e perche"         "agent-ready prompt"              output
 1 FD                    N SDD                             N agent
```

- **FD** = functional design document for humans (decisions, trade-offs, architecture)
- **SDD** = structured execution contract for agents (scope, interfaces, constraints, tests, acceptance criteria)
- **Constitution** = immutable project rules (style, security, conventions)

## Stack

- **CLI**: `cmd/forgia/` (Go, Cobra)
- **Task runner**: mise (install, run, status, doctor)
- **Agent runtime**: OpenHands (Docker, headless mode)
- **Vault**: Obsidian (dataview, kanban, templater)
- **AI commands**: Claude Code slash commands (`/fd-*`, `/sdd-*`, `/project-init`)

## Key Concepts

1. FD is a design document — humans decide
2. SDD is an agent-ready prompt — agents execute
3. 1 FD produces N SDD (one per component/crate/service)
4. Each SDD runs in isolation (worktree or container)
5. Work Log in SDD is mandatory — tracks decisions, output, retrospective
6. Constitution is the immutable ruleset applied to every change

## Conventions

- Italian for docs/communication, English for code
- Commit format: `feat|fix|docs|refactor: description`
- FD-linked commits: `feat(FD-NNN): description`
- `Co-Authored-By` when AI-generated
- Shell: `set -euo pipefail`, local vars, explicit errors
