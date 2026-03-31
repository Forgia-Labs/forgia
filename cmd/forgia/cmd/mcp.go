package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	forgia "github.com/Deepzima/forgia"
	"github.com/Deepzima/forgia/internal/config"
	"github.com/Deepzima/forgia/internal/guardrails"
	"github.com/Deepzima/forgia/internal/mcp"
	"github.com/Deepzima/forgia/internal/skill"
	"github.com/Deepzima/forgia/internal/vault"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server commands",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start Forgia as an MCP server on stdio",
	Long:  "Starts a JSON-RPC 2.0 MCP server reading from stdin and writing to stdout. Blocks until SIGINT/SIGTERM or stdin EOF.",
	RunE:  runMCPServe,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpServeCmd)
}

func runMCPServe(cmd *cobra.Command, args []string) error {
	logger := slog.With("command", "mcp-serve")

	// Context with signal-based cancellation.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 1. Open vault.
	v, err := vault.Open(".")
	if err != nil {
		return fmt.Errorf("mcp serve: open vault: %w", err)
	}
	logger.InfoContext(ctx, "vault opened", "dir", v.Dir())

	// 2. Load config.
	cfg, err := config.LoadConfig(ctx, v.Dir())
	if err != nil {
		return fmt.Errorf("mcp serve: load config: %w", err)
	}
	logger.InfoContext(ctx, "config loaded")

	// 3. Wire providers (create + start).
	providers, err := mcp.WireProviders(ctx, cfg)
	if err != nil {
		return fmt.Errorf("mcp serve: wire providers: %w", err)
	}
	defer providers.StopAll()
	logger.InfoContext(ctx, "providers wired")

	// 4. Load skill registry with embedded slash commands.
	skills := skill.NewRegistry()
	if err := skills.LoadEmbedded(forgia.SlashCommandFS()); err != nil {
		return fmt.Errorf("mcp serve: load embedded skills: %w", err)
	}
	logger.InfoContext(ctx, "embedded skills loaded")

	// 5. Register composite skills (knowledge graph wrappers).
	if err := skill.RegisterCompositeSkills(skills, providers, v.(skill.VaultReader)); err != nil {
		return fmt.Errorf("mcp serve: register composite skills: %w", err)
	}
	logger.InfoContext(ctx, "composite skills registered")

	// 6. Register vault provider (FD/SDD CRUD tools).
	var g *guardrails.Guardrails
	if gData, err := v.GuardrailsRaw(ctx); err == nil {
		g, _ = guardrails.Parse(gData)
	}
	vaultProvider := mcp.NewVaultProvider(v, g, providers)
	providers.Register(vaultProvider)
	logger.InfoContext(ctx, "vault provider registered")

	// 7. Create MCP server.
	transport := mcp.NewStdioTransport(os.Stdin, os.Stdout)
	server := mcp.NewMCPServer(transport, providers, skills)

	// 8. Serve — blocks until signal or stdin EOF.
	logger.InfoContext(ctx, "starting MCP server on stdio")

	if err := server.Serve(ctx); err != nil {
		if ctx.Err() != nil {
			// 9. Graceful shutdown via signal — providers stopped by defer.
			logger.InfoContext(ctx, "MCP server stopped (signal)")
			return nil
		}
		return fmt.Errorf("mcp serve: %w", err)
	}

	logger.InfoContext(ctx, "MCP server stopped (stdin EOF)")
	return nil
}
