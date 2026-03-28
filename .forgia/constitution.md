# Constitution

> EN: Immutable project rules. Every FD, SDD, and implementation must respect these principles.
> IT: Regole immutabili del progetto. Ogni FD, SDD e implementazione deve rispettare questi principi.
>
> Inspired by [GitHub Spec Kit](https://github.com/github/spec-kit).

---

## Principles / Principi

1. **Spec first, code second** — no implementation without an approved FD and generated SDD
2. **Fail-closed** — if a gate fails (review, verify), work stops until resolved
3. **Work Log is mandatory** — every SDD must have a completed Work Log before closing
4. **Constitution is immutable** — changes require explicit team consensus and versioning
5. **Agent isolation** — each SDD executes in its own worktree or container

## Code Standards

- Tests are required for every SDD (type and coverage defined in the SDD itself)
- No silent fallbacks — errors must be explicit
- No backwards-compatibility hacks — if deprecated, remove it

## Security

- **Guardrails are absolute** — `.forgia/guardrails/deny.toml` defines what agents CANNOT do
- No hardcoded secrets — use environment variables or secret managers
- No command injection — validate and sanitize all external input
- Validate at system boundaries (user input, external APIs), trust internal code
- Agents must use placeholders for secret values, never real credentials
- If an agent encounters a file that might contain secrets, it must SKIP and warn
- The deny list is fail-closed: if a pattern matches, the action is blocked

## Commit Conventions

- Format: `feat|fix|docs|refactor|test: description`
- FD-linked: `feat(FD-NNN): description`
- AI-generated commits include `Co-Authored-By`

## Communication

- Italian for documentation and communication
- English for code, variable names, comments in code

---

> EN: Edit this file to match your project's specific rules.
> IT: Modifica questo file con le regole specifiche del tuo progetto.
>
> This is loaded by `/fd-review` and `/fd-verify` as the baseline for all checks.
