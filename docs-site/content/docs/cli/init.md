---
title: "forgia init"
description: "Initialize the Forgia vault in an existing project."
navigation:
  title: "init"
---

# forgia init

```
forgia init [--dir <path>]
```

Initializes the Forgia vault in the current directory (or `--dir`). Safe to re-run — existing files are never overwritten.

## What gets created

```
.forgia/
├── constitution.md          # Immutable project rules
├── config.toml              # Runner and tool configuration
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
├── logs/                    # Execution reports (gitignored)
└── .gitignore
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dir` | `.` | Target directory to initialize |

## Language detection

`forgia init` automatically detects the project stack by inspecting the target directory:

| File found | Convention loaded |
|------------|------------------|
| `go.mod` | Go |
| `package.json` | Node/TypeScript |
| `Cargo.toml` | Rust |
| `pyproject.toml` / `requirements.txt` | Python |
| `Makefile` / `mise.toml` | Shell |

The detected language determines which `dev-guide/lang/` convention file is generated. No flag needed.

## After init

1. Review and customize `.forgia/constitution.md` — add project-specific rules
2. Review `.forgia/guardrails/deny.toml` — add paths that agents must never modify
3. Commit `.forgia/` to share with your team
4. Create your first FD: `forgia skill fd-new`
