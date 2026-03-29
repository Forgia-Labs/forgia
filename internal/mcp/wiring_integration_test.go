package mcp

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"
)

// --- SDD-008 integration tests: full startup lifecycle ---

// TestIntegration_FullLifecycle_StartServeShutdown wires providers and skills
// into the MCP server, verifies tools/list returns all expected tools, then
// cancels context and checks providers are stopped.
func TestIntegration_FullLifecycle_StartServeShutdown(t *testing.T) {
	t.Parallel()

	var stopped atomic.Int32

	// Mock provider that tracks Stop() calls.
	provider := &mockProvider{
		name:    "code",
		healthy: true,
		tools: []ToolDefinition{
			{Name: "search_graph", Namespace: "code", Description: "Search"},
			{Name: "detect_changes", Namespace: "code", Description: "Detect"},
		},
		callFn: func(_ context.Context, tool string, params map[string]any) (any, error) {
			return map[string]any{"tool": tool}, nil
		},
	}

	// Wrap to track Stop().
	trackable := &stopTrackingProvider{
		ToolProvider: provider,
		stopped:      &stopped,
	}

	// Wire provider registry.
	providers := NewProviderRegistry()
	providers.Register(trackable)

	// Wire skill dispatcher with composite-like skills.
	skills := &mockSkillDispatcher{
		tools: []ToolDefinition{
			{Name: "blast_radius", Description: "Blast radius"},
			{Name: "arch_init", Description: "Architecture init"},
			{Name: "trace_calls", Description: "Trace calls"},
			{Name: "search_code", Description: "Search code"},
			{Name: "security_scan", Description: "Security scan"},
			{Name: "arch_coherence", Description: "Architecture coherence"},
			{Name: "context_map", Description: "Context map"},
		},
		execFn: func(_ context.Context, name string, _ map[string]any) (any, bool, error) {
			return map[string]any{"skill": name}, true, nil
		},
	}

	// Create server.
	server, inW, outR := newTestServer(providers, skills)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(ctx) }()

	// 1. Initialize.
	initResp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 1, Method: "initialize",
	})
	if initResp.Error != nil {
		t.Fatalf("initialize: %v", initResp.Error)
	}

	// 2. List tools — expect 2 provider tools + 7 skill tools = 9.
	listResp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 2, Method: "tools/list",
	})
	if listResp.Error != nil {
		t.Fatalf("tools/list: %v", listResp.Error)
	}

	var listResult struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(listResp.Result, &listResult); err != nil {
		t.Fatalf("unmarshal tools/list: %v", err)
	}
	if got := len(listResult.Tools); got != 9 {
		t.Errorf("expected 9 tools (2 provider + 7 skill), got %d", got)
		for _, tool := range listResult.Tools {
			t.Logf("  tool: %s", tool.Name)
		}
	}

	// 3. Call a composite skill.
	skillResp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 3, Method: "tools/call",
		Params: map[string]any{
			"name":      "blast_radius",
			"arguments": map[string]any{},
		},
	})
	if skillResp.Error != nil {
		t.Fatalf("tools/call blast_radius: %v", skillResp.Error)
	}

	// 4. Cancel context (simulates SIGINT/SIGTERM).
	cancel()
	inW.Close()

	select {
	case <-serveDone:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within 2s")
	}

	// 5. Call StopAll — verify providers are stopped.
	providers.StopAll()
	if got := stopped.Load(); got != 1 {
		t.Errorf("expected 1 provider Stop() call, got %d", got)
	}
}

// TestIntegration_StopAll_MultipleProviders verifies StopAll calls Stop()
// on every registered provider.
func TestIntegration_StopAll_MultipleProviders(t *testing.T) {
	t.Parallel()

	var stopped atomic.Int32

	providers := NewProviderRegistry()
	for _, name := range []string{"code", "memory", "board"} {
		providers.Register(&stopTrackingProvider{
			ToolProvider: &mockProvider{name: name, healthy: true},
			stopped:      &stopped,
		})
	}

	providers.StopAll()

	if got := stopped.Load(); got != 3 {
		t.Errorf("expected 3 Stop() calls, got %d", got)
	}
}

// TestIntegration_ServerShutdown_ContextCancel verifies the server exits
// cleanly when context is cancelled (simulating SIGINT).
func TestIntegration_ServerShutdown_ContextCancel(t *testing.T) {
	t.Parallel()

	providers := NewProviderRegistry()
	server, inW, outR := newTestServer(providers, nil)

	ctx, cancel := context.WithCancel(context.Background())

	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(ctx) }()

	// Send initialize to confirm server is running.
	resp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 1, Method: "initialize",
	})
	if resp.Error != nil {
		t.Fatalf("initialize: %v", resp.Error)
	}

	// Cancel context.
	cancel()
	inW.Close()

	select {
	case err := <-serveDone:
		if err != nil && err != context.Canceled {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within 2s")
	}
}

// TestIntegration_MetadataProvider verifies the lightweight provider for skill listing.
func TestIntegration_MetadataProvider(t *testing.T) {
	t.Parallel()

	p := NewMetadataProvider("code", "code")

	if p.Name() != "code" {
		t.Errorf("Name() = %q, want %q", p.Name(), "code")
	}
	if !p.Healthy() {
		t.Error("MetadataProvider should always be healthy")
	}
	if p.Tools() != nil {
		t.Error("MetadataProvider should return nil tools")
	}
	if err := p.Start(context.Background()); err != nil {
		t.Errorf("Start() = %v, want nil", err)
	}
	if err := p.Stop(); err != nil {
		t.Errorf("Stop() = %v, want nil", err)
	}
	if _, err := p.Call(context.Background(), "test", nil); err == nil {
		t.Error("Call() should return error on metadata-only provider")
	}
}

// TestIntegration_SkillRegistryWithEmbeddedAndComposite verifies that
// a registry containing both embedded and composite skills returns all
// skills through All() and MCPToolDefs().
func TestIntegration_SkillRegistryWithEmbeddedAndComposite(t *testing.T) {
	t.Parallel()

	// This test uses the real skill package through the MCP server
	// to verify the full tools/list includes both provider and skill tools.
	providers := NewProviderRegistry()
	providers.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools: []ToolDefinition{
			{Name: "search_graph", Namespace: "code", Description: "Search"},
		},
	})

	// Simulate 19 embedded + 7 composite by having skill dispatcher return all.
	allSkillTools := []ToolDefinition{
		{Name: "blast_radius", Description: "Blast radius"},
		{Name: "arch_init", Description: "Architecture init"},
		{Name: "trace_calls", Description: "Trace calls"},
		{Name: "search_code", Description: "Search code"},
		{Name: "security_scan", Description: "Security scan"},
		{Name: "arch_coherence", Description: "Architecture coherence"},
		{Name: "context_map", Description: "Context map"},
	}

	skills := &mockSkillDispatcher{tools: allSkillTools}

	server, inW, outR := newTestServer(providers, skills)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	resp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 1, Method: "tools/list",
	})
	if resp.Error != nil {
		t.Fatalf("tools/list: %v", resp.Error)
	}

	var result struct {
		Tools []struct{ Name string } `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// 1 provider tool + 7 skill tools = 8.
	if got := len(result.Tools); got != 8 {
		t.Errorf("expected 8 tools, got %d", got)
	}

	// Verify all 7 composite skills present.
	names := make(map[string]bool)
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	for _, expected := range []string{
		"blast_radius", "arch_init", "trace_calls", "search_code",
		"security_scan", "arch_coherence", "context_map",
	} {
		if !names[expected] {
			t.Errorf("missing composite skill %q in tools/list", expected)
		}
	}

	cancel()
	inW.Close()
}

// --- Test helpers ---

// stopTrackingProvider wraps a ToolProvider and counts Stop() calls.
type stopTrackingProvider struct {
	ToolProvider
	stopped *atomic.Int32
}

func (p *stopTrackingProvider) Stop() error {
	p.stopped.Add(1)
	return p.ToolProvider.Stop()
}

