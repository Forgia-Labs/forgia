# Concepts

## The Three Layers

Forgia organizes software development into three distinct layers:

```
FD (Feature Design)     →     SDD (Spec-Driven Dev)     →     Code
human decides                  agent executes                   output
```

### FD — Feature Design

The FD is a **functional design document written for humans**. It captures:

- **What** needs to be built and **why**
- Trade-offs between alternative solutions
- Architecture at the component level
- Interfaces between components
- Verification criteria

The FD is technology-aware but implementation-agnostic. It says "we need a REST API for authentication" but doesn't say "use axum with tower middleware on port 8080".

**Lifecycle**: `planned → design → review → approved → in-progress → complete → closed`

### SDD — Spec-Driven Development

The SDD is an **execution contract written for agents**. It's not documentation — it's a structured prompt that contains everything an agent needs to work autonomously:

- **Scope**: exactly what to build (derived from the FD)
- **Interfaces**: types, traits, API contracts
- **Constraints**: language, framework, versions
- **Best Practices**: error handling, naming, style
- **Test Requirements**: what to test, expected coverage
- **Acceptance Criteria**: when the work is "done"
- **Context**: files to read, existing code to understand
- **Work Log**: mandatory record of execution (filled by the agent)

**Cardinality**: 1 FD → N SDDs. Each SDD covers one component, crate, or service.

**Lifecycle**: `planned → assigned → in-progress → done | failed`

### Constitution

The constitution is a set of **immutable project rules** that apply to every FD, SDD, and implementation. Inspired by [GitHub Spec Kit](https://github.com/github/spec-kit).

It defines: code standards, commit conventions, security requirements, and any project-specific invariants.

## The Workflow

```
/fd-new          Create FD (planned)
    ↓
/fd-deep         Optional: multi-agent exploration for complex problems
    ↓
/fd-review       GATE 1: mandatory review (planned → approved)
    ↓
/fd-sdd          Generate N SDDs from approved FD
    ↓
/sdd-assign      Assign each SDD to an agent
    ↓
mise run sdd     Execute SDD (OpenHands headless)
                 Or: Claude Code executes interactively
                 Or: developer implements manually
    ↓
/fd-verify       GATE 2: verify all SDDs complete + Work Logs filled
    ↓
/fd-close        GATE 3: archive, changelog, retrospective
```

## Agent Backends

| Backend | Mode | Best For |
|---------|------|----------|
| **OpenHands** | Autonomous, Docker container | Isolated SDDs, parallel execution, CI/CD |
| **Claude Code** | Interactive, local worktree | Complex SDDs, human-in-the-loop, exploration |
| **Manual** | Developer reads SDD as spec | When AI isn't the right tool |

## Work Log

Every SDD includes a mandatory Work Log with four sections:

1. **Agent**: who executed, when, how long
2. **Decisions**: deviations from plan, problems encountered
3. **Output**: commits, PRs, files created/modified
4. **Retrospective**: what worked, what didn't, suggestions

The retrospective feeds back into better FDs — it's the learning loop.

## Relationship to Other Frameworks

| Concept | Spec Kit | BMAD | Forgia |
|---------|----------|------|--------|
| Functional spec | spec.md | PRD | FD |
| Technical plan | plan.md | Architecture doc | SDD |
| Rules | constitution.md | Persona definitions | constitution.md |
| Agent roles | N/A | 12+ specialized agents | Flexible (any backend) |
| Task tracking | tasks/ | Scrum board | ops/ + Obsidian dataview |
| Execution | Manual | Multi-agent | OpenHands + Claude Code |
| Work history | N/A | N/A | Work Log (mandatory) |
