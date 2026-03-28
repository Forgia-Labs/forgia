---
id: "SDD-001"
fd: "FD-003"
title: "/fd-arch-review slash command"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-28"
started: "2026-03-28"
completed: "2026-03-28"
tags: ["phase:2-core"]
---

# SDD-001: /fd-arch-review slash command

> Parent FD: [[FD-003]]

## Scope

Create the file `modules/claude-commands/fd-arch-review.md` — a Claude Code slash command that performs architecture review of a Feature Design against the actual codebase.

The command accepts an FD identifier (`$ARGUMENTS` = `FD-NNN`) and produces a structured markdown report with 5 sections:

1. **Pattern Analysis** — scan codebase directory structure and detect known design patterns (Repository, Factory, Handler/Controller, Operator/Watcher, Builder, Strategy, Observer, Middleware, State Machine). Present as a table with: Pattern, Detected (where), Assessment.

2. **Anti-Pattern Detection** — check for: God Object (files/packages with too many responsibilities), Circular Dependency (bidirectional imports between packages), Missing Interface (concrete type dependencies where abstraction is warranted), Hardcoded Config (non-env configuration values), Missing Error Context (unwrapped errors). Present as a table with: Anti-Pattern, Found?, Where, Suggestion.

3. **Dependency Graph** — generate a Mermaid `graph TD` diagram from actual codebase package/module imports. Show which packages depend on which.

4. **SOLID Compliance** — score each principle (S, O, L, I, D) with: status icon, brief assessment, specific file/package reference if a violation is found. Base scoring on the rules in `.forgia/dev-guide/principles/solid.md`.

5. **Recommendations** — numbered list of actionable items. Each must reference a specific file or package, not generic advice.

### Deliverables

- `modules/claude-commands/fd-arch-review.md` — the complete slash command prompt

### What this SDD does NOT cover

- Updating documentation or review-process.md (SDD-002)
- Go code changes (none needed — embedded skill system auto-loads)
- Making `/fd-arch-review` a gate (it is advisory only)

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `$ARGUMENTS` input | string | FD identifier (e.g., `FD-003`). Validated by the command — must match existing file in `.forgia/fd/` |
| Report output | markdown (stdout) | Structured report with 5 sections as defined in Scope. Rendered to user in Claude Code conversation |
| Vault file reads | filesystem | Reads: FD file, constitution.md, dev-guide/principles/*.md, dev-guide/lang/*.md, guardrails/deny.toml |
| Codebase reads | filesystem | Scans project directory structure, reads package files, import statements. Must respect deny.toml patterns |

## Constraints / Vincoli

- Language / Linguaggio: Markdown (Claude Code slash command prompt)
- Framework: Claude Code slash command system (`$ARGUMENTS` substitution)
- Dependencies / Dipendenze: None — the command is a standalone markdown file
- Patterns / Pattern: Follow the structure of existing commands (see `fd-review.md`, `sdd-dry-run.md` for reference)

### Guardrails (from deny.toml)

The command instructions MUST include explicit directives to:
- Never read files matching `[read]` deny patterns (`**/.env`, `**/*.pem`, `**/*.key`, `**/credentials.json`, etc.)
- Never execute commands matching `[execute]` deny patterns
- Never write or modify any file — the command is read-only
- Before scanning a file path, check it against deny.toml patterns; if it matches, skip and note "blocked by guardrails"

### Additional Constraints

- The report header must include the FD identifier and title
- The command must read the actual codebase (Glob, Grep, Read tools) — not just analyze the FD text
- Pattern detection must be language-aware: use `.forgia/dev-guide/lang/` to know which patterns to look for (e.g., Go: interfaces, packages; Rust: traits, crates)
- Anti-pattern checks must be based on `.forgia/dev-guide/principles/` rules, not hardcoded heuristics
- The command must work on any project with a `.forgia/` vault, not just Forgia itself

## Best Practices

- Error handling: If the FD file doesn't exist, output a clear error: `"FD non trovato: FD-NNN. Verifica l'identificatore."` — do not produce a partial report
- Naming: Report section headers must match exactly: `## Pattern Analysis`, `## Anti-Pattern Detection`, `## Dependency Graph`, `## SOLID Compliance`, `## Recommendations`
- Style: Follow the output format shown in issue #22 (tables with icons, Mermaid graphs, bullet-point SOLID scores)
- The prompt must instruct Claude to load ALL files from `dev-guide/principles/` and `dev-guide/lang/` — not just specific ones, since language conventions are auto-detected during `forgia init`

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Manual | Run `/fd-arch-review FD-001` on the Forgia codebase — verify report has all 5 sections | All sections present |
| Manual | Run `/fd-arch-review FD-999` (non-existent) — verify error message | Error path |
| Manual | Verify report contains actual codebase patterns (e.g., Repository in `internal/vault/`, Runner interface in `internal/runner/`) | Pattern detection accuracy |
| Manual | Verify dependency graph is a valid Mermaid diagram | Mermaid syntax |
| Manual | Verify no files matching deny.toml patterns appear in the report | Guardrails compliance |
| Manual | Verify SOLID scoring references specific files/packages, not generic statements | Recommendation quality |
| Manual | Verify the command does not create, modify, or delete any file | Read-only constraint |

## Acceptance Criteria / Criteri di Accettazione

- [ ] File `modules/claude-commands/fd-arch-review.md` exists and is a valid Claude Code slash command
- [ ] Command accepts `$ARGUMENTS` as FD identifier and validates it against `.forgia/fd/`
- [ ] Report contains all 5 sections: Pattern Analysis, Anti-Pattern Detection, Dependency Graph, SOLID Compliance, Recommendations
- [ ] Pattern Analysis scans actual codebase directory structure using Glob/Grep/Read tools
- [ ] Anti-Pattern Detection checks for all 5 anti-patterns: God Object, Circular Dependency, Missing Interface, Hardcoded Config, Missing Error Context
- [ ] Dependency Graph is a valid Mermaid `graph TD` diagram generated from actual imports
- [ ] SOLID Compliance scores each principle individually (S, O, L, I, D) with file/package references
- [ ] Recommendations are numbered and reference specific files or packages
- [ ] Command reads constitution.md, dev-guide/principles/*.md, dev-guide/lang/*.md, guardrails/deny.toml
- [ ] Command respects deny.toml — never reads files matching deny patterns, notes "blocked by guardrails" when skipping
- [ ] Command is purely read-only — prompt explicitly forbids creating/modifying/deleting files
- [ ] Error handling: non-existent FD produces clear Italian error message, no partial report

## Context / Contesto

- [ ] `.forgia/fd/FD-003-fd-arch-review.md` — parent FD with full requirements
- [ ] `modules/claude-commands/fd-review.md` — existing review command, reference for structure and conventions
- [ ] `modules/claude-commands/sdd-dry-run.md` — existing complex slash command, reference for multi-pass analysis pattern
- [ ] `modules/claude-commands/fd-new.md` — reference for `$ARGUMENTS` parsing and error handling
- [ ] `.forgia/constitution.md` — must be loaded by the command
- [ ] `.forgia/dev-guide/principles/solid.md` — SOLID rules the command must check against
- [ ] `.forgia/dev-guide/principles/clean-code.md` — clean code rules for anti-pattern detection
- [ ] `.forgia/dev-guide/principles/design-patterns.md` — known patterns to detect
- [ ] `.forgia/dev-guide/lang/go.md` — Go-specific patterns and conventions
- [ ] `.forgia/dev-guide/lang/shell.md` — Shell conventions
- [ ] `.forgia/guardrails/deny.toml` — deny patterns the command must respect
- [ ] `internal/skill/embedded.go` — how slash commands are loaded (verify naming convention)
- [ ] GitHub issue: https://github.com/Deepzima/forgia/issues/22 — original requirements and example output format

## Constitution Check

- [ ] Respects code standards — command is a markdown prompt, follows existing slash command conventions
- [ ] Respects commit conventions — commits will use `feat(FD-003): description` format
- [ ] No hardcoded secrets — command reads no secrets, produces no secrets
- [ ] Tests defined and sufficient — 7 manual test scenarios covering all report sections, error paths, guardrails, and read-only constraint

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-28
- **Completed**: 2026-03-28
- **Duration / Durata**: ~10 min

### Decisions / Decisioni

1. Used 8-step analysis flow (resolve → load → scan → patterns → anti-patterns → dependency graph → SOLID → recommendations → report) modeled after `sdd-dry-run.md`'s multi-pass approach but adapted for architecture review
2. Made pattern detection language-aware by requiring the command to read `dev-guide/lang/` and adjust scanning heuristics per language (Go interfaces vs Rust traits vs Python classes)
3. Anti-pattern heuristics are guidelines not rigid rules — e.g., God Object threshold of 500 lines is a heuristic, not absolute
4. SOLID compliance table includes both "Current Codebase" and "FD Impact" columns so the review is useful for both existing architecture and proposed changes

### Output

- **Commit(s)**: pending
- **PR**: pending
- **Files created/modified**:
  - `modules/claude-commands/fd-arch-review.md` (created)

### Retrospective / Retrospettiva

- **What worked**: Using existing slash commands (fd-review, sdd-dry-run) as structural reference ensured consistency
- **What didn't**: N/A
- **Suggestions for future FDs**: Single-file slash command SDDs are fast to execute — good candidate for batch execution
