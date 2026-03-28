package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	forgia "github.com/Deepzima/forgia"
	"github.com/Deepzima/forgia/internal/beads"
	"github.com/Deepzima/forgia/internal/config"
	"github.com/Deepzima/forgia/internal/knowledge"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Forgia vault in the current project",
	Long:  "Scaffolds .forgia/ with templates, dev-guide, guardrails, and optional beads integration.",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().String("dir", ".", "Target directory")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	dir, _ := cmd.Flags().GetString("dir")
	dir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	logger := slog.With("command", "init")
	forgiaDir := filepath.Join(dir, ".forgia")

	// Check if vault already exists.
	if info, err := os.Stat(forgiaDir); err == nil && info.IsDir() {
		logger.InfoContext(ctx, "vault already exists, updating templates", "dir", forgiaDir)
	} else {
		logger.InfoContext(ctx, "scaffolding vault", "dir", forgiaDir)
	}

	// Copy all files from embedded vault template.
	templateFS := forgia.VaultTemplateFS()
	created, skipped, err := copyEmbeddedFS(ctx, templateFS, forgiaDir)
	if err != nil {
		return fmt.Errorf("copy templates: %w", err)
	}
	logger.InfoContext(ctx, "templates copied", "created", created, "skipped", skipped)

	// Create required directories that may be empty (not in templates).
	for _, d := range []string{"logs"} {
		if err := os.MkdirAll(filepath.Join(forgiaDir, d), 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", d, err)
		}
	}

	// Auto-detect project stack and copy language conventions.
	langs := detectStack(dir)
	logger.InfoContext(ctx, "detected languages", "languages", langs)

	// Create .gitignore for .forgia/.
	gitignorePath := filepath.Join(forgiaDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignore := "# Local-only files\nlogs/\nrun/\n*.pid\n.beads/\n"
		if err := os.WriteFile(gitignorePath, []byte(gitignore), 0o644); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
		logger.InfoContext(ctx, "created .gitignore")
	}

	// Create .github/CODEOWNERS only if .github/ already exists.
	githubDir := filepath.Join(dir, ".github")
	if info, err := os.Stat(githubDir); err == nil && info.IsDir() {
		codeownersPath := filepath.Join(githubDir, "CODEOWNERS")
		if _, err := os.Stat(codeownersPath); os.IsNotExist(err) {
			owner := gitUserName()
			if owner == "" {
				owner = "@owner"
			}
			codeowners := fmt.Sprintf(
				"# Forgia vault — critical files require review\n"+
					".forgia/constitution.md %s\n"+
					".forgia/guardrails/ %s\n"+
					".forgia/architecture/ %s\n"+
					".forgia/contexts/ %s\n"+
					".forgia/config.toml %s\n"+
					"# FDs are open to all contributors\n"+
					"# .forgia/fd/\n",
				owner, owner, owner, owner, owner,
			)
			os.WriteFile(codeownersPath, []byte(codeowners), 0o644)
			logger.InfoContext(ctx, "created CODEOWNERS")
		}
	}

	// Beads init (optional).
	initBeads(ctx, dir, logger)

	// Knowledge layer init (optional).
	initKnowledge(ctx, dir, forgiaDir, logger)

	// Summary.
	fmt.Println()
	fmt.Printf("✓ Vault scaffolded at %s\n", forgiaDir)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit .forgia/constitution.md with your project rules")
	fmt.Println("  2. Review .forgia/dev-guide/lang/ conventions for your stack")
	fmt.Println("  3. Commit .forgia/ to share with your team")
	fmt.Println("  4. Use: forgia skill fd-new to create your first Feature Design")

	return nil
}

// copyEmbeddedFS copies all files from an embedded FS to a target directory.
// Skips files that already exist (idempotent). Returns counts of created and skipped files.
func copyEmbeddedFS(ctx context.Context, src fs.FS, targetDir string) (created, skipped int, err error) {
	return created, skipped, fs.WalkDir(src, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		targetPath := filepath.Join(targetDir, path)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		// Skip if file already exists (idempotent).
		if _, statErr := os.Stat(targetPath); statErr == nil {
			skipped++
			return nil
		}

		// Ensure parent directory exists.
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("mkdir for %s: %w", path, err)
		}

		data, err := fs.ReadFile(src, path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", targetPath, err)
		}

		created++
		slog.InfoContext(ctx, "created", "file", path)
		return nil
	})
}

// detectStack checks for language-specific files and returns detected languages.
func detectStack(dir string) []string {
	detectors := map[string][]string{
		"go":     {"go.mod"},
		"rust":   {"Cargo.toml"},
		"python": {"pyproject.toml", "requirements.txt", "setup.py"},
		"node":   {"package.json"},
		"shell":  {"Makefile", "mise.toml"},
	}

	var langs []string
	for lang, files := range detectors {
		for _, f := range files {
			if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
				langs = append(langs, lang)
				break
			}
		}
	}

	if len(langs) == 0 {
		langs = append(langs, "shell") // fallback
	}

	return langs
}

// initBeads tries to initialize Beads if bd is available.
func initBeads(ctx context.Context, dir string, logger *slog.Logger) {
	bd := beads.NewClient()
	if !bd.Available() {
		logger.InfoContext(ctx, "beads not available, skipping")
		return
	}

	// Check if already initialized.
	beadsDir := filepath.Join(dir, ".beads")
	if _, err := os.Stat(beadsDir); err == nil {
		logger.InfoContext(ctx, "beads already initialized")
		return
	}

	logger.InfoContext(ctx, "initializing beads")
	initCmd := exec.CommandContext(ctx, "bd", "init")
	initCmd.Dir = dir
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		logger.WarnContext(ctx, "beads init failed (optional)", "error", err)
	}
}

// initKnowledge auto-indexes the codebase and creates .mcp.json if codebase-memory-mcp is available.
func initKnowledge(ctx context.Context, dir, forgiaDir string, logger *slog.Logger) {
	kc := knowledge.NewClient()
	if !kc.Available() {
		logger.InfoContext(ctx, "codebase-memory-mcp not available, skipping knowledge layer")
		return
	}

	// Load config to check auto_index setting.
	cfg, err := config.LoadConfig(ctx, forgiaDir)
	if err != nil {
		logger.WarnContext(ctx, "failed to load config for knowledge layer", "error", err)
		return
	}

	if cfg.Knowledge.AutoIndex {
		logger.InfoContext(ctx, "indexing codebase with codebase-memory-mcp")
		symbols, err := kc.Index(ctx)
		if err != nil {
			logger.WarnContext(ctx, "codebase indexing failed (optional)", "error", err)
		} else {
			fmt.Printf("  Knowledge: indexed %d symbols\n", symbols)
		}
	}

	// Create/merge .mcp.json.
	if err := knowledge.EnsureMCPJSON(ctx, dir); err != nil {
		logger.WarnContext(ctx, "failed to create/update .mcp.json", "error", err)
	} else {
		logger.InfoContext(ctx, "ensured .mcp.json has codebase-memory-mcp config")
	}
}

// stackString joins detected languages for display.
// gitUserName returns the current git user.name, or empty string if unavailable.
func gitUserName() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func stackString(langs []string) string {
	return strings.Join(langs, ", ")
}
