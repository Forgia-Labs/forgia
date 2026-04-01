---
title: "CLI Reference"
description: "All Forgia CLI commands: init, status, doctor, exec, batch."
navigation:
  title: "CLI Reference"
---

# CLI Reference

The `forgia` CLI manages the local vault, runs SDDs, and inspects project state.

## Installation

Install via mise (recommended):

```bash
mise install forgia
```

Or build from source:

```bash
git clone https://github.com/Forgia-Labs/forgia
cd forgia
go build -o forgia ./cmd/forgia/
```

## Commands

| Command | Description |
|---------|-------------|
| [`forgia init`](/docs/cli/init) | Initialize the Forgia vault in a project |
| [`forgia status`](/docs/cli/status) | Show FD and SDD status across the vault |
| [`forgia doctor`](/docs/cli/doctor) | Diagnose missing dependencies |
| [`forgia exec`](/docs/cli/exec) | Execute a single SDD against an agent runtime |
| `forgia batch FD-NNN` | Execute all SDDs for an FD in parallel |

## Claude Code slash commands

Most of the Forgia workflow runs through Claude Code slash commands, not the CLI:

| Command | Description |
|---------|-------------|
| `/fd-new` | Create a new Feature Design |
| `/fd-review` | Gate review before SDD generation |
| `/fd-sdd` | Generate SDDs from an approved FD |
| `/fd-verify` | Verify all SDDs are done before closing |
| `/fd-close` | Archive and close a completed FD |
| `/sdd-assign` | Assign an SDD to an agent backend |

The CLI handles infrastructure and automation; the slash commands handle the design and review workflow.
