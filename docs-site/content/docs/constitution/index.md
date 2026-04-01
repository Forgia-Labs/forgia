---
title: "Constitution"
description: "Immutable project rules that every agent must follow, every time."
navigation:
  title: "Constitution"
---

# Constitution

The constitution is the immutable ruleset for a Forgia project. Every SDD embeds a reference to it, and every agent must comply — no exceptions, no workarounds.

## Location

```
.forgia/constitution.md
```

Created by `forgia init` with sensible defaults. Customize it before your first FD.

## What it covers

A constitution typically defines:

**Code style**
- Formatting rules (line length, indentation, naming conventions)
- Language-specific idioms to use or avoid
- Comment and documentation standards

**Architecture constraints**
- Allowed and disallowed patterns (e.g., no global mutable state)
- Module organization rules
- Dependency management policies

**Security rules**
- No secrets in source code
- Input validation requirements
- Allowed crypto primitives

**Testing requirements**
- Minimum coverage thresholds
- Required test types (unit, integration, E2E)
- Test naming conventions

**Operational rules**
- Logging standards
- Error handling patterns
- Observability requirements

## Guardrails

Alongside the constitution, `.forgia/guardrails/deny.toml` lists file patterns that agents must never read, write, or execute:

```toml
[deny]
read  = [".env", "*.pem", "secrets/**"]
write = ["go.sum", "pnpm-lock.yaml", "*.lock"]
exec  = ["rm -rf", "DROP TABLE", "format C:"]
```

Both `/fd-review` and SDD generation check the constitution and deny list automatically.

## Updating the constitution

The constitution is not truly immutable in the technical sense — you can edit it. The immutability is a team commitment: the rules apply to everyone, including the humans who wrote them.

When you update the constitution:
1. Create an FD documenting the change and the reason
2. Update `.forgia/constitution.md`
3. Review open SDDs for compliance with the new rules

Never update the constitution silently mid-FD. Agents that were given an SDD before the change will not see the updated rules.
