package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Deepzima/forgia/internal/config"
	"github.com/Deepzima/forgia/internal/guardrails"
	"github.com/Deepzima/forgia/internal/sandbox"
	"github.com/Deepzima/forgia/internal/vault"
)

// SandboxExec wraps SDD execution in a sandbox if configured.
// Returns nil if no sandbox is configured (run on host).
func SandboxExec(ctx context.Context, sdd *vault.SDD, cfg *config.ClaudeRunnerConfig, v vault.Vault) (*ExecResult, error) {
	if cfg == nil || cfg.Sandbox == "" || cfg.Sandbox == "none" {
		return nil, nil // no sandbox — caller runs on host
	}

	logger := slog.With("runner", "claude-sandbox")

	// Resolve sandbox provider.
	provider, err := sandbox.Resolve(ctx, cfg.Sandbox)
	if err != nil {
		return nil, fmt.Errorf("sandbox: %w", err)
	}
	if provider == nil {
		return nil, nil
	}

	logger.InfoContext(ctx, "using sandbox", "provider", provider.Name())

	// Workspace.
	workDir, _ := os.Getwd()
	if cfg.SandboxWorkspaceMount != "" {
		workDir, _ = filepath.Abs(cfg.SandboxWorkspaceMount)
	}

	// Build mounts.
	mounts := sandbox.WorkspaceMounts(workDir)

	// Build environment.
	env := map[string]string{}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		env["ANTHROPIC_API_KEY"] = key
	}

	// Seccomp (Docker only).
	var seccompPath string
	if cfg.SandboxSeccompFromDeny && provider.Name() == "docker" {
		gData, err := v.GuardrailsRaw(ctx)
		if err == nil {
			g, err := guardrails.Parse(gData)
			if err == nil {
				seccompJSON, err := sandbox.GenerateSeccompJSON(g)
				if err == nil {
					tmpFile := filepath.Join(os.TempDir(), "forgia-seccomp.json")
					os.WriteFile(tmpFile, seccompJSON, 0o644)
					seccompPath = tmpFile
					defer os.Remove(tmpFile)
				}
			}
		}
	}

	// Audit logger.
	var auditLogger *sandbox.AuditLogger
	if cfg.SandboxAuditLog {
		logDir := filepath.Join(v.Dir(), "logs")
		os.MkdirAll(logDir, 0o755)
		auditPath := filepath.Join(logDir, fmt.Sprintf("audit-%s-%s.jsonl", sdd.ID, time.Now().Format("2006-01-02T15-04-05")))
		al, err := sandbox.NewAuditLogger(auditPath)
		if err == nil {
			auditLogger = al
			defer auditLogger.Close()
			logger.InfoContext(ctx, "audit log", "path", auditPath)
		}
	}

	// Services.
	if cfg.SandboxServicesFromSDD && len(sdd.Boundaries.Services) > 0 {
		var serviceDefs []sandbox.ServiceDef
		for _, s := range sdd.Boundaries.Services {
			serviceDefs = append(serviceDefs, sandbox.ServiceDef{
				Name:  s.Name,
				Image: s.Image,
				Ports: s.Ports,
				Env:   s.Env,
			})
		}
		mgr := sandbox.NewServiceManager(serviceDefs, provider.Name())
		cleanup, err := mgr.StartAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("start services: %w", err)
		}
		defer cleanup()
	}

	// Build system context.
	systemCtx, _ := BuildSystemContext(ctx, v)

	// Build command to run inside sandbox.
	taskPrompt := buildTaskPrompt(sdd)
	command := []string{
		"claude",
		"--dangerously-skip-permissions",
		"--append-system-prompt", systemCtx,
		"-p", taskPrompt,
	}

	image := cfg.SandboxImage
	if image == "" {
		image = "ghcr.io/anthropics/claude-code:latest"
	}

	started := time.Now()

	// Run in sandbox.
	result, err := provider.Run(ctx, sandbox.RunOpts{
		Image:       image,
		Command:     command,
		WorkDir:     "/workspace",
		Mounts:      mounts,
		Env:         env,
		NetworkMode: "none",
		SeccompPath: seccompPath,
	})

	completed := time.Now()
	duration := int(completed.Sub(started).Seconds())

	// Write audit entry.
	if auditLogger != nil {
		outputBytes := 0
		if result != nil {
			outputBytes = len(result.Output)
		}
		auditLogger.Log(sandbox.AuditEntry{
			Timestamp:   completed,
			Command:     "claude -p <sdd-task>",
			ExitCode:    result.ExitCode,
			DurationMs:  completed.Sub(started).Milliseconds(),
			OutputBytes: outputBytes,
		})
	}

	execResult := &ExecResult{
		SDD:          sdd.ID,
		FD:           sdd.FD,
		Runner:       "claude-sandbox:" + provider.Name(),
		Started:      started,
		Completed:    completed,
		DurationSecs: duration,
	}

	if err != nil {
		execResult.Status = "failed"
		execResult.ExitCode = 1
		return execResult, err
	}

	execResult.ExitCode = result.ExitCode
	if result.ExitCode == 0 {
		execResult.Status = "success"
	} else {
		execResult.Status = "failed"
	}

	return execResult, nil
}
