---
id: "SDD-001"
fd: "FD-002"
title: "/fd-compare slash command"
status: done
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [slash-command, compare, competitive]
---

# SDD-001: /fd-compare slash command

> Parent FD: [[FD-002]]

## Scope

Create `modules/claude-commands/fd-compare.md` — a Claude slash command that reads 2+ competing FDs and generates a structured comparison report.

The command:
1. Reads all specified FD files from the vault
2. Validates they share the same `upstream_issue` or have `competes_with` linking
3. Generates comparison table: dimensions (effort, impact, priority, author), architecture diagrams side-by-side, acceptance criteria overlap, technology choices
4. Produces a recommendation based on trade-offs
5. Optional `--post-to-issue` flag: posts the comparison as a comment on the linked GitHub issue via `gh api`

## Acceptance Criteria

- [ ] `/fd-compare FD-A FD-B` generates structured comparison table
- [ ] Comparison includes: dimensions, architecture, criteria overlap, recommendation
- [ ] Works with 2 or 3+ competing FDs
- [ ] `--post-to-issue` posts comparison to linked GitHub issue
- [ ] Error if FDs don't compete (no shared upstream_issue or competes_with)
- [ ] Read-only — never modifies FD files
