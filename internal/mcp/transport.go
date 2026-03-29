package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// maxRequestSize is the maximum allowed size for a single JSON-RPC request (10MB).
// Requests exceeding this limit are rejected to prevent memory exhaustion.
const maxRequestSize = 10 * 1024 * 1024

// StdioTransport handles JSON-RPC 2.0 read/write over stdio pipes.
// Used by both the MCP client (subprocess.go) and MCP server (server.go).
type StdioTransport struct {
	dec *json.Decoder // reads from io.Reader (stdin or subprocess stdout)
	enc *json.Encoder // writes to io.Writer (stdout or subprocess stdin)
	mu  sync.Mutex    // protects enc
}

// NewStdioTransport creates a transport that reads from r and writes to w.
func NewStdioTransport(r io.Reader, w io.Writer) *StdioTransport {
	return &StdioTransport{
		dec: json.NewDecoder(r),
		enc: json.NewEncoder(w),
	}
}

// ReadRequest reads a JSON-RPC request from the input stream.
// Rejects requests larger than maxRequestSize (10MB) to prevent memory exhaustion.
func (t *StdioTransport) ReadRequest() (*jsonrpcRequest, error) {
	raw, err := t.readLimited()
	if err != nil {
		return nil, err
	}
	var req jsonrpcRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("parse request: %w", err)
	}
	return &req, nil
}

// WriteResponse writes a JSON-RPC response to the output stream.
func (t *StdioTransport) WriteResponse(resp jsonrpcResponse) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.enc.Encode(resp)
}

// ReadResponse reads a JSON-RPC response from the input stream.
// Used by the client side (subprocess.go).
func (t *StdioTransport) ReadResponse() (*jsonrpcResponse, error) {
	var resp jsonrpcResponse
	if err := t.dec.Decode(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// WriteRequest writes a JSON-RPC request to the output stream.
// Used by the client side (subprocess.go).
func (t *StdioTransport) WriteRequest(req jsonrpcRequest) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.enc.Encode(req)
}

// WriteNotification writes a JSON-RPC notification to the output stream.
func (t *StdioTransport) WriteNotification(notif jsonrpcNotification) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.enc.Encode(notif)
}

// readLimited reads a single JSON token from the decoder, enforcing the size limit.
func (t *StdioTransport) readLimited() (json.RawMessage, error) {
	var raw json.RawMessage
	if err := t.dec.Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw) > maxRequestSize {
		return nil, fmt.Errorf("request too large: %d bytes (limit %d)", len(raw), maxRequestSize)
	}
	return raw, nil
}
