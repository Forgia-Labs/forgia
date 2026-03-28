package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Deepzima/forgia/internal/config"
	"github.com/Deepzima/forgia/internal/guardrails"
	"github.com/Deepzima/forgia/internal/vault"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health",
	Long:  "Verifies vault, tools, and configuration are set up correctly.",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	name   string
	status string // "pass", "fail", "skip"
	detail string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	logger := slog.With("command", "doctor")

	fmt.Println("=== Forgia Doctor ===")
	fmt.Println()

	var results []checkResult

	// 1. Vault exists and is valid.
	results = append(results, checkVault(ctx))

	// 2. Config parseable.
	results = append(results, checkConfig(ctx))

	// 3. Guardrails parseable.
	results = append(results, checkGuardrails(ctx))

	// 4. Required tools.
	results = append(results, checkTool("git", true))
	results = append(results, checkTool("claude", true))

	// 5. Optional tools.
	results = append(results, checkTool("docker", false))
	results = append(results, checkTool("bd", false))
	results = append(results, checkTool("fswatch", false))
	results = append(results, checkTool("yq", false))

	// 6. OpenHands checks.
	results = append(results, checkOpenHandsImage(ctx))
	results = append(results, checkOpenHandsContainer(ctx))

	// 7. Claude Code slash commands.
	results = append(results, checkClaudeCommands())

	// 8. Beads circuit breaker.
	results = append(results, checkBeadsCircuitBreaker())

	// 9. LLM API key presence (never log values).
	results = append(results, checkAPIKeys())

	// 10. Knowledge layer health.
	results = append(results, checkKnowledgeLayer(ctx))

	// Print results.
	var failures int
	for _, r := range results {
		icon := "✓"
		if r.status == "fail" {
			icon = "✗"
			failures++
		} else if r.status == "skip" {
			icon = "○"
		}
		if r.detail != "" {
			fmt.Printf("  %s %s — %s\n", icon, r.name, r.detail)
		} else {
			fmt.Printf("  %s %s\n", icon, r.name)
		}
	}

	fmt.Println()
	if failures > 0 {
		fmt.Printf("%d issue(s) found.\n", failures)
		logger.WarnContext(ctx, "doctor found issues", "failures", failures)
	} else {
		fmt.Println("All checks passed.")
	}

	// Doctor always exits 0 — it reports, doesn't fail.
	return nil
}

func checkVault(ctx context.Context) checkResult {
	_, err := vault.Open(".")
	if err != nil {
		return checkResult{"vault", "fail", "not found — run 'forgia init'"}
	}
	return checkResult{"vault", "pass", ".forgia/ found and valid"}
}

func checkConfig(ctx context.Context) checkResult {
	_, err := config.LoadConfig(ctx, ".forgia")
	if err != nil {
		return checkResult{"config", "fail", fmt.Sprintf("config.toml: %v", err)}
	}
	return checkResult{"config", "pass", "config.toml parsed"}
}

func checkGuardrails(ctx context.Context) checkResult {
	v, err := vault.Open(".")
	if err != nil {
		return checkResult{"guardrails", "skip", "vault not available"}
	}
	data, err := v.GuardrailsRaw(ctx)
	if err != nil {
		return checkResult{"guardrails", "fail", fmt.Sprintf("deny.toml: %v", err)}
	}
	_, err = guardrails.Parse(data)
	if err != nil {
		return checkResult{"guardrails", "fail", fmt.Sprintf("deny.toml parse: %v", err)}
	}
	return checkResult{"guardrails", "pass", "deny.toml parsed"}
}

func checkTool(name string, required bool) checkResult {
	_, err := exec.LookPath(name)
	if err != nil {
		if required {
			return checkResult{name, "fail", "not found in PATH"}
		}
		return checkResult{name, "skip", "not found (optional)"}
	}
	return checkResult{name, "pass", ""}
}

// checkOpenHandsImage verifies the OpenHands Docker image is pulled locally.
func checkOpenHandsImage(ctx context.Context) checkResult {
	if _, err := exec.LookPath("docker"); err != nil {
		return checkResult{"openhands image", "skip", "docker not available"}
	}

	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", "ghcr.io/openhands/openhands:latest")
	if err := cmd.Run(); err != nil {
		return checkResult{"openhands image", "fail", "not pulled (run: mise run openhands:install)"}
	}
	return checkResult{"openhands image", "pass", ""}
}

// checkOpenHandsContainer verifies an OpenHands container is running.
func checkOpenHandsContainer(ctx context.Context) checkResult {
	if _, err := exec.LookPath("docker"); err != nil {
		return checkResult{"openhands container", "skip", "docker not available"}
	}

	cmd := exec.CommandContext(ctx, "docker", "ps", "--filter", "name=openhands", "--format", "{{.Image}}")
	output, err := cmd.Output()
	if err != nil {
		return checkResult{"openhands container", "skip", "docker ps failed"}
	}
	if strings.TrimSpace(string(output)) == "" {
		return checkResult{"openhands container", "skip", "stopped"}
	}
	return checkResult{"openhands container", "pass", "running"}
}

// checkClaudeCommands verifies Claude Code slash commands are installed.
func checkClaudeCommands() checkResult {
	return checkClaudeCommandsIn(os.Getenv("HOME"))
}

// checkClaudeCommandsIn checks for Claude commands in the given home directory.
func checkClaudeCommandsIn(home string) checkResult {
	if home == "" {
		return checkResult{"claude-commands", "fail", "HOME not set"}
	}

	cmdDir := filepath.Join(home, ".claude", "commands")
	var count int

	for _, pattern := range []string{"fd-*.md", "sdd-*.md"} {
		matches, err := filepath.Glob(filepath.Join(cmdDir, pattern))
		if err != nil {
			continue
		}
		count += len(matches)
	}

	if count == 0 {
		return checkResult{"claude-commands", "fail", "not installed (run: mise run claude:install)"}
	}
	return checkResult{"claude-commands", "pass", fmt.Sprintf("%d commands", count)}
}

// circuitBreakerState represents the state field in a Beads circuit breaker JSON file.
type circuitBreakerState struct {
	State string `json:"state"`
}

// parseCircuitBreakerState parses a circuit breaker JSON and returns the state.
func parseCircuitBreakerState(data []byte) (string, error) {
	var cb circuitBreakerState
	if err := json.Unmarshal(data, &cb); err != nil {
		return "", fmt.Errorf("parse circuit breaker: %w", err)
	}
	return cb.State, nil
}

// checkBeadsCircuitBreaker checks for tripped Beads circuit breakers.
func checkBeadsCircuitBreaker() checkResult {
	// Check both /tmp and /private/tmp (macOS).
	patterns := []string{
		"/tmp/beads-dolt-circuit-*.json",
		"/private/tmp/beads-dolt-circuit-*.json",
	}

	var found bool
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, m := range matches {
			found = true
			data, err := os.ReadFile(m)
			if err != nil {
				continue
			}
			state, err := parseCircuitBreakerState(data)
			if err != nil {
				continue
			}
			if state == "open" {
				return checkResult{"beads circuit-breaker", "fail", "OPEN (run: mise run bd:reset)"}
			}
		}
	}

	if !found {
		return checkResult{"beads circuit-breaker", "skip", "no circuit breaker files"}
	}
	return checkResult{"beads circuit-breaker", "pass", "closed"}
}

// checkAPIKeys verifies at least one LLM API key is set. Never logs or prints values.
func checkAPIKeys() checkResult {
	if len(os.Getenv("ANTHROPIC_API_KEY")) > 0 || len(os.Getenv("OPENAI_API_KEY")) > 0 {
		return checkResult{"llm api key", "pass", ""}
	}
	return checkResult{"llm api key", "fail", "missing (set ANTHROPIC_API_KEY or OPENAI_API_KEY)"}
}

// checkKnowledgeLayer verifies codebase-memory-mcp is installed and reports stats.
func checkKnowledgeLayer(ctx context.Context) checkResult {
	if _, err := exec.LookPath("codebase-memory-mcp"); err != nil {
		return checkResult{"codebase-memory-mcp", "skip", "not installed (optional)"}
	}

	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "codebase-memory-mcp", "cli", "list_projects")
	output, err := cmd.Output()
	if err != nil {
		return checkResult{"codebase-memory-mcp", "skip", "installed but query failed"}
	}

	var projects []KnowledgeProject
	if err := json.Unmarshal(output, &projects); err != nil {
		return checkResult{"codebase-memory-mcp", "skip", "installed but invalid response"}
	}

	if len(projects) == 0 {
		return checkResult{"codebase-memory-mcp", "skip", "installed but no index (run forgia init)"}
	}

	p := projects[0]
	if p.NodeCount == 0 && p.EdgeCount == 0 {
		return checkResult{"codebase-memory-mcp", "skip", "installed but no index (run forgia init)"}
	}

	msg := fmt.Sprintf("%d symbols  %d edges", p.NodeCount, p.EdgeCount)
	if p.LastIndexed != "" {
		t, err := time.Parse(time.RFC3339, p.LastIndexed)
		if err == nil {
			msg += fmt.Sprintf("  synced %s", formatRelativeTime(time.Since(t)))
		}
	}
	return checkResult{"codebase-memory-mcp", "pass", msg}
}
