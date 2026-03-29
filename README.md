```
  ███████╗ ██████╗ ██████╗  ██████╗ ██╗ █████╗
  ██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██║██╔══██╗
  █████╗  ██║   ██║██████╔╝██║  ███╗██║███████║
  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██║██╔══██║
  ██║     ╚██████╔╝██║  ██║╚██████╔╝██║██║  ██║
  ╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝╚═╝  ╚═╝

  Forge specs into code. FD → SDD → Agent execution framework.
```

> Spec-driven development framework for AI agents.
> Forgia turns human design decisions into execution contracts for autonomous agents.

## The Model

```
You ──→ FD (what & why) ──→ SDD (how, for agents) ──→ Agent ──→ Code
         functional           execution contract        OpenHands
         design               agent-ready prompt        Claude Code
         1 per feature        N per FD                  parallel
```

| Layer | Who | What |
|-------|-----|------|
| **FD** (Feature Design) | Human + AI | Architecture, trade-offs, interfaces |
| **SDD** (Spec-Driven Dev) | AI generates, human approves | Scope, constraints, tests, acceptance criteria |
| **Constitution** | Human | Immutable project rules |
| **Code** | Agent | Implementation guided by SDD |

## Quick Start

```bash
# Prerequisites: mise, docker (optional for OpenHands)
git clone git@github.com:Deepzima/forgia.git
cd forgia

# Trust the mise config to allow running tasks
mise trust 

# Install Claude Code slash commands
mise run claude:install

# Install optional tools (fswatch, yq)
mise run tools:install

# Install Beads task tracker (optional)
mise run bd:install

# Initialize a new project
cd /path/to/your/project
forgia init
```

## Vault Structure

After `forgia init`, your project gets:

```
your-project/
  .forgia/
    config.toml               # runner and beads configuration
    constitution.md            # immutable project rules
    _dashboard.md              # Obsidian dataview overview
    fd/                        # Feature Designs (what & why)
    sdd/                       # Execution Specs (how, for agents)
      _templates/
        sdd-template.md        # markdown format
        sdd-template.yaml      # YAML format (machine-parseable)
    ops/                       # manual operational tasks
    dev-guide/
      principles/              # clean-code, SOLID, design-patterns
      lang/                    # per-language conventions (auto-detected)
      coding-conventions.md
      commit-conventions.md
      review-process.md
```

## Spells (Commands)

> Claude Code slash commands — the core workflow.

| Spell | What it does |
|-------|--------------|
| `/fd-new` | **Forge a new design** — create a Feature Design |
| `/fd-review` | **Trial by fire** — mandatory review gate |
| `/fd-sdd` | **Temper the specs** — generate N SDDs from an approved FD |
| `/fd-deep` | **Deep analysis** — 4 agents explore the problem in parallel |
| `/fd-explore` | **Study the piece** — load FD context |
| `/fd-verify` | **Quality check** — verify implementation vs spec |
| `/fd-close` | **Seal the work** — archive completed FD + retrospective |
| `/fd-compare` | **Weigh the options** — compare competing FDs side-by-side |
| `/fd-status` | **Forge status** — FD + SDD dashboard |

## CLI Tools

> `forgia` commands and `mise` tasks for workflow management.

| Command | What it does |
|---------|--------------|
| `forgia init` | Scaffold the vault in the current project |
| `forgia status` | Dashboard: FDs + SDDs + Beads ready tasks |
| `forgia doctor` | Health check (docker, bd, vault, mise, fswatch, yq) |
| `forgia validate <sdd>` | Validate an SDD before execution |
| `forgia exec <sdd>` | Execute an SDD (auto-detect runner from config) |
| `forgia batch <FD-NNN>` | Execute all SDDs for an FD |
| `forgia watch <FD-NNN>` | Watch and auto-execute new SDDs |

### Runners

```bash
# Claude Code — uses your Max subscription (free)
forgia exec .forgia/sdd/FD-001/SDD-001.md --runner=claude

# OpenHands — uses API keys, isolated Docker containers
forgia exec .forgia/sdd/FD-001/SDD-001.md --runner=openhands

# Default from config.toml
forgia exec .forgia/sdd/FD-001/SDD-001.md
```

| Runner | When to use |
|--------|-------------|
| **Claude Code** | Interactive, Max subscription, git worktree isolation |
| **OpenHands** | Autonomous, N parallel containers, API keys |
| **Manual** | Read the SDD and implement yourself |

## Beads Integration

[Beads (bd)](https://github.com/steveyegge/beads) manages the dependency graph between SDDs:

```bash
# After /fd-sdd: creates epic + subtasks with dependencies
bd ready              # show unblocked tasks
forgia batch FD-001   # execute only ready ones (dependency-aware)
```

## Knowledge Stack

Every agent automatically loads:

```
.forgia/dev-guide/
  principles/          ← ALWAYS loaded (clean-code, SOLID, design-patterns)
  lang/                ← auto-detected per project (rust, python, ts, go, shell)
  coding-conventions   ← operational rules
  commit-conventions   ← commit format
```

## Work Log

Every SDD includes a mandatory Work Log — filled by the agent or developer:

```yaml
work_log:
  executor: claude-code
  started: 2026-03-15T10:00:00
  completed: 2026-03-15T11:30:00
  decisions:
    - what: "Used tower middleware instead of manual auth"
      why: "Composable, testable, idiomatic axum"
  output:
    commits: ["abc123"]
    files_changed: ["src/auth.rs", "tests/auth_test.rs"]
  retrospective:
    worked: "Builder pattern for config was clean"
    suggestions: "Add integration test template to SDD"
```

## Prerequisites

| Tool | Required | Purpose | Install |
|------|----------|---------|---------|
| `mise` | Yes | Task runner | [mise.jdx.dev](https://mise.jdx.dev) |
| `claude` | Yes | Claude Code CLI | [claude.ai/claude-code](https://claude.ai/claude-code) |
| `docker` | Optional | OpenHands runner | [docker.com](https://docker.com) |
| `bd` | Optional | Beads task tracker | `mise run bd:install` |
| `fswatch` | Optional | `forgia watch` | `brew install fswatch` |
| `yq` | Optional | YAML SDD validation | `brew install yq` |

## Inspired By

- [AutoSpec](https://github.com/ariel-frischer/autospec) — YAML specs, auto-validation, session isolation
- [Beads](https://github.com/steveyegge/beads) — distributed graph issue tracker for AI agents
- [GitHub Spec Kit](https://github.com/github/spec-kit) — spec/plan separation, constitution
- [OpenHands](https://github.com/all-hands-ai/OpenHands) — sandboxed agent runtime

## License

MIT
