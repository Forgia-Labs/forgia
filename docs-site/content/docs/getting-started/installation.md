---
title: "Installation"
description: "Build the Forgia CLI and install Claude Code slash commands."
navigation:
  title: "Installation"
---

# Installation

## Install the CLI

```bash
git clone git@github.com:forgia-labs/forgia.git ~/forgia
cd ~/forgia

# Build the Go binary
mise run go:build
# Or install directly:
# go install github.com/forgia-labs/forgia/cmd/forgia@latest
```

## Install Claude Code slash commands

```bash
mise run claude:install
```

This copies the `/fd-new`, `/fd-review`, `/fd-sdd`, `/fd-verify`, `/fd-close`, and other commands into your Claude Code configuration.

## Install OpenHands (optional)

OpenHands runs SDDs autonomously in a Docker container. Skip this if you plan to use Claude Code interactively.

```bash
mise run openhands:install
```

## Initialize a project

Run this inside any project directory:

```bash
# Via CLI
forgia init

# Via Claude Code
/project-init
```

This creates `.forgia/` with:

| Path | Purpose |
|------|---------|
| `constitution.md` | Immutable project rules — edit to match your standards |
| `fd/` | Feature Designs |
| `sdd/` | Execution specs |
| `dev-guide/` | Coding, commit, and review conventions |
| `guardrails/deny.toml` | Files and commands agents are never allowed to touch |

## Health check

```bash
forgia doctor
```

Checks: vault structure, Docker, OpenHands, mise, Claude commands, LLM API key.
