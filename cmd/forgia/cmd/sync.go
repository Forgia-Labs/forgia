package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Deepzima/forgia/internal/board"
	"github.com/Deepzima/forgia/internal/boardsync"
	"github.com/Deepzima/forgia/internal/config"
	"github.com/Deepzima/forgia/internal/vault"
)

var (
	syncDirection string // "push", "pull", or "" (both)
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync FD/SDD status between vault and project board",
	Long: `Bidirectional sync between .forgia/ files and the project board (GitHub Projects).

  forgia sync          # push + pull (default)
  forgia sync --push   # vault → board only
  forgia sync --pull   # board → vault only`,
	RunE: runSync,
}

func init() {
	syncCmd.Flags().StringVar(&syncDirection, "direction", "", "sync direction: push, pull, or both (default)")
	syncCmd.Flags().Bool("push", false, "push vault state to board (shorthand for --direction=push)")
	syncCmd.Flags().Bool("pull", false, "pull board state to vault (shorthand for --direction=pull)")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	// Find vault.
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	v, err := vault.Open(dir)
	if err != nil {
		return err
	}

	// Load config.
	forgiaDir := dir + "/.forgia"
	cfg, err := config.LoadConfig(ctx, forgiaDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Resolve board.
	b, err := board.Resolve(ctx,
		cfg.Board.Provider,
		cfg.Board.GitHub.Owner,
		cfg.Board.GitHub.ProjectNumber,
	)
	if err != nil {
		return fmt.Errorf("resolve board: %w", err)
	}

	// Wire adapter.
	adapter := boardsync.NewVaultAdapter(v)
	if gb, ok := b.(*board.GitHubBoard); ok {
		gb.SetVault(adapter, adapter)
	}

	// Determine direction.
	push, _ := cmd.Flags().GetBool("push")
	pull, _ := cmd.Flags().GetBool("pull")
	if syncDirection == "push" {
		push = true
	}
	if syncDirection == "pull" {
		pull = true
	}
	// Default: both.
	if !push && !pull {
		push = true
		pull = true
	}

	if push {
		fmt.Fprintln(cmd.OutOrStdout(), "Syncing vault → board...")
		if err := b.SyncFromVault(ctx); err != nil {
			return fmt.Errorf("push sync: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Push complete.")
	}

	if pull {
		fmt.Fprintln(cmd.OutOrStdout(), "Syncing board → vault...")
		if err := b.SyncToVault(ctx); err != nil {
			return fmt.Errorf("pull sync: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Pull complete.")
	}

	return nil
}
