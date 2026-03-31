package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgia-labs/forgia/internal/beads"
	"github.com/forgia-labs/forgia/internal/config"
	"github.com/forgia-labs/forgia/internal/guardrails"
	"github.com/forgia-labs/forgia/internal/runner"
	"github.com/forgia-labs/forgia/internal/vault"
	"github.com/spf13/cobra"
)

var (
	execDryRun bool
	execRunner string
	execMode   string
)

var execCmd = &cobra.Command{
	Use:   "exec <sdd-file>",
	Short: "Execute an SDD (or simulate with --dry-run)",
	Long:  "Validates, checks guardrails, then executes an SDD via the configured runner.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return execSDD(cmd, args[0])
	},
}

func init() {
	execCmd.Flags().BoolVar(&execDryRun, "dry-run", false, "Simulate execution without modifying files")
	execCmd.Flags().StringVar(&execRunner, "runner", "", "Runner backend (claude, dry-run)")
	execCmd.Flags().StringVar(&execMode, "mode", "", "Guardrail mode (off, careful, freeze, guard)")
	rootCmd.AddCommand(execCmd)
}

// execSDD is the shared execution logic used by both exec and batch commands.
func execSDD(cmd *cobra.Command, sddFile string) error {
	ctx := cmd.Context()
	logger := slog.With("command", "exec")

	if _, err := os.Stat(sddFile); os.IsNotExist(err) {
		return fmt.Errorf("SDD not found: %s", sddFile)
	}

	// Open vault.
	v, err := vault.Open(".")
	if err != nil {
		return fmt.Errorf("vault: %w", err)
	}

	// Load config.
	cfg, err := config.LoadConfig(ctx, v.Dir())
	if err != nil {
		logger.WarnContext(ctx, "config not loaded, using defaults", "error", err)
	}

	// Pre-exec validation.
	gData, _ := v.GuardrailsRaw(ctx)
	g, _ := guardrails.Parse(gData)
	if g == nil {
		g = &guardrails.Guardrails{}
	}

	errors := validateSDD(ctx, v, g, sddFile)
	if len(errors) > 0 {
		fmt.Println("Validation failed:")
		for _, e := range errors {
			fmt.Printf("  %s\n", e)
		}
		return fmt.Errorf("validation failed: %d error(s)", len(errors))
	}

	// Build SDD struct.
	data, _ := os.ReadFile(sddFile)
	content := string(data)
	sddID := strings.Trim(extractFrontmatterValue(content, "id:"), "\"' ")
	fdID := strings.Trim(extractFrontmatterValue(content, "fd:"), "\"' ")

	sdd, err := v.GetSDD(ctx, fdID, sddID)
	if err != nil {
		sdd = &vault.SDD{ID: sddID, FD: fdID, FilePath: sddFile}
	}

	// Dry-run path.
	if execDryRun {
		slashCmd := filepath.Join("modules", "claude-commands", "sdd-dry-run.md")
		if _, err := os.Stat(slashCmd); os.IsNotExist(err) {
			// Fallback: use DryRunRunner.
			r := runner.NewDryRunRunner()
			result, err := r.Execute(ctx, sdd, runner.ExecOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("Dry-run: %s — %s\n", result.SDD, result.Status)
			return nil
		}
		result, err := runner.DryRun(ctx, sdd, runner.DryRunOptions{SlashCommandPath: slashCmd})
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), result.ReportText)
		if !result.IsGo() {
			return fmt.Errorf("dry-run: NO-GO")
		}
		return nil
	}

	// Guardrails enforcement.
	mode := guardrails.ParseMode(execMode)
	if mode != guardrails.ModeOff {
		violations := g.Enforce(ctx, mode, guardrails.EnforceOpts{
			WriteDirs:     sdd.Boundaries.WriteDirs,
			ForbiddenDirs: sdd.Boundaries.ForbiddenDirs,
		})
		if len(violations) > 0 {
			fmt.Println("Guardrail violations:")
			for _, v := range violations {
				fmt.Printf("  %s\n", v.Error())
			}
			return fmt.Errorf("guardrails blocked execution: %d violation(s)", len(violations))
		}
	}

	// Resolve runner.
	resolvedRunner := execRunner
	if resolvedRunner == "" && cfg != nil {
		resolvedRunner = cfg.Runner.Default
	}
	if resolvedRunner == "" {
		resolvedRunner = "claude"
	}

	var ohCfg config.OpenHandsConfig
	if cfg != nil {
		ohCfg = cfg.Runner.OpenHands
	}
	r, err := runner.Resolve(resolvedRunner, ohCfg)
	if err != nil {
		return err
	}

	// Build system context.
	systemCtx, err := runner.BuildSystemContext(ctx, v)
	if err != nil {
		logger.WarnContext(ctx, "failed to build system context", "error", err)
	}

	maxTurns := 200
	if cfg != nil && cfg.Runner.Claude.MaxTurns > 0 {
		maxTurns = cfg.Runner.Claude.MaxTurns
	}

	opts := runner.ExecOptions{
		PermissionMode: "auto",
		MaxTurns:       maxTurns,
		SystemContext:   systemCtx,
	}

	fmt.Printf("→ Executing %s with runner: %s\n", sddID, r.Name())
	result, execErr := r.Execute(ctx, sdd, opts)

	if result != nil {
		fmt.Printf("\n=== Execution Complete ===\n")
		fmt.Printf("  SDD:      %s\n", result.SDD)
		fmt.Printf("  Duration: %ds\n", result.DurationSecs)
		fmt.Printf("  Status:   %s\n", result.Status)

		// Ensure File is set on the result for the JSON report.
		if result.File == "" {
			result.File = sddFile
		}

		// Write JSON execution report.
		logsDir := filepath.Join(".forgia", "logs")
		reportPath, reportErr := writeExecReport(result, logsDir)
		if reportErr != nil {
			logger.WarnContext(ctx, "failed to write exec report", "error", reportErr)
		} else {
			fmt.Printf("  Report:   %s\n", reportPath)
		}
	}

	// Update SDD status.
	if result != nil && result.Status == "success" {
		sdd.Status = vault.SDDDone
		if err := v.UpdateSDD(ctx, sdd); err != nil {
			logger.WarnContext(ctx, "failed to update SDD status", "error", err)
		}

		// Close matching Beads task.
		bc := beads.NewClient()
		closeBeadsTask(ctx, bc, sddID)
	}

	if execErr != nil {
		return fmt.Errorf("execution failed: %w", execErr)
	}
	return nil
}
