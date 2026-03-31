package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	forgia "github.com/forgia-labs/forgia"
	"github.com/forgia-labs/forgia/internal/config"
	"github.com/forgia-labs/forgia/internal/mcp"
	"github.com/forgia-labs/forgia/internal/runner"
	"github.com/forgia-labs/forgia/internal/skill"
	"github.com/forgia-labs/forgia/internal/vault"
	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill <name> [args...]",
	Short: "Execute an embedded slash command via Claude",
	Long:  "Runs a slash command (e.g., fd-new, fd-review) with system context automatically injected.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSkill,
}

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "List all available skills (slash commands + MCP tools)",
	RunE:  runSkills,
}

func init() {
	rootCmd.AddCommand(skillCmd)
	rootCmd.AddCommand(skillsCmd)
}

func loadRegistry() (*skill.Registry, error) {
	reg := skill.NewRegistry()
	if err := reg.LoadEmbedded(forgia.SlashCommandFS()); err != nil {
		return nil, fmt.Errorf("load embedded skills: %w", err)
	}
	return reg, nil
}

// loadFullRegistry loads embedded skills and, if a vault + config exist,
// also registers composite skills backed by MCP providers.
func loadFullRegistry(ctx context.Context) (*skill.Registry, error) {
	reg, err := loadRegistry()
	if err != nil {
		return nil, err
	}

	v, vErr := vault.Open(".")
	if vErr != nil {
		return reg, nil
	}

	cfg, cfgErr := config.LoadConfig(ctx, v.Dir())
	if cfgErr != nil {
		return reg, nil
	}

	providers := mcp.NewProviderRegistry()
	// Register providers without starting them — we only need tool metadata for listing.
	for name, pcfg := range cfg.MCP.Providers {
		if pcfg.Namespace == "" {
			pcfg.Namespace = name
		}
		providers.Register(mcp.NewMetadataProvider(name, pcfg.Namespace))
	}
	if cfg.Knowledge.Provider != "" {
		if _, exists := cfg.MCP.Providers["code"]; !exists {
			providers.Register(mcp.NewMetadataProvider("code", "code"))
		}
	}

	_ = skill.RegisterCompositeSkills(reg, providers, v.(skill.VaultReader))
	return reg, nil
}

func runSkills(cmd *cobra.Command, args []string) error {
	reg, err := loadFullRegistry(cmd.Context())
	if err != nil {
		return err
	}

	fmt.Println("=== Available Skills ===")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "NAME\tMODE\tCATEGORY\tDESCRIPTION\n")
	fmt.Fprintf(w, "────\t────\t────────\t───────────\n")

	for _, s := range reg.All() {
		mode := "Slash Command"
		if s.Mode() == skill.ModeMCPTool {
			mode = "MCP Tool"
		} else if s.Mode() == skill.ModeBoth {
			mode = "Both"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name(), mode, s.Category(), s.Description())
	}
	w.Flush()

	fmt.Printf("\nTotal: %d skills\n", len(reg.All()))
	fmt.Println("Usage: forgia skill <name> [args]")
	return nil
}

func runSkill(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	logger := slog.With("command", "skill")

	name := args[0]
	skillArgs := ""
	if len(args) > 1 {
		skillArgs = strings.Join(args[1:], " ")
	}

	reg, err := loadRegistry()
	if err != nil {
		return err
	}

	s, ok := reg.Get(name)
	if !ok {
		fmt.Printf("Unknown skill: %q\n\nAvailable skills:\n", name)
		for _, s := range reg.All() {
			fmt.Printf("  %s\n", s.Name())
		}
		return fmt.Errorf("skill %q not found", name)
	}

	// Get the embedded content.
	embedded, ok := s.(*skill.EmbeddedSkill)
	if !ok {
		return fmt.Errorf("skill %q is not an embedded slash command", name)
	}

	// Build content with args substituted.
	content := embedded.Content(skillArgs)

	// Build system context from vault (if available).
	var systemCtx string
	v, err := vault.Open(".")
	if err == nil {
		systemCtx, _ = runner.BuildSystemContext(ctx, v)
	} else {
		logger.InfoContext(ctx, "vault not available, running skill without system context")
	}

	// Verify claude is available.
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found: %w", err)
	}

	// Build claude args.
	claudeArgs := []string{}
	if systemCtx != "" {
		claudeArgs = append(claudeArgs, "--append-system-prompt", systemCtx)
	}
	claudeArgs = append(claudeArgs, "-p", content)

	logger.InfoContext(ctx, "executing skill", "name", name, "args", skillArgs)

	// Run claude.
	c := exec.CommandContext(ctx, claudePath, claudeArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	return c.Run()
}
