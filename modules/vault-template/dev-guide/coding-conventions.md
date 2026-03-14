# Coding Conventions

Rules for AI agents and developers. Loaded automatically by `/fd-review` and `/fd-verify`.

## General Principles

1. **No silent fallback**: never use default values that hide errors
2. **DRY**: extract helpers for repeated code (but don't abstract prematurely)
3. **No backwards compatibility hacks**: if deprecated, remove it
4. **Explicit errors**: use appropriate return codes and error types
5. **Validate at boundaries**: user input, external APIs, config files

## Language-Specific

### Rust
- `anyhow` for applications, `thiserror` for libraries
- Async: tokio runtime
- Config: TOML files, env vars for secrets

### Python
- Type hints on all public functions
- `uv` for dependency management
- Async: asyncio
- Config: TOML or env vars

### TypeScript
- Strict mode enabled
- Prefer `const` over `let`
- Use explicit return types on exported functions

### Shell (Bash/Zsh)
- `set -euo pipefail` in every script
- Local variables with `local`
- Usage strings on stderr
- Tool check: `command -v tool >/dev/null 2>&1`

## Testing

- Every SDD defines its own test requirements
- Minimum: unit tests for business logic, integration tests for boundaries
- No mocking of databases in integration tests (use real instances)
- Test names describe the behavior, not the method

## Security

- No hardcoded secrets (use env vars or secret managers)
- No command injection (quote all variables in shell)
- No SQL injection (use parameterized queries)
- Validate and sanitize at system boundaries
