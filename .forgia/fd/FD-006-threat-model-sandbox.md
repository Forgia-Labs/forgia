---
id: "FD-006-TM"
title: "Threat Model — Dual-profile sandbox runner"
parent_fd: "FD-006"
created: "2026-04-01"
author: "DeepZima"
methodology: STRIDE
---

# Threat Model: Dual-profile sandbox (FD-006)

## Components under analysis

| # | Component | Trust boundary |
|---|-----------|---------------|
| C1 | `sandbox_exec.go` — orchestrates sandbox lifecycle | Host → Container |
| C2 | `apple.go` — Apple Container provider | Host → VM |
| C3 | `docker.go` — Docker provider | Host → Container |
| C4 | `seccomp.go` — deny.toml → seccomp.json | Host-side, trusted |
| C5 | `services.go` — SDD service lifecycle | Host → Service containers |
| C6 | `audit.go` — command logging | Container → Host filesystem |
| C7 | `config.go` — sandbox_env, sandbox_mount_claude | Config file, user-controlled |

## STRIDE Analysis

### S — Spoofing

| # | Threat | Component | Risk | Mitigation |
|---|--------|-----------|------|-----------|
| S1 | Malicious OCI image in `sandbox_image` config | C1, C7 | **HIGH** — attacker controls config.toml, sets image to a compromised image that exfiltrates workspace | Validate image against allowlist, or warn on non-default images. Currently no validation — any image is accepted. |
| S2 | Agent impersonates host auth via mounted `~/.claude` | C1 | **MEDIUM** — `sandbox_mount_claude = true` gives the agent read access to OAuth tokens. Agent could extract and exfiltrate the token if network is not `none`. | Mount is read-only + opt-in. But if NetworkAllow is ever implemented with non-empty list, token could be exfiltrated via allowed domains. **Recommendation: when sandbox_mount_claude=true, force NetworkMode=none.** |

### T — Tampering

| # | Threat | Component | Risk | Mitigation |
|---|--------|-----------|------|-----------|
| T1 | Agent modifies workspace files outside SDD boundaries | C1, C2, C3 | **MEDIUM** — workspace is mounted read-write. Agent can write anywhere in the project. SDD boundaries are checked pre-flight but not enforced at filesystem level. | Would require read-only mount + specific write dirs mounted separately. Not implemented — current mitigation is guardrails pre-flight check only. |
| T2 | Agent tampers with audit log | C6 | **LOW** — audit log is written by the host-side `sandbox_exec.go`, not by the agent inside the container. Agent cannot modify audit entries. | Correct architecture — audit writes happen outside sandbox. |
| T3 | Seccomp profile tampering | C4 | **LOW** — seccomp.json is generated host-side from deny.toml, written to temp file, passed to Docker. Agent never sees the profile. | Temp file deleted after use (`defer os.Remove`). |

### R — Repudiation

| # | Threat | Component | Risk | Mitigation |
|---|--------|-----------|------|-----------|
| R1 | Agent claims it didn't execute a destructive command | C6 | **LOW** — audit logger captures command, exit code, duration, output size. But it only logs what `sandbox_exec.go` sees (the top-level claude invocation), not individual tool calls inside Claude's agent loop. | **Gap: audit captures "claude -p task" as one entry, not the 50 tool calls inside.** For full repudiation protection, need PostToolUse hooks logging inside the container. |

### I — Information Disclosure

| # | Threat | Component | Risk | Mitigation |
|---|--------|-----------|------|-----------|
| I1 | `ANTHROPIC_API_KEY` leaked via env inspection inside container | C1 | **MEDIUM** — API key passed as env var. Any process inside container can read `/proc/1/environ`. | Standard practice for container auth. Risk accepted — key is needed for execution. Alternative: mount key as file with restrictive permissions. |
| I2 | `sandbox_env` contains secrets in plaintext config | C7 | **HIGH** — user puts `ANTHROPIC_API_KEY=sk-ant-...` in config.toml which is committed to git. | **Recommendation: warn if sandbox_env values match secret patterns (reuse guardrails.ScanForSecrets on config values).** |
| I3 | `ExcludedHostPaths` not exhaustive | C1 | **MEDIUM** — list covers ~/.ssh, ~/.gnupg, ~/.aws etc. but misses: ~/.claude (unless mount_claude=true), ~/.config/gh/hosts.yml (GitHub CLI token), ~/.kube/config (K8s credentials). | **Recommendation: add ~/.config/gh, ~/.kube, ~/.docker/config.json to ExcludedHostPaths.** |
| I4 | Workspace mount exposes `.env` files | C1 | **MEDIUM** — project `.env` files with secrets are mounted into sandbox. Agent can read them. | deny.toml already has `**/.env` in read deny list. But this is prompt-level, not filesystem-level. **Recommendation: add `.env*` to ExcludedHostPaths or mount-exclude.** |

### D — Denial of Service

| # | Threat | Component | Risk | Mitigation |
|---|--------|-----------|------|-----------|
| D1 | Agent fills disk inside container | C2, C3 | **LOW** — workspace mount is shared with host. Agent could fill host disk via workspace writes. | Docker: use `--storage-opt size=10G`. Apple Container: no equivalent currently. **Recommendation: add disk quota config.** |
| D2 | Service manager starts many containers | C5 | **LOW** — SDD declares services, ServiceManager starts them all. A malicious SDD could declare 100 services. | **Recommendation: max_services limit in config (default 5).** |

### E — Elevation of Privilege

| # | Threat | Component | Risk | Mitigation |
|---|--------|-----------|------|-----------|
| E1 | Docker container escape | C3 | **MEDIUM** — namespace isolation is weaker than VM. Known Docker escape CVEs. Seccomp mitigates but doesn't eliminate. | Apple Container (VM) eliminates this. Docker seccomp blocks ptrace, module loading. **Recommendation: prefer Apple Container when available.** |
| E2 | Agent runs `--dangerously-skip-permissions` | C1 | **BY DESIGN** — the whole point of sandbox is that `--dangerously-skip-permissions` is safe because isolation is at the container/VM level, not at the permission level. | Correct architecture. This is not a vulnerability — it's the design. |

## Risk Summary

| Risk | Count | Highest |
|------|-------|---------|
| HIGH | 2 | S1 (malicious image), I2 (secrets in config) |
| MEDIUM | 5 | S2, T1, I1, I3, E1 |
| LOW | 4 | T2, T3, R1, D1, D2 |

## Recommendations for SDD constraints

| # | Recommendation | Applies to | Priority |
|---|---------------|-----------|----------|
| TM-1 | Validate `sandbox_image` against allowlist or warn on non-default | sandbox_exec.go | HIGH |
| TM-2 | When `sandbox_mount_claude=true`, force `NetworkMode=none` | sandbox_exec.go | HIGH |
| TM-3 | Scan `sandbox_env` values against secret patterns before writing config | config.go or exec.go | HIGH |
| TM-4 | Add `~/.config/gh`, `~/.kube` to ExcludedHostPaths | seccomp.go | MEDIUM |
| TM-5 | Mount-exclude `.env*` files or warn | sandbox_exec.go | MEDIUM |
| TM-6 | Add disk quota for sandbox (`--storage-opt size=10G`) | docker.go | LOW |
| TM-7 | Add `max_services` limit in config (default 5) | services.go | LOW |
| TM-8 | Audit: log PostToolUse hooks inside container for full command trace | Future — needs Claude Code hooks | LOW |
| TM-9 | Prefer Apple Container in Resolve() fallback messaging | provider.go | LOW |
