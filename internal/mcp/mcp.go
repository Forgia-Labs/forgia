// Package mcp implements the Model Context Protocol integration.
// Forgia acts as both MCP server (serving Claude) and MCP client (chaining to providers).
package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// ToolProvider abstracts any external tool source (MCP server, API, in-process).
type ToolProvider interface {
	// Name returns the provider identifier.
	Name() string

	// Tools returns the tool definitions this provider exposes.
	Tools() []ToolDefinition

	// Call invokes a tool and returns the result.
	Call(ctx context.Context, tool string, params map[string]any) (any, error)

	// Start initializes the provider (spawn subprocess, connect API).
	Start(ctx context.Context) error

	// Stop gracefully shuts down the provider.
	Stop() error

	// Healthy returns true if the provider is ready.
	Healthy() bool
}

// ToolDefinition describes a tool exposed by a provider.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Namespace   string         `json:"-"` // prefix when re-exposed (e.g., "forgia_code_")
}

// ProviderRegistry manages multiple ToolProviders with namespace routing.
// Thread-safe — MCP server handles concurrent tool calls.
type ProviderRegistry struct {
	providers map[string]ToolProvider
	mu        sync.RWMutex
}

// NewProviderRegistry creates an empty registry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]ToolProvider),
	}
}

// Register adds a provider to the registry.
func (r *ProviderRegistry) Register(p ToolProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// AllTools returns merged tool list from all providers with namespace prefix.
func (r *ProviderRegistry) AllTools() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []ToolDefinition
	for _, p := range r.providers {
		if !p.Healthy() {
			continue
		}
		for _, t := range p.Tools() {
			t.Name = ToolName(t.Namespace, t.Name)
			tools = append(tools, t)
		}
	}
	return tools
}

// StopAll stops all registered providers. Called during graceful shutdown.
// Logs errors but does not fail — best-effort cleanup.
func (r *ProviderRegistry) StopAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, p := range r.providers {
		if err := p.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, "provider %s stop error: %v\n", name, err)
		}
	}
}

// Call routes a namespaced tool call to the right provider.
// Parses "forgia_code_search_graph" → provider="code", tool="search_graph".
func (r *ProviderRegistry) Call(ctx context.Context, tool string, params map[string]any) (any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	namespace, toolName, err := parseNamespacedTool(tool)
	if err != nil {
		return nil, err
	}

	p, ok := r.providers[namespace]
	if !ok {
		return nil, fmt.Errorf("unknown provider namespace: %s", namespace)
	}

	return p.Call(ctx, toolName, params)
}

// ToolName builds a namespaced tool name: ToolName("code", "search_graph") → "forgia_code_search_graph".
// Single source of truth for the naming convention — used by ProviderRegistry and CompositeSkill.
func ToolName(namespace, tool string) string {
	return "forgia_" + namespace + "_" + tool
}

// parseNamespacedTool splits "forgia_code_search_graph" → ("code", "search_graph").
func parseNamespacedTool(tool string) (namespace, name string, err error) {
	const prefix = "forgia_"
	if !strings.HasPrefix(tool, prefix) {
		return "", "", fmt.Errorf("tool %q missing forgia_ prefix", tool)
	}

	rest := strings.TrimPrefix(tool, prefix)
	idx := strings.Index(rest, "_")
	if idx < 0 {
		return "", "", fmt.Errorf("tool %q missing namespace separator", tool)
	}

	return rest[:idx], rest[idx+1:], nil
}

// MetadataProvider is a lightweight ToolProvider that exists only for
// metadata purposes (listing composite skills without starting real processes).
// Used by `forgia skills` to register composite skills without launching providers.
type MetadataProvider struct {
	name      string
	namespace string
}

// Compile-time interface check.
var _ ToolProvider = (*MetadataProvider)(nil)

// NewMetadataProvider creates a provider that only has a name and namespace.
func NewMetadataProvider(name, namespace string) *MetadataProvider {
	return &MetadataProvider{name: name, namespace: namespace}
}

func (p *MetadataProvider) Name() string                                                   { return p.name }
func (p *MetadataProvider) Tools() []ToolDefinition                                        { return nil }
func (p *MetadataProvider) Call(_ context.Context, _ string, _ map[string]any) (any, error) { return nil, fmt.Errorf("metadata-only provider") }
func (p *MetadataProvider) Start(_ context.Context) error                                  { return nil }
func (p *MetadataProvider) Stop() error                                                    { return nil }
func (p *MetadataProvider) Healthy() bool                                                  { return true }
