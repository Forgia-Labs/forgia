# Go Conventions

> Auto-loaded when `go.mod` is detected in the project.

## Project Structure

- Follow [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- `cmd/<app>/main.go` for entry points
- `internal/` for private packages (not importable by other modules)
- `pkg/` for public packages (if publishing a library)
- Tests: colocated `*_test.go` files

## Error Handling

- Always check errors — never ignore with `_`
- Wrap errors with context: `fmt.Errorf("failed to connect: %w", err)`
- Use sentinel errors for expected conditions: `var ErrNotFound = errors.New("not found")`
- Use custom error types for complex errors
- Errors are values, not exceptions — handle them at every level

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("load config %s: %w", path, err)
    }
    var cfg Config
    if err := toml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config %s: %w", path, err)
    }
    return &cfg, nil
}
```

## Concurrency

- Use goroutines for concurrent work
- Communicate via channels, not shared memory
- Use `context.Context` for cancellation and timeouts
- Use `sync.WaitGroup` for fan-out/fan-in
- Use `errgroup.Group` for goroutines that can fail

## Interfaces

- Keep interfaces small (1-3 methods)
- Define interfaces where they're consumed, not where they're implemented
- Use `io.Reader`, `io.Writer`, `fmt.Stringer` as examples of good interfaces
- Accept interfaces, return structs

```go
// Good — small, defined at consumer
type Authenticator interface {
    Authenticate(ctx context.Context, token string) (*Agent, error)
}

// Bad — too large, defined at implementation
type UserService interface {
    Create(ctx context.Context, u User) error
    Get(ctx context.Context, id string) (*User, error)
    Update(ctx context.Context, u User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter Filter) ([]User, error)
    // ...20 more methods
}
```

## Dependencies

- HTTP: `net/http` (stdlib) or `chi` router
- CLI: `cobra` + `viper`
- Database: `sqlx` or `pgx`
- Config: `viper` or `envconfig`
- Logging: `slog` (stdlib, Go 1.21+)
- Testing: stdlib `testing` + `testify`

## Style

- Formatter: `gofmt` or `goimports` — mandatory, no exceptions
- Linter: `golangci-lint` with default + `errcheck`, `govet`, `staticcheck`
- Naming: `camelCase` unexported, `PascalCase` exported, short variable names in small scopes
- Package names: short, lowercase, no underscores (e.g., `auth`, `store`, `gateway`)
- Comments: `// FunctionName does X` for exported functions (godoc style)

## Testing

- Table-driven tests for multiple cases
- `testify/assert` for assertions (or stdlib)
- `httptest.NewServer` for HTTP integration tests
- Use `t.Parallel()` for independent tests
- Test names: `Test<Function>_<scenario>`

```go
func TestAuthenticate_ExpiredToken(t *testing.T) {
    t.Parallel()
    auth := NewAuthenticator(testKey)
    _, err := auth.Authenticate(ctx, expiredToken)
    assert.ErrorIs(t, err, ErrTokenExpired)
}
```

## What NOT to Do

- Don't use `init()` functions — explicit initialization is clearer
- Don't use package-level variables for state — use dependency injection
- Don't use `panic` for error handling — return errors
- Don't use `interface{}` / `any` when a concrete type works
- Don't create packages named `util`, `common`, `helper` — name by purpose
