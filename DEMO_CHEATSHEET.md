# Forgia Demo Cheatsheet

## Demo 1 — Greenfield (Go e-commerce)

```bash
# Setup
mkdir /tmp/demo-ecommerce && cd /tmp/demo-ecommerce
git init && echo "module demo" > go.mod && echo "package main" > main.go
git add . && git commit -m "init"

# Scaffold vault
forgia init
forgia doctor
forgia status
forgia skills
```

```bash
# Architecture
forgia skill arch-init "E-commerce platform: user auth (OAuth2), product catalog, order management, payment integration"
cat .forgia/architecture/system-context.yaml
ls .forgia/contexts/
forgia skill arch-review
```

```bash
# Feature Design
forgia skill fd-new "User authentication with OAuth2 (Google, GitHub) and JWT session management"
cat .forgia/fd/FD-*.md
forgia skill fd-arch-review FD-001
forgia skill fd-threat-model FD-001
forgia skill fd-review FD-001
```

```bash
# SDD Generation
forgia skill fd-sdd FD-001
ls .forgia/sdd/FD-001/
forgia status
forgia validate .forgia/sdd/FD-001/SDD-001-*.md
```

```bash
# Execution
forgia exec .forgia/sdd/FD-001/SDD-001-*.md
git diff --stat
go build ./...
go test ./...

# Remaining SDDs
forgia batch FD-001

# Verify + close
forgia skill fd-verify FD-001
forgia status
```

```bash
# Competitive FD (bonus)
forgia skill fd-new "Auth with Passkeys/WebAuthn instead of OAuth2"
forgia skill fd-compare FD-001 FD-002
```

---

## Demo 2 — Brownfield: furyctl (Go CLI)

```bash
# Setup
git clone git@github.com:sighupio/furyctl.git /tmp/demo-furyctl
cd /tmp/demo-furyctl
forgia init
forgia doctor
forgia status
```

```bash
# Architecture — analyze existing codebase
forgia skill arch-init "Kubernetes distribution lifecycle manager. Cobra CLI with plugin system for cluster create/delete/upgrade. Supports EKS, GKE, on-prem via provider plugins."
cat .forgia/architecture/system-context.yaml
ls .forgia/contexts/
```

```bash
# Feature Design — real feature on existing code
forgia skill fd-new "Add structured logging with slog — replace all log.Printf with slog.InfoContext across the codebase for observability and correlation IDs"
forgia skill fd-arch-review FD-001
forgia skill fd-review FD-001
```

```bash
# SDD Generation
forgia skill fd-sdd FD-001
ls .forgia/sdd/FD-001/
forgia status
```

```bash
# Execution — agent modifies real furyctl code
forgia exec .forgia/sdd/FD-001/SDD-001-*.md
git diff --stat
go build ./...
go test ./...

# Batch remaining + integration wiring
forgia batch FD-001

# Verify
forgia skill fd-verify FD-001
forgia status
git log --oneline -10
```

---

## Demo 3 — Brownfield: module-logging (K8s/Shell)

```bash
# Setup
git clone git@github.com:sighupio/module-logging.git /tmp/demo-logging
cd /tmp/demo-logging
forgia init
forgia doctor
forgia status
```

```bash
# Architecture — Kubernetes module
forgia skill arch-init "Kubernetes logging module: Loki + Promtail + Grafana stack deployed via Kustomize. Collects pod logs, ships to Loki, dashboards in Grafana. Part of SIGHUP Fury distribution."
cat .forgia/architecture/system-context.yaml
ls .forgia/contexts/
```

```bash
# Feature Design
forgia skill fd-new "Add log retention policies — configure Loki compactor with configurable retention period (default 30d), add Kustomize overlay for retention settings, Grafana dashboard for storage usage"
forgia skill fd-arch-review FD-001
forgia skill fd-threat-model FD-001
forgia skill fd-review FD-001
```

```bash
# SDD Generation
forgia skill fd-sdd FD-001
ls .forgia/sdd/FD-001/
forgia status
```

```bash
# Execution — agent modifies K8s manifests
forgia exec .forgia/sdd/FD-001/SDD-001-*.md
git diff --stat

# Show changes: Kustomize patches, Loki config, Grafana dashboard JSON
cat katalog/loki/configs/loki.yaml  # or wherever config lives

# Batch remaining
forgia batch FD-001

# Verify
forgia skill fd-verify FD-001
forgia status
```

### FD-002: Version update (Loki/Promtail/Grafana) + Kind test

```bash
# New FD — classic version bump with E2E verification on Kind
forgia skill fd-new "Update Loki from 2.9 to 3.x, Promtail to Alloy migration, Grafana to 11.x — update Kustomize manifests, image tags, config breaking changes. Include Kind cluster E2E test that deploys the stack, sends test logs, verifies query returns results in Grafana."
forgia skill fd-review FD-002
```

```bash
# SDD Generation
forgia skill fd-sdd FD-002
ls .forgia/sdd/FD-002/
# Expect: SDD-001 (Loki 3.x config), SDD-002 (Alloy migration),
#         SDD-003 (Grafana 11.x), SDD-004 (Kind E2E test),
#         SDD-005 (Integration Wiring)
forgia status
```

```bash
# Execute — agent updates manifests + writes E2E test
forgia batch FD-002
git diff --stat

# Show key changes
git diff -- katalog/loki/          # new Loki 3.x config
git diff -- katalog/promtail/      # Alloy migration
git diff -- katalog/grafana/       # Grafana 11.x

# Run E2E on Kind
kind create cluster --name demo-logging
kubectl apply -k katalog/
# Wait for pods
kubectl wait --for=condition=ready pod -l app=loki -n logging --timeout=120s
kubectl wait --for=condition=ready pod -l app=grafana -n logging --timeout=120s

# Send test logs
kubectl run loggen --image=busybox --restart=Never -- sh -c 'for i in $(seq 1 10); do echo "demo log line $i"; sleep 1; done'

# Verify logs in Loki via Grafana API
kubectl port-forward svc/grafana 3000:3000 -n logging &
curl -s "http://localhost:3000/api/ds/query" \
  -H "Content-Type: application/json" \
  -d '{"queries":[{"expr":"{pod=\"loggen\"}","refId":"A"}]}' | jq '.results.A.frames[0].data.values | length'
# → should return > 0

# Cleanup
kind delete cluster --name demo-logging

# Verify + close
forgia skill fd-verify FD-002
forgia status
```

---

## Quick reference

| Step | Command | Gate? |
|------|---------|-------|
| Scaffold | `forgia init` | |
| Health | `forgia doctor` | |
| Dashboard | `forgia status` | |
| Architecture | `forgia skill arch-init "brief"` | |
| Arch review | `forgia skill arch-review` | |
| New FD | `forgia skill fd-new "desc"` | |
| Arch review FD | `forgia skill fd-arch-review FD-N` | optional |
| Threat model | `forgia skill fd-threat-model FD-N` | optional |
| FD review | `forgia skill fd-review FD-N` | **GATE 1** |
| Generate SDD | `forgia skill fd-sdd FD-N` | |
| Validate | `forgia validate path/to/SDD.md` | |
| Execute | `forgia exec path/to/SDD.md` | |
| Batch | `forgia batch FD-N` | |
| Dry run | `forgia exec path --dry-run` | |
| Guard mode | `forgia exec path --mode=guard` | |
| Verify | `forgia skill fd-verify FD-N` | **GATE 2** |
| Close | `forgia skill fd-close FD-N` | **GATE 3** |
| Compare | `forgia skill fd-compare FD-A FD-B` | |
| MCP server | `forgia mcp serve` | |
| Board sync | `forgia sync` | |
| Skills list | `forgia skills` | |
