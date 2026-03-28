package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Deepzima/forgia/internal/beads"
	"github.com/Deepzima/forgia/internal/runner"
	"github.com/Deepzima/forgia/internal/vault"
	"github.com/spf13/cobra"
)

var batchDryRun bool
var batchRunner string

var batchCmd = &cobra.Command{
	Use:   "batch <FD-NNN> [--runner=claude|openhands] [--dry-run]",
	Short: "Execute all pending SDDs for an FD (or simulate with --dry-run)",
	Long:  "Execute all pending SDDs for the given FD. With --dry-run, runs feasibility simulation on each SDD without modifying any files.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fdID := args[0]
		sddDir := filepath.Join(".forgia", "sdd", fdID)

		if _, err := os.Stat(sddDir); os.IsNotExist(err) {
			return fmt.Errorf("nessun SDD trovato per %s in %s", fdID, sddDir)
		}

		entries, err := os.ReadDir(sddDir)
		if err != nil {
			return fmt.Errorf("lettura directory %s: %w", sddDir, err)
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
			fmt.Fprintln(cmd.OutOrStdout(), "Nessun SDD pendente per", fdID)
			return nil
		}

		if batchDryRun {
			if batchRunner != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Nota: --runner ignorato durante dry-run (la simulazione usa Claude)")
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

		var completed, failed int
		for _, sddFile := range pendingSDDs {
			data, _ := os.ReadFile(sddFile)
			content := string(data)
			sddID := strings.Trim(extractFrontmatterValue(content, "id:"), "\"' ")
			sddFD := strings.Trim(extractFrontmatterValue(content, "fd:"), "\"' ")

			sdd := &vault.SDD{ID: sddID, FD: sddFD, FilePath: sddFile}
			result, execErr := r.Execute(ctx, sdd, runner.ExecOptions{})

			// Write JSON report for every execution.
			if result != nil {
				if result.File == "" {
					result.File = sddFile
				}
				if _, reportErr := writeExecReport(result, logsDir); reportErr != nil {
					logger.WarnContext(ctx, "failed to write exec report", "sdd", sddID, "error", reportErr)
				}
			}

			if execErr != nil || (result != nil && result.ExitCode != 0) {
				failed++
				fmt.Fprintf(cmd.ErrOrStderr(), "Avviso: %s fallito\n", sddFile)
			} else {
				completed++
				closeBeadsTask(ctx, bc, sddID)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Completati: %d, Falliti: %d\n", completed, failed)
		if failed > 0 {
			return fmt.Errorf("%d SDD falliti", failed)
		}
		return nil
	},
}

func runBatchDryRun(cmd *cobra.Command, fdID string, sddFiles []string) error {
	slashCmd := filepath.Join("modules", "claude-commands", "sdd-dry-run.md")
	if _, err := os.Stat(slashCmd); os.IsNotExist(err) {
		return fmt.Errorf("sdd-dry-run.md non trovato — esegui prima SDD-001 di FD-012")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "=== FD Dry Run: %s (%d SDDs) ===\n\n", fdID, len(sddFiles))

	var nogoCount int
	for _, sddFile := range sddFiles {
		sdd := &vault.SDD{FilePath: sddFile}
		result, err := runner.DryRun(cmd.Context(), sdd, runner.DryRunOptions{
			SlashCommandPath: slashCmd,
		})
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Errore dry-run %s: %v\n", sddFile, err)
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
		return fmt.Errorf("dry-run: %d SDD con NO-GO", nogoCount)
	}
	return nil
}

func init() {
	batchCmd.Flags().BoolVar(&batchDryRun, "dry-run", false, "Simula esecuzione senza modificare file")
	batchCmd.Flags().StringVar(&batchRunner, "runner", "", "Runner da usare (claude|openhands)")
	rootCmd.AddCommand(batchCmd)
}
