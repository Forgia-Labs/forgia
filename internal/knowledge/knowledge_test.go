package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// --- Available() tests ---

func TestAvailable_False(t *testing.T) {
	t.Parallel()
	c := &Client{available: false}
	if c.Available() {
		t.Error("expected Available() = false")
	}
}

func TestAvailable_True(t *testing.T) {
	t.Parallel()
	c := &Client{available: true}
	if !c.Available() {
		t.Error("expected Available() = true")
	}
}

// --- call() guard tests ---

func TestCall_Unavailable(t *testing.T) {
	t.Parallel()
	c := &Client{available: false}
	_, err := c.call(context.Background(), "cli", "list_projects")
	if !errors.Is(err, ErrUnavailable) {
		t.Errorf("expected ErrUnavailable, got: %v", err)
	}
}

// --- Stats JSON parsing tests ---

func TestParseListProjects_ValidJSON(t *testing.T) {
	t.Parallel()
	input := `[{"node_count": 1234, "edge_count": 5678, "last_indexed": "2026-03-28T10:00:00"}]`

	var projects []Stats
	if err := json.Unmarshal([]byte(input), &projects); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Symbols != 1234 {
		t.Errorf("Symbols = %d, want 1234", projects[0].Symbols)
	}
	if projects[0].Edges != 5678 {
		t.Errorf("Edges = %d, want 5678", projects[0].Edges)
	}
	if projects[0].LastIndexed != "2026-03-28T10:00:00" {
		t.Errorf("LastIndexed = %q, want '2026-03-28T10:00:00'", projects[0].LastIndexed)
	}
}

func TestParseListProjects_EmptyArray(t *testing.T) {
	t.Parallel()
	input := `[]`

	var projects []Stats
	if err := json.Unmarshal([]byte(input), &projects); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("expected 0 projects, got %d", len(projects))
	}
}

func TestParseListProjects_MissingFields(t *testing.T) {
	t.Parallel()
	input := `[{"node_count": 42}]`

	var projects []Stats
	if err := json.Unmarshal([]byte(input), &projects); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if projects[0].Symbols != 42 {
		t.Errorf("Symbols = %d, want 42", projects[0].Symbols)
	}
	if projects[0].Edges != 0 {
		t.Errorf("Edges = %d, want 0 (zero value)", projects[0].Edges)
	}
	if projects[0].LastIndexed != "" {
		t.Errorf("LastIndexed = %q, want empty", projects[0].LastIndexed)
	}
}

func TestParseIndexResponse_ValidJSON(t *testing.T) {
	t.Parallel()
	input := `{"symbols_indexed": 999}`

	var result struct {
		SymbolsIndexed int `json:"symbols_indexed"`
	}
	if err := json.Unmarshal([]byte(input), &result); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if result.SymbolsIndexed != 999 {
		t.Errorf("SymbolsIndexed = %d, want 999", result.SymbolsIndexed)
	}
}

// --- EnsureMCPJSON tests ---

func TestEnsureMCPJSON_CreateNew(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ctx := context.Background()

	if err := EnsureMCPJSON(ctx, dir); err != nil {
		t.Fatalf("EnsureMCPJSON() error: %v", err)
	}

	mcpPath := filepath.Join(dir, ".mcp.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}

	var doc mcpJSON
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}

	raw, exists := doc.MCPServers[binaryName]
	if !exists {
		t.Fatal("expected codebase-memory-mcp entry in .mcp.json")
	}

	var entry mcpServerEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		t.Fatalf("parse entry: %v", err)
	}
	if entry.Type != "stdio" {
		t.Errorf("Type = %q, want 'stdio'", entry.Type)
	}
	if entry.Command != binaryName {
		t.Errorf("Command = %q, want %q", entry.Command, binaryName)
	}
}

func TestEnsureMCPJSON_MergeExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ctx := context.Background()

	// Write an existing .mcp.json with another server.
	existing := `{
  "mcpServers": {
    "other-server": {
      "type": "stdio",
      "command": "other-tool"
    }
  }
}
`
	mcpPath := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(mcpPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write existing .mcp.json: %v", err)
	}

	if err := EnsureMCPJSON(ctx, dir); err != nil {
		t.Fatalf("EnsureMCPJSON() error: %v", err)
	}

	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}

	var doc mcpJSON
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}

	// Verify both servers exist.
	if _, exists := doc.MCPServers["other-server"]; !exists {
		t.Error("existing 'other-server' entry was lost during merge")
	}
	if _, exists := doc.MCPServers[binaryName]; !exists {
		t.Error("codebase-memory-mcp entry was not added during merge")
	}
}

func TestEnsureMCPJSON_AlreadyPresent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ctx := context.Background()

	// Write .mcp.json that already has codebase-memory-mcp.
	existing := `{
  "mcpServers": {
    "codebase-memory-mcp": {
      "type": "stdio",
      "command": "codebase-memory-mcp",
      "custom_field": "preserve-me"
    }
  }
}
`
	mcpPath := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(mcpPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write existing .mcp.json: %v", err)
	}

	if err := EnsureMCPJSON(ctx, dir); err != nil {
		t.Fatalf("EnsureMCPJSON() error: %v", err)
	}

	// Verify file was NOT overwritten (custom_field preserved).
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}

	var doc mcpJSON
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}

	raw := doc.MCPServers[binaryName]
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse entry: %v", err)
	}
	if m["custom_field"] != "preserve-me" {
		t.Error("existing codebase-memory-mcp entry was overwritten — should be preserved")
	}
}

func TestEnsureMCPJSON_InvalidExistingJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ctx := context.Background()

	mcpPath := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(mcpPath, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("write broken .mcp.json: %v", err)
	}

	err := EnsureMCPJSON(ctx, dir)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// --- NewClient with manipulated PATH ---

func TestNewClient_BinaryNotInPATH(t *testing.T) {
	// t.Setenv cannot be used with t.Parallel(), so this test runs sequentially.
	t.Setenv("PATH", t.TempDir())

	c := NewClient()
	if c.Available() {
		t.Error("expected Available() = false when binary not in PATH")
	}
}
