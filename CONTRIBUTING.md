# Contributing to Forgia

## Getting Started

```bash
git clone git@github.com:forgia-labs/forgia.git
cd forgia
mise trust
mise run go:build
mise run go:test
```

### Prerequisites

| Tool | Required | Install |
|------|----------|---------|
| Go 1.25+ | Yes | `mise install` |
| mise | Yes | [mise.jdx.dev](https://mise.jdx.dev) |
| Docker | Optional | For OpenHands runner |

## Development Workflow

### Branches

- `main` — stable, protected
- `feat/<name>` — new features
- `fix/<name>` — bug fixes
- `chore/<name>` — maintenance, CI, docs

### Making Changes

1. Create a branch from `main`
2. Write code + tests
3. Run `go build ./...` and `go test ./...`
4. Run `mise run go:lint` (if available)
5. Open a PR against `main`

### PR Requirements

- CI must pass (Go test + build + E2E)
- At least 1 reviewer approval
- Commits should be logically grouped (squash if noisy)

## Code Conventions

### Go

- **Go 1.25+** — use `iter.Seq`, `slog`, `testing/synctest` where appropriate
- **Context propagation** — all I/O functions take `ctx context.Context` as first parameter
- **Structured logging** — use `slog.InfoContext(ctx, ...)`, never `log.Printf` or bare `slog.Info`
- **Error wrapping** — always `fmt.Errorf("context: %w", err)`, never `fmt.Errorf("...%v", err)`
- **Interface checks** — add `var _ Interface = (*Struct)(nil)` compile-time checks
- **Type assertions** — always use comma-ok pattern: `v, ok := x.(Type)`
- **Testing** — table-driven tests, `t.Helper()` in helpers, `t.TempDir()` for temp files
- **No `init()`** — explicit initialization only

### YAML / TOML / JSON

- 2-space indent for all

## Project Structure

```
cmd/forgia/           # CLI entry point (Cobra)
internal/
  vault/              # FileVault — .forgia/ read/write
  board/              # GitHub Projects sync
  boardsync/          # Vault ↔ Board adapter
  config/             # TOML config loading
  mcp/                # MCP protocol (ToolProvider, subprocess)
  runner/             # Runner abstraction (Claude, OpenHands)
  guardrails/         # Security rules (deny.toml)
  skill/              # Slash command system
  beads/              # Beads (bd) client
  process/            # Subprocess management
  scm/                # Git/GitHub/GitLab abstraction
modules/
  claude-commands/    # Claude Code slash commands
  vault-template/     # Template for forgia init
tests/                # E2E tests (Bash)
docs/                 # Architecture documentation
```

## Testing

### Go tests

```bash
go test ./...              # all packages
go test ./internal/vault/  # single package
go test -count=1 ./...     # skip cache
go vet ./...               # static analysis
```

### E2E tests

E2E tests are Go tests in the `tests/` package. They build the forgia binary and run
the full command suite against a temp directory.

```bash
go test ./tests/  # E2E suite only
```

### Writing Tests

- Place tests in `*_test.go` alongside the code
- Use `t.TempDir()` for filesystem tests — never write to the repo
- Mock external dependencies (no real GitHub API calls in tests)
- For MCP subprocess tests, use the `TestMain` mock server pattern (see `internal/mcp/subprocess_test.go`)

## Review Process

### For Humans

1. Check that CI passes
2. Verify Go conventions (ctx propagation, error wrapping, slog usage)
3. Look for security issues (path traversal, injection, bare type assertions)
4. Check test coverage for new code paths

### For AI Agents

Agents contributing via Forgia SDD execution must:

1. Follow the SDD scope strictly — no out-of-scope changes
2. Update the Work Log in the SDD file when done
3. Commit with message referencing the SDD ID: `feat(FD-xxxx): SDD-xxx — description`
4. Run `go build ./...` and `go test ./...` before committing
5. Never modify `.forgia/` meta files (config.toml, constitution.md) unless the SDD explicitly requires it

## Security

- Never commit secrets, API keys, or credentials
- Never read files in `~/.gnupg/`, `~/.ssh/`, `~/.password-store/`
- Use `deny.toml` for project-level guardrails
- Pin all CI action versions by commit SHA
- Pin Dockerfile base images by digest

## License

MIT — see [LICENSE](LICENSE)
