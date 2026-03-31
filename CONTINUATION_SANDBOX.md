# Forgia #32 — Dual-Profile Sandbox — Continuation Prompt

> Copia questo prompt in una nuova sessione Claude dalla directory forgia.

---

## Context

Stai lavorando su **Forgia** — spec-driven development framework per AI agents. Repo: `forgia-labs/forgia`.

### Stato attuale del progetto

- **v0.7.0** rilasciata — Go binary completo, 19 slash commands, 11 CLI commands
- **bin/forgia** (Bash) rimosso — Go binary è l'unico CLI
- **28 issue chiuse, 15 aperte**
- **Zero PR aperte**
- Branch `feat/dual-profile-sandbox` creato da main (aggiornato)

### Issue #32 — Dual-profile runner

Il problema: oggi `forgia exec` rulla Claude con `--dangerously-skip-permissions` sulla macchina host. L'agente vede tutto — SSH keys, secrets, network.

La soluzione: **due profili + sandbox isolato**.

**Host (architect)**: tu interattivo, `/fd-new`, `/fd-review`. Permessi normali.
**Sandbox (builder)**: agente autonomo, `forgia exec/batch/watch`. Isolato con 7 layer:

| Layer | Cosa | Come |
|-------|------|------|
| L1 Prompt | deny.toml + constitution | Già implementato (guardrails pkg) |
| L2 Permissions | `--dangerously-skip-permissions` | Safe perché isolato |
| L3 Filesystem | Solo workspace montato | Volume mount |
| L4 Kernel | deny.toml → seccomp (Docker) / VM isolation (Apple Container) | Da implementare |
| L5 Network | Solo domini approvati | `--network=none` + proxy (Docker) / VM rules (Apple Container) |
| L6 Token | RTK compression 60-90% | Da implementare |
| L7 Audit | Ogni comando loggato | Da implementare |

### Sandbox providers verificati

| Provider | Stato | CLI |
|----------|-------|-----|
| **Apple Container** | Installato v0.10.0, testato | `container run ubuntu:latest uname -a` → funziona |
| **Docker** | Installato v29.3.0 | Già usato da OpenHandsRunner |
| **Ona** | Cloud, issue #2 separata | Non per ora |

Apple Container è il **preferito** — VM-level isolation, nativo Apple Silicon, no seccomp needed.

### Cosa esiste già nel Go binary

```
internal/runner/runner.go      — Runner interface
internal/runner/claude.go      — ClaudeRunner (host, no sandbox)
internal/runner/openhands.go   — OpenHandsRunner (Docker-based)
internal/runner/dryrun.go      — DryRunRunner
internal/guardrails/           — Parse, Enforce, 3 modes, deny.toml
internal/config/config.go      — [runner.claude] section
internal/process/spawn.go      — Spawn with SIGTERM→SIGKILL
internal/vault/sdd.go          — SDDBoundaries (write_dirs, forbidden_dirs, banned_imports)
```

### Cosa va creato

```
internal/sandbox/              — nuovo package
internal/sandbox/provider.go   — SandboxProvider interface
internal/sandbox/apple.go      — Apple Container provider
internal/sandbox/docker.go     — Docker provider (seccomp)
internal/sandbox/seccomp.go    — deny.toml → seccomp.json converter
internal/sandbox/services.go   — service lifecycle da SDD boundaries
internal/sandbox/audit.go      — command audit logger
internal/sandbox/rtk.go        — RTK integration
```

### Config target

```toml
[runner.claude]
sandbox = "apple-container"        # apple-container | docker | none
sandbox_network_allow = [
  "api.anthropic.com",
  "github.com",
  "registry.npmjs.org",
  "proxy.golang.org",
]
sandbox_workspace_mount = "."
sandbox_seccomp_from_deny = true   # Docker only
sandbox_audit_log = true
sandbox_services_from_sdd = true
use_rtk = true
rtk_track_savings = true
```

### Apple Container CLI reference

```bash
# Run a command in a container
container run <image> <command> [args...]

# Run with volume mount
container run --mount type=bind,source=/host/path,target=/container/path <image> <command>

# Run with env vars
container run -e KEY=VALUE <image> <command>

# Run interactive
container run -it <image> /bin/bash

# System management
container system start/stop/status

# Image management
container images list
container images pull <image>
```

### Workflow Forgia

Il progetto usa Forgia su se stesso (dogfooding). Workflow:
1. Crea FD in `.forgia/fd/`
2. `/fd-review` — gate obbligatorio
3. `/fd-sdd` — genera SDD (ultimo deve essere Integration Wiring)
4. Implementa ogni SDD
5. Commit con `feat(FD-NNN): SDD-NNN — description` + `Signed-off-by: Eugenio Uccheddu <deepzima@outlook.com>` + `Co-Authored-By`

### Go conventions

- Go 1.25+, ctx come primo parametro, slog.InfoContext, fmt.Errorf %w
- Compile-time interface checks: `var _ Interface = (*Struct)(nil)`
- Table-driven tests, t.Parallel(), t.TempDir()
- No init(), no package-level state

### Team

- **deepzima** (Eugenio Uccheddu) — lead
- **ferruvich** — assegnato a #17, #18, #58 (guardrails enforcement, compliance, MCP serve)
- **federicoibba** (ibba) — assegnato a #6 (Kiro)

### Prossimo comando

```bash
cd /Users/deepzima/zimafiles/dev/github/zima/forgia
git checkout feat/dual-profile-sandbox
```

Poi crea l'FD, review, genera SDD, implementa.
