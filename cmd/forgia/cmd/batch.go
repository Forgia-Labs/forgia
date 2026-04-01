package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgia-labs/forgia/internal/beads"
	"github.com/forgia-labs/forgia/internal/config"
	"github.com/forgia-labs/forgia/internal/runner"
	"github.com/forgia-labs/forgia/internal/vault"
	"github.com/spf13/cobra"
)

var batchDryRun bool
var batchRunner string
var batchSandbox string

var batchCmd = &cobra.Command{
	Use:   "batch <FD-NNN> [--runner=claude|openhands] [--dry-run]",
	Short: "Execute all pending SDDs for an FD (or simulate with --dry-run)",
	Long:  "Execute all pending SDDs for the given FD. With --dry-run, runs feasibility simulation on each SDD without modifying any files.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fdID := args[0]
		sddDir := filepath.Join(".forgia", "sdd", fdID)

		if _, err := os.Stat(sddDir); os.IsNotExist(err) {
			return fmt.Errorf("no SDDs found for %s in %s", fdID, sddDir)
		}

		entries, err := os.ReadDir(sddDir)
		if err != nil {
			return fmt.Errorf("read directory %s: %w", sddDir, err)
		}

		var pendingSDDs []string
		for _, e := range entries {
			if e.IsDir() || !strings.HasPrefix(e.Name(), "SDD-") || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			sddPath := filepath.Join(sddDir, e.Name())
			data, err := os.ReadFile(sddPath)
			if err != nil {
				continue
			}
			// Simple status check — skip done SDDs
			if strings.Contains(string(data), "status: done") {
				continue
			}
			pendingSDDs = append(pendingSDDs, sddPath)
		}

		if len(pendingSDDs) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No pending SDDs for", fdID)
			return nil
		}

		if batchDryRun {
			if batchRunner != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Note: --runner ignored during dry-run (simulation uses Claude)")
			}
			return runBatchDryRun(cmd, fdID, pendingSDDs)
		}

		// Normal execution
		resolvedRunner := batchRunner
		if resolvedRunner == "" {
			resolvedRunner = "claude"
		}

		r, err := runner.Resolve(resolvedRunner)
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		logger := slog.With("command", "batch")
		logsDir := filepath.Join(".forgia", "logs")
		bc := beads.NewClient()

		// Build exec options matching exec.go — load config for max_turns.
		v, vErr := vault.Open(".")
		if vErr != nil {
			slog.WarnContext(ctx, "could not open vault, using defaults", "error", vErr)
		}
		var systemCtx string
		maxTurns := 200
		if v != nil {
			systemCtx, _ = runner.BuildSystemContext(ctx, v)
			cfg, cfgErr := config.LoadConfig(ctx, v.Dir())
			if cfgErr == nil && cfg.Runner.Claude.MaxTurns > 0 {
				maxTurns = cfg.Runner.Claude.MaxTurns
			}
		}
		opts := runner.ExecOptions{
			PermissionMode: "auto",
			MaxTurns:       maxTurns,
			SystemContext:   systemCtx,
		}

		total := len(pendingSDDs)
		var completed, failed int

		for i, sddFile := range pendingSDDs {
			data, _ := os.ReadFile(sddFile)
			content := string(data)
			sddID := strings.Trim(extractFrontmatterValue(content, "id:"), "\"' ")
			sddFD := strings.Trim(extractFrontmatterValue(content, "fd:"), "\"' ")

			fmt.Fprintf(cmd.OutOrStdout(), "\n=== [%d/%d] Executing %s ===\n", i+1, total, sddID)
			fmt.Fprintf(cmd.OutOrStdout(), "  File: %s\n", sddFile)
			fmt.Fprintf(cmd.OutOrStdout(), "  Runner: %s\n\n", r.Name())

			sdd := &vault.SDD{ID: sddID, FD: sddFD, FilePath: sddFile}
			result, execErr := r.Execute(ctx, sdd, opts)

			// Write JSON report for every execution.
			var reportPath string
			if result != nil {
				if result.File == "" {
					result.File = sddFile
				}
				rp, reportErr := writeExecReport(result, logsDir)
				if reportErr != nil {
					logger.WarnContext(ctx, "failed to write exec report", "sdd", sddID, "error", reportErr)
				} else {
					reportPath = rp
				}
			}

			// Log result summary.
			if result != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  Status:   %s\n", result.Status)
				fmt.Fprintf(cmd.OutOrStdout(), "  Exit:     %d\n", result.ExitCode)
				fmt.Fprintf(cmd.OutOrStdout(), "  Duration: %ds\n", result.DurationSecs)
				if reportPath != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  Report:   %s\n", reportPath)
				}
				if len(result.FilesCreated) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "  Created:  %s\n", strings.Join(result.FilesCreated, ", "))
				}
				if len(result.FilesModified) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "  Modified: %s\n", strings.Join(result.FilesModified, ", "))
				}
			}

			if execErr != nil || (result != nil && result.ExitCode != 0) {
				failed++
				reason := "non-zero exit code"
				if execErr != nil {
					reason = execErr.Error()
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "\n  ✗ %s FAILED: %s\n", sddID, reason)

				// Check log file for permission issues or other blocking errors.
				if reportPath != "" {
					fmt.Fprintf(cmd.ErrOrStderr(), "  Check log: %s\n", reportPath)
				}

				// Stop batch on failure — don't execute remaining SDDs.
				fmt.Fprintf(cmd.ErrOrStderr(), "\n  Batch stopped. Fix %s before continuing.\n", sddID)
				fmt.Fprintf(cmd.ErrOrStderr(), "  Remaining: %d SDDs not executed.\n", total-i-1)
				fmt.Fprintf(cmd.OutOrStdout(), "\nCompleted: %d, Failed: %d, Remaining: %d\n", completed, failed, total-i-1)
				return fmt.Errorf("%s failed: %s", sddID, reason)
			}

			completed++
			closeBeadsTask(ctx, bc, sddID)
			fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s completed\n", sddID)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\n=== Batch Complete ===\nCompleted: %d, Failed: %d\n", completed, failed)
		return nil
	},
}

func runBatchDryRun(cmd *cobra.Command, fdID string, sddFiles []string) error {
	slashCmd := filepath.Join("modules", "claude-commands", "sdd-dry-run.md")
	if _, err := os.Stat(slashCmd); os.IsNotExist(err) {
		return fmt.Errorf("sdd-dry-run.md not found — run SDD-001 of FD-012 first")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "=== FD Dry Run: %s (%d SDDs) ===\n\n", fdID, len(sddFiles))

	var nogoCount int
	for _, sddFile := range sddFiles {
		sdd := &vault.SDD{FilePath: sddFile}
		result, err := runner.DryRun(cmd.Context(), sdd, runner.DryRunOptions{
			SlashCommandPath: slashCmd,
		})
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Dry-run error %s: %v\n", sddFile, err)
			nogoCount++
			continue
		}

		fmt.Fprintf(cmd.OutOrStdout(), "--- %s ---\n", filepath.Base(sddFile))
		fmt.Fprintln(cmd.OutOrStdout(), result.ReportText)
		fmt.Fprintln(cmd.OutOrStdout())

		if !result.IsGo() {
			nogoCount++
		}
	}

	if nogoCount > 0 {
		return fmt.Errorf("dry-run: %d SDDs with NO-GO", nogoCount)
	}
	return nil
}

func init() {
	batchCmd.Flags().BoolVar(&batchDryRun, "dry-run", false, "Simulate execution without modifying files")
	batchCmd.Flags().StringVar(&batchRunner, "runner", "", "Runner backend (claude, openhands)")
	batchCmd.Flags().StringVar(&batchSandbox, "sandbox", "", "Sandbox provider (none, docker, apple-container)")
	rootCmd.AddCommand(batchCmd)
}
