# Forgia

> Forge specs into code. Spec-driven development framework for AI agents.

Forgia is an opinionated SDD (Spec-Driven Development) framework that bridges the gap between human design decisions and autonomous agent execution.

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

# Install OpenHands runtime
mise run openhands:install

# Initialize a new project
cd /path/to/your/project
mise run --cd /path/to/forgia init

# Or use Claude Code slash commands
/project-init
/fd-new "my feature"
/fd-review FD-001
/fd-sdd FD-001        # generate N SDDs
/fd-verify FD-001
/fd-close FD-001
```

## Vault Structure

After `forgia init`, your project gets:

```
your-project/
  .forgia/
    constitution.md         # immutable project rules
    _dashboard.md           # Obsidian dataview overview
    fd/
      _templates/
        fd-template.md
      FD-001-feature.md
    sdd/
      _templates/
        sdd-template.md
      FD-001/
        SDD-001-component.md
    ops/
      _templates/
        ops-task.md
      active/
      done/
    dev-guide/
      coding-conventions.md
      commit-conventions.md
      review-process.md
```

## Commands

### mise tasks

| Task | Description |
|------|-------------|
| `mise run init` | Scaffold `.forgia/` vault in current project |
| `mise run openhands:install` | Pull and configure OpenHands container |
| `mise run openhands:up` | Start OpenHands UI on :3000 |
| `mise run openhands:down` | Stop OpenHands |
| `mise run sdd <file>` | Execute an SDD with OpenHands headless |
| `mise run status` | Dashboard of all FD + SDD |
| `mise run doctor` | Health check (docker, openhands, vault) |

### Claude Code slash commands

| Command | Description |
|---------|-------------|
| `/project-init` | Scaffold vault (interactive) |
| `/fd-new` | Create new Feature Design |
| `/fd-review` | Mandatory review gate |
| `/fd-sdd` | Generate N SDDs from approved FD |
| `/fd-deep` | Multi-agent deep exploration |
| `/fd-explore` | Load FD context |
| `/fd-verify` | Verify implementation vs spec |
| `/fd-close` | Archive completed FD |
| `/fd-status` | FD + SDD dashboard |
| `/sdd-assign` | Assign SDD to agent |
| `/sdd-status` | SDD execution status |

## SDD Work Log

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

| Backend | When |
|---------|------|
| **Claude Code** | Interactive, you're at the keyboard |
| **OpenHands** | Autonomous, N agents in parallel containers |
| **Manual** | You read the SDD and implement yourself |

## Inspired By

- [GitHub Spec Kit](https://github.com/github/spec-kit) — spec/plan separation, constitution
- [BMAD Method](https://github.com/bmad-code-org/BMAD-METHOD) — multi-agent roles
- [OpenHands](https://github.com/OpenHands/OpenHands) — sandboxed agent runtime

## License

MIT
