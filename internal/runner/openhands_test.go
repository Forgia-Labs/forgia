package runner

import (
	"os"
	"strings"
	"testing"

	"github.com/Deepzima/forgia/internal/config"
	"github.com/Deepzima/forgia/internal/vault"
)

func TestNewOpenHandsRunner_Defaults(t *testing.T) {
	t.Parallel()

	r := NewOpenHandsRunner(config.OpenHandsConfig{})

	if r.Name() != "openhands" {
		t.Errorf("Name() = %q, want %q", r.Name(), "openhands")
	}
	if r.cfg.Image != "ghcr.io/openhands/openhands:latest" {
		t.Errorf("Image = %q, want default", r.cfg.Image)
	}
	if r.cfg.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model = %q, want default", r.cfg.Model)
	}
	if r.cfg.WorkspaceMount != "/workspace" {
		t.Errorf("WorkspaceMount = %q, want /workspace", r.cfg.WorkspaceMount)
	}
	if r.cfg.MaxIterations != 100 {
		t.Errorf("MaxIterations = %d, want 100", r.cfg.MaxIterations)
	}
}

func TestNewOpenHandsRunner_CustomConfig(t *testing.T) {
	t.Parallel()

	cfg := config.OpenHandsConfig{
		Image:          "custom/openhands:v1",
		Model:          "gpt-4o",
		WorkspaceMount: "/opt/work",
		MaxIterations:  50,
		UIPort:         8080,
	}

	r := NewOpenHandsRunner(cfg)

	if r.cfg.Image != "custom/openhands:v1" {
		t.Errorf("Image = %q, want custom", r.cfg.Image)
	}
	if r.cfg.Model != "gpt-4o" {
		t.Errorf("Model = %q, want gpt-4o", r.cfg.Model)
	}
	if r.cfg.WorkspaceMount != "/opt/work" {
		t.Errorf("WorkspaceMount = %q, want /opt/work", r.cfg.WorkspaceMount)
	}
	if r.cfg.MaxIterations != 50 {
		t.Errorf("MaxIterations = %d, want 50", r.cfg.MaxIterations)
	}
	if r.cfg.UIPort != 8080 {
		t.Errorf("UIPort = %d, want 8080", r.cfg.UIPort)
	}
}

func TestOpenHandsRunner_BuildDockerArgs_VolumeMounts(t *testing.T) {
	t.Parallel()

	r := NewOpenHandsRunner(config.OpenHandsConfig{
		Image:          "ghcr.io/openhands/openhands:latest",
		Model:          "claude-sonnet-4-20250514",
		WorkspaceMount: "/workspace",
		MaxIterations:  100,
		UIPort:         3000,
	})

	sdd := &vault.SDD{ID: "SDD-004", FD: "FD-004", FilePath: ".forgia/sdd/FD-004/SDD-004.md"}
	args := r.buildDockerArgs(sdd, "/tmp/test-prompt.md")

	argsStr := strings.Join(args, " ")

	// Verify volume mounts.
	cwd, _ := os.Getwd()
	expectedWorkspaceMount := cwd + ":/workspace"
	if !strings.Contains(argsStr, expectedWorkspaceMount) {
		t.Errorf("args missing workspace volume mount %q, got: %s", expectedWorkspaceMount, argsStr)
	}
	if !strings.Contains(argsStr, "/tmp/test-prompt.md:/tmp/prompt.md:ro") {
		t.Errorf("args missing prompt volume mount, got: %s", argsStr)
	}
}

func TestOpenHandsRunner_BuildDockerArgs_PortBinding(t *testing.T) {
	t.Parallel()

	r := NewOpenHandsRunner(config.OpenHandsConfig{
		UIPort: 3000,
	})

	sdd := &vault.SDD{ID: "SDD-004"}
	args := r.buildDockerArgs(sdd, "/tmp/prompt.md")
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-p 3000:3000") {
		t.Errorf("args missing port binding, got: %s", argsStr)
	}
}

func TestOpenHandsRunner_BuildDockerArgs_NoPortWhenZero(t *testing.T) {
	t.Parallel()

	r := NewOpenHandsRunner(config.OpenHandsConfig{
		UIPort: 0,
	})

	sdd := &vault.SDD{ID: "SDD-004"}
	args := r.buildDockerArgs(sdd, "/tmp/prompt.md")
	argsStr := strings.Join(args, " ")

	if strings.Contains(argsStr, "-p ") {
		t.Errorf("args should not contain port binding when UIPort=0, got: %s", argsStr)
	}
}

func TestOpenHandsRunner_BuildDockerArgs_EnvVars(t *testing.T) {
	t.Parallel()

	r := NewOpenHandsRunner(config.OpenHandsConfig{
		Model:         "claude-sonnet-4-20250514",
		MaxIterations: 100,
	})

	sdd := &vault.SDD{ID: "SDD-004"}
	args := r.buildDockerArgs(sdd, "/tmp/prompt.md")
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "LLM_MODEL=claude-sonnet-4-20250514") {
		t.Errorf("args missing LLM_MODEL env var, got: %s", argsStr)
	}
	if !strings.Contains(argsStr, "MAX_ITERATIONS=100") {
		t.Errorf("args missing MAX_ITERATIONS env var, got: %s", argsStr)
	}
	if !strings.Contains(argsStr, "WORKSPACE_BASE=") {
		t.Errorf("args missing WORKSPACE_BASE env var, got: %s", argsStr)
	}
}

func TestOpenHandsRunner_BuildDockerArgs_ContainerName(t *testing.T) {
	t.Parallel()

	r := NewOpenHandsRunner(config.OpenHandsConfig{})

	sdd := &vault.SDD{ID: "SDD-004"}
	args := r.buildDockerArgs(sdd, "/tmp/prompt.md")
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "--name forgia-sdd-004") {
		t.Errorf("args missing container name, got: %s", argsStr)
	}
}

func TestOpenHandsRunner_BuildDockerArgs_ImageAndTaskFile(t *testing.T) {
	t.Parallel()

	r := NewOpenHandsRunner(config.OpenHandsConfig{
		Image: "ghcr.io/openhands/openhands:latest",
	})

	sdd := &vault.SDD{ID: "SDD-004"}
	args := r.buildDockerArgs(sdd, "/tmp/prompt.md")

	// Image and --task-file should be at the end.
	last3 := args[len(args)-3:]
	if last3[0] != "ghcr.io/openhands/openhands:latest" {
		t.Errorf("expected image as third-to-last arg, got %q", last3[0])
	}
	if last3[1] != "--task-file" {
		t.Errorf("expected --task-file as second-to-last arg, got %q", last3[1])
	}
	if last3[2] != "/tmp/prompt.md" {
		t.Errorf("expected /tmp/prompt.md as last arg, got %q", last3[2])
	}
}

func TestOpenHandsRunner_BuildDockerArgs_APIKeyNotInArgs(t *testing.T) {
	t.Parallel()

	r := NewOpenHandsRunner(config.OpenHandsConfig{})

	sdd := &vault.SDD{ID: "SDD-004"}
	args := r.buildDockerArgs(sdd, "/tmp/prompt.md")
	argsStr := strings.Join(args, " ")

	// API key values should never appear in arguments — only the env var name.
	for _, arg := range args {
		if strings.Contains(arg, "ANTHROPIC_API_KEY=sk-") || strings.Contains(arg, "OPENAI_API_KEY=sk-") {
			t.Errorf("API key value leaked into docker args: %s", argsStr)
		}
	}
}

func TestResolve_OpenHands(t *testing.T) {
	t.Parallel()

	r, err := Resolve("openhands")
	if err != nil {
		t.Fatalf("Resolve(openhands) error: %v", err)
	}
	if r.Name() != "openhands" {
		t.Errorf("Name() = %q, want %q", r.Name(), "openhands")
	}
}

func TestResolve_OpenHandsWithConfig(t *testing.T) {
	t.Parallel()

	cfg := config.OpenHandsConfig{
		Image: "custom:v2",
		Model: "gpt-4o",
	}

	r, err := Resolve("openhands", cfg)
	if err != nil {
		t.Fatalf("Resolve(openhands, cfg) error: %v", err)
	}

	oh, ok := r.(*OpenHandsRunner)
	if !ok {
		t.Fatal("expected *OpenHandsRunner")
	}
	if oh.cfg.Image != "custom:v2" {
		t.Errorf("Image = %q, want custom:v2", oh.cfg.Image)
	}
	if oh.cfg.Model != "gpt-4o" {
		t.Errorf("Model = %q, want gpt-4o", oh.cfg.Model)
	}
}

func TestResolve_ExistingRunners_StillWork(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		wantName string
	}{
		{"claude", "claude"},
		{"claude-code", "claude"},
		{"", "claude"},
		{"dry-run", "dry-run"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := Resolve(tt.name)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tt.name, err)
			}
			if r.Name() != tt.wantName {
				t.Errorf("Name() = %q, want %q", r.Name(), tt.wantName)
			}
		})
	}
}

func TestBuildOpenHandsPrompt_IncludesSystemContext(t *testing.T) {
	t.Parallel()

	sdd := &vault.SDD{ID: "SDD-004", FD: "FD-004", FilePath: ".forgia/sdd/FD-004/SDD-004.md"}
	opts := ExecOptions{
		SystemContext: "## Constitution\ntest constitution",
	}

	prompt := buildOpenHandsPrompt(sdd, opts)

	if !strings.Contains(prompt, "test constitution") {
		t.Error("prompt missing system context")
	}
	if !strings.Contains(prompt, "SDD (Spec-Driven Development)") {
		t.Error("prompt missing SDD preamble")
	}
}

func TestBuildOpenHandsPrompt_UsesCustomTaskPrompt(t *testing.T) {
	t.Parallel()

	sdd := &vault.SDD{ID: "SDD-004", FilePath: ".forgia/sdd/FD-004/SDD-004.md"}
	opts := ExecOptions{
		TaskPrompt: "Custom task instruction",
	}

	prompt := buildOpenHandsPrompt(sdd, opts)

	if !strings.Contains(prompt, "Custom task instruction") {
		t.Error("prompt missing custom task prompt")
	}
}

func TestWritePromptFile_CreatesAndCleans(t *testing.T) {
	t.Parallel()

	content := "test prompt content"
	path, err := writePromptFile(content)
	if err != nil {
		t.Fatalf("writePromptFile error: %v", err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read prompt file: %v", err)
	}
	if string(data) != content {
		t.Errorf("content = %q, want %q", string(data), content)
	}

	// Verify cleanup works.
	os.Remove(path)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("prompt file not cleaned up")
	}
}
