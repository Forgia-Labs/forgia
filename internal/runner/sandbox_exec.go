package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forgia-labs/forgia/internal/config"
	"github.com/forgia-labs/forgia/internal/guardrails"
	"github.com/forgia-labs/forgia/internal/sandbox"
	"github.com/forgia-labs/forgia/internal/vault"
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

	// Validate workspace is not an excluded path.
	excluded := sandbox.ExcludedHostPaths()
	homeDir, _ := os.UserHomeDir()
	for _, ep := range excluded {
		// Skip ~/.claude check if mount_claude is explicitly enabled.
		if ep == "~/.claude" && cfg.SandboxMountClaude {
			continue
		}
		expanded := strings.Replace(ep, "~", homeDir, 1)
		if strings.HasPrefix(workDir, expanded) {
			return nil, fmt.Errorf("sandbox: workspace %q is inside excluded path %q", workDir, ep)
		}
	}

	// Build mounts.
	mounts := sandbox.WorkspaceMounts(workDir)

	// Optional: mount ~/.claude read-only for Max OAuth auth.
	// TM-2: when mount_claude=true, network MUST be none to prevent token exfiltration.
	if cfg.SandboxMountClaude {
		claudeDir := filepath.Join(homeDir, ".claude")
		if _, err := os.Stat(claudeDir); err == nil {
			mounts = append(mounts, sandbox.Mount{
				Source:   claudeDir,
				Target:   "/root/.claude",
				ReadOnly: true,
			})
			logger.InfoContext(ctx, "mounting ~/.claude read-only for auth")
			// Force network=none when OAuth tokens are mounted.
			if len(cfg.SandboxNetworkAllow) > 0 {
				logger.WarnContext(ctx, "sandbox_mount_claude=true forces network=none — ignoring sandbox_network_allow")
			}
		}
	}

	// Build environment.
	env := map[string]string{}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		env["ANTHROPIC_API_KEY"] = key
	}

	// TM-3: scan sandbox_env for accidental secret exposure.
	for _, e := range cfg.SandboxEnv {
		k, v, ok := strings.Cut(e, "=")
		if ok {
			if looksLikeSecret(v) {
				logger.WarnContext(ctx, "sandbox_env value looks like a secret — consider using env var reference instead of plaintext in config.toml",
					"key", k)
			}
			env[k] = v
		}
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

	// TM-1: validate sandbox image — warn on non-default images.
	image := cfg.SandboxImage
	if image == "" {
		image = "ghcr.io/anthropics/claude-code:latest"
	}
	defaultImages := map[string]bool{
		"ghcr.io/anthropics/claude-code:latest": true,
		"ubuntu:latest":                         true,
		"debian:latest":                         true,
	}
	if !defaultImages[image] {
		logger.WarnContext(ctx, "non-default sandbox image — ensure it is trusted", "image", image)
	}

	// TM-2: force network=none when OAuth tokens are mounted.
	networkMode := "none"
	if !cfg.SandboxMountClaude && len(cfg.SandboxNetworkAllow) > 0 {
		networkMode = "filtered" // future: proxy-based egress filtering
	}

	started := time.Now()

	// Run in sandbox.
	result, err := provider.Run(ctx, sandbox.RunOpts{
		Image:       image,
		Command:     command,
		WorkDir:     "/workspace",
		Mounts:      mounts,
		Env:         env,
		NetworkMode: networkMode,
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

// looksLikeSecret checks if a value matches common API key / credential patterns.
// Used by TM-3 to warn when sandbox_env contains plaintext secrets.
func looksLikeSecret(v string) bool {
	prefixes := []string{
		"sk-ant-", "sk-", "ghp_", "gho_", "github_pat_",
		"glpat-", "AKIA", "xoxb-", "xoxp-",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(v, p) {
			return true
		}
	}
	if strings.Contains(v, "PRIVATE KEY") {
		return true
	}
	return false
}
