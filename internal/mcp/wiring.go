package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/forgia-labs/forgia/internal/config"
)

// shellMetachars contains characters that indicate shell injection attempts.
// If any of these appear in a provider command, it is rejected.
const shellMetachars = ";|&$`"

// ValidateCommand checks a provider command for shell metacharacters and resolves
// it to an absolute path via exec.LookPath.
// Returns the resolved absolute path or an error.
func ValidateCommand(command string) (string, error) {
	if command == "" {
		return "", fmt.Errorf("empty command")
	}

	for _, ch := range shellMetachars {
		if strings.ContainsRune(command, ch) {
			return "", fmt.Errorf("command %q contains shell metacharacter %q", command, string(ch))
		}
	}

	resolved, err := exec.LookPath(command)
	if err != nil {
		return "", fmt.Errorf("command %q not found in PATH: %w", command, err)
	}

	return resolved, nil
}

// WireProviders reads MCP provider configuration, validates commands, creates
// MCPSubprocessProvider instances, registers them in a ProviderRegistry, and
// starts all providers. If a provider fails to start, it is logged and skipped.
//
// When cfg.Knowledge.Provider is set, it auto-registers a codebase-memory-mcp
// provider under the "code" namespace.
func WireProviders(ctx context.Context, cfg *config.Config) (*ProviderRegistry, error) {
	registry := NewProviderRegistry()
	logger := slog.With("component", "provider-wiring")

	// Collect all provider configs: explicit + auto-registered from knowledge.
	providers := make(map[string]config.MCPProviderConfig)
	for name, pcfg := range cfg.MCP.Providers {
		providers[name] = pcfg
	}

	// Auto-register codebase-memory-mcp from [knowledge] config.
	if cfg.Knowledge.Provider != "" {
		if _, exists := providers["code"]; !exists {
			providers["code"] = config.MCPProviderConfig{
				Command:   cfg.Knowledge.Provider,
				Namespace: "code",
			}
			logger.InfoContext(ctx, "auto-registering knowledge provider", "provider", cfg.Knowledge.Provider, "namespace", "code")
		}
	}

	for name, pcfg := range providers {
		// Ensure namespace is set.
		if pcfg.Namespace == "" {
			pcfg.Namespace = name
		}

		// Validate and resolve command path.
		resolved, err := ValidateCommand(pcfg.Command)
		if err != nil {
			logger.WarnContext(ctx, "skipping provider: command validation failed",
				"provider", name, "command", pcfg.Command, "error", err)
			continue
		}

		logger.InfoContext(ctx, "resolved provider command",
			"provider", name, "command", pcfg.Command, "resolved", resolved)

		// Use the resolved absolute path.
		pcfg.Command = resolved

		provider := NewMCPSubprocessProvider(pcfg)
		registry.Register(provider)

		if err := provider.Start(ctx); err != nil {
			logger.ErrorContext(ctx, "provider failed to start, continuing with others",
				"provider", name, "error", err)
			continue
		}
	}

	return registry, nil
}
