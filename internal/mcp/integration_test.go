package mcp

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// --- In-process server integration tests (SDD-007 suite 1 & 2) ---

// TestIntegration_ServerFullStack_ToolsCallProvider wires a mock provider
// and skill dispatcher into the server, then verifies tools/call dispatches
// to the correct handler and returns the expected result.
func TestIntegration_ServerFullStack_ToolsCallProvider(t *testing.T) {
	t.Parallel()

	// Wire mock provider.
	providers := NewProviderRegistry()
	providers.Register(&mockProvider{
		name:    "code",
		healthy: true,
		tools: []ToolDefinition{
			{Name: "search_graph", Namespace: "code", Description: "Search the code graph"},
		},
		callFn: func(_ context.Context, tool string, params map[string]any) (any, error) {
			return map[string]any{
				"tool":   tool,
				"query":  params["query"],
				"source": "mock_provider",
			}, nil
		},
	})

	// Wire mock skill dispatcher.
	skills := &mockSkillDispatcher{
		tools: []ToolDefinition{
			{Name: "blast_radius", Description: "Blast radius analysis"},
		},
		execFn: func(_ context.Context, name string, params map[string]any) (any, bool, error) {
			if name == "blast_radius" {
				return map[string]any{"skill": name, "source": "mock_skill"}, true, nil
			}
			return nil, false, nil
		},
	}

	server, inW, outR := newTestServer(providers, skills)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	// 1. Initialize handshake.
	initResp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 1, Method: "initialize",
	})
	if initResp.Error != nil {
		t.Fatalf("initialize error: %v", initResp.Error)
	}

	// 2. List tools — should contain provider + skill tools.
	listResp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 2, Method: "tools/list",
	})
	if listResp.Error != nil {
		t.Fatalf("tools/list error: %v", listResp.Error)
	}
	var listResult struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(listResp.Result, &listResult); err != nil {
		t.Fatalf("unmarshal tools/list: %v", err)
	}
	if len(listResult.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(listResult.Tools))
	}

	// 3. Call provider tool.
	callResp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 3, Method: "tools/call",
		Params: map[string]any{
			"name":      "forgia_code_search_graph",
			"arguments": map[string]any{"query": "auth"},
		},
	})
	if callResp.Error != nil {
		t.Fatalf("tools/call provider error: %v", callResp.Error)
	}
	var provResult map[string]any
	if err := json.Unmarshal(callResp.Result, &provResult); err != nil {
		t.Fatalf("unmarshal provider result: %v", err)
	}
	if provResult["tool"] != "search_graph" {
		t.Errorf("expected tool 'search_graph', got %v", provResult["tool"])
	}
	if provResult["source"] != "mock_provider" {
		t.Errorf("expected source 'mock_provider', got %v", provResult["source"])
	}

	// 4. Call skill tool.
	skillResp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 4, Method: "tools/call",
		Params: map[string]any{
			"name":      "blast_radius",
			"arguments": map[string]any{},
		},
	})
	if skillResp.Error != nil {
		t.Fatalf("tools/call skill error: %v", skillResp.Error)
	}
	var skillResult map[string]any
	if err := json.Unmarshal(skillResp.Result, &skillResult); err != nil {
		t.Fatalf("unmarshal skill result: %v", err)
	}
	if skillResult["source"] != "mock_skill" {
		t.Errorf("expected source 'mock_skill', got %v", skillResult["source"])
	}

	// 5. Call unknown tool → error.
	unknownResp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 5, Method: "tools/call",
		Params: map[string]any{
			"name":      "forgia_unknown_tool",
			"arguments": map[string]any{},
		},
	})
	if unknownResp.Error == nil {
		t.Error("expected error for unknown tool")
	}

	cancel()
	inW.Close()
}

// TestIntegration_Server_OversizedRequest verifies the server rejects
// requests exceeding the 10MB size limit with a parse error response.
func TestIntegration_Server_OversizedRequest(t *testing.T) {
	t.Parallel()

	// Build an oversized JSON payload (>10MB).
	bigStr := `"` + strings.Repeat("x", maxRequestSize+1) + `"`

	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	transport := NewStdioTransport(inR, outW)
	providers := NewProviderRegistry()
	server := NewMCPServer(transport, providers, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	// Write oversized payload then a valid request.
	// The server should reject the oversized one (parse error) and continue.
	go func() {
		inW.Write([]byte(bigStr + "\n"))
		// Follow with a valid request to prove server survived.
		enc := json.NewEncoder(inW)
		enc.Encode(jsonrpcRequest{JSONRPC: jsonrpcVersion, ID: 99, Method: "initialize"})
		// Don't close — let ctx cancel handle it.
	}()

	dec := json.NewDecoder(outR)

	// First response: parse error for oversized request.
	var errResp jsonrpcResponse
	if err := dec.Decode(&errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errResp.Error == nil {
		t.Fatal("expected parse error for oversized request")
	}
	if errResp.Error.Code != -32700 {
		t.Errorf("expected error code -32700, got %d", errResp.Error.Code)
	}

	// Second response: valid initialize response (server recovered).
	var okResp jsonrpcResponse
	if err := dec.Decode(&okResp); err != nil {
		t.Fatalf("decode ok response: %v", err)
	}
	if okResp.Error != nil {
		t.Fatalf("expected successful response after recovery, got error: %v", okResp.Error)
	}
	if okResp.ID == nil || *okResp.ID != 99 {
		t.Errorf("expected ID=99, got %v", okResp.ID)
	}

	cancel()
	inW.Close()
}

// TestIntegration_Server_SkillPriorityOverProvider verifies that when both
// a skill and a provider could handle a tool name, skills are tried first.
func TestIntegration_Server_SkillPriorityOverProvider(t *testing.T) {
	t.Parallel()

	providers := NewProviderRegistry()
	providers.Register(&mockProvider{
		name:    "code",
		healthy: true,
		callFn: func(_ context.Context, _ string, _ map[string]any) (any, error) {
			return map[string]any{"source": "provider"}, nil
		},
	})

	skills := &mockSkillDispatcher{
		execFn: func(_ context.Context, name string, _ map[string]any) (any, bool, error) {
			// Skills claim to handle this tool.
			if name == "forgia_code_search" {
				return map[string]any{"source": "skill"}, true, nil
			}
			return nil, false, nil
		},
	}

	server, inW, outR := newTestServer(providers, skills)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Serve(ctx)

	resp := serverTestHelper(t, server, inW, outR, jsonrpcRequest{
		JSONRPC: jsonrpcVersion, ID: 1, Method: "tools/call",
		Params: map[string]any{
			"name":      "forgia_code_search",
			"arguments": map[string]any{},
		},
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Skill should win over provider.
	if result["source"] != "skill" {
		t.Errorf("expected source 'skill' (priority), got %v", result["source"])
	}

	cancel()
	inW.Close()
}
