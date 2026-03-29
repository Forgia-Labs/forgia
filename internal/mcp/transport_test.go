package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestStdioTransport_WriteAndReadRequest(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	writer := NewStdioTransport(nil, &buf)
	req := jsonrpcRequest{JSONRPC: jsonrpcVersion, ID: 1, Method: "initialize"}

	if err := writer.WriteRequest(req); err != nil {
		t.Fatalf("WriteRequest: %v", err)
	}

	reader := NewStdioTransport(&buf, nil)
	got, err := reader.ReadRequest()
	if err != nil {
		t.Fatalf("ReadRequest: %v", err)
	}

	if got.ID != 1 || got.Method != "initialize" || got.JSONRPC != jsonrpcVersion {
		t.Errorf("got %+v, want ID=1 Method=initialize", got)
	}
}

func TestStdioTransport_WriteAndReadResponse(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	id := 42
	result := json.RawMessage(`{"ok":true}`)
	resp := jsonrpcResponse{JSONRPC: jsonrpcVersion, ID: &id, Result: result}

	writer := NewStdioTransport(nil, &buf)
	if err := writer.WriteResponse(resp); err != nil {
		t.Fatalf("WriteResponse: %v", err)
	}

	reader := NewStdioTransport(&buf, nil)
	got, err := reader.ReadResponse()
	if err != nil {
		t.Fatalf("ReadResponse: %v", err)
	}

	if got.ID == nil || *got.ID != 42 {
		t.Errorf("expected ID=42, got %v", got.ID)
	}
	if string(got.Result) != `{"ok":true}` {
		t.Errorf("expected result {\"ok\":true}, got %s", got.Result)
	}
}

func TestStdioTransport_ReadRequestRejectsOversized(t *testing.T) {
	t.Parallel()

	// Create a JSON payload larger than 10MB.
	// A JSON string of 11MB should exceed the limit.
	bigStr := `"` + strings.Repeat("x", maxRequestSize+1) + `"`
	reader := NewStdioTransport(strings.NewReader(bigStr), nil)

	_, err := reader.ReadRequest()
	if err == nil {
		t.Fatal("expected error for oversized request")
	}
	if !strings.Contains(err.Error(), "request too large") {
		t.Errorf("expected 'request too large' error, got: %v", err)
	}
}

func TestStdioTransport_ReadRequestAcceptsNormalSize(t *testing.T) {
	t.Parallel()

	req := jsonrpcRequest{JSONRPC: jsonrpcVersion, ID: 1, Method: "tools/list"}
	data, _ := json.Marshal(req)

	reader := NewStdioTransport(bytes.NewReader(data), nil)
	got, err := reader.ReadRequest()
	if err != nil {
		t.Fatalf("ReadRequest: %v", err)
	}
	if got.Method != "tools/list" {
		t.Errorf("expected method tools/list, got %s", got.Method)
	}
}

func TestStdioTransport_ReadRequestEOF(t *testing.T) {
	t.Parallel()

	reader := NewStdioTransport(strings.NewReader(""), nil)
	_, err := reader.ReadRequest()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestStdioTransport_WriteNotification(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	transport := NewStdioTransport(nil, &buf)

	notif := jsonrpcNotification{
		JSONRPC: jsonrpcVersion,
		Method:  "notifications/initialized",
	}
	if err := transport.WriteNotification(notif); err != nil {
		t.Fatalf("WriteNotification: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["method"] != "notifications/initialized" {
		t.Errorf("expected method notifications/initialized, got %v", got["method"])
	}
}
