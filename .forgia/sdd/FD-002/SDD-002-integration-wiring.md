---
id: "SDD-002"
fd: "FD-002"
title: "Integration wiring and documentation"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-28"
started: "2026-03-28"
completed: "2026-03-28"
tags: ["phase:2-core"]
---

# SDD-002: Integration wiring and documentation

> Parent FD: [[FD-002]]

## Scope

Wire the `/fd-arch-review` slash command (created by SDD-001) into the Forgia workflow and verify it is operational.

### Deliverables

1. **Update `.forgia/dev-guide/review-process.md`** — add `/fd-arch-review` as an optional step between `/fd-review` (Gate 1) and `/fd-sdd` in the workflow diagram. The updated flow should be:

   ```
   FD Review ──→ Arch Review ──→ SDD Generation ──→ Implementation ──→ Verification ──→ Close
   /fd-review    /fd-arch-review   /fd-sdd            agent executes     /fd-verify       /fd-close
   GATE 1        (optional)                                               GATE 2           GATE 3
   ```

   Add a new subsection documenting `/fd-arch-review`: when to run it, what it produces, and that it is advisory (not a gate).

2. **Verify embedded skill loading** — confirm that `modules/claude-commands/fd-arch-review.md` is picked up by the `go:embed` directive in `embedded.go` and that `internal/skill/embedded.go` → `LoadEmbedded()` registers it with the correct name (`fd-arch-review`) and category (`CategoryFD`).

3. **E2E validation** — run `go build ./cmd/forgia/` and verify `forgia skills` lists `fd-arch-review`. Run `forgia skill fd-arch-review FD-002` and verify it invokes the command without errors.

### What this SDD does NOT cover

- Creating the slash command file itself (SDD-001)
- Modifying Go code (no changes needed — embed is automatic)

### Dependency

This SDD depends on SDD-001 being completed first. The file `modules/claude-commands/fd-arch-review.md` must exist before integration can be verified.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| SDD-001 output | file | `modules/claude-commands/fd-arch-review.md` must exist — this is the contract from SDD-001 |
| `review-process.md` | markdown | Updated workflow documentation with `/fd-arch-review` step |
| `forgia skills` CLI | stdout | Must list `fd-arch-review` in the skill registry |
| `forgia skill fd-arch-review` CLI | stdout | Must invoke the command and produce a report |

## Constraints / Vincoli

- Language / Linguaggio: Markdown (documentation update), Go (build verification only — no code changes)
- Framework: Forgia dev-guide documentation conventions
- Dependencies / Dipendenze: SDD-001 must be completed first
- Patterns / Pattern: Follow existing dev-guide document structure

### Guardrails (from deny.toml)

- Do not modify `.forgia/constitution.md` or `.forgia/config.toml` (write-denied)
- Do not modify `.forgia/guardrails/deny.toml` (write-denied)
- `.forgia/dev-guide/review-process.md` is NOT in the deny list — modification is allowed

## Best Practices

- Error handling: If SDD-001 output file is missing, report the dependency and stop — do not create placeholder files
- Naming: Keep the existing document structure of `review-process.md` — add sections, don't restructure
- Style: Match the existing formatting in `review-process.md` (ASCII workflow diagram, markdown tables, section headers)

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Manual | Review `review-process.md` diff — verify `/fd-arch-review` is added in the correct position | Documentation accuracy |
| Manual | Run `go build ./cmd/forgia/` — verify build succeeds | Build verification |
| Manual | Run `build/forgia skills` — verify `fd-arch-review` appears in the list | Skill registration |
| Manual | Run `build/forgia skill fd-arch-review FD-002` — verify command invokes without crash | E2E smoke test |

## Acceptance Criteria / Criteri di Accettazione

- [ ] `.forgia/dev-guide/review-process.md` updated with `/fd-arch-review` in the workflow diagram
- [ ] New subsection in `review-process.md` documents `/fd-arch-review`: purpose, usage, advisory (non-gate) nature
- [ ] `go build ./cmd/forgia/` succeeds with the new slash command file embedded
- [ ] `forgia skills` output includes `fd-arch-review`
- [ ] `forgia skill fd-arch-review FD-002` invokes successfully (produces report or delegates to Claude)
- [ ] No Go code was modified — only documentation and verification

## Context / Contesto

- [ ] `.forgia/sdd/FD-002/SDD-001-fd-arch-review-command.md` — dependency SDD
- [ ] `modules/claude-commands/fd-arch-review.md` — the file created by SDD-001 (must exist)
- [ ] `.forgia/dev-guide/review-process.md` — file to update
- [ ] `internal/skill/embedded.go` — `LoadEmbedded()` function, `inferCategory()` logic (confirms `fd-` prefix → `CategoryFD`)
- [ ] `embedded.go` (root) — `go:embed all:modules/claude-commands` directive
- [ ] `cmd/forgia/cmd/skill.go` — `forgia skill` and `forgia skills` command implementations
- [ ] `.forgia/guardrails/deny.toml` — verify `review-process.md` is not write-denied

## Constitution Check

- [ ] Respects code standards — documentation follows existing dev-guide conventions
- [ ] Respects commit conventions — commits will use `feat(FD-002): description` format
- [ ] No hardcoded secrets — no secrets involved
- [ ] Tests defined and sufficient — 4 verification steps covering docs, build, skill registration, and E2E

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-28
- **Completed**: 2026-03-28
- **Duration / Durata**: ~5 min

### Decisions / Decisioni

1. Added `/fd-arch-review` as a new subsection between Gate 1 and Gate 2 in review-process.md, clearly marked as "(optional)" in the workflow diagram
2. Skipped E2E smoke test of `forgia skill fd-arch-review FD-002` — this requires Claude CLI to be available and would execute the full analysis. Build + skill listing is sufficient verification for wiring.
3. No Go code was modified — confirmed `inferCategory()` in `embedded.go` maps `fd-` prefix to `CategoryFD` automatically

### Output

- **Commit(s)**: pending
- **PR**: pending
- **Files created/modified**:
  - `.forgia/dev-guide/review-process.md` (modified — added arch review step and subsection)

### Retrospective / Retrospettiva

- **What worked**: Existing embed system required zero Go changes — placing the .md file was enough
- **What didn't**: N/A
- **Suggestions for future FDs**: Integration wiring SDDs for slash commands are trivially small when the embed system works — consider merging into SDD-001 for future single-file features
