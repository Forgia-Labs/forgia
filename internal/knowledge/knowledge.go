// Package knowledge provides a resilient client for codebase-memory-mcp.
// The knowledge layer is ALWAYS optional — Forgia works 100% without it.
// When available, it provides codebase indexing and semantic graph stats.
package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	// ErrUnavailable means codebase-memory-mcp is not installed.
	ErrUnavailable = errors.New("codebase-memory-mcp unavailable")

	// ErrTimeout means a codebase-memory-mcp command timed out.
	ErrTimeout = errors.New("codebase-memory-mcp call timed out")
)

const binaryName = "codebase-memory-mcp"

// Stats holds knowledge graph statistics from list_projects.
type Stats struct {
	Symbols     int    `json:"node_count"`
	Edges       int    `json:"edge_count"`
	LastIndexed string `json:"last_indexed"`
}

// Client wraps codebase-memory-mcp CLI with timeout and graceful fallback.
type Client struct {
	available bool
	timeout   time.Duration
}

// NewClient creates a knowledge client, checking availability.
func NewClient() *Client {
	c := &Client{timeout: 30 * time.Second}
	if _, err := exec.LookPath(binaryName); err != nil {
		c.available = false
		return c
	}
	c.available = true
	return c
}

// Available returns true if codebase-memory-mcp is in PATH.
func (c *Client) Available() bool {
	return c.available
}

// call executes a codebase-memory-mcp CLI command with timeout.
func (c *Client) call(ctx context.Context, args ...string) (string, error) {
	if !c.available {
		return "", ErrUnavailable
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			slog.Warn("codebase-memory-mcp call timed out", "args", args)
			return "", ErrTimeout
		}
		return "", fmt.Errorf("%s %v: %w", binaryName, args, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Index triggers repository indexing and returns the symbol count.
func (c *Client) Index(ctx context.Context) (int, error) {
	output, err := c.call(ctx, "cli", "index_repository", "path=./")
	if err != nil {
		return 0, fmt.Errorf("index repository: %w", err)
	}

	var result struct {
		SymbolsIndexed int `json:"symbols_indexed"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return 0, fmt.Errorf("parse index response: %w", err)
	}

	return result.SymbolsIndexed, nil
}

// GetStats retrieves knowledge graph statistics from list_projects.
func (c *Client) GetStats(ctx context.Context) (*Stats, error) {
	output, err := c.call(ctx, "cli", "list_projects")
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	var projects []Stats
	if err := json.Unmarshal([]byte(output), &projects); err != nil {
		return nil, fmt.Errorf("parse list_projects response: %w", err)
	}

	if len(projects) == 0 {
		return &Stats{}, nil
	}

	return &projects[0], nil
}

// mcpServerEntry is the structure for a single MCP server in .mcp.json.
type mcpServerEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// mcpJSON is the top-level structure of .mcp.json.
type mcpJSON struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
}

// EnsureMCPJSON creates or merges .mcp.json with codebase-memory-mcp server config.
// If the file does not exist, it creates it. If it exists, it merges without overwriting.
func EnsureMCPJSON(_ context.Context, dir string) error {
	mcpPath := filepath.Join(dir, ".mcp.json")

	entry := mcpServerEntry{
		Type:    "stdio",
		Command: binaryName,
	}
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal mcp entry: %w", err)
	}

	data, err := os.ReadFile(mcpPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read %s: %w", mcpPath, err)
		}
		// File does not exist — create fresh.
		doc := mcpJSON{
			MCPServers: map[string]json.RawMessage{
				binaryName: entryBytes,
			},
		}
		return writeMCPJSON(mcpPath, doc)
	}

	// File exists — parse and merge.
	var doc mcpJSON
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parse %s: %w", mcpPath, err)
	}

	if doc.MCPServers == nil {
		doc.MCPServers = make(map[string]json.RawMessage)
	}

	// Don't overwrite if already present.
	if _, exists := doc.MCPServers[binaryName]; exists {
		slog.Info(".mcp.json: codebase-memory-mcp already configured")
		return nil
	}

	doc.MCPServers[binaryName] = entryBytes
	return writeMCPJSON(mcpPath, doc)
}

// writeMCPJSON writes the mcpJSON structure to disk with indentation.
func writeMCPJSON(path string, doc mcpJSON) error {
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal .mcp.json: %w", err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
