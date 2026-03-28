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

	"github.com/Deepzima/forgia/internal/vault"
)

// Compile-time interface check.
var _ Runner = (*ClaudeRunner)(nil)

// ClaudeRunner executes SDDs via the claude CLI.
type ClaudeRunner struct {
	logger *slog.Logger
}

// NewClaudeRunner creates a Claude Code runner.
func NewClaudeRunner() *ClaudeRunner {
	return &ClaudeRunner{
		logger: slog.With("runner", "claude"),
	}
}

// Name returns "claude".
func (r *ClaudeRunner) Name() string {
	return "claude"
}

// Execute runs an SDD via the claude CLI.
func (r *ClaudeRunner) Execute(ctx context.Context, sdd *vault.SDD, opts ExecOptions) (*ExecResult, error) {
	// Verify claude is available.
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return nil, fmt.Errorf("claude CLI not found in PATH: %w (install: https://claude.ai/claude-code)", err)
	}

	started := time.Now()
	result := &ExecResult{
		SDD:     sdd.ID,
		FD:      sdd.FD,
		File:    sdd.FilePath,
		Runner:  "claude",
		Started: started,
	}

	// Build task prompt.
	taskPrompt := opts.TaskPrompt
	if taskPrompt == "" {
		taskPrompt = buildTaskPrompt(sdd)
	}

	// Build claude args.
	args := []string{}
	if opts.PermissionMode == "auto" || opts.PermissionMode == "bypassPermissions" {
		args = append(args, "--dangerously-skip-permissions")
	}
	if opts.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", opts.MaxTurns))
	}
	if opts.SystemContext != "" {
		args = append(args, "--append-system-prompt", opts.SystemContext)
	}
	args = append(args, "-p", taskPrompt)

	// Setup logging.
	logDir := filepath.Join(filepath.Dir(filepath.Dir(sdd.FilePath)), "..", "logs")
	os.MkdirAll(logDir, 0o755)
	logFile := filepath.Join(logDir, fmt.Sprintf("exec-%s-%s.log", sdd.ID, started.Format("2006-01-02T15:04:05")))

	r.logger.InfoContext(ctx, "executing SDD",
		"sdd", sdd.ID, "claude", claudePath, "log", logFile)

	// Run claude.
	cmd := exec.CommandContext(ctx, claudePath, args...)
	output, execErr := cmd.CombinedOutput()

	completed := time.Now()
	result.Completed = completed
	result.DurationSecs = int(completed.Sub(started).Seconds())

	// Write log.
	logContent := fmt.Sprintf("=== Forgia Exec Log ===\nSDD: %s\nFile: %s\nStarted: %s\nRunner: claude-code\n---\n%s\n---\nCompleted: %s\nDuration: %ds\n",
		sdd.ID, sdd.FilePath, started.Format(time.RFC3339), string(output), completed.Format(time.RFC3339), result.DurationSecs)
	os.WriteFile(logFile, []byte(logContent), 0o644)

	if execErr != nil {
		result.ExitCode = 1
		if exitErr, ok := execErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Status = "failed"
		r.logger.WarnContext(ctx, "execution failed", "sdd", sdd.ID, "exit_code", result.ExitCode, "error", execErr)
	} else {
		result.ExitCode = 0
		result.Status = "success"
		r.logger.InfoContext(ctx, "execution complete", "sdd", sdd.ID, "duration", result.DurationSecs)
	}

	return result, execErr
}

// BuildSystemContext assembles the system context from vault files.
// Exported so other commands (skill, dry-run) can reuse it.
func BuildSystemContext(ctx context.Context, v vault.Vault) (string, error) {
	var parts []string

	// Constitution.
	constitution, err := v.Constitution(ctx)
	if err == nil && len(constitution) > 0 {
		parts = append(parts, "## Constitution\n"+string(constitution))
	}

	// Guardrails.
	guardrailsRaw, err := v.GuardrailsRaw(ctx)
	if err == nil && len(guardrailsRaw) > 0 {
		parts = append(parts, "## Security Guardrails — MANDATORY\n"+string(guardrailsRaw)+
			"\nNEVER read secrets, NEVER execute key export commands, NEVER write to .env or forgia meta files.")
	}

	// Dev-guide principles (walk the directory).
	vaultDir := v.Dir()
	principlesDir := filepath.Join(vaultDir, "dev-guide", "principles")
	if entries, err := os.ReadDir(principlesDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				data, err := os.ReadFile(filepath.Join(principlesDir, e.Name()))
				if err == nil {
					parts = append(parts, string(data))
				}
			}
		}
	}

	// Language conventions.
	langDir := filepath.Join(vaultDir, "dev-guide", "lang")
	if entries, err := os.ReadDir(langDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				data, err := os.ReadFile(filepath.Join(langDir, e.Name()))
				if err == nil {
					parts = append(parts, string(data))
				}
			}
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

// buildTaskPrompt creates the default task prompt for an SDD.
func buildTaskPrompt(sdd *vault.SDD) string {
	return fmt.Sprintf(
		"Execute this SDD. Read the file %s for the full spec. "+
			"Implement everything in the Scope section, follow Constraints and Best Practices, "+
			"write tests as specified, verify all Acceptance Criteria. "+
			"When done, update the Work Log section in %s with agent=claude-code, "+
			"date=%s, decisions, output files, and retrospective. "+
			"Commit with a message referencing %s.",
		sdd.FilePath, sdd.FilePath, time.Now().Format("2006-01-02"), sdd.ID,
	)
}

// Resolve picks the right runner from a name string.
func Resolve(runnerName string) (Runner, error) {
	switch strings.ToLower(strings.TrimSpace(runnerName)) {
	case "claude", "claude-code", "":
		return NewClaudeRunner(), nil
	case "dry-run":
		return NewDryRunRunner(), nil
	default:
		return nil, fmt.Errorf("unknown runner %q (available: claude, dry-run)", runnerName)
	}
}
