package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Deepzima/forgia/internal/runner"
	"github.com/Deepzima/forgia/internal/vault"
	"github.com/spf13/cobra"
)

var execDryRun bool
var execRunner string

var execCmd = &cobra.Command{
	Use:   "exec <sdd-file> [--runner=claude|openhands] [--dry-run]",
	Short: "Execute an SDD (or simulate with --dry-run)",
	Long:  "Execute an SDD using the specified runner. With --dry-run, runs a feasibility simulation via Claude without modifying any files.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sddFile := args[0]

		if _, err := os.Stat(sddFile); os.IsNotExist(err) {
			return fmt.Errorf("SDD non trovato: %s", sddFile)
		}

		if execDryRun {
			if execRunner != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Nota: --runner ignorato durante dry-run (la simulazione usa Claude)")
			}
			return runExecDryRun(cmd, sddFile)
		}

		// Normal execution
		resolvedRunner := execRunner
		if resolvedRunner == "" {
			resolvedRunner = "claude"
		}

		r, err := runner.Resolve(resolvedRunner)
		if err != nil {
			return err
		}

		sdd := &vault.SDD{FilePath: sddFile}
		result, err := r.Execute(cmd.Context(), sdd, runner.ExecOptions{})
		if err != nil {
			return fmt.Errorf("esecuzione fallita: %w", err)
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("runner terminato con codice %d", result.ExitCode)
		}
		return nil
	},
}

func runExecDryRun(cmd *cobra.Command, sddFile string) error {
	slashCmd := filepath.Join("modules", "claude-commands", "sdd-dry-run.md")
	if _, err := os.Stat(slashCmd); os.IsNotExist(err) {
		return fmt.Errorf("sdd-dry-run.md non trovato — esegui prima SDD-001 di FD-012")
	}

	sdd := &vault.SDD{FilePath: sddFile}

	result, err := runner.DryRun(cmd.Context(), sdd, runner.DryRunOptions{
		SlashCommandPath: slashCmd,
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), result.ReportText)

	if !result.IsGo() {
		return fmt.Errorf("dry-run: NO-GO")
	}
	return nil
}

func init() {
	execCmd.Flags().BoolVar(&execDryRun, "dry-run", false, "Simula esecuzione senza modificare file")
	execCmd.Flags().StringVar(&execRunner, "runner", "", "Runner da usare (claude|openhands)")
	rootCmd.AddCommand(execCmd)
}
