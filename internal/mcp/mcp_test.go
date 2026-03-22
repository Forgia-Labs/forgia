package mcp

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// mockProvider implements ToolProvider for testing.
type mockProvider struct {
	name    string
	tools   []ToolDefinition
	healthy bool
	callFn  func(ctx context.Context, tool string, params map[string]any) (any, error)
}

func (m *mockProvider) Name() string            { return m.name }
func (m *mockProvider) Tools() []ToolDefinition { return m.tools }
func (m *mockProvider) Healthy() bool           { return m.healthy }
func (m *mockProvider) Start(context.Context) error { return nil }
func (m *mockProvider) Stop() error                 { return nil }
func (m *mockProvider) Call(ctx context.Context, tool string, params map[string]any) (any, error) {
	if m.callFn != nil {
		return m.callFn(ctx, tool, params)
	}
	return fmt.Sprintf("called %s", tool), nil
}

func TestNewProviderRegistry(t *testing.T) {
	r := NewProviderRegistry()
	if r == nil {
		t.Fatal("NewProviderRegistry returned nil")
	}
	if len(r.providers) != 0 {
		t.Errorf("new registry should be empty, got %d providers", len(r.providers))
	}
}

func TestRegister(t *testing.T) {
	r := NewProviderRegistry()
	p := &mockProvider{name: "code", healthy: true}

	r.Register(p)

	r.mu.RLock()
	defer r.mu.RUnlock()
	if got, ok := r.providers["code"]; !ok {
		t.Error("provider 'code' not found after Register")
	} else if got != p {
		t.Error("provider 'code' does not match registered instance")
	}
}

func TestRegisterOverwrites(t *testing.T) {
	r := NewProviderRegistry()
	p1 := &mockProvider{name: "code", healthy: true, tools: []ToolDefinition{{Name: "old"}}}
	p2 := &mockProvider{name: "code", healthy: true, tools: []ToolDefinition{{Name: "new"}}}

	r.Register(p1)
	r.Register(p2)

	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.providers) != 1 {
		t.Errorf("expected 1 provider after overwrite, got %d", len(r.providers))
	}
	if r.providers["code"].Tools()[0].Name != "new" {
		t.Error("Register should overwrite existing provider with same name")
	}
}

func TestAllToolsEmpty(t *testing.T) {
	r := NewProviderRegistry()
	tools := r.AllTools()
	if len(tools) != 0 {
		t.Errorf("expected 0 tools from empty registry, got %d", len(tools))
	}
}

func TestAllToolsNamespacing(t *testing.T) {
	r := NewProviderRegistry()
	r.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools: []ToolDefinition{
			{Name: "search_graph", Namespace: "code", Description: "search"},
			{Name: "detect_changes", Namespace: "code", Description: "detect"},
		},
	})
	r.Register(&mockProvider{
		name:    "mem",
		healthy: true,
		tools: []ToolDefinition{
			{Name: "recall", Namespace: "mem", Description: "recall"},
		},
	})

	tools := r.AllTools()
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}

	names := make(map[string]bool)
	for _, td := range tools {
		names[td.Name] = true
	}
	for _, want := range []string{"forgia_code_search_graph", "forgia_code_detect_changes", "forgia_mem_recall"} {
		if !names[want] {
			t.Errorf("expected tool %q in AllTools, got %v", want, names)
		}
	}
}

func TestAllToolsSkipsUnhealthy(t *testing.T) {
	r := NewProviderRegistry()
	r.Register(&mockProvider{
		name:    "healthy",
		healthy: true,
		tools:   []ToolDefinition{{Name: "ok", Namespace: "healthy"}},
	})
	r.Register(&mockProvider{
		name:    "down",
		healthy: false,
		tools:   []ToolDefinition{{Name: "nope", Namespace: "down"}},
	})

	tools := r.AllTools()
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool (unhealthy skipped), got %d", len(tools))
	}
	if tools[0].Name != "forgia_healthy_ok" {
		t.Errorf("expected forgia_healthy_ok, got %s", tools[0].Name)
	}
}

func TestCallRoutesToProvider(t *testing.T) {
	r := NewProviderRegistry()
	var calledTool string
	r.Register(&mockProvider{
		name:    "code",
		healthy: true,
		callFn: func(_ context.Context, tool string, _ map[string]any) (any, error) {
			calledTool = tool
			return "result", nil
		},
	})

	result, err := r.Call(context.Background(), "forgia_code_search_graph", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calledTool != "search_graph" {
		t.Errorf("expected provider to receive tool 'search_graph', got %q", calledTool)
	}
	if result != "result" {
		t.Errorf("expected 'result', got %v", result)
	}
}

func TestCallUnknownNamespace(t *testing.T) {
	r := NewProviderRegistry()
	r.Register(&mockProvider{name: "code", healthy: true})

	_, err := r.Call(context.Background(), "forgia_unknown_tool", nil)
	if err == nil {
		t.Fatal("expected error for unknown namespace")
	}
	if want := "unknown provider namespace: unknown"; err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestCallInvalidToolName(t *testing.T) {
	r := NewProviderRegistry()

	_, err := r.Call(context.Background(), "bad_tool_name", nil)
	if err == nil {
		t.Fatal("expected error for invalid tool name")
	}
}

func TestCallPassesParams(t *testing.T) {
	r := NewProviderRegistry()
	var gotParams map[string]any
	r.Register(&mockProvider{
		name:    "code",
		healthy: true,
		callFn: func(_ context.Context, _ string, params map[string]any) (any, error) {
			gotParams = params
			return nil, nil
		},
	})

	params := map[string]any{"query": "test", "limit": 10}
	_, _ = r.Call(context.Background(), "forgia_code_search", params)

	if gotParams["query"] != "test" || gotParams["limit"] != 10 {
		t.Errorf("params not passed through: got %v", gotParams)
	}
}

func TestCallProviderError(t *testing.T) {
	r := NewProviderRegistry()
	r.Register(&mockProvider{
		name:    "code",
		healthy: true,
		callFn: func(_ context.Context, _ string, _ map[string]any) (any, error) {
			return nil, fmt.Errorf("provider exploded")
		},
	})

	_, err := r.Call(context.Background(), "forgia_code_search", nil)
	if err == nil {
		t.Fatal("expected error from provider")
	}
	if err.Error() != "provider exploded" {
		t.Errorf("error = %q, want 'provider exploded'", err.Error())
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	r := NewProviderRegistry()
	for i := range 10 {
		r.Register(&mockProvider{
			name:    fmt.Sprintf("p%d", i),
			healthy: true,
			tools:   []ToolDefinition{{Name: "t", Namespace: fmt.Sprintf("p%d", i)}},
		})
	}

	var wg sync.WaitGroup
	// Concurrent reads (AllTools + Call) while registering new providers.
	for i := range 50 {
		wg.Add(3)
		go func(i int) {
			defer wg.Done()
			r.AllTools()
		}(i)
		go func(i int) {
			defer wg.Done()
			r.Call(context.Background(), fmt.Sprintf("forgia_p%d_t", i%10), nil)
		}(i)
		go func(i int) {
			defer wg.Done()
			r.Register(&mockProvider{
				name:    fmt.Sprintf("new%d", i),
				healthy: true,
				tools:   []ToolDefinition{{Name: "t", Namespace: fmt.Sprintf("new%d", i)}},
			})
		}(i)
	}
	wg.Wait()
}

func TestToolName(t *testing.T) {
	tests := []struct {
		namespace string
		tool      string
		want      string
	}{
		{"code", "search_graph", "forgia_code_search_graph"},
		{"code", "detect_changes", "forgia_code_detect_changes"},
		{"custom", "my_tool", "forgia_custom_my_tool"},
	}

	for _, tt := range tests {
		got := ToolName(tt.namespace, tt.tool)
		if got != tt.want {
			t.Errorf("ToolName(%q, %q) = %q, want %q", tt.namespace, tt.tool, got, tt.want)
		}
	}
}

func TestParseNamespacedTool(t *testing.T) {
	tests := []struct {
		tool      string
		wantNS    string
		wantName  string
		wantError bool
	}{
		{"forgia_code_search_graph", "code", "search_graph", false},
		{"forgia_code_detect_changes", "code", "detect_changes", false},
		{"forgia_custom_my_tool", "custom", "my_tool", false},
		{"forgia_ns_a_b_c", "ns", "a_b_c", false},

		// Error cases.
		{"search_graph", "", "", true},           // missing forgia_ prefix
		{"forgia_nounderscoreafter", "", "", true}, // no separator after namespace
		{"", "", "", true},                        // empty
		{"other_prefix_tool", "", "", true},       // wrong prefix
	}

	for _, tt := range tests {
		ns, name, err := parseNamespacedTool(tt.tool)
		if tt.wantError {
			if err == nil {
				t.Errorf("parseNamespacedTool(%q) expected error, got ns=%q name=%q", tt.tool, ns, name)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseNamespacedTool(%q) unexpected error: %v", tt.tool, err)
			continue
		}
		if ns != tt.wantNS || name != tt.wantName {
			t.Errorf("parseNamespacedTool(%q) = (%q, %q), want (%q, %q)", tt.tool, ns, name, tt.wantNS, tt.wantName)
		}
	}
}
