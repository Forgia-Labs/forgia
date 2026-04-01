---
title: "Feature Lifecycle"
description: "How a feature moves from idea to merged code — and which spell triggers each transition."
navigation:
  title: "Feature Lifecycle"
---

# Feature Lifecycle

Every feature in Forgia follows a fixed pipeline. Humans own the design phase; agents own the execution phase. The slash commands (`/fd-*`, `/sdd-*`) are the gates that move work from one phase to the next.

## Pipeline Overview

```
 Idea
  │
  ▼
/fd-new ──────────────► FD [planned]
                              │
                    ┌─────────┴──────────┐
                    │ (optional, advised) │
                    ▼                    ▼
            /fd-arch-review      /fd-threat-model
            architecture         security analysis
            report               threat model file
                    │                    │
                    └─────────┬──────────┘
                              │ findings inform
                              ▼
                        /fd-review
                         │       │
                       fail     pass
                         │       │
                         ▼       ▼
                      (edit) FD [approved]
                                 │
                                 ▼
                           /fd-sdd
                     (reads arch + threat model)
                                 │
                    ┌────────────┼────────────┐
                    ▼            ▼            ▼
                 SDD-001      SDD-002      SDD-003
                    │            │            │
                    └────────────┴────────────┘
                                 │
                           /sdd-assign
                                 │
                                 ▼
                         agent executes
                         fills Work Log
                                 │
                           /fd-verify
                         │           │
                       fail         pass
                         │           │
                         ▼           ▼
                   (re-assign)  /fd-close
                                     │
                                     ▼
                                  Closed
```

---

## Phase 1 — Design (Human)

### `/fd-new` — Create a Feature Design

The starting point. Give it a free-text description or a GitHub issue number.

```bash
/fd-new "Add OAuth2 login"
# or
/fd-new #42
```

Creates `.forgia/fd/FD-NNN-<slug>.md` with status `planned`. Claude pre-fills the template with a problem statement, candidate solutions, and a skeleton architecture — **you fill in the decisions**.

> The FD is a design document for humans. The goal is to capture *what* and *why*, not *how*.

---

### `/fd-arch-review` — Architecture Analysis (advisory)

Run after drafting the FD's architecture section to validate it against the codebase, design patterns, and SOLID principles.

```bash
/fd-arch-review FD-001
```

Produces a structured report covering:

```
1. Pattern Analysis ......... detected patterns in the codebase vs. proposed
2. Anti-Pattern Detection ... God Object, circular deps, missing interfaces, etc.
3. Dependency Graph ......... import/package relationships
4. SOLID Compliance ......... score per principle (S, O, L, I, D)
5. Recommendations .......... actionable, file-specific suggestions
```

This command is **read-only and advisory** — it does not change the FD status. Run it iteratively while authoring the architecture section. Its findings inform `/fd-review` and its recommendations are automatically picked up by `/fd-sdd` when generating SDD constraints.

---

### `/fd-threat-model` — Security Analysis (advisory)

Run after the architecture is stable to identify threats before implementation begins.

```bash
/fd-threat-model FD-001
```

Writes `.forgia/fd/FD-001-threat-model.md` containing:

```
Assets ...................... what the feature protects or exposes
Threat Actors ............... who might attack and how
STRIDE Analysis ............. per-component threat table
  S — Spoofing
  T — Tampering
  R — Repudiation
  I — Information Disclosure
  D — Denial of Service
  E — Elevation of Privilege
SDD Recommendations ......... mitigations mapped to specific SDDs
Guardrails Suggestions ...... proposed deny.toml additions (advisory)
```

Also **advisory and read-only** (only creates the threat model file). If present, `/fd-review` flags its existence and `/fd-sdd` automatically injects the mitigations into each SDD's Constraints section.

---

### `/fd-review` — Gate: Design Quality

A mandatory review that blocks progress until the FD meets the bar.

```bash
/fd-review FD-001
```

The review checks five things in order:

```
1. Problem clearly defined? .............. (what + why, not how)
2. At least 2 solutions with pros/cons? .. (decision record)
3. Architecture diagrams present? ........ (non-trivial)
4. Interfaces between components defined?  (inputs, outputs, protocols)
5. Verification criteria testable? ....... (concrete, not vague)

All pass → status: approved
Any fail → status stays: planned  (fix and re-run)
```

---

## Phase 2 — Spec (Bridge)

### `/fd-sdd` — Generate Execution Specs

Translates the approved FD into one SDD per component.

```bash
/fd-sdd FD-001
```

```
FD-001 (approved)
    │
    ├──► .forgia/sdd/FD-001/SDD-001-auth-service.md
    ├──► .forgia/sdd/FD-001/SDD-002-token-store.md
    └──► .forgia/sdd/FD-001/SDD-003-login-ui.md
```

Each SDD is a self-contained agent contract: scope, interfaces, constraints, best practices, test requirements, and acceptance criteria. An agent reading only the SDD has everything it needs to implement its component.

---

## Phase 3 — Execution (Agent)

### `/sdd-assign` — Run an Agent

Assigns one SDD to an agent for implementation.

```bash
# Interactive (Claude Code in this session)
/sdd-assign SDD-001 claude

# Autonomous (OpenHands in Docker)
mise run sdd .forgia/sdd/FD-001/SDD-001.md

# Batch all SDDs for an FD in parallel
forgia batch FD-001
```

The agent flow per SDD:

```
Human: /sdd-assign SDD-001
  │
  ▼
Agent reads: scope, interfaces, constraints, acceptance criteria
  │
  ▼
Agent implements code
  │
  ▼
Agent fills Work Log: decisions made, output produced, issues found
  │
  ▼
SDD status → done
```

The Work Log is mandatory. It records what the agent decided, what it produced, and any issues encountered.

---

## Phase 4 — Close (Human)

### `/fd-verify` — Gate: Completion Check

Before closing, verify all SDDs are done and acceptance criteria are met.

```bash
/fd-verify FD-001
```

Checks:
- All SDDs have status `done`
- All Work Logs are filled
- Acceptance criteria from the FD pass

Failures send the relevant SDD back to `/sdd-assign`.

---

### `/fd-close` — Archive the Feature

Final step. Archives the FD, updates the changelog, aggregates retrospectives.

```bash
/fd-close FD-001
```

Status becomes `closed`. The retrospective data from all Work Logs is consolidated into the FD for future reference.

---

## Full Lifecycle at a Glance

```
FD status:    planned ──► approved ──► (SDDs generated)
                 ▲              
                 │ fail         
              /fd-review        
                                
SDD status:   ready ──► in-progress ──► done
                             ▲               │
                             │ fail          │ pass
                          /fd-verify ◄───────┘

FD status:    (all SDDs done) ──► verified ──► closed
```

## Spell Reference

| Spell | Phase | Who | Gate? | Purpose |
|-------|-------|-----|-------|---------|
| `/fd-new` | Design | Human | — | Create FD from idea or issue |
| `/fd-explore` | Design | Human | — | Load FD context for editing |
| `/fd-arch-review` | Design | Claude | advisory | Validate architecture: patterns, SOLID, anti-patterns |
| `/fd-threat-model` | Design | Claude | advisory | STRIDE security analysis, writes threat model file |
| `/fd-review` | Design | Claude | **gate** | Enforce design quality, blocks on failure |
| `/fd-sdd` | Spec | Claude | — | Generate SDDs from approved FD |
| `/sdd-assign` | Execution | Human/Auto | — | Assign SDD to agent |
| `/fd-verify` | Close | Claude | **gate** | Check all SDDs done and criteria met |
| `/fd-close` | Close | Human | — | Archive and consolidate |
| `/fd-status` | Any | Human | — | Dashboard of all FDs and SDDs |

---

Next: dive into [Creating an FD](/docs/fd/creating) for the full FD authoring guide.
