package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"testing"

	"github.com/forgia-labs/forgia/internal/guardrails"
	"github.com/forgia-labs/forgia/internal/vault"
)

// --- mock vault ---

type mockVaultForProvider struct {
	dir  string
	fds  []*vault.FD
	sdds map[string][]*vault.SDD // fdID → SDDs

	// Write capture fields for testing write tools.
	createdFDs  []*vault.FD
	updatedFDs  []*vault.FD
	createdSDDs []*vault.SDD
	updatedSDDs []*vault.SDD
}

func (m *mockVaultForProvider) Dir() string { return m.dir }
func (m *mockVaultForProvider) Init(_ context.Context, _ vault.InitOptions) error {
	return nil
}
func (m *mockVaultForProvider) FDs() iter.Seq[*vault.FD] {
	return func(yield func(*vault.FD) bool) {
		for _, fd := range m.fds {
			if !yield(fd) {
				return
			}
		}
	}
}
func (m *mockVaultForProvider) ListFDs(_ context.Context) ([]*vault.FD, error) {
	return m.fds, nil
}
func (m *mockVaultForProvider) GetFD(_ context.Context, id string) (*vault.FD, error) {
	for _, fd := range m.fds {
		if fd.ID == id {
			return fd, nil
		}
	}
	return nil, fmt.Errorf("FD %q not found", id)
}
func (m *mockVaultForProvider) CreateFD(_ context.Context, fd *vault.FD) error {
	m.createdFDs = append(m.createdFDs, fd)
	return nil
}
func (m *mockVaultForProvider) UpdateFD(_ context.Context, fd *vault.FD) error {
	m.updatedFDs = append(m.updatedFDs, fd)
	return nil
}
func (m *mockVaultForProvider) SDDs(fdID string) iter.Seq[*vault.SDD] {
	return func(yield func(*vault.SDD) bool) {
		for _, sdd := range m.sdds[fdID] {
			if !yield(sdd) {
				return
			}
		}
	}
}
func (m *mockVaultForProvider) ListSDDs(_ context.Context, fdID string) ([]*vault.SDD, error) {
	return m.sdds[fdID], nil
}
func (m *mockVaultForProvider) GetSDD(_ context.Context, fdID, sddID string) (*vault.SDD, error) {
	for _, sdd := range m.sdds[fdID] {
		if sdd.ID == sddID {
			return sdd, nil
		}
	}
	return nil, fmt.Errorf("SDD %q not found in FD %q", sddID, fdID)
}
func (m *mockVaultForProvider) CreateSDD(_ context.Context, sdd *vault.SDD) error {
	m.createdSDDs = append(m.createdSDDs, sdd)
	return nil
}
func (m *mockVaultForProvider) UpdateSDD(_ context.Context, sdd *vault.SDD) error {
	m.updatedSDDs = append(m.updatedSDDs, sdd)
	return nil
}
func (m *mockVaultForProvider) GetArchitecture(_ context.Context) (*vault.Architecture, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockVaultForProvider) ListContexts(_ context.Context) ([]*vault.BoundedContext, error) {
	return nil, nil
}
func (m *mockVaultForProvider) GetContext(_ context.Context, _ string) (*vault.BoundedContext, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockVaultForProvider) Constitution(_ context.Context) (string, error) {
	return "# Constitution", nil
}
func (m *mockVaultForProvider) GuardrailsRaw(_ context.Context) ([]byte, error) {
	return []byte("# deny.toml"), nil
}
func (m *mockVaultForProvider) Render(_ context.Context, _ string) error { return nil }

// --- mock code provider for KG enrichment ---

type mockCodeProvider struct {
	callFn func(ctx context.Context, tool string, params map[string]any) (any, error)
}

func (m *mockCodeProvider) Name() string            { return "code" }
func (m *mockCodeProvider) Tools() []ToolDefinition { return nil }
func (m *mockCodeProvider) Call(ctx context.Context, tool string, params map[string]any) (any, error) {
	if m.callFn != nil {
		return m.callFn(ctx, tool, params)
	}
	return nil, fmt.Errorf("not implemented")
}
func (m *mockCodeProvider) Start(_ context.Context) error { return nil }
func (m *mockCodeProvider) Stop() error                   { return nil }
func (m *mockCodeProvider) Healthy() bool                 { return true }

// --- test helpers ---

func newTestVault(t *testing.T) *mockVaultForProvider {
	t.Helper()
	dir := t.TempDir()
	forgiaDir := filepath.Join(dir, ".forgia")

	// Create required subdirectories.
	for _, d := range []string{"fd", "sdd/FD-001", "ops/active", "logs"} {
		if err := os.MkdirAll(filepath.Join(forgiaDir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	return &mockVaultForProvider{
		dir: forgiaDir,
		fds: []*vault.FD{
			{ID: "FD-001", Title: "Feature One", Status: vault.FDApproved, Priority: "high", Author: "alice"},
			{ID: "FD-002", Title: "Feature Two", Status: vault.FDPlanned, Priority: "low", Author: "bob"},
		},
		sdds: map[string][]*vault.SDD{
			"FD-001": {
				{ID: "SDD-001", FD: "FD-001", Title: "Spec One", Status: vault.SDDPlanned},
				{ID: "SDD-002", FD: "FD-001", Title: "Spec Two", Status: vault.SDDDone},
			},
		},
	}
}

// --- tests ---

func TestVaultProvider_InterfaceSatisfaction(t *testing.T) {
	t.Parallel()
	var _ ToolProvider = (*VaultProvider)(nil)
}

func TestVaultProvider_Name(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	if p.Name() != "vault" {
		t.Errorf("Name() = %q, want 'vault'", p.Name())
	}
}

func TestVaultProvider_Tools(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	tools := p.Tools()
	if len(tools) != 11 {
		t.Fatalf("expected 11 tools, got %d", len(tools))
	}

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
		if tool.Namespace != "vault" {
			t.Errorf("tool %q namespace = %q, want 'vault'", tool.Name, tool.Namespace)
		}
		if tool.Parameters == nil {
			t.Errorf("tool %q has nil parameters", tool.Name)
		}
	}

	expected := []string{
		"fd_get", "fd_list", "fd_create", "fd_update",
		"sdd_get", "sdd_list", "sdd_create", "sdd_update",
		"status", "validate", "threat_model_get",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing tool %q", name)
		}
	}
}

func TestVaultProvider_Healthy(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	if !p.Healthy() {
		t.Error("Healthy() = false, want true")
	}

	nilP := NewVaultProvider(nil, nil, nil)
	if nilP.Healthy() {
		t.Error("Healthy() = true with nil vault, want false")
	}
}

func TestVaultProvider_FDGet_Existing(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "fd_get", map[string]any{"id": "FD-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["id"] != "FD-001" {
		t.Errorf("id = %v, want FD-001", r["id"])
	}
	if r["title"] != "Feature One" {
		t.Errorf("title = %v, want Feature One", r["title"])
	}
	if r["status"] != "approved" {
		t.Errorf("status = %v, want approved", r["status"])
	}
	if r["author"] != "alice" {
		t.Errorf("author = %v, want alice", r["author"])
	}
}

func TestVaultProvider_FDGet_WithBody(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)

	// Write an FD file with frontmatter and body.
	fdPath := filepath.Join(mv.dir, "fd", "FD-001.md")
	content := "---\nid: FD-001\ntitle: Feature One\nstatus: approved\n---\n\n# Feature One\n\nThis is the body.\n"
	if err := os.WriteFile(fdPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	mv.fds[0].FilePath = fdPath

	p := NewVaultProvider(mv, nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "fd_get", map[string]any{"id": "FD-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	body, _ := r["body"].(string)
	if body == "" {
		t.Error("body is empty, expected content")
	}
	if !contains(body, "This is the body.") {
		t.Errorf("body = %q, expected to contain 'This is the body.'", body)
	}
}

func TestVaultProvider_FDGet_Missing(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	_, err := p.Call(ctx, "fd_get", map[string]any{"id": "FD-999"})
	if err == nil {
		t.Fatal("expected error for missing FD")
	}
}

func TestVaultProvider_FDGet_MissingID(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	_, err := p.Call(ctx, "fd_get", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing id param")
	}
}

func TestVaultProvider_FDList_All(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "fd_list", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["count"] != 2 {
		t.Errorf("count = %v, want 2", r["count"])
	}
	fds := r["fds"].([]map[string]any)
	if len(fds) != 2 {
		t.Errorf("len(fds) = %d, want 2", len(fds))
	}
}

func TestVaultProvider_FDList_Filtered(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "fd_list", map[string]any{"status": "planned"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["count"] != 1 {
		t.Errorf("count = %v, want 1", r["count"])
	}
	fds := r["fds"].([]map[string]any)
	if fds[0]["id"] != "FD-002" {
		t.Errorf("filtered FD id = %v, want FD-002", fds[0]["id"])
	}
}

func TestVaultProvider_FDList_NoMatch(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "fd_list", map[string]any{"status": "closed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["count"] != 0 {
		t.Errorf("count = %v, want 0", r["count"])
	}
}

func TestVaultProvider_SDDGet_Existing(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "sdd_get", map[string]any{"fd": "FD-001", "id": "SDD-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["id"] != "SDD-001" {
		t.Errorf("id = %v, want SDD-001", r["id"])
	}
	if r["fd"] != "FD-001" {
		t.Errorf("fd = %v, want FD-001", r["fd"])
	}
}

func TestVaultProvider_SDDGet_Missing(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	_, err := p.Call(ctx, "sdd_get", map[string]any{"fd": "FD-001", "id": "SDD-999"})
	if err == nil {
		t.Fatal("expected error for missing SDD")
	}
}

func TestVaultProvider_SDDList(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "sdd_list", map[string]any{"fd": "FD-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["count"] != 2 {
		t.Errorf("count = %v, want 2", r["count"])
	}
}

func TestVaultProvider_SDDList_Filtered(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "sdd_list", map[string]any{"fd": "FD-001", "status": "done"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["count"] != 1 {
		t.Errorf("count = %v, want 1", r["count"])
	}
	sdds := r["sdds"].([]map[string]any)
	if sdds[0]["id"] != "SDD-002" {
		t.Errorf("filtered SDD id = %v, want SDD-002", sdds[0]["id"])
	}
}

func TestVaultProvider_SDDList_EmptyFD(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "sdd_list", map[string]any{"fd": "FD-002"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["count"] != 0 {
		t.Errorf("count = %v, want 0", r["count"])
	}
}

func TestVaultProvider_Status(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "status", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	fds := r["fds"].([]map[string]any)
	if len(fds) != 2 {
		t.Errorf("fds count = %d, want 2", len(fds))
	}

	// First FD should have SDDs.
	fd1SDDs := fds[0]["sdds"].([]map[string]any)
	if len(fd1SDDs) != 2 {
		t.Errorf("FD-001 SDDs count = %d, want 2", len(fd1SDDs))
	}

	// FD-002 has no SDDs — should be empty array.
	fd2SDDs := fds[1]["sdds"].([]map[string]any)
	if len(fd2SDDs) != 0 {
		t.Errorf("FD-002 SDDs count = %d, want 0", len(fd2SDDs))
	}
}

func TestVaultProvider_Status_WithExecLogs(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)

	// Write an exec log.
	logData := map[string]any{
		"sdd": "SDD-001", "fd": "FD-001",
		"status": "done", "started": "2026-03-01T10:00:00Z",
		"duration_seconds": 120, "exit_code": 0,
	}
	logBytes, _ := json.Marshal(logData)
	logPath := filepath.Join(mv.dir, "logs", "exec-001.json")
	if err := os.WriteFile(logPath, logBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewVaultProvider(mv, nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "status", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	logs, ok := r["exec_logs"].([]map[string]any)
	if !ok || len(logs) == 0 {
		t.Error("expected exec_logs in status result")
	}
}

func TestVaultProvider_Status_NoKG(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "status", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if _, ok := r["knowledge"]; ok {
		t.Error("knowledge field should not be present without KG")
	}
}

func TestVaultProvider_Status_WithKG(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)

	reg := NewProviderRegistry()
	reg.Register(&mockCodeProvider{
		callFn: func(_ context.Context, tool string, _ map[string]any) (any, error) {
			if tool == "list_projects" {
				return map[string]any{
					"node_count": 150, "edge_count": 300, "last_indexed": "2026-03-29T10:00:00Z",
				}, nil
			}
			return nil, fmt.Errorf("unknown tool")
		},
	})

	p := NewVaultProvider(mv, nil, reg)
	ctx := context.Background()

	result, err := p.Call(ctx, "status", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	kg, ok := r["knowledge"]
	if !ok {
		t.Fatal("expected knowledge field in status result with KG")
	}
	kgMap := kg.(map[string]any)
	if kgMap["node_count"] != 150 {
		t.Errorf("node_count = %v, want 150", kgMap["node_count"])
	}
}

func TestVaultProvider_Validate_Valid(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)

	// Create a file that the context section references.
	projectRoot := filepath.Dir(mv.dir)
	ctxFilePath := filepath.Join(projectRoot, "internal", "mcp", "board_provider.go")
	if err := os.MkdirAll(filepath.Dir(ctxFilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ctxFilePath, []byte("package mcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write a valid SDD file.
	sddContent := `---
id: SDD-001
fd: FD-001
title: Spec One
status: planned
---

# SDD-001: Spec One

## Scope
Do something.

## Interfaces / Interfacce
| Name | Type |
|------|------|
| Foo  | Bar  |

## Constraints / Vincoli
- Language: Go

## Test Requirements
| Type | What |
|------|------|
| Unit | test |

## Acceptance Criteria / Criteri di Accettazione
- [ ] It works

## Context / Contesto
- [ ] ` + "`internal/mcp/board_provider.go`" + `

## Constitution Check
- [ ] OK

## Work Log / Diario di Lavoro
> Mandatory.
`
	sddPath := filepath.Join(mv.dir, "sdd", "FD-001", "SDD-001.md")
	if err := os.WriteFile(sddPath, []byte(sddContent), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewVaultProvider(mv, nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "validate", map[string]any{"path": sddPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	files := r["files"].([]map[string]any)
	if len(files) != 1 {
		t.Fatalf("expected 1 file result, got %d", len(files))
	}
	if files[0]["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors=%v", files[0]["valid"], files[0]["errors"])
	}
}

func TestVaultProvider_Validate_Invalid(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)

	// Write an invalid SDD file (missing sections).
	sddContent := `---
id: SDD-BAD
fd: FD-001
title: Bad Spec
status: planned
---

# Missing required sections
`
	sddPath := filepath.Join(mv.dir, "sdd", "FD-001", "SDD-BAD.md")
	if err := os.WriteFile(sddPath, []byte(sddContent), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewVaultProvider(mv, nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "validate", map[string]any{"path": sddPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	files := r["files"].([]map[string]any)
	if files[0]["valid"] != false {
		t.Error("expected valid=false for invalid SDD")
	}
	errs, ok := files[0]["errors"].([]map[string]any)
	if !ok || len(errs) == 0 {
		t.Error("expected validation errors for invalid SDD")
	}
}

func TestVaultProvider_Validate_ByFD(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)

	// Write an SDD file under FD-001.
	sddContent := `---
id: SDD-001
fd: FD-001
title: Spec One
status: planned
---

# Minimal — missing sections
`
	sddPath := filepath.Join(mv.dir, "sdd", "FD-001", "SDD-001.md")
	if err := os.WriteFile(sddPath, []byte(sddContent), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewVaultProvider(mv, nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "validate", map[string]any{"fd": "FD-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	files := r["files"].([]map[string]any)
	if len(files) == 0 {
		t.Fatal("expected at least 1 file result")
	}
}

func TestVaultProvider_Validate_MissingParams(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	_, err := p.Call(ctx, "validate", map[string]any{})
	if err == nil {
		t.Fatal("expected error when both path and fd are missing")
	}
}

func TestVaultProvider_Validate_WithKG(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)

	// Write an SDD with a context path that the KG can verify.
	sddContent := `---
id: SDD-001
fd: FD-001
title: Spec One
status: planned
---

## Scope
Do something.

## Interfaces
OK.

## Constraints
Go.

## Test Requirements
Unit tests.

## Acceptance Criteria
- [ ] Works

## Context
- [ ] ` + "`nonexistent/file.go`" + `

## Constitution Check
OK.

## Work Log
Done.
`
	sddPath := filepath.Join(mv.dir, "sdd", "FD-001", "SDD-001.md")
	if err := os.WriteFile(sddPath, []byte(sddContent), 0o644); err != nil {
		t.Fatal(err)
	}

	reg := NewProviderRegistry()
	reg.Register(&mockCodeProvider{
		callFn: func(_ context.Context, tool string, params map[string]any) (any, error) {
			if tool == "search" {
				return map[string]any{"count": float64(0)}, nil
			}
			return nil, fmt.Errorf("unknown tool")
		},
	})

	p := NewVaultProvider(mv, nil, reg)
	ctx := context.Background()

	result, err := p.Call(ctx, "validate", map[string]any{"path": sddPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	files := r["files"].([]map[string]any)
	errs, _ := files[0]["errors"].([]map[string]any)

	// Should have a context path warning from KG.
	found := false
	for _, e := range errs {
		if e["category"] == "context" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected context path warning from KG enrichment")
	}
}

func TestVaultProvider_Validate_WithGuardrails(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)

	// Write an SDD that references a denied path in context.
	sddContent := `---
id: SDD-001
fd: FD-001
title: Spec One
status: planned
---

## Scope
OK.

## Interfaces
OK.

## Constraints
Go.

## Test Requirements
Unit.

## Acceptance Criteria
- [ ] Works

## Context
- [ ] ` + "`.env`" + `

## Constitution Check
OK.

## Work Log
Done.
`
	sddPath := filepath.Join(mv.dir, "sdd", "FD-001", "SDD-001.md")
	if err := os.WriteFile(sddPath, []byte(sddContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Parse guardrails that deny .env files.
	g, err := guardrails.Parse([]byte(`[read]
patterns = ["**/.env"]
`))
	if err != nil {
		t.Fatal(err)
	}

	p := NewVaultProvider(mv, g, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "validate", map[string]any{"path": sddPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	files := r["files"].([]map[string]any)
	errs, _ := files[0]["errors"].([]map[string]any)

	found := false
	for _, e := range errs {
		if e["category"] == "guardrail" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected guardrail violation for .env in context paths")
	}
}

func TestVaultProvider_ThreatModelGet_Exists(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)

	// Write a threat model file.
	tmContent := "# Threat Model for FD-001\n\n## STRIDE\n...\n"
	tmPath := filepath.Join(mv.dir, "fd", "FD-001-threat-model.md")
	if err := os.WriteFile(tmPath, []byte(tmContent), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewVaultProvider(mv, nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "threat_model_get", map[string]any{"fd": "FD-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["fd"] != "FD-001" {
		t.Errorf("fd = %v, want FD-001", r["fd"])
	}
	content, ok := r["content"].(string)
	if !ok || content == "" {
		t.Error("expected content in threat model result")
	}
	if !contains(content, "STRIDE") {
		t.Errorf("content = %q, expected STRIDE", content)
	}
}

func TestVaultProvider_ThreatModelGet_Missing(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	result, err := p.Call(ctx, "threat_model_get", map[string]any{"fd": "FD-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["content"] != nil {
		t.Errorf("content = %v, want nil for missing threat model", r["content"])
	}
}

func TestVaultProvider_ThreatModelGet_MissingFD(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	_, err := p.Call(ctx, "threat_model_get", map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing fd param")
	}
}

func TestVaultProvider_UnknownTool(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	_, err := p.Call(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestVaultProvider_Registry(t *testing.T) {
	t.Parallel()
	reg := NewProviderRegistry()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	reg.Register(p)

	tools := reg.AllTools()
	found := 0
	for _, tool := range tools {
		if tool.Namespace == "" || tool.Namespace == "vault" {
			found++
		}
	}
	if found != 11 {
		t.Errorf("expected 11 vault tools in registry, found %d", found)
	}
}

func TestVaultProvider_NilParams(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)
	ctx := context.Background()

	// Status should work with nil params.
	_, err := p.Call(ctx, "status", nil)
	if err != nil {
		t.Fatalf("status with nil params: %v", err)
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
