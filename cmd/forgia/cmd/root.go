package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "forgia",
	Short: "Spec-driven development framework",
	Long:  "Forgia — spec-driven development framework. Forge specs into code.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
