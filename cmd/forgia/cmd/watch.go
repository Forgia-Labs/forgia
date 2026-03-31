package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/forgia-labs/forgia/internal/vault"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	watchDebounce time.Duration
	watchRunner   string
	watchMode     string
)

var watchCmd = &cobra.Command{
	Use:   "watch <FD-NNN>",
	Short: "Watch and auto-execute new SDDs",
	Long:  "Watches .forgia/sdd/<FD-NNN>/ for new or modified SDDs and executes them automatically.",
	Args:  cobra.ExactArgs(1),
	RunE:  runWatch,
}

func init() {
	watchCmd.Flags().DurationVar(&watchDebounce, "debounce", 5*time.Second, "Debounce period before execution")
	watchCmd.Flags().StringVar(&watchRunner, "runner", "", "Runner backend (claude, dry-run)")
	watchCmd.Flags().StringVar(&watchMode, "mode", "", "Guardrail mode (off, careful, freeze, guard)")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	fdID := args[0]
	logger := slog.With("command", "watch", "fd", fdID)

	v, err := vault.Open(".")
	if err != nil {
		return fmt.Errorf("vault: %w", err)
	}

	sddDir := filepath.Join(v.Dir(), "sdd", fdID)
	if _, err := os.Stat(sddDir); os.IsNotExist(err) {
		return fmt.Errorf("SDD directory not found: %s", sddDir)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(sddDir); err != nil {
		return fmt.Errorf("watch directory: %w", err)
	}

	fmt.Println("=== Forgia Watcher ===")
	fmt.Printf("  Watching:  %s\n", sddDir)
	fmt.Printf("  Runner:    %s\n", watchRunner)
	fmt.Printf("  Debounce:  %s\n", watchDebounce)
	fmt.Println()
	fmt.Println("  Waiting for new SDDs... (Ctrl+C to stop)")
	fmt.Println()

	// Execute any pending SDDs at startup.
	executePending(ctx, cmd, v, sddDir, logger)

	// Setup signal handling.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Debounce timer.
	var debounceTimer *time.Timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only react to create/write on SDD .md files.
			if !isSDDFile(event.Name) {
				continue
			}
			if event.Op&(fsnotify.Create|fsnotify.Write) == 0 {
				continue
			}

			logger.InfoContext(ctx, "file change detected", "file", filepath.Base(event.Name))

			// Reset debounce timer.
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(watchDebounce, func() {
				executePending(ctx, cmd, v, sddDir, logger)
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			logger.WarnContext(ctx, "watcher error", "error", err)

		case <-sigCh:
			fmt.Println("\n  Stopping watcher...")
			return nil
		}
	}
}

// executePending finds and executes all planned SDDs in the directory.
func executePending(ctx context.Context, cmd *cobra.Command, v vault.Vault, sddDir string, logger *slog.Logger) {
	entries, err := os.ReadDir(sddDir)
	if err != nil {
		logger.WarnContext(ctx, "failed to read SDD dir", "error", err)
		return
	}

	var pending []string
	for _, e := range entries {
		if !isSDDFile(e.Name()) {
			continue
		}
		path := filepath.Join(sddDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		status := extractFrontmatterValue(string(data), "status:")
		status = strings.Trim(status, "\"' ")
		if status == "planned" || status == "validated" {
			pending = append(pending, path)
		}
	}

	if len(pending) == 0 {
		return
	}

	fmt.Printf("  Found %d pending SDD(s) — executing now...\n\n", len(pending))

	for _, sddFile := range pending {
		// Temporarily set exec flags for shared logic.
		oldRunner, oldMode := execRunner, execMode
		execRunner, execMode = watchRunner, watchMode

		if err := execSDD(cmd, sddFile); err != nil {
			logger.WarnContext(ctx, "execution failed", "sdd", filepath.Base(sddFile), "error", err)
		}

		execRunner, execMode = oldRunner, oldMode
	}
}

// isSDDFile returns true for SDD markdown files, false for temp files.
func isSDDFile(name string) bool {
	base := filepath.Base(name)
	if !strings.HasPrefix(base, "SDD-") {
		return false
	}
	if !strings.HasSuffix(base, ".md") && !strings.HasSuffix(base, ".yaml") {
		return false
	}
	// Skip temp files.
	if strings.HasSuffix(base, ".swp") || strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".tmp") {
		return false
	}
	return true
}
