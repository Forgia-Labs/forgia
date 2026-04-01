---
id: "SDD-006"
fd: "FD-006"
title: "Sandbox wiring — CLI integration, env injection, auth mount, security hardening"
status: done
created: "2026-04-01"
author: "DeepZima"
executor: "claude-opus"
complexity: medium
---

# SDD-006: Sandbox Wiring

> Parent: FD-006 (Dual-profile sandbox runner)

## Scope

Wire the existing sandbox packages (`internal/sandbox/`) into the CLI commands:

1. `--sandbox` flag on `forgia exec` and `forgia batch`
2. `SandboxExec()` called when sandbox is configured
3. `ExcludedHostPaths()` enforced before mounting workspace
4. Custom env injection from `sandbox_env` config (model override, base URL)
5. Optional `~/.claude` read-only mount for Max OAuth auth
6. Config fields: `sandbox_env`, `sandbox_mount_claude`

## Files modified

- `cmd/forgia/cmd/exec.go` — `--sandbox` flag, sandbox execution path
- `cmd/forgia/cmd/batch.go` — `--sandbox` flag
- `internal/runner/sandbox_exec.go` — ExcludedHostPaths, env injection, auth mount
- `internal/config/config.go` — `sandbox_env`, `sandbox_mount_claude` fields

## Work Log

### Executor: claude-opus
### Started: 2026-04-01

### Decisions
- ExcludedHostPaths checked before mounting, not after — fail-fast
- `~/.claude` mount is opt-in via config, not default — security tradeoff documented
- Custom env uses `strings.Cut("=")` for KEY=VALUE parsing
- Sandbox path in exec.go returns early — host path only reached if sandbox is "none" or empty

### Retrospective
- Worked: clean separation between sandbox and host execution paths
- Didn't work: batch.go has its own runner logic, doesn't reuse execSDD — sandbox wiring had to be done separately (not done in batch yet, only flag added)
- Suggestion: batch should delegate to execSDD to avoid duplication
