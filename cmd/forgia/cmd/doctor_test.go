package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckAPIKeys_BothSet(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel in Go.

	t.Setenv("ANTHROPIC_API_KEY", "test-key-value")
	t.Setenv("OPENAI_API_KEY", "test-key-value")

	r := checkAPIKeys()
	if r.status != "pass" {
		t.Errorf("expected pass when both keys set, got %s", r.status)
	}
	if r.name != "llm api key" {
		t.Errorf("expected name 'llm api key', got %q", r.name)
	}
}

func TestCheckAPIKeys_AnthropicOnly(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel in Go.

	t.Setenv("ANTHROPIC_API_KEY", "test-key-value")
	t.Setenv("OPENAI_API_KEY", "")

	r := checkAPIKeys()
	if r.status != "pass" {
		t.Errorf("expected pass when ANTHROPIC_API_KEY set, got %s", r.status)
	}
}

func TestCheckAPIKeys_OpenAIOnly(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel in Go.

	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "test-key-value")

	r := checkAPIKeys()
	if r.status != "pass" {
		t.Errorf("expected pass when OPENAI_API_KEY set, got %s", r.status)
	}
}

func TestCheckAPIKeys_NoneSet(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel in Go.

	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")

	r := checkAPIKeys()
	if r.status != "fail" {
		t.Errorf("expected fail when no keys set, got %s", r.status)
	}
}

func TestCheckAPIKeys_NeverLogsValue(t *testing.T) {
	// t.Setenv is incompatible with t.Parallel in Go.

	t.Setenv("ANTHROPIC_API_KEY", "super-secret-key-12345")
	t.Setenv("OPENAI_API_KEY", "")

	r := checkAPIKeys()
	if r.status != "pass" {
		t.Fatalf("expected pass, got %s", r.status)
	}
	// Verify the detail never contains the key value.
	if r.detail != "" && r.detail != "super-secret-key-12345" {
		// detail is empty on pass — this is correct
	}
	if r.detail == "super-secret-key-12345" {
		t.Error("API key value leaked into check detail")
	}
}

func TestParseCircuitBreakerState_Open(t *testing.T) {
	t.Parallel()

	data := []byte(`{"state":"open","failures":3,"last_failure":"2026-03-28T10:00:00Z"}`)
	state, err := parseCircuitBreakerState(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != "open" {
		t.Errorf("expected state 'open', got %q", state)
	}
}

func TestParseCircuitBreakerState_Closed(t *testing.T) {
	t.Parallel()

	data := []byte(`{"state":"closed","failures":0}`)
	state, err := parseCircuitBreakerState(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != "closed" {
		t.Errorf("expected state 'closed', got %q", state)
	}
}

func TestParseCircuitBreakerState_HalfOpen(t *testing.T) {
	t.Parallel()

	data := []byte(`{"state":"half-open"}`)
	state, err := parseCircuitBreakerState(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != "half-open" {
		t.Errorf("expected state 'half-open', got %q", state)
	}
}

func TestParseCircuitBreakerState_InvalidJSON(t *testing.T) {
	t.Parallel()

	data := []byte(`not json`)
	_, err := parseCircuitBreakerState(data)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseCircuitBreakerState_EmptyState(t *testing.T) {
	t.Parallel()

	data := []byte(`{}`)
	state, err := parseCircuitBreakerState(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != "" {
		t.Errorf("expected empty state, got %q", state)
	}
}

func TestCheckClaudeCommandsIn_WithCommands(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	cmdDir := filepath.Join(home, ".claude", "commands")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create mock command files.
	for _, name := range []string{"fd-new.md", "fd-review.md", "sdd-assign.md"} {
		if err := os.WriteFile(filepath.Join(cmdDir, name), []byte("# test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	r := checkClaudeCommandsIn(home)
	if r.status != "pass" {
		t.Errorf("expected pass, got %s: %s", r.status, r.detail)
	}
	if r.detail != "3 commands" {
		t.Errorf("expected '3 commands', got %q", r.detail)
	}
}

func TestCheckClaudeCommandsIn_NoCommands(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	// Don't create any command files.

	r := checkClaudeCommandsIn(home)
	if r.status != "fail" {
		t.Errorf("expected fail when no commands, got %s", r.status)
	}
}

func TestCheckClaudeCommandsIn_EmptyHome(t *testing.T) {
	t.Parallel()

	r := checkClaudeCommandsIn("")
	if r.status != "fail" {
		t.Errorf("expected fail when HOME empty, got %s", r.status)
	}
}

func TestCheckClaudeCommandsIn_OnlyFDCommands(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	cmdDir := filepath.Join(home, ".claude", "commands")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(cmdDir, "fd-new.md"), []byte("# test"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := checkClaudeCommandsIn(home)
	if r.status != "pass" {
		t.Errorf("expected pass, got %s", r.status)
	}
	if r.detail != "1 commands" {
		t.Errorf("expected '1 commands', got %q", r.detail)
	}
}

func TestCheckBeadsCircuitBreaker_NoFiles(t *testing.T) {
	t.Parallel()

	// If no circuit breaker files exist, should skip.
	r := checkBeadsCircuitBreaker()
	// Result depends on system state, but it must not panic.
	if r.name != "beads circuit-breaker" {
		t.Errorf("expected name 'beads circuit-breaker', got %q", r.name)
	}
}

func TestCheckOpenHandsImage_NoDocker(t *testing.T) {
	t.Parallel()

	// If docker is missing from PATH, this should skip gracefully.
	// We can't mock exec.LookPath easily, so just verify it doesn't panic.
	r := checkOpenHandsImage(t.Context())
	if r.name != "openhands image" {
		t.Errorf("expected name 'openhands image', got %q", r.name)
	}
}

func TestCheckOpenHandsContainer_NoDocker(t *testing.T) {
	t.Parallel()

	r := checkOpenHandsContainer(t.Context())
	if r.name != "openhands container" {
		t.Errorf("expected name 'openhands container', got %q", r.name)
	}
}

func TestCheckKnowledgeLayer_NoTool(t *testing.T) {
	t.Parallel()

	// Should degrade gracefully if codebase-memory-mcp is not installed.
	r := checkKnowledgeLayer(t.Context())
	if r.name != "codebase-memory-mcp" {
		t.Errorf("expected name 'codebase-memory-mcp', got %q", r.name)
	}
}
