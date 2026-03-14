# Rust Conventions

> Auto-loaded when `Cargo.toml` is detected in the project.

## Project Structure

- Workspace: `Cargo.toml` at root with `[workspace]` for multi-crate projects
- Crate naming: `kebab-case` for crate names, `snake_case` for modules
- Binary crates: `src/main.rs`, library crates: `src/lib.rs`
- Tests: `tests/` for integration, `#[cfg(test)] mod tests` for unit

## Error Handling

- **Applications** (`main.rs`, CLI, servers): use `anyhow::Result` for ergonomic error propagation
- **Libraries** (reusable crates): use `thiserror::Error` for typed, descriptive errors
- Never use `.unwrap()` in production code — use `.expect("reason")` only when panic is intentional
- Propagate errors with `?` operator, avoid manual `match` on `Result` unless adding context

```rust
// Library error
#[derive(Debug, thiserror::Error)]
pub enum GatewayError {
    #[error("LDAP connection failed: {0}")]
    LdapConnection(#[from] ldap3::LdapError),
    #[error("token expired for agent {agent_id}")]
    TokenExpired { agent_id: String },
}

// Application error
fn main() -> anyhow::Result<()> {
    let config = Config::load().context("failed to load config")?;
    Ok(())
}
```

## Async

- Runtime: `tokio` (multi-threaded by default)
- Use `#[tokio::main]` for entry points
- Prefer `tokio::spawn` for concurrent tasks, `tokio::select!` for racing
- Channels: `tokio::sync::mpsc` for multi-producer, `tokio::sync::watch` for broadcast

## Traits & Generics

- Use `async_trait` for async trait methods (until native async traits stabilize)
- Prefer `impl Trait` in argument position over explicit generics when there's only one generic
- Use `where` clauses for complex bounds

## Dependencies

- HTTP: `axum` (server), `reqwest` (client)
- Serialization: `serde` + `serde_json` / `toml`
- CLI: `clap` with derive
- Logging: `tracing` (not `log`)
- Config: TOML files, env vars via `std::env` or `dotenvy`

## Style

- `cargo fmt` — always, no exceptions
- `cargo clippy` — treat warnings as errors in CI (`-D warnings`)
- Edition: 2021 (or latest stable)
- Imports: group by `std`, external crates, internal modules (separated by blank lines)
- Naming: `snake_case` functions/variables, `PascalCase` types/traits, `SCREAMING_SNAKE` constants
- Documentation: `///` for public items, `//!` for module-level docs

## Testing

- Unit tests: `#[cfg(test)] mod tests` in the same file
- Integration tests: `tests/` directory
- Use `#[test]` for sync, `#[tokio::test]` for async
- Test names: `test_<behavior>_<scenario>` (e.g., `test_auth_rejects_expired_token`)
- Assertions: prefer `assert_eq!` / `assert_ne!` over `assert!` for better error messages

## Common Patterns

```rust
// Builder pattern for config
let server = ServerBuilder::new()
    .port(8080)
    .tls(true)
    .build()?;

// Type-state pattern for state machines
struct Connection<S: State> { inner: TcpStream, _state: PhantomData<S> }
```

## What NOT to Do

- Don't use `String` where `&str` suffices
- Don't clone to satisfy the borrow checker — restructure instead
- Don't use `Rc<RefCell<T>>` unless you have a specific reason (prefer ownership transfer)
- Don't write `unsafe` without a `// SAFETY:` comment explaining the invariant
