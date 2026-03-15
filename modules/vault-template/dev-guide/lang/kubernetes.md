# Kubernetes Conventions

> Auto-loaded when `kustomization.yaml`, `Chart.yaml`, `Tiltfile`, or CRD manifests are detected.

## Project Structure

- `deploy/` or `manifests/` for K8s resources
- `deploy/crd/` for Custom Resource Definitions
- `deploy/operator/` for operator resources (RBAC, Deployment, Service)
- `charts/` for Helm charts
- `tests/` for integration and validation tests
- `local/` for local dev (Tiltfile, kind/minikube config)

## YAML Style

- Indent: 2 spaces, no tabs
- Always quote strings that could be misinterpreted (`"true"`, `"yes"`, `"null"`, `"1.0"`)
- Use `---` to separate multiple resources in a single file
- Order: `apiVersion`, `kind`, `metadata`, `spec` (follow K8s convention)
- Labels: always include `app.kubernetes.io/name`, `app.kubernetes.io/version`, `app.kubernetes.io/managed-by`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: todo-operator
  labels:
    app.kubernetes.io/name: todo-operator
    app.kubernetes.io/version: "0.1.0"
    app.kubernetes.io/managed-by: forgia
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: todo-operator
```

## Linting

- **kubeconform**: validate manifests against K8s OpenAPI schemas — mandatory
- **kube-linter**: static analysis for security and best practices
- **yamllint**: basic YAML syntax and style

```bash
# Validate all manifests
kubeconform -strict -summary deploy/

# Lint for security issues
kube-linter lint deploy/

# YAML style
yamllint -d relaxed deploy/
```

Install:
```bash
brew install kubeconform kube-linter yamllint
```

## Security

- Never use `privileged: true` — use specific capabilities instead
- Always set `runAsNonRoot: true` and `readOnlyRootFilesystem: true`
- Always set resource requests and limits
- Use `securityContext` on both pod and container level
- Never use `latest` tag — pin image versions
- Use distroless or scratch base images for production

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65534
  fsGroup: 65534
containers:
  - name: operator
    image: todo-operator:0.1.0  # pinned, never :latest
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop: ["ALL"]
    resources:
      requests:
        memory: "64Mi"
        cpu: "50m"
      limits:
        memory: "256Mi"
        cpu: "200m"
```

## RBAC

- Least privilege: only the permissions the operator actually needs
- Separate ClusterRole for CRD-level access, Role for namespace-level
- Always specify `resourceNames` when possible
- Use `verbs` explicitly (never `["*"]`)

```yaml
rules:
  - apiGroups: ["todo.grafana.app"]
    resources: ["todos"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["todo.grafana.app"]
    resources: ["todos/status"]
    verbs: ["get", "update", "patch"]
```

## Testing

### BATS (Bash Automated Testing System)

Use BATS for integration tests that validate K8s resources:

```bash
#!/usr/bin/env bats
# tests/integration/crd.bats

setup() {
  kubectl apply -f deploy/crd/todo-crd.yaml
}

teardown() {
  kubectl delete -f deploy/crd/todo-crd.yaml --ignore-not-found
}

@test "CRD is registered" {
  run kubectl get crd todos.todo.grafana.app
  [ "$status" -eq 0 ]
}

@test "can create a Todo resource" {
  run kubectl apply -f - <<EOF
apiVersion: todo.grafana.app/v1
kind: Todo
metadata:
  name: test-todo
spec:
  title: "Test"
  status: "open"
EOF
  [ "$status" -eq 0 ]
}

@test "rejects invalid status" {
  run kubectl apply -f - <<EOF
apiVersion: todo.grafana.app/v1
kind: Todo
metadata:
  name: bad-todo
spec:
  title: "Test"
  status: "invalid"
EOF
  [ "$status" -ne 0 ]
}
```

Install: `brew install bats-core`

### Validation scripts

```bash
#!/usr/bin/env bash
# tests/deployment/validate.sh

set -euo pipefail

errors=0

# Check all YAML files are valid
for f in deploy/**/*.yaml; do
  if ! kubeconform -strict "$f" 2>/dev/null; then
    echo "FAIL: $f"
    errors=$((errors + 1))
  fi
done

# Check security context
for f in deploy/**/*.yaml; do
  if grep -q 'kind: Deployment' "$f"; then
    if ! grep -q 'runAsNonRoot' "$f"; then
      echo "WARN: $f missing runAsNonRoot"
    fi
    if ! grep -q 'readOnlyRootFilesystem' "$f"; then
      echo "WARN: $f missing readOnlyRootFilesystem"
    fi
    if grep -q ':latest' "$f"; then
      echo "FAIL: $f uses :latest tag"
      errors=$((errors + 1))
    fi
  fi
done

exit $errors
```

## CRD Development

- Define schema with OpenAPI v3 validation (`openAPIV3Schema`)
- Always mark required fields explicitly
- Use `enum` for constrained string values
- Use `default` for fields with sensible defaults
- Add `description` to every field (improves `kubectl explain`)
- Version CRDs: `v1alpha1` → `v1beta1` → `v1`

## Helm (if applicable)

- `values.yaml`: document every value with comments
- Templates: use `include` for reusable snippets, not copy-paste
- Always support `nameOverride` and `fullnameOverride`
- Use `helm lint` before every commit
- Test with `helm template` (no cluster needed)

## Kustomize (if applicable)

- Base in `deploy/base/`, overlays in `deploy/overlays/{dev,staging,prod}/`
- Use `namePrefix` and `nameSuffix` for environment separation
- Use `configMapGenerator` and `secretGenerator` instead of raw resources
- Validate with `kustomize build deploy/overlays/dev/ | kubeconform -strict`

## Local Development

- Use `kind` or `minikube` for local clusters
- Use `Tiltfile` for live reload during development
- Use `ctlptl` for declarative cluster management
- Mount local code into the cluster for fast iteration

## What NOT to Do

- Don't use `kubectl apply` with `--force` in scripts
- Don't hardcode namespaces in manifests (use kustomize or helm values)
- Don't store secrets in manifests (use external-secrets, sealed-secrets, or SOPS)
- Don't use `hostPath` volumes in production
- Don't skip resource limits (causes noisy neighbor problems)
- Don't use `cluster-admin` ClusterRoleBinding for operators
