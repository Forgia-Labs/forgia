package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/Deepzima/forgia/internal/beads"
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

  forgia sync                  # push + pull (default, both directions)
  forgia sync --push           # vault → board only (upload local changes)
  forgia sync --pull           # board → vault only (download board changes)
  forgia sync --direction=push # same as --push (long form)

Requires [board] section in .forgia/config.toml with provider, owner, and project number.`,
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
	logger := slog.With("command", "sync")

	// Find vault.
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	v, err := vault.Open(dir)
	if err != nil {
		return fmt.Errorf("open vault at %s: %w", dir, err)
	}

	// Load config from vault directory.
	cfg, err := config.LoadConfig(ctx, v.Dir())
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Validate board config.
	if cfg.Board.Provider == "" {
		return fmt.Errorf("board provider not configured in .forgia/config.toml (set [board] provider)")
	}
	logger.Info("resolving board",
		"provider", cfg.Board.Provider,
		"owner", cfg.Board.GitHub.Owner,
		"project", cfg.Board.GitHub.ProjectNumber)

	// Resolve board.
	b, err := board.Resolve(ctx,
		cfg.Board.Provider,
		cfg.Board.GitHub.Owner,
		cfg.Board.GitHub.ProjectNumber,
	)
	if err != nil {
		return fmt.Errorf("resolve %s board: %w", cfg.Board.Provider, err)
	}

	// Wire adapter.
	adapter := boardsync.NewVaultAdapter(v)
	if gb, ok := b.(*board.GitHubBoard); ok {
		gb.SetVault(adapter, adapter)
	} else {
		logger.Warn("board type does not support vault sync via adapter", "type", fmt.Sprintf("%T", b))
	}

	// Initialize Beads cache (optional — works without it).
	bd := beads.NewClient()
	if bd.Available() {
		logger.Info("beads cache available")
	}

	// Determine direction.
	push, err := cmd.Flags().GetBool("push")
	if err != nil {
		return fmt.Errorf("parse --push flag: %w", err)
	}
	pull, err := cmd.Flags().GetBool("pull")
	if err != nil {
		return fmt.Errorf("parse --pull flag: %w", err)
	}
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
		logger.Info("syncing vault → board")
		if err := b.SyncFromVault(ctx); err != nil {
			return fmt.Errorf("push sync: %w", err)
		}
		// Cache card mappings in Beads after push.
		cacheCardMappings(ctx, bd, b)
		logger.Info("push complete")
	}

	if pull {
		logger.Info("syncing board → vault")
		if err := b.SyncToVault(ctx); err != nil {
			return fmt.Errorf("pull sync: %w", err)
		}
		// Update cache after pull.
		cacheCardMappings(ctx, bd, b)
		logger.Info("pull complete")
	}

	return nil
}

// cacheCardMappings stores vault→board ID mappings in Beads for offline access.
// Fetches actual board cards to get real card IDs (not vault IDs).
func cacheCardMappings(ctx context.Context, bd *beads.Client, b board.ProjectBoard) {
	if !bd.Available() {
		return
	}

	cards, err := b.GetCards(ctx, board.CardFilter{})
	if err != nil {
		slog.WarnContext(ctx, "failed to read board cards for beads cache", "error", err)
		return
	}

	for _, card := range cards {
		vaultID := board.ExtractIDFromTitle(card.Title)
		if vaultID == "" {
			continue
		}
		if err := bd.CacheCardMapping(ctx, beads.CardMapping{
			VaultID: vaultID,
			CardID:  card.ID,
			Column:  card.Column,
		}); err != nil {
			continue
		}
	}
}
