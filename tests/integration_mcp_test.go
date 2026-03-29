package e2e

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// --- Server E2E subprocess tests (SDD-007 suite 1) ---

// mcpServer spawns `forgia mcp serve` as subprocess and returns
// JSON encoder (stdin), decoder (stdout), and cleanup function.
func mcpServer(t *testing.T) (*json.Encoder, *json.Decoder, func()) {
	t.Helper()

	dir := t.TempDir()
	cmd := exec.Command(binaryPath, "mcp", "serve")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+dir)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}

	// Discard stderr (server logs go there).
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		t.Fatalf("start mcp serve: %v", err)
	}

	enc := json.NewEncoder(stdin)
	dec := json.NewDecoder(stdout)

	cleanup := func() {
		stdin.Close()
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			<-done
		}
	}

	return enc, dec, cleanup
}

// mcpRequest sends a JSON-RPC request and reads the response with a timeout.
func mcpRequest(t *testing.T, enc *json.Encoder, dec *json.Decoder, id int, method string, params any) map[string]any {
	t.Helper()

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(id),
		"method":  method,
	}
	if params != nil {
		req["params"] = params
	}

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode request: %v", err)
	}

	type result struct {
		resp map[string]any
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		var resp map[string]any
		err := dec.Decode(&resp)
		ch <- result{resp, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("decode response: %v", r.err)
		}
		return r.resp
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for MCP response")
		return nil
	}
}

func TestIntegration_ServerE2E_Initialize(t *testing.T) {
	t.Parallel()

	enc, dec, cleanup := mcpServer(t)
	defer cleanup()

	resp := mcpRequest(t, enc, dec, 1, "initialize", nil)

	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result in response: %v", resp)
	}

	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("missing serverInfo in result")
	}
	if serverInfo["name"] != "forgia" {
		t.Errorf("expected server name 'forgia', got %v", serverInfo["name"])
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocolVersion '2024-11-05', got %v", result["protocolVersion"])
	}
	if _, ok := result["capabilities"]; !ok {
		t.Error("missing capabilities in result")
	}
}

func TestIntegration_ServerE2E_ToolsList(t *testing.T) {
	t.Parallel()

	enc, dec, cleanup := mcpServer(t)
	defer cleanup()

	resp := mcpRequest(t, enc, dec, 1, "tools/list", nil)

	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result: %v", resp)
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("missing tools array in result")
	}

	// No providers configured → empty tools list.
	if len(tools) != 0 {
		t.Errorf("expected 0 tools (no providers configured), got %d", len(tools))
	}
}

func TestIntegration_ServerE2E_ToolsCallUnknown(t *testing.T) {
	t.Parallel()

	enc, dec, cleanup := mcpServer(t)
	defer cleanup()

	params := map[string]any{
		"name":      "forgia_unknown_nonexistent",
		"arguments": map[string]any{},
	}
	resp := mcpRequest(t, enc, dec, 1, "tools/call", params)

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error response for unknown tool, got: %v", resp)
	}

	msg, _ := errObj["message"].(string)
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestIntegration_ServerE2E_UnknownMethod(t *testing.T) {
	t.Parallel()

	enc, dec, cleanup := mcpServer(t)
	defer cleanup()

	resp := mcpRequest(t, enc, dec, 1, "nonexistent/method", nil)

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error for unknown method, got: %v", resp)
	}

	code, _ := errObj["code"].(float64)
	if code != -32601 {
		t.Errorf("expected error code -32601 (method not found), got %v", code)
	}
}

func TestIntegration_ServerE2E_MultipleRequests(t *testing.T) {
	t.Parallel()

	enc, dec, cleanup := mcpServer(t)
	defer cleanup()

	// Send initialize.
	resp1 := mcpRequest(t, enc, dec, 1, "initialize", nil)
	if resp1["error"] != nil {
		t.Fatalf("initialize failed: %v", resp1["error"])
	}

	// Send tools/list.
	resp2 := mcpRequest(t, enc, dec, 2, "tools/list", nil)
	if resp2["error"] != nil {
		t.Fatalf("tools/list failed: %v", resp2["error"])
	}

	// Send unknown tool call.
	params := map[string]any{
		"name":      "forgia_fake_tool",
		"arguments": map[string]any{},
	}
	resp3 := mcpRequest(t, enc, dec, 3, "tools/call", params)
	if resp3["error"] == nil {
		t.Error("expected error for unknown tool")
	}

	// Verify IDs match.
	if id, _ := resp1["id"].(float64); id != 1 {
		t.Errorf("expected ID=1, got %v", id)
	}
	if id, _ := resp2["id"].(float64); id != 2 {
		t.Errorf("expected ID=2, got %v", id)
	}
	if id, _ := resp3["id"].(float64); id != 3 {
		t.Errorf("expected ID=3, got %v", id)
	}
}

func TestIntegration_ServerE2E_OversizedRequest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cmd := exec.Command(binaryPath, "mcp", "serve")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+dir)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}

	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		t.Fatalf("start mcp serve: %v", err)
	}
	defer func() {
		stdin.Close()
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			<-done
		}
	}()

	// Build a JSON-RPC request with a payload >10MB.
	// Use a very large "params" string value.
	bigValue := strings.Repeat("x", 11*1024*1024) // 11MB string
	oversized := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"data":"` + bigValue + `"}}` + "\n"

	go func() {
		io.WriteString(stdin, oversized)
	}()

	dec := json.NewDecoder(stdout)
	type result struct {
		resp map[string]any
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		var resp map[string]any
		err := dec.Decode(&resp)
		ch <- result{resp, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			// Server may close connection on oversized request — acceptable.
			return
		}
		// If we get a response, it should be an error.
		if r.resp["error"] == nil {
			t.Error("expected error response for oversized request")
		}
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for oversized request response")
	}
}

func TestIntegration_ServerE2E_StdinEOF(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cmd := exec.Command(binaryPath, "mcp", "serve")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+dir)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		t.Fatalf("start mcp serve: %v", err)
	}

	// Close stdin immediately → server should exit cleanly.
	stdin.Close()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected clean exit on stdin EOF, got: %v", err)
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		t.Fatal("server did not exit on stdin EOF")
	}
}
