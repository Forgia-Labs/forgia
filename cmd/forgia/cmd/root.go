package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "forgia",
	Short: "Spec-driven development framework",
	Long: `Usage: forgia <command> [arguments]

Commands:
  init                          Initialize .forgia/ vault
  status                        Show FD + SDD dashboard
  doctor                        Check health (docker, openhands, vault)
  validate <sdd-file|FD-NNN>    Validate an SDD or all SDDs for an FD
  exec <sdd-file> [--runner=X]  Execute an SDD
  batch <FD-NNN> [--runner=X]   Execute all SDDs for an FD
  watch <FD-NNN> [--runner=X]   Watch and auto-execute new SDDs
  version                       Show version`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
