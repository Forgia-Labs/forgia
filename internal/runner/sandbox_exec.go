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
		expanded := strings.Replace(ep, "~", homeDir, 1)
		if strings.HasPrefix(workDir, expanded) {
			return nil, fmt.Errorf("sandbox: workspace %q is inside excluded path %q", workDir, ep)
		}
	}

	// Build mounts.
	mounts := sandbox.WorkspaceMounts(workDir)

	// Build environment.
	env := map[string]string{}

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

	// --- Auth resolution: keychain → API key → error ---
	var authCleanup func()
	authMethod := resolveAuth(ctx, logger, homeDir, env, &mounts, &authCleanup)
	if authCleanup != nil {
		defer authCleanup()
	}
	if authMethod == "none" {
		return nil, fmt.Errorf("sandbox: no credentials available\n  ✗ No macOS keychain credentials found\n  ✗ No ANTHROPIC_API_KEY in environment\n\n  Fix: run 'claude /login' on host, or export ANTHROPIC_API_KEY")
	}
	logger.InfoContext(ctx, "auth resolved", "method", authMethod)

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

	// Build command to run inside sandbox.
	// Claude runs inside /workspace which is the project root.
	// It reads the SDD file directly from the mounted filesystem —
	// no need to pass the full spec as a command-line argument.
	sddPath := sdd.FilePath
	if sddPath == "" {
		sddPath = filepath.Join(".forgia", "sdd", sdd.FD, sdd.ID+".md")
	}

	prompt := fmt.Sprintf("Read and execute the SDD at %s — first read .forgia/constitution.md and .forgia/guardrails/deny.toml for project rules, then implement the Scope section exactly, verify Acceptance Criteria, and update the Work Log when done.", sddPath)

	// Note: the Docker image ENTRYPOINT is ["claude"], so we only pass arguments here.
	command := []string{
		"--dangerously-skip-permissions",
		"-p",
		prompt,
	}

	// TM-1: validate sandbox image — warn on non-default images.
	image := cfg.SandboxImage
	if image == "" {
		image = "forgia-sandbox:latest"
	}
	defaultImages := map[string]bool{
		"forgia-sandbox:latest": true,
		"node:22-slim":         true,
	}
	if !defaultImages[image] {
		logger.WarnContext(ctx, "non-default sandbox image — ensure it is trusted", "image", image)
	}

	// Network: Claude needs API access. Use host network for now.
	// TODO: implement egress proxy that allows only api.anthropic.com
	networkMode := "host"

	started := time.Now()
	fmt.Printf("→ Executing %s in sandbox...\n", sdd.ID)

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

	fmt.Printf("→ Auth: credentials cleaned up ✓\n")

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

// resolveAuth tries: 1) sandbox-auth dir (forgia sandbox login), 2) ANTHROPIC_API_KEY env, 3) none.
// Returns the method used ("sandbox-auth", "api-key", "none") and optionally adds mounts/env.
func resolveAuth(ctx context.Context, logger *slog.Logger, homeDir string, env map[string]string, mounts *[]sandbox.Mount, cleanup *func()) string {
	// 1. Try persisted sandbox auth (from `forgia sandbox login`).
	authDir := filepath.Join(homeDir, ".forgia", "sandbox-auth", "claude")

	// Check for .env file (API key stored via `forgia sandbox login`).
	envPath := filepath.Join(authDir, ".env")
	if data, err := os.ReadFile(envPath); err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			k, v, ok := strings.Cut(line, "=")
			if ok && k != "" && v != "" {
				env[k] = v
			}
		}
		fmt.Printf("→ Auth: using sandbox credentials (API key) ✓\n")
		return "sandbox-auth"
	}

	// Check for OAuth credentials (from `forgia sandbox login`).
	if info, err := os.Stat(authDir); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(authDir)
		if len(entries) > 0 {
			// Mount ~/.claude directory with OAuth session data.
			// Claude needs read-write — it writes session/cache files.
			// .claude.json must NOT be mounted separately as a file —
			// Docker file mounts break when Claude truncates and rewrites.
			*mounts = append(*mounts, sandbox.Mount{
				Source:   authDir,
				Target:   "/home/forgia/.claude",
				ReadOnly: false,
			})
			fmt.Printf("→ Auth: using sandbox credentials (OAuth) ✓\n")
			return "sandbox-auth"
		}
	}

	// 2. Try ANTHROPIC_API_KEY from host environment.
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		env["ANTHROPIC_API_KEY"] = key
		fmt.Printf("→ Auth: using ANTHROPIC_API_KEY from env ✓\n")
		return "api-key"
	}

	// 3. Check if already set via sandbox_env config.
	if _, ok := env["ANTHROPIC_API_KEY"]; ok {
		fmt.Printf("→ Auth: using ANTHROPIC_API_KEY from config ✓\n")
		return "api-key"
	}

	return "none"
}

// looksLikeSecret checks if a value matches common API key / credential patterns.
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

