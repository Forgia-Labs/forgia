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
# Install with mise (recommended)
mise use github:forgia-labs/forgia

# Or download binary from GitHub Releases
# https://github.com/forgia-labs/forgia/releases

# Or build from source
git clone git@github.com:forgia-labs/forgia.git
cd forgia
mise run go:build
# Or: go install github.com/forgia-labs/forgia/cmd/forgia@latest

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

## Workflow

Forgia enforces a 3-gate workflow with optional analysis steps:

```
FD Review ──→ Arch Review ──→ Threat Model ──→ SDD Generation ──→ Implementation ──→ Verification ──→ Close
 /fd-review   /fd-arch-review  /fd-threat-model  /fd-sdd           agent executes     /fd-verify       /fd-close
 GATE 1        (optional)       (optional)                                              GATE 2           GATE 3
```

- **Gate 1** (`/fd-review`): FD must pass review before SDDs can be generated
- **Arch Review** (`/fd-arch-review`): Pattern detection, SOLID compliance, dependency graph (advisory)
- **Threat Model** (`/fd-threat-model`): STRIDE security analysis, feeds mitigations into SDD constraints (advisory)
- **Gate 2** (`/fd-verify`): Acceptance criteria met, tests pass, Work Logs filled
- **Gate 3** (`/fd-close`): Archive, retrospective, changelog

## Vault Structure

After `forgia init`, your project gets:

```
your-project/
  .forgia/
    config.toml               # runner, beads, knowledge, watcher config
    constitution.md            # immutable project rules
    _dashboard.md              # Obsidian dataview overview
    guardrails/
      deny.toml                # security deny patterns (read/write/execute)
      ignore                   # files to exclude from agent context
    fd/                        # Feature Designs (what & why)
    sdd/                       # Execution Specs (how, for agents)
      _templates/
        sdd-template.md        # markdown format
        sdd-template.yaml      # YAML format (machine-parseable)
    ops/                       # manual operational tasks
    architecture/              # C4/DDD system design (YAML)
    contexts/                  # bounded context definitions
    learnings/                 # retrospective feedback loop
    dev-guide/
      principles/              # clean-code, SOLID, design-patterns
      lang/                    # per-language conventions (auto-detected)
      coding-conventions.md
      commit-conventions.md
      review-process.md
```

## Spells (Slash Commands)

> Claude Code slash commands — the core workflow. Run via `/command` in Claude Code or `forgia skill <name>`.

### Feature Design

| Spell | What it does |
|-------|--------------|
| `/fd-new` | **Forge a new design** — create FD from issue or description |
| `/fd-review` | **Trial by fire** — mandatory review gate (GATE 1) |
| `/fd-arch-review` | **Inspect the blueprint** — pattern detection, SOLID compliance, dependency graph |
| `/fd-threat-model` | **Test the armor** — STRIDE security analysis, guardrails suggestions |
| `/fd-compare` | **Weigh the options** — compare competing FDs side-by-side |
| `/fd-sdd` | **Temper the specs** — generate N SDDs from approved FD |
| `/fd-deep` | **Deep analysis** — parallel exploration for complex problems |
| `/fd-explore` | **Study the piece** — load FD context into session |
| `/fd-verify` | **Quality check** — verify implementation vs spec (GATE 2) |
| `/fd-close` | **Seal the work** — archive FD + retrospective (GATE 3) |
| `/fd-feedback` | **Learn from the forge** — route Work Log insights to architecture/learnings |
| `/fd-status` | **Forge status** — FD + SDD dashboard |

### SDD Execution

| Spell | What it does |
|-------|--------------|
| `/sdd-dry-run` | **Test the fire** — simulate SDD execution, feasibility report (GO/NO-GO) |
| `/sdd-assign` | **Hand the hammer** — assign SDD to agent backend |
| `/sdd-status` | **Check the anvil** — execution status of all SDDs |

### Architecture (DDD)

| Spell | What it does |
|-------|--------------|
| `/arch-init` | **Lay the foundation** — initialize DDD architecture from brief |
| `/arch-review` | **Survey the structure** — review coherence across vault |
| `/arch-update` | **Record the build** — update architecture from SDD Work Logs |

### Setup

| Spell | What it does |
|-------|--------------|
| `/project-init` | **Light the forge** — scaffold vault in current project |

## CLI Commands

| Command | What it does |
|---------|--------------|
| `forgia init` | Scaffold the vault in the current project |
| `forgia status` | Dashboard: FDs, SDDs, OPS tasks, exec logs, Beads, knowledge stats |
| `forgia doctor` | Health check: vault, tools, Docker, API keys, knowledge layer |
| `forgia validate <sdd>` | Validate an SDD before execution |
| `forgia exec <sdd>` | Execute an SDD (`--runner`, `--dry-run`, `--mode`) |
| `forgia batch <FD-NNN>` | Execute all pending SDDs for an FD |
| `forgia watch <FD-NNN>` | Watch and auto-execute new SDDs |
| `forgia sync` | Sync vault with project board (GitHub Projects) |
| `forgia skill <name>` | Execute a slash command via Claude CLI |
| `forgia skills` | List all available slash commands |
| `forgia version` | Show version |

### Runners

```bash
# Claude Code — uses your Max subscription
forgia exec .forgia/sdd/FD-001/SDD-001.md --runner=claude

# OpenHands — autonomous, N parallel containers
forgia exec .forgia/sdd/FD-001/SDD-001.md --runner=openhands

# Dry run — simulate without executing
forgia exec .forgia/sdd/FD-001/SDD-001.md --dry-run

# Guardrail enforcement mode
forgia exec .forgia/sdd/FD-001/SDD-001.md --mode=guard

# Default from config.toml
forgia exec .forgia/sdd/FD-001/SDD-001.md
```

| Runner | When to use |
|--------|-------------|
| **Claude Code** | Interactive, Max subscription, git worktree isolation |
| **OpenHands** | Autonomous, N parallel containers, API keys |
| **Dry Run** | Feasibility check before committing to execution |
| **Manual** | Read the SDD and implement yourself |

## Guardrails

Security deny patterns in `.forgia/guardrails/deny.toml` control what agents can and cannot do:

```toml
[read]    # Files agents must NEVER read (.env, *.pem, .ssh/*, etc.)
[execute] # Commands agents must NEVER run (pass show, gpg --export-secret, etc.)
[write]   # Files agents must NEVER modify (constitution.md, config.toml, deny.toml)
```

Guardrails are **fail-closed** — if a pattern matches, the action is blocked. Three enforcement modes:
- **careful** — warns on destructive commands
- **freeze** — restricts writes to SDD boundaries
- **guard** — careful + freeze combined

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
  lang/                ← auto-detected per project (rust, python, ts, go, k8s, shell)
  coding-conventions   ← operational rules
  commit-conventions   ← commit format
  review-process       ← workflow gates
```

Optional: [codebase-memory-mcp](https://github.com/nicobailey-ai/codebase-memory-mcp) provides Tier 3 semantic indexing (symbols, edges, dependency graph). Auto-indexed during `forgia init` if installed.

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
| `go` | Yes | Build the CLI | [go.dev](https://go.dev/dl/) (1.25+) |
| `mise` | Yes | Task runner | [mise.jdx.dev](https://mise.jdx.dev) |
| `claude` | Yes | Claude Code CLI | [claude.ai/claude-code](https://claude.ai/claude-code) |
| `docker` | Optional | OpenHands runner | [docker.com](https://docker.com) |
| `bd` | Optional | Beads task tracker | `mise run bd:install` |
| `fswatch` | Optional | `forgia watch` | `brew install fswatch` |
| `yq` | Optional | YAML SDD validation | `brew install yq` |
| `codebase-memory-mcp` | Optional | Knowledge layer (Tier 3) | [GitHub](https://github.com/nicobailey-ai/codebase-memory-mcp) |

## Inspired By

- [AutoSpec](https://github.com/ariel-frischer/autospec) — YAML specs, auto-validation, session isolation
- [Beads](https://github.com/steveyegge/beads) — distributed graph issue tracker for AI agents
- [GitHub Spec Kit](https://github.com/github/spec-kit) — spec/plan separation, constitution
- [OpenHands](https://github.com/all-hands-ai/OpenHands) — sandboxed agent runtime

## License

MIT
