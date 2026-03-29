package skill

import (
	"context"
	"testing"

	"github.com/Deepzima/forgia/internal/mcp"
	"github.com/Deepzima/forgia/internal/vault"
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
		"blast-radius",
		"context-lookup",
		"guardrails-check",
		"constitution-read",
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

func TestCompositeSkill_ExecuteWithVaultAwarePreProcess(t *testing.T) {
	t.Parallel()

	providerReg := mcp.NewProviderRegistry()
	providerReg.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools: []mcp.ToolDefinition{
			{Name: "detect_changes", Namespace: "code"},
		},
		callFn: func(_ context.Context, tool string, params map[string]any) (any, error) {
			return map[string]any{"tool": tool, "constitution": params["constitution"]}, nil
		},
	})

	v := &mockVaultReader{
		constitution: "# Test Constitution\nNo secrets allowed.",
	}

	s := &CompositeSkill{
		name:         "blast-radius",
		description:  "test skill",
		category:     CategoryDesign,
		mode:         ModeMCPTool,
		ProviderName: "code",
		ProviderTool: "detect_changes",
		registry:     providerReg,
		vault:        v,
		PreProcess: func(ctx context.Context, params map[string]any) (map[string]any, error) {
			// Vault-aware pre-processing: inject constitution into params.
			constitution, err := v.Constitution(ctx)
			if err != nil {
				return nil, err
			}
			params["constitution"] = constitution
			return params, nil
		},
	}

	result, err := s.Execute(context.Background(), map[string]any{"path": "cmd/main.go"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if resultMap["constitution"] != "# Test Constitution\nNo secrets allowed." {
		t.Errorf("expected constitution in result, got %v", resultMap["constitution"])
	}
}

func TestCompositeSkill_VaultFieldAccessible(t *testing.T) {
	t.Parallel()

	v := &mockVaultReader{
		constitution: "test constitution",
		guardrails:   []byte("[read]\npatterns = []"),
	}

	s := &CompositeSkill{
		name:     "test-skill",
		category: CategoryDesign,
		mode:     ModeMCPTool,
		vault:    v,
	}

	if s.vault == nil {
		t.Fatal("expected vault to be set")
	}

	ctx := context.Background()
	constitution, err := s.vault.Constitution(ctx)
	if err != nil {
		t.Fatalf("Constitution() failed: %v", err)
	}
	if constitution != "test constitution" {
		t.Errorf("unexpected constitution: %q", constitution)
	}

	guardrails, err := s.vault.GuardrailsRaw(ctx)
	if err != nil {
		t.Fatalf("GuardrailsRaw() failed: %v", err)
	}
	if string(guardrails) != "[read]\npatterns = []" {
		t.Errorf("unexpected guardrails: %q", string(guardrails))
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
		t.Errorf("expected 2 architecture skills, got %d", len(archSkills))
	}

	designSkills := reg.ByCategory(CategoryDesign)
	if len(designSkills) != 5 {
		t.Errorf("expected 5 design skills, got %d", len(designSkills))
	}
}
