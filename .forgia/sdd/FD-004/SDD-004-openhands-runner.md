---
id: "SDD-004"
fd: "FD-004"
title: "OpenHands runner — Docker-based Runner implementation"
status: done
agent: "claude-code"
assigned_to: "claude-code"
created: "2026-03-28"
started: "2026-03-28"
completed: "2026-03-28"
tags: ["enhancement"]
---

# SDD-004: OpenHands runner — Docker-based Runner implementation

> Parent FD: [[FD-004]]

## Scope

Implement `OpenHandsRunner` in `internal/runner/` that satisfies the `Runner` interface. This is the Go equivalent of bash `modules/runners/openhands.sh`.

### Behavior

1. **Check Docker** — verify Docker daemon is running (`docker info`)
2. **Pull image** — if `config.OpenHands.Image` not present locally, pull it (`docker pull`)
3. **Build prompt** — assemble SDD content + system context (constitution, guardrails, principles) into a task prompt
4. **Write prompt to temp file** — write prompt to a temp file that will be mounted into the container
5. **Launch container** — `docker run` with:
   - Volume mount: current directory → `/workspace`
   - Volume mount: temp prompt file → `/tmp/prompt.md`
   - Port binding: `config.OpenHands.UIPort` → container port 3000 (skip if UIPort=0)
   - Environment: API key (`ANTHROPIC_API_KEY` or `OPENAI_API_KEY`), model, max iterations
   - Image: `config.OpenHands.Image` (default: `ghcr.io/openhands/openhands:latest`)
6. **Wait for completion** — block until container exits, capture exit code
7. **Return ExecResult** — populate with SDD ID, runner name, duration, exit code, status

### Config

Read from `[runner.openhands]` section in `config.toml`:
```toml
[runner.openhands]
image = "ghcr.io/openhands/openhands:latest"
model = "claude-sonnet-4-20250514"
workspace_mount = "/workspace"
max_iterations = 100
ui_port = 3000
```

### Registration

Register `OpenHandsRunner` in `runner.Resolve()` so `--runner=openhands` works.

## Interfaces / Interfacce

| Interface / Interfaccia | Type / Tipo | Description / Descrizione |
|-------------------------|-------------|---------------------------|
| `Runner` interface | Go interface | `Execute(ctx, sdd, opts) → ExecResult`, `Name() → "openhands"` |
| Docker CLI | subprocess | `docker info`, `docker pull`, `docker run` via `exec.CommandContext` |
| Config | struct | `config.OpenHandsConfig` — image, model, workspace, max_iterations, ui_port |
| System context builder | function | `runner.BuildSystemContext(ctx, vault)` — reuse existing function |

## Constraints / Vincoli

- Language / Linguaggio: Go
- Framework: `os/exec` for Docker commands, `internal/runner` for interface
- Dependencies / Dipendenze: Docker daemon must be running
- Patterns / Pattern: Follow `ClaudeRunner` implementation pattern in `internal/runner/claude.go`

### Guardrails (from deny.toml)

- API keys passed via `-e` env flag to Docker — never written to files or logged
- Prompt file is a temp file, cleaned up after execution

### Security

- API keys injected as Docker environment variables (`-e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY`)
- Never write API keys to the prompt file, log output, or any persistent storage
- Temp prompt file deleted after container exits

## Best Practices

- Error handling: Check Docker availability first — return clear error: `"Docker non disponibile. Installa Docker: https://docs.docker.com/get-docker/"`
- Error handling: Image pull failure → return error with image name and pull output
- Naming: `OpenHandsRunner` struct, `NewOpenHandsRunner(cfg config.OpenHandsConfig)` constructor
- Style: Use `context.Context` for cancellation, `slog` for structured logging

## Test Requirements

| Type / Tipo | What / Cosa | Coverage |
|-------------|-------------|----------|
| Unit | `NewOpenHandsRunner()` — verify config defaults | Constructor |
| Unit | Docker command construction — verify volume mounts, env vars, port binding | Command building |
| Unit | `runner.Resolve("openhands")` returns `OpenHandsRunner` | Registration |
| Integration | Full execution with Docker (requires Docker daemon) | End-to-end (CI-optional) |

## Acceptance Criteria / Criteri di Accettazione

- [x] `OpenHandsRunner` implements `Runner` interface (`Execute`, `Name`)
- [x] `runner.Resolve("openhands")` returns `OpenHandsRunner`
- [x] Docker availability checked before execution — clear error if missing
- [x] Image pulled if not present locally
- [x] Container launched with correct volume mounts, env vars, port binding
- [x] API keys passed via environment variables, never written to files
- [x] `ExecResult` populated with SDD ID, duration, exit code, status
- [x] Temp prompt file cleaned up after execution
- [x] Config read from `[runner.openhands]` with sensible defaults
- [x] All new functions have unit tests with `t.Parallel()`

## Context / Contesto

- [x] `internal/runner/runner.go` — `Runner` interface, `ExecResult`, `ExecOptions`
- [x] `internal/runner/claude.go` — `ClaudeRunner` as reference implementation
- [x] `internal/config/config.go` — `OpenHandsConfig` struct (already defined)
- [x] `modules/runners/openhands.sh` — bash implementation as reference
- [x] `internal/runner/dryrun.go` — `BuildSystemContext()` function to reuse

## Constitution Check

- [x] Respects code standards — Go conventions, error wrapping, context propagation
- [x] Respects commit conventions — `feat(FD-004): description`
- [x] No hardcoded secrets — API keys via env vars only
- [x] Tests defined and sufficient — unit + integration (Docker-dependent tests CI-optional)

---

## Work Log / Diario di Lavoro

> This section is **mandatory**. Must be filled by the agent or developer during and after execution.

### Agent / Agente

- **Executor**: claude-code
- **Started**: 2026-03-28
- **Completed**: 2026-03-28
- **Duration / Durata**: ~15 min

### Decisions / Decisioni

1. Extended `Resolve()` with variadic `config.OpenHandsConfig` parameter instead of changing the signature to require config — this keeps backward compatibility for callers that don't need OpenHands config (e.g., batch.go) while allowing exec.go to pass the loaded config.
2. API keys passed to Docker via `-e ANTHROPIC_API_KEY` (env var name only, not value) — Docker inherits the value from the host environment. This prevents key leakage in logs or process listings.
3. Followed ClaudeRunner pattern closely: compile-time interface check, same logging style, same exec log format, same ExecResult population.
4. Prompt file uses `os.CreateTemp` with `defer os.Remove` for guaranteed cleanup.
5. Added OpenHands defaults in `config.DefaultConfig()` so zero-config usage works out of the box.

### Output

- **Commit(s)**: (pending)
- **PR**: (pending)
- **Files created/modified**:
  - `internal/runner/openhands.go` — new: OpenHandsRunner implementation
  - `internal/runner/openhands_test.go` — new: 18 unit tests
  - `internal/runner/claude.go` — modified: Resolve() extended with openhands case + config import
  - `internal/config/config.go` — modified: DefaultConfig() includes OpenHands defaults
  - `cmd/forgia/cmd/exec.go` — modified: passes OpenHands config to Resolve()

### Retrospective / Retrospettiva

- **What worked / Cosa ha funzionato**: Il pattern ClaudeRunner ha fornito un template chiaro; la struttura OpenHandsConfig era già definita nel package config; il bash runner ha chiarito il flusso Docker atteso.
- **What didn't / Cosa non ha funzionato**: Nulla di bloccante.
- **Suggestions for future FDs / Suggerimenti per FD futuri**: Considerare un'interfaccia `RunnerConfig` generica per evitare la proliferazione di parametri variadic in `Resolve()` quando si aggiungono nuovi runner.
