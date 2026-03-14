```
  ███████╗ ██████╗ ██████╗  ██████╗ ██╗ █████╗
  ██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██║██╔══██╗
  █████╗  ██║   ██║██████╔╝██║  ███╗██║███████║
  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██║██╔══██║
  ██║     ╚██████╔╝██║  ██║╚██████╔╝██║██║  ██║
  ╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝╚═╝  ╚═╝
```

> Forge specs into code. Spec-driven development framework for AI agents.

Forgia is an opinionated SDD (Spec-Driven Development) framework that turns human design decisions into execution contracts for autonomous agents.

## The Model

```
You ──→ FD (what & why) ──→ SDD (how, for agents) ──→ Agent ──→ Code
         functional            execution contract        OpenHands
         design                agent-ready prompt        Claude Code
         1 per feature         N per FD                  parallel
```

| Layer | Who | What |
|-------|-----|------|
| **FD** (Feature Design) | Human + AI | Architecture, trade-offs, interfaces |
| **SDD** (Spec-Driven Dev) | AI generates, human approves | Scope, constraints, tests, acceptance criteria |
| **Constitution** | Human | Immutable project rules |
| **Code** | Agent | Implementation guided by SDD |

## Quick Start

```bash
# Prerequisites: mise, docker
git clone git@github.com:Deepzima/forgia.git
cd forgia

# Install Claude Code commands
mise run claude:install

# Install OpenHands runtime
mise run openhands:install

# Initialize a new project
cd /path/to/your/project
forgia init
```

## Vault Structure

After `forgia init`, your project gets:

```
your-project/
  .forgia/
    constitution.md         # immutable project rules
    _dashboard.md           # Obsidian dataview overview
    fd/                     # Feature Designs
    sdd/                    # Execution Specs (agent-ready)
    ops/                    # Operational tasks
    dev-guide/              # Conventions
```

## Incantesimi (Spells)

> Forgia's commands — Claude Code slash commands.

| Incantesimo | What it does |
|-------------|--------------|
| `/project-init` | Prepare the forge — scaffold the `.forgia/` vault |
| `/fd-new` | Forge a new design — create a Feature Design |
| `/fd-review` | Trial by fire — mandatory review gate |
| `/fd-sdd` | Temper the specs — generate N SDDs from an approved FD |
| `/fd-deep` | Deep analysis — 4 agents explore the problem in parallel |
| `/fd-explore` | Study the piece — load FD context |
| `/fd-verify` | Quality check — verify implementation vs spec |
| `/fd-close` | Seal the work — archive completed FD + retrospective |
| `/fd-status` | Forge status — FD + SDD dashboard |
| `/sdd-assign` | Assign the work — send an SDD to an agent |
| `/sdd-status` | Agent status — SDD execution progress |

## Attrezzi della Fucina (Forge Tools)

> Mise tasks for infrastructure management.

| Attrezzo | What it does |
|----------|--------------|
| `mise run init` | Scaffold vault in current project |
| `mise run openhands:install` | Pull OpenHands container |
| `mise run openhands:up` | Light the forge — start OpenHands on :3000 |
| `mise run openhands:down` | Shut down the forge |
| `mise run sdd <file>` | Execute an SDD with OpenHands headless |
| `mise run sdd:batch FD-001` | Execute all SDDs for an FD in parallel |
| `mise run status` | Dashboard of all FD + SDD |
| `mise run doctor` | Health check (docker, openhands, vault) |
| `mise run claude:install` | Install commands in Claude Code |

## Work Log

Every SDD includes a mandatory Work Log section:

```markdown
## Work Log
### Agent
- who executed, when, duration
### Decisions
- deviations from plan, problems encountered
### Output
- commit hash, PR link, files created/modified
### Retrospective
- what worked, what didn't, suggestions for future FDs
```

## Agent Backends

| Agent | When |
|-------|------|
| **Claude Code** | Interactive, you're at the keyboard |
| **OpenHands** | Autonomous, N agents in parallel containers |
| **Manual** | You read the SDD and implement yourself |

## Inspired By

- [GitHub Spec Kit](https://github.com/github/spec-kit) — spec/plan separation, constitution
- [BMAD Method](https://github.com/bmad-code-org/BMAD-METHOD) — multi-agent roles
- [OpenHands](https://github.com/OpenHands/OpenHands) — sandboxed agent runtime

## License

MIT
