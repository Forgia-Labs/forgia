package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/forgia-labs/forgia/internal/config"
	"github.com/forgia-labs/forgia/internal/vault"
)

// Compile-time interface check.
var _ Runner = (*OpenHandsRunner)(nil)

// OpenHandsRunner executes SDDs via OpenHands in a Docker container.
type OpenHandsRunner struct {
	cfg    config.OpenHandsConfig
	logger *slog.Logger
}

// NewOpenHandsRunner creates an OpenHands runner with the given config.
// Applies sensible defaults for zero-value fields.
func NewOpenHandsRunner(cfg config.OpenHandsConfig) *OpenHandsRunner {
	if cfg.Image == "" {
		cfg.Image = "ghcr.io/openhands/openhands:latest"
	}
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-20250514"
	}
	if cfg.WorkspaceMount == "" {
		cfg.WorkspaceMount = "/workspace"
	}
	if cfg.MaxIterations == 0 {
		cfg.MaxIterations = 100
	}
	return &OpenHandsRunner{
		cfg:    cfg,
		logger: slog.With("runner", "openhands"),
	}
}

// Name returns "openhands".
func (r *OpenHandsRunner) Name() string {
	return "openhands"
}

// Execute runs an SDD via OpenHands in a Docker container.
func (r *OpenHandsRunner) Execute(ctx context.Context, sdd *vault.SDD, opts ExecOptions) (*ExecResult, error) {
	if err := checkDocker(ctx); err != nil {
		return nil, err
	}

	started := time.Now()
	result := &ExecResult{
		SDD:     sdd.ID,
		FD:      sdd.FD,
		File:    sdd.FilePath,
		Runner:  "openhands",
		Started: started,
	}

	// Pull image if not present locally.
	if err := ensureImage(ctx, r.cfg.Image); err != nil {
		return nil, err
	}

	// Build prompt content.
	promptContent := buildOpenHandsPrompt(sdd, opts)

	// Write prompt to temp file.
	promptFile, err := writePromptFile(promptContent)
	if err != nil {
		return nil, fmt.Errorf("write prompt file: %w", err)
	}
	defer os.Remove(promptFile)

	// Build and run Docker command.
	args := r.buildDockerArgs(sdd, promptFile)

	r.logger.InfoContext(ctx, "executing SDD",
		"sdd", sdd.ID, "image", r.cfg.Image, "model", r.cfg.Model)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, execErr := cmd.CombinedOutput()

	completed := time.Now()
	result.Completed = completed
	result.DurationSecs = int(completed.Sub(started).Seconds())

	// Write execution log.
	writeExecLog(sdd, started, completed, result.DurationSecs, output)

	if execErr != nil {
		result.ExitCode = 1
		if exitErr, ok := execErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Status = "failed"
		r.logger.WarnContext(ctx, "execution failed",
			"sdd", sdd.ID, "exit_code", result.ExitCode, "error", execErr)
	} else {
		result.ExitCode = 0
		result.Status = "success"
		r.logger.InfoContext(ctx, "execution complete",
			"sdd", sdd.ID, "duration", result.DurationSecs)
	}

	return result, execErr
}

// buildDockerArgs constructs the docker run arguments.
func (r *OpenHandsRunner) buildDockerArgs(sdd *vault.SDD, promptFile string) []string {
	cwd, _ := os.Getwd()
	containerName := fmt.Sprintf("forgia-%s", strings.ToLower(sdd.ID))

	args := []string{
		"run",
		"--name", containerName,
		"--rm",
		"-v", fmt.Sprintf("%s:%s", cwd, r.cfg.WorkspaceMount),
		"-v", fmt.Sprintf("%s:/tmp/prompt.md:ro", promptFile),
	}

	// Port binding (skip if UIPort=0).
	if r.cfg.UIPort > 0 {
		args = append(args, "-p", fmt.Sprintf("%d:3000", r.cfg.UIPort))
	}

	// API key — passed via environment, never written to files.
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		args = append(args, "-e", "ANTHROPIC_API_KEY")
	} else if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		args = append(args, "-e", "OPENAI_API_KEY")
	}

	// Model and iteration config.
	args = append(args,
		"-e", fmt.Sprintf("LLM_MODEL=%s", r.cfg.Model),
		"-e", fmt.Sprintf("MAX_ITERATIONS=%d", r.cfg.MaxIterations),
		"-e", fmt.Sprintf("WORKSPACE_BASE=%s", r.cfg.WorkspaceMount),
	)

	args = append(args, r.cfg.Image, "--task-file", "/tmp/prompt.md")

	return args
}

// checkDocker verifies the Docker daemon is running.
func checkDocker(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker non disponibile. Installa Docker: https://docs.docker.com/get-docker/")
	}
	return nil
}

// ensureImage pulls the Docker image if not present locally.
func ensureImage(ctx context.Context, image string) error {
	// Check if image exists locally.
	check := exec.CommandContext(ctx, "docker", "image", "inspect", image)
	check.Stdout = nil
	check.Stderr = nil
	if check.Run() == nil {
		return nil // already present
	}

	// Pull image.
	pull := exec.CommandContext(ctx, "docker", "pull", image)
	output, err := pull.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker pull %s failed: %s: %w", image, strings.TrimSpace(string(output)), err)
	}
	return nil
}

// buildOpenHandsPrompt assembles the task prompt for the OpenHands container.
func buildOpenHandsPrompt(sdd *vault.SDD, opts ExecOptions) string {
	var parts []string

	parts = append(parts, "You are executing an SDD (Spec-Driven Development) task. Follow the spec exactly.")

	if opts.SystemContext != "" {
		parts = append(parts, opts.SystemContext)
	}

	taskPrompt := opts.TaskPrompt
	if taskPrompt == "" {
		taskPrompt = buildTaskPrompt(sdd)
	}
	parts = append(parts, "## SDD Task\n"+taskPrompt)

	return strings.Join(parts, "\n\n")
}

// writePromptFile writes the prompt content to a temporary file.
func writePromptFile(content string) (string, error) {
	f, err := os.CreateTemp("", "forgia-prompt-*.md")
	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// writeExecLog writes the execution log to the logs directory.
func writeExecLog(sdd *vault.SDD, started, completed time.Time, durationSecs int, output []byte) {
	logDir := filepath.Join(filepath.Dir(filepath.Dir(sdd.FilePath)), "..", "logs")
	os.MkdirAll(logDir, 0o755)
	logFile := filepath.Join(logDir, fmt.Sprintf("exec-%s-%s.log", sdd.ID, started.Format("2006-01-02T15:04:05")))

	logContent := fmt.Sprintf(
		"=== Forgia Exec Log ===\nSDD: %s\nFile: %s\nStarted: %s\nRunner: openhands\n---\n%s\n---\nCompleted: %s\nDuration: %ds\n",
		sdd.ID, sdd.FilePath, started.Format(time.RFC3339), string(output), completed.Format(time.RFC3339), durationSecs,
	)
	os.WriteFile(logFile, []byte(logContent), 0o644)
}
