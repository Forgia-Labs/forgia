package mcp

import (
	"context"
	"os/exec"
	"runtime"
	"testing"

	"github.com/forgia-labs/forgia/internal/config"
)

func TestValidateCommand_RejectsShellMetachars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		command string
		char    string
	}{
		{"echo; rm -rf /", ";"},
		{"cmd | cat", "|"},
		{"cmd & bg", "&"},
		{"echo $HOME", "$"},
		{"cmd `whoami`", "`"},
		{";leading", ";"},
		{"trailing|", "|"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			t.Parallel()
			_, err := ValidateCommand(tt.command)
			if err == nil {
				t.Errorf("ValidateCommand(%q) should reject command containing %q", tt.command, tt.char)
			}
		})
	}
}

func TestValidateCommand_RejectsEmptyCommand(t *testing.T) {
	t.Parallel()

	_, err := ValidateCommand("")
	if err == nil {
		t.Error("ValidateCommand should reject empty command")
	}
}

func TestValidateCommand_RejectsMissingCommand(t *testing.T) {
	t.Parallel()

	_, err := ValidateCommand("nonexistent-binary-xyz-12345")
	if err == nil {
		t.Error("ValidateCommand should reject command not found in PATH")
	}
}

func TestValidateCommand_ResolvesValidCommand(t *testing.T) {
	t.Parallel()

	// "echo" is available on all platforms.
	cmd := "echo"
	if runtime.GOOS == "windows" {
		cmd = "cmd"
	}

	resolved, err := ValidateCommand(cmd)
	if err != nil {
		t.Fatalf("ValidateCommand(%q) unexpected error: %v", cmd, err)
	}

	// Verify the resolved path actually exists via LookPath.
	expected, _ := exec.LookPath(cmd)
	if resolved != expected {
		t.Errorf("ValidateCommand(%q) = %q, want %q", cmd, resolved, expected)
	}
}

func TestWireProviders_WithMockConfig(t *testing.T) {
	t.Parallel()

	// Use real binaries that exist on every system for testing.
	echoBin := "echo"
	if runtime.GOOS == "windows" {
		echoBin = "cmd"
	}

	cfg := &config.Config{
		MCP: config.MCPConfig{
			Providers: map[string]config.MCPProviderConfig{
				"alpha": {
					Command:   echoBin,
					Namespace: "alpha",
					Lazy:      true, // lazy=true so Start won't spawn
				},
				"beta": {
					Command:   echoBin,
					Namespace: "beta",
					Lazy:      true,
				},
			},
		},
	}

	registry, err := WireProviders(context.Background(), cfg)
	if err != nil {
		t.Fatalf("WireProviders unexpected error: %v", err)
	}

	// Both providers should be registered.
	registry.mu.RLock()
	count := len(registry.providers)
	_, hasAlpha := registry.providers["alpha"]
	_, hasBeta := registry.providers["beta"]
	registry.mu.RUnlock()

	if count != 2 {
		t.Errorf("expected 2 providers registered, got %d", count)
	}
	if !hasAlpha {
		t.Error("provider 'alpha' not registered")
	}
	if !hasBeta {
		t.Error("provider 'beta' not registered")
	}
}

func TestWireProviders_MissingCommandSkipped(t *testing.T) {
	t.Parallel()

	echoBin := "echo"
	if runtime.GOOS == "windows" {
		echoBin = "cmd"
	}

	cfg := &config.Config{
		MCP: config.MCPConfig{
			Providers: map[string]config.MCPProviderConfig{
				"good": {
					Command:   echoBin,
					Namespace: "good",
					Lazy:      true,
				},
				"bad": {
					Command:   "nonexistent-binary-xyz-12345",
					Namespace: "bad",
					Lazy:      true,
				},
			},
		},
	}

	registry, err := WireProviders(context.Background(), cfg)
	if err != nil {
		t.Fatalf("WireProviders should not return error when a provider is skipped: %v", err)
	}

	registry.mu.RLock()
	count := len(registry.providers)
	_, hasGood := registry.providers["good"]
	_, hasBad := registry.providers["bad"]
	registry.mu.RUnlock()

	if count != 1 {
		t.Errorf("expected 1 provider (bad one skipped), got %d", count)
	}
	if !hasGood {
		t.Error("provider 'good' should be registered")
	}
	if hasBad {
		t.Error("provider 'bad' should NOT be registered (missing command)")
	}
}

func TestWireProviders_ShellMetacharCommandSkipped(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		MCP: config.MCPConfig{
			Providers: map[string]config.MCPProviderConfig{
				"malicious": {
					Command:   "echo; rm -rf /",
					Namespace: "malicious",
					Lazy:      true,
				},
			},
		},
	}

	registry, err := WireProviders(context.Background(), cfg)
	if err != nil {
		t.Fatalf("WireProviders should not return error when a provider is skipped: %v", err)
	}

	registry.mu.RLock()
	count := len(registry.providers)
	registry.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 providers (malicious one skipped), got %d", count)
	}
}

func TestWireProviders_KnowledgeAutoRegisters(t *testing.T) {
	t.Parallel()

	echoBin := "echo"
	if runtime.GOOS == "windows" {
		echoBin = "cmd"
	}

	cfg := &config.Config{
		Knowledge: config.KnowledgeConfig{
			Provider: echoBin,
		},
	}

	registry, err := WireProviders(context.Background(), cfg)
	if err != nil {
		t.Fatalf("WireProviders unexpected error: %v", err)
	}

	registry.mu.RLock()
	_, hasCode := registry.providers["code"]
	registry.mu.RUnlock()

	if !hasCode {
		t.Error("knowledge provider should be auto-registered under 'code' namespace")
	}
}

func TestWireProviders_KnowledgeDoesNotOverrideExplicit(t *testing.T) {
	t.Parallel()

	echoBin := "echo"
	if runtime.GOOS == "windows" {
		echoBin = "cmd"
	}

	cfg := &config.Config{
		MCP: config.MCPConfig{
			Providers: map[string]config.MCPProviderConfig{
				"code": {
					Command:   echoBin,
					Namespace: "code",
					Args:      []string{"explicit"},
					Lazy:      true,
				},
			},
		},
		Knowledge: config.KnowledgeConfig{
			Provider: echoBin,
		},
	}

	registry, err := WireProviders(context.Background(), cfg)
	if err != nil {
		t.Fatalf("WireProviders unexpected error: %v", err)
	}

	registry.mu.RLock()
	count := len(registry.providers)
	registry.mu.RUnlock()

	// Should only have 1 provider — the explicit one, not doubled.
	if count != 1 {
		t.Errorf("expected 1 provider (explicit 'code'), got %d", count)
	}
}

func TestWireProviders_EmptyConfig(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	registry, err := WireProviders(context.Background(), cfg)
	if err != nil {
		t.Fatalf("WireProviders unexpected error: %v", err)
	}

	registry.mu.RLock()
	count := len(registry.providers)
	registry.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 providers from empty config, got %d", count)
	}
}

func TestWireProviders_NamespaceDefaultsToMapKey(t *testing.T) {
	t.Parallel()

	echoBin := "echo"
	if runtime.GOOS == "windows" {
		echoBin = "cmd"
	}

	cfg := &config.Config{
		MCP: config.MCPConfig{
			Providers: map[string]config.MCPProviderConfig{
				"myns": {
					Command: echoBin,
					// Namespace intentionally empty — should default to "myns".
					Lazy: true,
				},
			},
		},
	}

	registry, err := WireProviders(context.Background(), cfg)
	if err != nil {
		t.Fatalf("WireProviders unexpected error: %v", err)
	}

	registry.mu.RLock()
	_, hasMyNS := registry.providers["myns"]
	registry.mu.RUnlock()

	if !hasMyNS {
		t.Error("provider should be registered under map key 'myns' when namespace is empty")
	}
}
