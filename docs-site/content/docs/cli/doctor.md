---
title: "forgia doctor"
description: "Diagnose missing dependencies and configuration issues."
navigation:
  title: "doctor"
---

# forgia doctor

```
forgia doctor
```

Checks that all Forgia dependencies are installed and configured correctly.

## What gets checked

```
✓ mise          1.8.3
✓ go            1.22.0
✓ gh            2.48.0   (authenticated as federicoibba)
✓ docker        26.1.1
✓ openhands     0.9.0    (image: ghcr.io/all-hands-ai/openhands:main)
✗ glab                   not found — optional, needed for GitLab issues
```

## Checks

| Dependency | Required | Purpose |
|------------|----------|---------|
| `mise` | Yes | Task runner and tool version manager |
| `go` | Yes | Builds the `forgia` CLI |
| `gh` | Recommended | Fetches GitHub issues for `/fd-new #N` |
| `docker` | Recommended | Runs OpenHands for autonomous SDD execution |
| `openhands` image | Recommended | Agent runtime for `forgia exec` |
| `glab` | Optional | Fetches GitLab issues |

## Common fixes

**mise not found**

```bash
curl https://mise.run | sh
```

**gh not authenticated**

```bash
gh auth login
```

**Docker not running**

Start Docker Desktop, or on Linux:

```bash
sudo systemctl start docker
```

**OpenHands image not pulled**

```bash
mise run openhands:up
```

Running `forgia doctor` after fixing issues confirms everything is in order before you start work.
