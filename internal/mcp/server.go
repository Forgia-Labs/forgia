package mcp

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
)

// SkillDispatcher provides skill tool listing and execution for the MCP server.
// Defined in the mcp package to avoid an import cycle (skill imports mcp).
type SkillDispatcher interface {
	// MCPToolDefs returns tool definitions for all MCP-capable skills.
	MCPToolDefs() []ToolDefinition
	// GetAndExecute looks up a skill by name and executes it.
	// Returns (result, true, nil) if found and executed, (nil, false, nil) if not found.
	GetAndExecute(ctx context.Context, name string, params map[string]any) (any, bool, error)
}

// MCPServer reads JSON-RPC requests from stdin, dispatches to ProviderRegistry
// and SkillDispatcher, and writes responses to stdout.
type MCPServer struct {
	transport *StdioTransport
	providers *ProviderRegistry
	skills    SkillDispatcher
	logger    *slog.Logger
}

// NewMCPServer creates a server that dispatches to the given registries.
func NewMCPServer(transport *StdioTransport, providers *ProviderRegistry, skills SkillDispatcher) *MCPServer {
	return &MCPServer{
		transport: transport,
		providers: providers,
		skills:    skills,
		logger:    slog.With("component", "mcp-server"),
	}
}

// Serve blocks until ctx is cancelled or stdin reaches EOF.
// It reads JSON-RPC requests and dispatches them to the appropriate handler.
func (s *MCPServer) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := s.transport.ReadRequest()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			// NOTE: json.Decoder does not recover after a genuine parse error (truncated
			// or malformed JSON mid-stream). The continue here is best-effort — subsequent
			// reads on the same stream will likely also fail and exit via io.EOF.
			s.writeError(nil, -32700, "Parse error")
			continue
		}

		s.handleRequest(ctx, req)
	}
}

// handleRequest dispatches a single JSON-RPC request to the appropriate handler.
func (s *MCPServer) handleRequest(ctx context.Context, req *jsonrpcRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "notifications/initialized":
		// Acknowledge, no response required.
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolsCall(ctx, req)
	default:
		s.writeError(&req.ID, -32601, "Method not found")
	}
}

// handleInitialize returns server info and capabilities.
func (s *MCPServer) handleInitialize(req *jsonrpcRequest) {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo": map[string]any{
			"name":    "forgia",
			"version": "0.1.0",
		},
	}
	s.writeResult(req.ID, result)
}

// handleToolsList returns merged tools from providers and skills.
func (s *MCPServer) handleToolsList(req *jsonrpcRequest) {
	var tools []map[string]any

	// Provider tools.
	for _, t := range s.providers.AllTools() {
		tool := map[string]any{
			"name":        t.Name,
			"description": t.Description,
		}
		if t.Parameters != nil {
			tool["inputSchema"] = t.Parameters
		}
		tools = append(tools, tool)
	}

	// Skill tools.
	if s.skills != nil {
		for _, t := range s.skills.MCPToolDefs() {
			tool := map[string]any{
				"name":        t.Name,
				"description": t.Description,
			}
			if t.Parameters != nil {
				tool["inputSchema"] = t.Parameters
			}
			tools = append(tools, tool)
		}
	}

	if tools == nil {
		tools = []map[string]any{}
	}

	s.writeResult(req.ID, map[string]any{"tools": tools})
}

// handleToolsCall dispatches a tool call to the appropriate provider or skill.
func (s *MCPServer) handleToolsCall(ctx context.Context, req *jsonrpcRequest) {
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		s.writeError(&req.ID, -32602, "Invalid params")
		return
	}

	var callParams struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(paramsBytes, &callParams); err != nil {
		s.writeError(&req.ID, -32602, "Invalid params")
		return
	}

	// Log tool invocation: tool name + timestamp, NOT parameters (security requirement).
	s.logger.InfoContext(ctx, "tools/call", "tool", callParams.Name)

	// Try skill dispatcher first.
	if s.skills != nil {
		result, found, err := s.skills.GetAndExecute(ctx, callParams.Name, callParams.Arguments)
		if err != nil {
			s.writeError(&req.ID, -32603, err.Error())
			return
		}
		if found {
			s.writeResult(req.ID, result)
			return
		}
	}

	// Try provider registry.
	result, err := s.providers.Call(ctx, callParams.Name, callParams.Arguments)
	if err != nil {
		s.writeError(&req.ID, -32602, err.Error())
		return
	}
	s.writeResult(req.ID, result)
}

// writeResult sends a successful JSON-RPC response.
func (s *MCPServer) writeResult(id int, result any) {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		s.writeError(&id, -32603, "failed to marshal result")
		return
	}
	resp := jsonrpcResponse{
		JSONRPC: jsonrpcVersion,
		ID:      &id,
		Result:  json.RawMessage(resultBytes),
	}
	if err := s.transport.WriteResponse(resp); err != nil {
		s.logger.Error("failed to write response", "error", err)
	}
}

// writeError sends a JSON-RPC error response.
func (s *MCPServer) writeError(id *int, code int, message string) {
	resp := jsonrpcResponse{
		JSONRPC: jsonrpcVersion,
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
	if err := s.transport.WriteResponse(resp); err != nil {
		s.logger.Error("failed to write error response", "error", err)
	}
}
