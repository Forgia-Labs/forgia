---
title: "forgia init"
description: "Initialize the Forgia vault in an existing project."
navigation:
  title: "init"
---

# forgia init

```
forgia init [--lang <language>] [--vault <path>]
```

Initializes the Forgia vault in the current directory.

## What gets created

```
.forgia/
├── constitution.md          # Immutable project rules
├── dev-guide/
│   ├── principles/
│   │   ├── clean-code.md
│   │   ├── SOLID.md
│   │   └── design-patterns.md
│   └── lang/
│       └── <detected>.md    # Language-specific conventions
├── fd/
│   └── _templates/
│       └── fd-template.md
├── sdd/
│   └── _templates/
│       └── sdd-template.md
├── guardrails/
│   └── deny.toml            # Denied file patterns and operations
└── learnings/               # Retrospective feed for future FDs
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--lang` | auto-detected | Primary language (`go`, `ts`, `rust`, `python`) |
| `--vault` | `.forgia/` | Custom vault path |
| `--force` | false | Overwrite existing vault |

## Language detection

Without `--lang`, `forgia init` inspects the project root for:

- `go.mod` → Go
- `package.json` → TypeScript/JavaScript
- `Cargo.toml` → Rust
- `pyproject.toml` / `requirements.txt` → Python

The detected language determines which `lang/` convention file is generated.

## After init

1. Review and customize `.forgia/constitution.md` — add project-specific rules
2. Review `.forgia/guardrails/deny.toml` — add paths that agents must never modify
3. Create your first FD: `/fd-new "description"`
