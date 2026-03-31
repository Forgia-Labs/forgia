package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/forgia-labs/forgia/internal/board"
)

// BoardProvider exposes project board operations as MCP tools.
// Tools: forgia_board_sync, forgia_board_status.
type BoardProvider struct {
	board board.ProjectBoard
}

// Compile-time interface satisfaction check.
var _ ToolProvider = (*BoardProvider)(nil)

// NewBoardProvider creates an MCP provider for board operations.
func NewBoardProvider(b board.ProjectBoard) *BoardProvider {
	return &BoardProvider{board: b}
}

func (p *BoardProvider) Name() string { return "board" }

func (p *BoardProvider) Tools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "sync",
			Description: "Sync FD/SDD status between vault and project board",
			Namespace:   "board",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"direction": map[string]any{
						"type":        "string",
						"enum":        []string{"push", "pull", "both"},
						"description": "Sync direction: push (vault→board), pull (board→vault), or both",
						"default":     "both",
					},
				},
			},
		},
		{
			Name:        "status",
			Description: "Get current board cards with status",
			Namespace:   "board",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"column": map[string]any{
						"type":        "string",
						"description": "Filter by status column (e.g., planned, in-progress, done)",
					},
				},
			},
		},
	}
}

func (p *BoardProvider) Call(ctx context.Context, tool string, params map[string]any) (any, error) {
	switch tool {
	case "sync":
		return p.callSync(ctx, params)
	case "status":
		return p.callStatus(ctx, params)
	default:
		return nil, fmt.Errorf("unknown board tool: %s", tool)
	}
}

func (p *BoardProvider) Start(_ context.Context) error { return nil }
func (p *BoardProvider) Stop() error                   { return nil }
func (p *BoardProvider) Healthy() bool                 { return p.board != nil }

func (p *BoardProvider) callSync(ctx context.Context, params map[string]any) (any, error) {
	direction, _ := params["direction"].(string)
	if direction == "" {
		direction = "both"
	}

	result := map[string]any{"direction": direction}

	if direction == "push" || direction == "both" {
		slog.InfoContext(ctx, "MCP sync: vault → board", "component", "mcp-board")
		if err := p.board.SyncFromVault(ctx); err != nil {
			return nil, fmt.Errorf("push sync: %w", err)
		}
		result["push"] = "complete"
	}

	if direction == "pull" || direction == "both" {
		slog.InfoContext(ctx, "MCP sync: board → vault", "component", "mcp-board")
		if err := p.board.SyncToVault(ctx); err != nil {
			return nil, fmt.Errorf("pull sync: %w", err)
		}
		result["pull"] = "complete"
	}

	return result, nil
}

func (p *BoardProvider) callStatus(ctx context.Context, params map[string]any) (any, error) {
	column, _ := params["column"].(string)
	filter := board.CardFilter{Column: column}

	cards, err := p.board.GetCards(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get cards: %w", err)
	}

	return map[string]any{
		"count": len(cards),
		"cards": cards,
	}, nil
}
