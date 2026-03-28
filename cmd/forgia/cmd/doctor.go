package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

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
