package skill

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Deepzima/forgia/internal/mcp"
	"github.com/Deepzima/forgia/internal/vault"
)

// --- Composite skill integration tests (SDD-007 suite 3) ---

// TestIntegration_SearchCode_GuardrailsFilter registers a mock provider
// with fixture results containing .env and .pem paths, then executes
// search_code through the full stack and verifies deny.toml-matched
// paths are stripped from the result.
func TestIntegration_SearchCode_GuardrailsFilter(t *testing.T) {
	t.Parallel()

	denyToml := `[read]
patterns = [
  "**/.env",
  "**/.env.*",
  "!**/.env.example",
  "!**/.env.template",
  "**/*.pem",
  "**/*.key",
  "**/*.p12",
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
					map[string]any{"path": "internal/mcp/server.go", "name": "MCPServer"},
					map[string]any{"path": "config/.env", "name": "SECRET_DB_URL"},
					map[string]any{"path": "deploy/.env.production", "name": "PROD_KEY"},
					map[string]any{"path": "certs/server.pem", "name": "tls_cert"},
					map[string]any{"path": "certs/private.key", "name": "tls_key"},
					map[string]any{"path": "auth/credentials.json", "name": "service_creds"},
					map[string]any{"path": "internal/skill/skill.go", "name": "Registry"},
					map[string]any{"path": "templates/.env.example", "name": "example_env"},
					map[string]any{"path": "templates/.env.template", "name": "template_env"},
				},
			}, nil
		},
	})

	v := &mockVaultReader{guardrails: []byte(denyToml)}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills: %v", err)
	}

	s, ok := reg.Get("search_code")
	if !ok {
		t.Fatal("search_code not registered")
	}

	result, err := s.Execute(context.Background(), map[string]any{"query": "config"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	results, ok := resultMap["results"].([]any)
	if !ok {
		t.Fatalf("expected results array, got %T", resultMap["results"])
	}

	// Should keep: server.go, skill.go, .env.example, .env.template (allow-listed).
	// Should strip: .env, .env.production, .pem, .key, credentials.json.
	kept := make(map[string]bool)
	for _, r := range results {
		if rm, ok := r.(map[string]any); ok {
			if p, ok := rm["path"].(string); ok {
				kept[p] = true
			}
		}
	}

	wantKept := []string{
		"internal/mcp/server.go",
		"internal/skill/skill.go",
		"templates/.env.example",
		"templates/.env.template",
	}
	wantStripped := []string{
		"config/.env",
		"deploy/.env.production",
		"certs/server.pem",
		"certs/private.key",
		"auth/credentials.json",
	}

	for _, p := range wantKept {
		if !kept[p] {
			t.Errorf("expected %q to be kept after guardrails filtering", p)
		}
	}
	for _, p := range wantStripped {
		if kept[p] {
			t.Errorf("expected %q to be stripped by guardrails", p)
		}
	}

	if len(results) != len(wantKept) {
		t.Errorf("expected %d results after filtering, got %d", len(wantKept), len(results))
	}
}

// TestIntegration_SecurityScan_NoSecretContent verifies that security_scan
// returns only file paths and pattern names — never actual secret values.
// This is a security requirement from the FD-006 threat model.
func TestIntegration_SecurityScan_NoSecretContent(t *testing.T) {
	t.Parallel()

	denyToml := `[read]
patterns = [
  "**/.env",
  "**/*.pem",
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
			// Simulate provider returning results with secret-like content.
			return map[string]any{
				"results": []any{
					map[string]any{
						"path":    "internal/auth/handler.go",
						"name":    "ValidateToken",
						"content": "func ValidateToken(token string) { if token == \"sk-secret-key-12345\" { ... } }",
						"value":   "AKIA1234567890ABCDEF",
					},
					map[string]any{
						"path":    "config/db.go",
						"name":    "ConnectDB",
						"content": "password := \"super_secret_p@ssword\"",
						"value":   "postgresql://admin:secret@localhost:5432/db",
					},
					map[string]any{
						"path":    "deploy/.env",
						"name":    "ENV_FILE",
						"content": "API_KEY=sk-ant-very-secret-key",
					},
				},
			}, nil
		},
	})

	v := &mockVaultReader{guardrails: []byte(denyToml)}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills: %v", err)
	}

	s, ok := reg.Get("security_scan")
	if !ok {
		t.Fatal("security_scan not registered")
	}

	result, err := s.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	// Verify findings contain only path and name — no content or value fields.
	findings, ok := resultMap["findings"].([]map[string]any)
	if !ok {
		t.Fatalf("expected findings as []map[string]any, got %T", resultMap["findings"])
	}

	for i, f := range findings {
		if _, hasContent := f["content"]; hasContent {
			t.Errorf("finding[%d]: must not contain 'content' field (secret leak risk)", i)
		}
		if _, hasValue := f["value"]; hasValue {
			t.Errorf("finding[%d]: must not contain 'value' field (secret leak risk)", i)
		}
		if _, hasPath := f["path"]; !hasPath {
			t.Errorf("finding[%d]: must contain 'path' field", i)
		}
	}

	// Verify no secret content appears anywhere in the serialized output.
	serialized, _ := jsonMarshalAny(resultMap)
	secretStrings := []string{
		"sk-secret-key-12345",
		"AKIA1234567890ABCDEF",
		"super_secret_p@ssword",
		"postgresql://admin:secret",
		"sk-ant-very-secret-key",
	}
	for _, secret := range secretStrings {
		if strings.Contains(serialized, secret) {
			t.Errorf("serialized output contains secret value %q — security_scan must not leak secret content", secret)
		}
	}

	// Verify guardrail_gaps identifies unprotected security-relevant files.
	gaps, _ := resultMap["guardrail_gaps"].([]map[string]any)
	// handler.go and db.go are not in deny.toml → should be flagged as unprotected.
	unprotectedPaths := make(map[string]bool)
	for _, g := range gaps {
		if p, ok := g["path"].(string); ok {
			unprotectedPaths[p] = true
		}
	}
	for _, want := range []string{"internal/auth/handler.go", "config/db.go"} {
		if !unprotectedPaths[want] {
			t.Errorf("expected %q to be flagged as unprotected guardrail gap", want)
		}
	}
}

// TestIntegration_BlastRadius_EnrichedResult verifies blast_radius
// adds risk labels and bounded context annotations to results.
func TestIntegration_BlastRadius_EnrichedResult(t *testing.T) {
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
						"path":       "internal/mcp/server.go",
						"dependents": []any{"a", "b", "c", "d", "e", "f"}, // 6 → High
					},
					map[string]any{
						"path":       "internal/skill/composite.go",
						"dependents": []any{"x", "y", "z"}, // 3 → Medium
					},
					map[string]any{
						"path":       "cmd/forgia/cmd/mcp.go",
						"dependents": []any{"single"}, // 1 → Low
					},
				},
			}, nil
		},
	})

	v := &mockVaultReader{
		contexts: []*vault.BoundedContext{
			{
				Name:             "MCP",
				Responsibilities: []string{"internal/mcp"},
			},
			{
				Name:             "Skills",
				Responsibilities: []string{"internal/skill"},
			},
			{
				Name:             "CLI",
				Responsibilities: []string{"cmd/forgia"},
			},
		},
	}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills: %v", err)
	}

	s, ok := reg.Get("blast_radius")
	if !ok {
		t.Fatal("blast_radius not registered")
	}

	result, err := s.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	changes, ok := resultMap["changes"].([]any)
	if !ok {
		t.Fatalf("expected changes array, got %T", resultMap["changes"])
	}

	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(changes))
	}

	// Verify risk labels.
	risks := map[string]string{}
	contexts := map[string]string{}
	for _, c := range changes {
		cm, ok := c.(map[string]any)
		if !ok {
			continue
		}
		path, _ := cm["path"].(string)
		risk, _ := cm["risk"].(string)
		ctx, _ := cm["context"].(string)
		risks[path] = risk
		contexts[path] = ctx
	}

	if risks["internal/mcp/server.go"] != "High" {
		t.Errorf("server.go: expected risk High, got %q", risks["internal/mcp/server.go"])
	}
	if risks["internal/skill/composite.go"] != "Medium" {
		t.Errorf("composite.go: expected risk Medium, got %q", risks["internal/skill/composite.go"])
	}
	if risks["cmd/forgia/cmd/mcp.go"] != "Low" {
		t.Errorf("mcp.go: expected risk Low, got %q", risks["cmd/forgia/cmd/mcp.go"])
	}

	// Verify context annotations.
	if contexts["internal/mcp/server.go"] != "MCP" {
		t.Errorf("server.go: expected context MCP, got %q", contexts["internal/mcp/server.go"])
	}
	if contexts["internal/skill/composite.go"] != "Skills" {
		t.Errorf("composite.go: expected context Skills, got %q", contexts["internal/skill/composite.go"])
	}
	if contexts["cmd/forgia/cmd/mcp.go"] != "CLI" {
		t.Errorf("mcp.go: expected context CLI, got %q", contexts["cmd/forgia/cmd/mcp.go"])
	}
}

// TestIntegration_HigherLevelSkill_CompositionChain verifies that a
// higher-level skill (arch_coherence) calls the underlying primitive
// provider tool and enriches the result with drift detection.
func TestIntegration_HigherLevelSkill_CompositionChain(t *testing.T) {
	t.Parallel()

	var calledTool string
	var calledParams map[string]any

	providerReg := mcp.NewProviderRegistry()
	providerReg.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools:   []mcp.ToolDefinition{{Name: "trace_call_path", Namespace: "code"}},
		callFn: func(_ context.Context, tool string, params map[string]any) (any, error) {
			calledTool = tool
			calledParams = params
			return map[string]any{
				"call_path": []any{
					map[string]any{"file": "internal/mcp/server.go", "function": "Serve"},
					map[string]any{"file": "internal/skill/composite.go", "function": "Execute"},
					map[string]any{"file": "internal/guardrails/guardrails.go", "function": "CheckFilePaths"},
				},
			}, nil
		},
	})

	v := &mockVaultReader{
		arch: &vault.Architecture{
			SystemContext: &vault.SystemContext{Name: "Forgia"},
		},
		contexts: []*vault.BoundedContext{
			{
				Name:             "MCP",
				Responsibilities: []string{"internal/mcp"},
				Dependencies: vault.ContextDependencies{
					DependsOn: []string{"Skills"},
				},
			},
			{
				Name:             "Skills",
				Responsibilities: []string{"internal/skill"},
				// Skills does NOT document dependency on Guardrails.
			},
			{
				Name:             "Guardrails",
				Responsibilities: []string{"internal/guardrails"},
			},
		},
	}

	reg := NewRegistry()
	if err := RegisterCompositeSkills(reg, providerReg, v); err != nil {
		t.Fatalf("RegisterCompositeSkills: %v", err)
	}

	s, ok := reg.Get("arch_coherence")
	if !ok {
		t.Fatal("arch_coherence not registered")
	}

	result, err := s.Execute(context.Background(), map[string]any{"symbol": "Serve"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Verify the primitive provider was called.
	if calledTool != "trace_call_path" {
		t.Errorf("expected underlying tool 'trace_call_path' called, got %q", calledTool)
	}
	if calledParams["symbol"] != "Serve" {
		t.Errorf("expected params passed through, got %v", calledParams)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	// Verify drift detection enrichment.
	drift, ok := resultMap["drift"].([]map[string]any)
	if !ok {
		t.Fatalf("expected drift array, got %T", resultMap["drift"])
	}

	// Skills → Guardrails is NOT documented → should be flagged as drift.
	foundDrift := false
	for _, d := range drift {
		from, _ := d["from"].(string)
		to, _ := d["to"].(string)
		if from == "Skills" && to == "Guardrails" {
			foundDrift = true
			if d["status"] != "undocumented" {
				t.Errorf("expected status 'undocumented', got %v", d["status"])
			}
		}
	}
	if !foundDrift {
		t.Error("expected drift detection for undocumented Skills → Guardrails dependency")
	}

	// MCP → Skills IS documented → should NOT appear in drift.
	for _, d := range drift {
		from, _ := d["from"].(string)
		to, _ := d["to"].(string)
		if from == "MCP" && to == "Skills" {
			t.Error("MCP → Skills is documented, should NOT appear in drift")
		}
	}
}

// --- Slash command fallback tests (SDD-007 suite 4) ---

// TestIntegration_SlashFallback_AllSixCommands verifies that the 6 slash
// commands updated in SDD-006 contain MCP availability check and fallback
// instructions, ensuring they produce useful output when MCP is unavailable.
func TestIntegration_SlashFallback_AllSixCommands(t *testing.T) {
	t.Parallel()

	// The 6 commands updated with MCP/fallback in SDD-006.
	commands := []struct {
		name    string
		mcpTool string // the MCP tool referenced in the availability check
	}{
		{"fd-arch-review", "forgia_search_code"},
		{"fd-threat-model", "forgia_security_scan"},
		{"arch-init", "forgia_arch_init"},
		{"arch-review", "forgia_arch_coherence"},
		{"arch-update", "forgia_arch_coherence"},
		{"sdd-dry-run", "forgia_blast_radius"},
	}

	// Read actual command files from disk.
	root := findProjectRoot(t)
	cmdDir := filepath.Join(root, "modules", "claude-commands")

	for _, cmd := range commands {
		t.Run(cmd.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(cmdDir, cmd.name+".md")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", cmd.name, err)
			}
			content := string(data)

			// Verify MCP availability check exists.
			if !strings.Contains(content, "MCP") {
				t.Errorf("%s: missing MCP availability check", cmd.name)
			}

			// Verify fallback path exists (references direct scanning or alternative).
			hasFallback := strings.Contains(content, "MCP_AVAILABLE=false") ||
				strings.Contains(content, "fallback") ||
				strings.Contains(content, "not available")
			if !hasFallback {
				t.Errorf("%s: missing fallback path for MCP unavailability", cmd.name)
			}

			// Verify the specific MCP tool is referenced.
			if !strings.Contains(content, cmd.mcpTool) || !strings.Contains(content, "tools/list") {
				t.Errorf("%s: should reference %s or tools/list for availability check", cmd.name, cmd.mcpTool)
			}

			// Verify timeout is specified (3-second timeout per SDD-006).
			if !strings.Contains(content, "timeout") && !strings.Contains(content, "3-second") {
				t.Errorf("%s: should specify timeout for MCP availability check", cmd.name)
			}
		})
	}
}

// TestIntegration_SlashFallback_EmbeddedSkillsLoaded verifies the 6 slash
// commands can be loaded as EmbeddedSkills and their content is accessible.
func TestIntegration_SlashFallback_EmbeddedSkillsLoaded(t *testing.T) {
	t.Parallel()

	// Build a mock FS with the 6 command files.
	root := findProjectRoot(t)
	cmdDir := filepath.Join(root, "modules", "claude-commands")

	commandNames := []string{
		"fd-arch-review",
		"fd-threat-model",
		"arch-init",
		"arch-review",
		"arch-update",
		"sdd-dry-run",
	}

	mockFS := fstest.MapFS{}
	for _, name := range commandNames {
		data, err := os.ReadFile(filepath.Join(cmdDir, name+".md"))
		if err != nil {
			t.Fatalf("read %s.md: %v", name, err)
		}
		mockFS[name+".md"] = &fstest.MapFile{Data: data}
	}

	reg := NewRegistry()
	if err := reg.LoadEmbedded(mockFS); err != nil {
		t.Fatalf("LoadEmbedded: %v", err)
	}

	for _, name := range commandNames {
		s, ok := reg.Get(name)
		if !ok {
			t.Errorf("skill %q not loaded", name)
			continue
		}

		if s.Mode() != ModeSlashCommand {
			t.Errorf("%s: expected mode SlashCommand, got %v", name, s.Mode())
		}

		// Content should be non-empty and contain MCP references.
		es, ok := s.(*EmbeddedSkill)
		if !ok {
			t.Errorf("%s: expected *EmbeddedSkill, got %T", name, s)
			continue
		}
		content := es.Content("")
		if len(content) == 0 {
			t.Errorf("%s: content is empty", name)
		}
		if !strings.Contains(content, "MCP") {
			t.Errorf("%s: content should reference MCP", name)
		}
	}
}

// --- Test helpers ---

// findProjectRoot walks up from cwd to find go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find project root (no go.mod)")
		}
		dir = parent
	}
}

// jsonMarshalAny converts a value to string for content inspection.
func jsonMarshalAny(v any) (string, error) {
	return marshalToString(v), nil
}

// marshalToString recursively converts a value to string for content checking.
func marshalToString(v any) string {
	switch val := v.(type) {
	case map[string]any:
		var parts []string
		for k, v := range val {
			parts = append(parts, k+":"+marshalToString(v))
		}
		return strings.Join(parts, ",")
	case []map[string]any:
		var parts []string
		for _, m := range val {
			parts = append(parts, marshalToString(m))
		}
		return strings.Join(parts, ",")
	case []any:
		var parts []string
		for _, item := range val {
			parts = append(parts, marshalToString(item))
		}
		return strings.Join(parts, ",")
	case string:
		return val
	default:
		return ""
	}
}
