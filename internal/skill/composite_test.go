package skill

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Deepzima/forgia/internal/guardrails"
	"github.com/Deepzima/forgia/internal/mcp"
	"github.com/Deepzima/forgia/internal/vault"
	"gopkg.in/yaml.v3"
)

// Compile-time check: *vault.FileVault satisfies VaultReader.
var _ VaultReader = (*vault.FileVault)(nil)

// mockVaultReader implements VaultReader for testing.
type mockVaultReader struct {
	constitution string
	guardrails   []byte
	arch         *vault.Architecture
	contexts     []*vault.BoundedContext
}

func (m *mockVaultReader) Constitution(_ context.Context) (string, error) {
	return m.constitution, nil
}

func (m *mockVaultReader) GuardrailsRaw(_ context.Context) ([]byte, error) {
	return m.guardrails, nil
}

func (m *mockVaultReader) GetArchitecture(_ context.Context) (*vault.Architecture, error) {
	return m.arch, nil
}

func (m *mockVaultReader) ListContexts(_ context.Context) ([]*vault.BoundedContext, error) {
	return m.contexts, nil
}

// mockVaultReaderWithDir extends mockVaultReader with Dir() for arch_init file writing.
type mockVaultReaderWithDir struct {
	mockVaultReader
	dir string
}

func (m *mockVaultReaderWithDir) Dir() string {
	return m.dir
}

// mockProvider implements mcp.ToolProvider for testing.
type mockProvider struct {
	name    string
	tools   []mcp.ToolDefinition
	healthy bool
	callFn  func(ctx context.Context, tool string, params map[string]any) (any, error)
}

func (m *mockProvider) Name() string                { return m.name }
func (m *mockProvider) Tools() []mcp.ToolDefinition { return m.tools }
func (m *mockProvider) Healthy() bool               { return m.healthy }
func (m *mockProvider) Start(_ context.Context) error { return nil }
func (m *mockProvider) Stop() error                   { return nil }

func (m *mockProvider) Call(ctx context.Context, tool string, params map[string]any) (any, error) {
	if m.callFn != nil {
		return m.callFn(ctx, tool, params)
	}
	return "ok", nil
}

// --- Registration tests ---

func TestRegisterCompositeSkills_Registers7Skills(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	providerReg := mcp.NewProviderRegistry()
	v := &mockVaultReader{constitution: "test"}

	err := RegisterCompositeSkills(reg, providerReg, v)
	if err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	mcpTools := reg.MCPTools()
	if len(mcpTools) != 7 {
		t.Errorf("expected 7 MCP tools, got %d", len(mcpTools))
		for _, s := range mcpTools {
			t.Logf("  registered: %s", s.Name())
		}
	}

	expected := []string{
		"arch_init",
		"blast_radius",
		"trace_calls",
		"search_code",
		"impact-analysis",
		"arch-validate",
		"design-assist",
	}
	for _, name := range expected {
		if _, ok := reg.Get(name); !ok {
			t.Errorf("expected skill %q to be registered", name)
		}
	}
}

func TestRegisterCompositeSkills_NilProviderReg(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	v := &mockVaultReader{}

	err := RegisterCompositeSkills(reg, nil, v)
	if err == nil {
		t.Fatal("expected error for nil providerReg")
	}
}

func TestRegisterCompositeSkills_NilVault(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	providerReg := mcp.NewProviderRegistry()

	err := RegisterCompositeSkills(reg, providerReg, nil)
	if err != nil {
		t.Fatalf("nil vault should be allowed (some skills may not need it): %v", err)
	}

	if len(reg.MCPTools()) != 7 {
		t.Errorf("expected 7 skills even with nil vault, got %d", len(reg.MCPTools()))
	}
}

func TestCompositeSkill_SkillCategories(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	providerReg := mcp.NewProviderRegistry()

	if err := RegisterCompositeSkills(reg, providerReg, &mockVaultReader{}); err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	archSkills := reg.ByCategory(CategoryArchitecture)
	if len(archSkills) != 2 {
		t.Errorf("expected 2 architecture skills (arch_init, arch-validate), got %d", len(archSkills))
	}

	designSkills := reg.ByCategory(CategoryDesign)
	if len(designSkills) != 3 {
		t.Errorf("expected 3 design skills (blast_radius, impact-analysis, design-assist), got %d", len(designSkills))
	}

	knowledgeSkills := reg.ByCategory(CategoryKnowledge)
	if len(knowledgeSkills) != 2 {
		t.Errorf("expected 2 knowledge skills (trace_calls, search_code), got %d", len(knowledgeSkills))
	}
}

// --- Primitive skill routing tests ---

func TestPrimitiveSkills_CorrectProviderToolNames(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	providerReg := mcp.NewProviderRegistry()

	if err := RegisterCompositeSkills(reg, providerReg, &mockVaultReader{}); err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	tests := []struct {
		skillName    string
		providerTool string
	}{
		{"arch_init", "get_architecture"},
		{"blast_radius", "detect_changes"},
		{"trace_calls", "trace_call_path"},
		{"search_code", "search_graph"},
	}

	for _, tt := range tests {
		s, ok := reg.Get(tt.skillName)
		if !ok {
			t.Errorf("skill %q not registered", tt.skillName)
			continue
		}
		cs, ok := s.(*CompositeSkill)
		if !ok {
			t.Errorf("skill %q is not a *CompositeSkill", tt.skillName)
			continue
		}
		if cs.ProviderTool != tt.providerTool {
			t.Errorf("skill %q: expected ProviderTool %q, got %q", tt.skillName, tt.providerTool, cs.ProviderTool)
		}
		if cs.ProviderName != "code" {
			t.Errorf("skill %q: expected ProviderName %q, got %q", tt.skillName, "code", cs.ProviderName)
		}
	}
}

// --- search_code PostProcess tests ---

func TestSearchCodePostProcess_StripsDenyPaths(t *testing.T) {
	t.Parallel()

	denyToml := `[read]
patterns = [
  "**/.env",
  "**/.env.*",
  "!**/.env.example",
  "**/*.pem",
  "**/*.key",
]

[write]
patterns = []
`

	providerReg := mcp.NewProviderRegistry()
	providerReg.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools:   []mcp.ToolDefinition{{Name: "search_graph", Namespace: "code"}},
		callFn: func(_ context.Context, _ string, _ map[string]any) (any, error) {
			return map[string]any{
				"results": []any{
					map[string]any{"path": "internal/skill/skill.go", "name": "Skill"},
					map[string]any{"path": "config/.env", "name": "SECRET_KEY"},
					map[string]any{"path": "certs/server.pem", "name": "certificate"},
					map[string]any{"path": "internal/mcp/mcp.go", "name": "ToolProvider"},
					map[string]any{"path": ".env.example", "name": "EXAMPLE"},
				},
			}, nil
		},
	})

	v := &mockVaultReader{guardrails: []byte(denyToml)}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	s, _ := reg.Get("search_code")
	result, err := s.Execute(context.Background(), map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	results, ok := resultMap["results"].([]any)
	if !ok {
		t.Fatalf("expected results array, got %T", resultMap["results"])
	}

	// Should keep: skill.go, mcp.go, .env.example (allow-listed).
	// Should strip: .env, .pem.
	if len(results) != 3 {
		t.Errorf("expected 3 results after filtering, got %d", len(results))
		for _, r := range results {
			if rm, ok := r.(map[string]any); ok {
				t.Logf("  kept: %v", rm["path"])
			}
		}
	}

	// Verify specific paths are kept.
	kept := make(map[string]bool)
	for _, r := range results {
		if rm, ok := r.(map[string]any); ok {
			if p, ok := rm["path"].(string); ok {
				kept[p] = true
			}
		}
	}
	for _, want := range []string{"internal/skill/skill.go", "internal/mcp/mcp.go", ".env.example"} {
		if !kept[want] {
			t.Errorf("expected %q to be kept after filtering", want)
		}
	}
}

func TestSearchCodePostProcess_StripsDotEnvAndPem(t *testing.T) {
	t.Parallel()

	// Use realistic deny.toml patterns matching project guardrails.
	denyToml := `[read]
patterns = [
  "**/.env",
  "**/.env.*",
  "!**/.env.example",
  "!**/.env.template",
  "**/*.pem",
  "**/*.key",
  "**/*.p12",
  "**/*.pfx",
  "**/credentials.json",
]

[write]
patterns = []
`

	providerReg := mcp.NewProviderRegistry()
	providerReg.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools:   []mcp.ToolDefinition{{Name: "search_graph", Namespace: "code"}},
		callFn: func(_ context.Context, _ string, _ map[string]any) (any, error) {
			return map[string]any{
				"results": []any{
					map[string]any{"path": "deploy/.env", "name": "env_file"},
					map[string]any{"path": "deploy/.env.production", "name": "env_prod"},
					map[string]any{"path": "tls/cert.pem", "name": "tls_cert"},
					map[string]any{"path": "tls/private.key", "name": "tls_key"},
					map[string]any{"path": "auth/credentials.json", "name": "creds"},
					map[string]any{"path": "internal/config/config.go", "name": "safe_file"},
					map[string]any{"path": "templates/.env.example", "name": "env_example"},
				},
			}, nil
		},
	})

	v := &mockVaultReader{guardrails: []byte(denyToml)}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	s, _ := reg.Get("search_code")
	result, err := s.Execute(context.Background(), map[string]any{"query": "credentials"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]any)
	results := resultMap["results"].([]any)

	// Should keep: config.go, .env.example.
	// Should strip: .env, .env.production, cert.pem, private.key, credentials.json.
	if len(results) != 2 {
		t.Errorf("expected 2 results after filtering, got %d", len(results))
		for _, r := range results {
			if rm, ok := r.(map[string]any); ok {
				t.Logf("  kept: %v", rm["path"])
			}
		}
	}
}

func TestSearchCodePostProcess_FailOpenOnBadGuardrails(t *testing.T) {
	t.Parallel()

	providerReg := mcp.NewProviderRegistry()
	providerReg.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools:   []mcp.ToolDefinition{{Name: "search_graph", Namespace: "code"}},
		callFn: func(_ context.Context, _ string, _ map[string]any) (any, error) {
			return map[string]any{
				"results": []any{
					map[string]any{"path": "secret/.env"},
				},
			}, nil
		},
	})

	// Invalid TOML → guardrails parse fails → fail-open, return unfiltered.
	v := &mockVaultReader{guardrails: []byte("!!! invalid toml !!!")}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	s, _ := reg.Get("search_code")
	result, err := s.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute should not fail on bad guardrails: %v", err)
	}

	resultMap := result.(map[string]any)
	results := resultMap["results"].([]any)
	if len(results) != 1 {
		t.Errorf("expected 1 unfiltered result on guardrails parse failure, got %d", len(results))
	}
}

// --- arch_init PostProcess tests ---

func TestArchInitPostProcess_MergesNotOverwrites(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Write existing architecture file.
	archDir := filepath.Join(tmpDir, "architecture")
	if err := os.MkdirAll(archDir, 0o755); err != nil {
		t.Fatalf("create arch dir: %v", err)
	}
	existingArch := &vault.Architecture{
		SystemContext: &vault.SystemContext{
			Name:        "Forgia",
			Description: "Existing description",
		},
		Containers: &vault.Containers{
			Items: []vault.Container{
				{Name: "CLI", Technology: "Go", Purpose: "Command line interface"},
			},
		},
		TechnologyDecisions: &vault.TechnologyDecisions{
			Decisions: []vault.TechDecision{
				{Area: "Language", Choice: "Go", Rationale: "Team expertise"},
			},
		},
	}
	existingData, _ := yaml.Marshal(existingArch)
	if err := os.WriteFile(filepath.Join(archDir, "system.yaml"), existingData, 0o644); err != nil {
		t.Fatalf("write existing arch: %v", err)
	}

	providerReg := mcp.NewProviderRegistry()
	providerReg.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools:   []mcp.ToolDefinition{{Name: "get_architecture", Namespace: "code"}},
		callFn: func(_ context.Context, _ string, _ map[string]any) (any, error) {
			// Provider returns new data with an updated container and a new decision.
			return map[string]any{
				"system_context": map[string]any{
					"name": "Forgia",
				},
				"containers": []any{
					map[string]any{"name": "CLI", "technology": "Go", "purpose": "Updated CLI purpose"},
					map[string]any{"name": "MCP Server", "technology": "Go", "purpose": "MCP protocol server"},
				},
				"technology_decisions": []any{
					map[string]any{"area": "Protocol", "choice": "MCP", "rationale": "Agent interop"},
				},
			}, nil
		},
	})

	v := &mockVaultReaderWithDir{
		mockVaultReader: mockVaultReader{
			arch: existingArch,
		},
		dir: tmpDir,
	}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	s, _ := reg.Get("arch_init")
	result, err := s.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	arch, ok := result.(*vault.Architecture)
	if !ok {
		t.Fatalf("expected *vault.Architecture, got %T", result)
	}

	// Verify merge: existing description preserved (incoming had empty description).
	if arch.SystemContext.Description != "Existing description" {
		t.Errorf("expected existing description preserved, got %q", arch.SystemContext.Description)
	}

	// Verify merge: CLI container updated, MCP Server added.
	if len(arch.Containers.Items) != 2 {
		t.Errorf("expected 2 containers after merge, got %d", len(arch.Containers.Items))
	}
	for _, c := range arch.Containers.Items {
		if c.Name == "CLI" && c.Purpose != "Updated CLI purpose" {
			t.Errorf("expected CLI purpose updated, got %q", c.Purpose)
		}
	}

	// Verify merge: existing decision preserved, new decision added.
	if len(arch.TechnologyDecisions.Decisions) != 2 {
		t.Errorf("expected 2 decisions after merge, got %d", len(arch.TechnologyDecisions.Decisions))
	}

	// Verify file was written (not just returned).
	writtenData, err := os.ReadFile(filepath.Join(archDir, "system.yaml"))
	if err != nil {
		t.Fatalf("read written architecture: %v", err)
	}
	var writtenArch vault.Architecture
	if err := yaml.Unmarshal(writtenData, &writtenArch); err != nil {
		t.Fatalf("parse written architecture: %v", err)
	}
	if len(writtenArch.Containers.Items) != 2 {
		t.Errorf("written file should have 2 containers, got %d", len(writtenArch.Containers.Items))
	}
}

func TestArchInitPostProcess_CreatesNewWhenNoExisting(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	providerReg := mcp.NewProviderRegistry()
	providerReg.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools:   []mcp.ToolDefinition{{Name: "get_architecture", Namespace: "code"}},
		callFn: func(_ context.Context, _ string, _ map[string]any) (any, error) {
			return map[string]any{
				"system_context": map[string]any{
					"name":        "NewProject",
					"description": "A new project",
				},
			}, nil
		},
	})

	v := &mockVaultReaderWithDir{
		mockVaultReader: mockVaultReader{arch: nil},
		dir:             tmpDir,
	}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	s, _ := reg.Get("arch_init")
	result, err := s.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	arch := result.(*vault.Architecture)
	if arch.SystemContext.Name != "NewProject" {
		t.Errorf("expected name %q, got %q", "NewProject", arch.SystemContext.Name)
	}

	// Verify file was created.
	if _, err := os.Stat(filepath.Join(tmpDir, "architecture", "system.yaml")); err != nil {
		t.Errorf("expected architecture file to be created: %v", err)
	}
}

// --- blast_radius PostProcess tests ---

func TestBlastRadiusPostProcess_AddsRiskLabelsAndContextAnnotations(t *testing.T) {
	t.Parallel()

	providerReg := mcp.NewProviderRegistry()
	providerReg.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools:   []mcp.ToolDefinition{{Name: "detect_changes", Namespace: "code"}},
		callFn: func(_ context.Context, _ string, _ map[string]any) (any, error) {
			return map[string]any{
				"changes": []any{
					map[string]any{
						"path":       "internal/skill/skill.go",
						"dependents": []any{"a.go", "b.go", "c.go", "d.go", "e.go"},
					},
					map[string]any{
						"path":       "internal/mcp/mcp.go",
						"dependents": []any{"x.go", "y.go"},
					},
					map[string]any{
						"path":       "cmd/forgia/main.go",
						"dependents": []any{},
					},
				},
			}, nil
		},
	})

	v := &mockVaultReader{
		contexts: []*vault.BoundedContext{
			{
				Name:             "skill-system",
				Responsibilities: []string{"internal/skill"},
			},
			{
				Name:             "mcp-protocol",
				Responsibilities: []string{"internal/mcp"},
			},
		},
	}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	s, _ := reg.Get("blast_radius")
	result, err := s.Execute(context.Background(), map[string]any{"path": "internal/skill"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]any)
	changes := resultMap["changes"].([]any)

	// Verify risk labels.
	tests := []struct {
		idx      int
		wantRisk string
		wantCtx  string
	}{
		{0, "High", "skill-system"},    // 5 dependents
		{1, "Medium", "mcp-protocol"},  // 2 dependents
		{2, "Low", ""},                 // 0 dependents, no context match
	}

	for _, tt := range tests {
		cm := changes[tt.idx].(map[string]any)
		if cm["risk"] != tt.wantRisk {
			t.Errorf("change[%d]: expected risk %q, got %q", tt.idx, tt.wantRisk, cm["risk"])
		}
		ctx, _ := cm["context"].(string)
		if ctx != tt.wantCtx {
			t.Errorf("change[%d]: expected context %q, got %q", tt.idx, tt.wantCtx, ctx)
		}
	}
}

// --- trace_calls PostProcess tests ---

func TestTraceCallsPostProcess_AddsContextAnnotations(t *testing.T) {
	t.Parallel()

	providerReg := mcp.NewProviderRegistry()
	providerReg.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools:   []mcp.ToolDefinition{{Name: "trace_call_path", Namespace: "code"}},
		callFn: func(_ context.Context, _ string, _ map[string]any) (any, error) {
			return map[string]any{
				"call_path": []any{
					map[string]any{"name": "main", "file": "cmd/forgia/main.go"},
					map[string]any{"name": "Execute", "file": "internal/skill/skill.go"},
					map[string]any{"name": "Call", "file": "internal/mcp/mcp.go"},
				},
			}, nil
		},
	})

	v := &mockVaultReader{
		contexts: []*vault.BoundedContext{
			{
				Name:             "skill-system",
				Responsibilities: []string{"internal/skill"},
			},
			{
				Name:             "mcp-protocol",
				Responsibilities: []string{"internal/mcp"},
			},
		},
	}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills failed: %v", err)
	}

	s, _ := reg.Get("trace_calls")
	result, err := s.Execute(context.Background(), map[string]any{"function": "main"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]any)
	callPath := resultMap["call_path"].([]any)

	tests := []struct {
		idx     int
		wantCtx string
	}{
		{0, ""},               // cmd/forgia/main.go — no context match
		{1, "skill-system"},   // internal/skill/skill.go
		{2, "mcp-protocol"},   // internal/mcp/mcp.go
	}

	for _, tt := range tests {
		nm := callPath[tt.idx].(map[string]any)
		ctx, _ := nm["context"].(string)
		if ctx != tt.wantCtx {
			t.Errorf("call_path[%d]: expected context %q, got %q", tt.idx, tt.wantCtx, ctx)
		}
	}
}

// --- Helper function tests ---

func TestMapToArchitecture_EmptyResult(t *testing.T) {
	t.Parallel()

	arch := mapToArchitecture("not a map")
	if arch == nil {
		t.Fatal("expected non-nil Architecture")
	}
	if arch.SystemContext != nil {
		t.Error("expected nil SystemContext for non-map input")
	}
}

func TestMergeArchitecture_NilCases(t *testing.T) {
	t.Parallel()

	incoming := &vault.Architecture{
		SystemContext: &vault.SystemContext{Name: "Test"},
	}

	if got := mergeArchitecture(nil, incoming); got != incoming {
		t.Error("expected incoming when existing is nil")
	}
	if got := mergeArchitecture(incoming, nil); got != incoming {
		t.Error("expected existing when incoming is nil")
	}
}

func TestAddRiskLabels_Thresholds(t *testing.T) {
	t.Parallel()

	result := map[string]any{
		"changes": []any{
			map[string]any{"dependents": []any{"a", "b", "c", "d", "e", "f"}}, // 6 → High
			map[string]any{"dependents": []any{"a", "b", "c", "d", "e"}},      // 5 → High
			map[string]any{"dependents": []any{"a", "b", "c", "d"}},           // 4 → Medium
			map[string]any{"dependents": []any{"a", "b"}},                     // 2 → Medium
			map[string]any{"dependents": []any{"a"}},                          // 1 → Low
			map[string]any{"dependents": []any{}},                             // 0 → Low
			map[string]any{},                                                  // no dependents key → Low
		},
	}

	addRiskLabels(result)

	expected := []string{"High", "High", "Medium", "Medium", "Low", "Low", "Low"}
	changes := result["changes"].([]any)
	for i, want := range expected {
		cm := changes[i].(map[string]any)
		if cm["risk"] != want {
			t.Errorf("change[%d]: expected risk %q, got %q", i, want, cm["risk"])
		}
	}
}

func TestFilterResultPaths_NoResultsKey(t *testing.T) {
	t.Parallel()

	// Result without "results" key → pass through.
	result := map[string]any{"data": "something"}
	got := filterResultPaths(context.Background(), result, &guardrails.Guardrails{})
	if _, ok := got.(map[string]any); !ok {
		t.Errorf("expected map result, got %T", got)
	}
}
