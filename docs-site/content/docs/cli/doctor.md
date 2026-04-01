---
title: "forgia doctor"
description: "Diagnose missing dependencies and configuration issues."
navigation:
  title: "doctor"
---

# forgia doctor

```
forgia doctor
```

Checks that the vault, tools, and configuration are set up correctly. Always exits `0` — it reports issues without failing, so it's safe to run in CI as a diagnostic step.

## Output format

```
=== Forgia Doctor ===

  ✓ vault — .forgia/ found and valid
  ✓ config — config.toml parsed
  ✓ guardrails — deny.toml parsed
  ✓ git
  ✓ claude
  ○ docker — not found (optional)
  ○ bd — not found (optional)
  ○ fswatch — not found (optional)
  ○ yq — not found (optional)
  ○ openhands image — docker not available
  ○ openhands container — stopped
  ✓ claude-commands — 12 commands
  ○ beads circuit-breaker — no circuit breaker files
  ✓ llm api key
  ○ codebase-memory-mcp — not installed (optional)

All checks passed.
```

Icons: `✓` pass, `✗` fail, `○` skipped/optional.

## Checks

| Check | Required | Purpose |
|-------|----------|---------|
| `vault` | Yes | `.forgia/` exists and is readable |
| `config` | Yes | `config.toml` parses without error |
| `guardrails` | Yes | `deny.toml` parses without error |
| `git` | Yes | `git` is in PATH |
| `claude` | Yes | `claude` CLI is in PATH |
| `docker` | Optional | Needed for OpenHands autonomous agents |
| `bd` | Optional | Beads task tracker |
| `fswatch` | Optional | File watcher for `forgia watch` |
| `yq` | Optional | YAML processing |
| `openhands image` | Optional | Docker image pulled for OpenHands |
| `openhands container` | Optional | OpenHands container currently running |
| `claude-commands` | — | `/fd-*` and `/sdd-*` slash commands installed |
| `beads circuit-breaker` | — | Beads circuit breaker is closed (not tripped) |
| `llm api key` | Yes | `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` is set |
| `codebase-memory-mcp` | Optional | Knowledge layer for codebase indexing |

## Common fixes

**claude not found**

Install Claude Code: [claude.ai/claude-code](https://claude.ai/claude-code)

**claude-commands not installed**

```bash
mise run claude:install
```

**Docker not running**

Start Docker Desktop, or on Linux:

```bash
sudo systemctl start docker
```

**OpenHands image not pulled**

```bash
mise run openhands:install
```

**llm api key missing**

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

Running `forgia doctor` after fixing issues confirms everything is in order before you start work.
