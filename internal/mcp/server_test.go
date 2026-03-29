package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// mockSkillDispatcher implements SkillDispatcher for testing.
type mockSkillDispatcher struct {
	tools    []ToolDefinition
	execFn   func(ctx context.Context, name string, params map[string]any) (any, bool, error)
}

func (m *mockSkillDispatcher) MCPToolDefs() []ToolDefinition {
	return m.tools
}

func (m *mockSkillDispatcher) GetAndExecute(ctx context.Context, name string, params map[string]any) (any, bool, error) {
	if m.execFn != nil {
		return m.execFn(ctx, name, params)
	}
	return nil, false, nil
}

// serverTestHelper sends a request to the server and reads the response.
func serverTestHelper(t *testing.T, _ *MCPServer, in *io.PipeWriter, out *io.PipeReader, req jsonrpcRequest) jsonrpcResponse {
	t.Helper()

	enc := json.NewEncoder(in)
	dec := json.NewDecoder(out)

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode request: %v", err)
	}

	var resp jsonrpcResponse
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func newTestServer(providers *ProviderRegistry, skills SkillDispatcher) (*MCPServer, *io.PipeWriter, *io.PipeReader) {
	inR, inW := io.Pipe()   // test writes → server reads
	outR, outW := io.Pipe() // server writes → test reads

	transport := NewStdioTransport(inR, outW)
	server := NewMCPServer(transport, providers, skills)

	return server, inW, outR
}

func TestMCPServer_Initialize(t *testing.T) {
	t.Parallel()

	providers := NewProviderRegistry()
	server, inW, outR := newTestServer(providers, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	req := jsonrpcRequest{JSONRPC: jsonrpcVersion, ID: 1, Method: "initialize"}
	resp := serverTestHelper(t, server, inW, outR, req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if resp.ID == nil || *resp.ID != 1 {
		t.Fatalf("expected ID=1, got %v", resp.ID)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("missing serverInfo in response")
	}
	if serverInfo["name"] != "forgia" {
		t.Errorf("expected server name 'forgia', got %v", serverInfo["name"])
	}

	cancel()
	inW.Close()
}

func TestMCPServer_ToolsList(t *testing.T) {
	t.Parallel()

	providers := NewProviderRegistry()
	providers.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools: []ToolDefinition{
			{Name: "search_graph", Namespace: "code", Description: "Search the graph"},
		},
	})

	skills := &mockSkillDispatcher{
		tools: []ToolDefinition{
			{Name: "forgia_blast_radius", Description: "Blast radius analysis"},
		},
	}

	server, inW, outR := newTestServer(providers, skills)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	req := jsonrpcRequest{JSONRPC: jsonrpcVersion, ID: 2, Method: "tools/list"}
	resp := serverTestHelper(t, server, inW, outR, req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result struct {
		Tools []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(result.Tools))
	}

	names := make(map[string]bool)
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	if !names["forgia_code_search_graph"] {
		t.Error("missing provider tool forgia_code_search_graph")
	}
	if !names["forgia_blast_radius"] {
		t.Error("missing skill tool forgia_blast_radius")
	}

	cancel()
	inW.Close()
}

func TestMCPServer_ToolsCallProvider(t *testing.T) {
	t.Parallel()

	providers := NewProviderRegistry()
	providers.Register(&mockProvider{
		name:    "code",
		healthy: true,
		callFn: func(_ context.Context, tool string, params map[string]any) (any, error) {
			return map[string]any{"tool": tool, "called": true}, nil
		},
	})

	server, inW, outR := newTestServer(providers, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	callParams := map[string]any{
		"name":      "forgia_code_search_graph",
		"arguments": map[string]any{"query": "test"},
	}
	req := jsonrpcRequest{JSONRPC: jsonrpcVersion, ID: 3, Method: "tools/call", Params: callParams}
	resp := serverTestHelper(t, server, inW, outR, req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["tool"] != "search_graph" {
		t.Errorf("expected tool 'search_graph', got %v", result["tool"])
	}
	if result["called"] != true {
		t.Error("expected called=true")
	}

	cancel()
	inW.Close()
}

func TestMCPServer_ToolsCallSkill(t *testing.T) {
	t.Parallel()

	providers := NewProviderRegistry()
	skills := &mockSkillDispatcher{
		execFn: func(_ context.Context, name string, params map[string]any) (any, bool, error) {
			if name == "forgia_blast_radius" {
				return map[string]any{"skill": name, "executed": true}, true, nil
			}
			return nil, false, nil
		},
	}

	server, inW, outR := newTestServer(providers, skills)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	callParams := map[string]any{
		"name":      "forgia_blast_radius",
		"arguments": map[string]any{},
	}
	req := jsonrpcRequest{JSONRPC: jsonrpcVersion, ID: 4, Method: "tools/call", Params: callParams}
	resp := serverTestHelper(t, server, inW, outR, req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["skill"] != "forgia_blast_radius" {
		t.Errorf("expected skill 'forgia_blast_radius', got %v", result["skill"])
	}

	cancel()
	inW.Close()
}

func TestMCPServer_UnknownMethod(t *testing.T) {
	t.Parallel()

	providers := NewProviderRegistry()
	server, inW, outR := newTestServer(providers, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	req := jsonrpcRequest{JSONRPC: jsonrpcVersion, ID: 5, Method: "unknown/method"}
	resp := serverTestHelper(t, server, inW, outR, req)

	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", resp.Error.Code)
	}
	if !strings.Contains(resp.Error.Message, "Method not found") {
		t.Errorf("expected 'Method not found', got %q", resp.Error.Message)
	}

	cancel()
	inW.Close()
}

func TestMCPServer_MalformedJSON(t *testing.T) {
	t.Parallel()

	// Send malformed JSON — server should return parse error, not crash.
	providers := NewProviderRegistry()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	transport := NewStdioTransport(inR, outW)
	server := NewMCPServer(transport, providers, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	// Send malformed JSON then close stdin.
	go func() {
		inW.Write([]byte("{invalid json}\n"))
		inW.Close()
	}()

	dec := json.NewDecoder(outR)

	// Should get a parse error response.
	var errResp jsonrpcResponse
	if err := dec.Decode(&errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errResp.Error == nil {
		t.Fatal("expected parse error response")
	}
	if errResp.Error.Code != -32700 {
		t.Errorf("expected error code -32700, got %d", errResp.Error.Code)
	}
}

func TestMCPServer_StdinEOF(t *testing.T) {
	t.Parallel()

	// Server should exit cleanly on stdin EOF.
	input := bytes.NewReader(nil) // empty → immediate EOF
	var output bytes.Buffer

	transport := NewStdioTransport(input, &output)
	providers := NewProviderRegistry()
	server := NewMCPServer(transport, providers, nil)

	err := server.Serve(context.Background())
	if err != nil {
		t.Fatalf("expected nil error on EOF, got: %v", err)
	}
}

func TestMCPServer_ToolsListEmpty(t *testing.T) {
	t.Parallel()

	providers := NewProviderRegistry()
	server, inW, outR := newTestServer(providers, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	req := jsonrpcRequest{JSONRPC: jsonrpcVersion, ID: 1, Method: "tools/list"}
	resp := serverTestHelper(t, server, inW, outR, req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result struct {
		Tools []any `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Tools == nil {
		t.Error("tools should be empty array, not null")
	}
	if len(result.Tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(result.Tools))
	}

	cancel()
	inW.Close()
}
