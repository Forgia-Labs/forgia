package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Deepzima/forgia/internal/mcp"
	"github.com/Deepzima/forgia/internal/skill"
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

	// Create provider registry.
	providers := mcp.NewProviderRegistry()

	// Create skill registry.
	// Future SDDs will register composite skills here.
	skills := skill.NewRegistry()

	// Create transport on stdio.
	transport := mcp.NewStdioTransport(os.Stdin, os.Stdout)

	// Create and start server.
	server := mcp.NewMCPServer(transport, providers, skills)

	// Context with signal-based cancellation.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.InfoContext(ctx, "starting MCP server on stdio")

	if err := server.Serve(ctx); err != nil {
		if ctx.Err() != nil {
			// Graceful shutdown via signal.
			logger.InfoContext(ctx, "MCP server stopped")
			return nil
		}
		return fmt.Errorf("mcp serve: %w", err)
	}

	logger.InfoContext(ctx, "MCP server stopped (stdin EOF)")
	return nil
}
