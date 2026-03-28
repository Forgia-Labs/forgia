package mcp

import (
	"context"
	"testing"

	"github.com/Deepzima/forgia/internal/board"
)

// mockBoard implements board.ProjectBoard for testing MCP tools.
type mockBoard struct {
	syncFromCalled bool
	syncToCalled   bool
	cards          []board.BoardItem
}

func (m *mockBoard) NextID(_ context.Context, _ string) (string, error) { return "", nil }
func (m *mockBoard) CreateCard(_ context.Context, item board.BoardItem) (string, error) {
	return item.ID, nil
}
func (m *mockBoard) UpdateCard(_ context.Context, _ string, _ map[string]any) error { return nil }
func (m *mockBoard) MoveCard(_ context.Context, _ string, _ string) error           { return nil }
func (m *mockBoard) GetCards(_ context.Context, filter board.CardFilter) ([]board.BoardItem, error) {
	if filter.Column != "" {
		var filtered []board.BoardItem
		for _, c := range m.cards {
			if c.Column == filter.Column {
				filtered = append(filtered, c)
			}
		}
		return filtered, nil
	}
	return m.cards, nil
}
func (m *mockBoard) SyncFromVault(_ context.Context) error {
	m.syncFromCalled = true
	return nil
}
func (m *mockBoard) SyncToVault(_ context.Context) error {
	m.syncToCalled = true
	return nil
}

func TestBoardProvider_InterfaceSatisfaction(t *testing.T) {
	var _ ToolProvider = (*BoardProvider)(nil)
}

func TestBoardProvider_Tools(t *testing.T) {
	p := NewBoardProvider(&mockBoard{})
	tools := p.Tools()
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name] = true
	}
	if !names["sync"] {
		t.Error("missing 'sync' tool")
	}
	if !names["status"] {
		t.Error("missing 'status' tool")
	}
}

func TestBoardProvider_Name(t *testing.T) {
	p := NewBoardProvider(&mockBoard{})
	if p.Name() != "board" {
		t.Errorf("Name() = %q, want 'board'", p.Name())
	}
}

func TestBoardProvider_SyncPush(t *testing.T) {
	mb := &mockBoard{}
	p := NewBoardProvider(mb)
	ctx := context.Background()

	result, err := p.Call(ctx, "sync", map[string]any{"direction": "push"})
	if err != nil {
		t.Fatalf("Call sync push: %v", err)
	}
	if !mb.syncFromCalled {
		t.Error("SyncFromVault not called on push")
	}
	if mb.syncToCalled {
		t.Error("SyncToVault should not be called on push-only")
	}

	r := result.(map[string]any)
	if r["push"] != "complete" {
		t.Errorf("result push = %v, want 'complete'", r["push"])
	}
}

func TestBoardProvider_SyncPull(t *testing.T) {
	mb := &mockBoard{}
	p := NewBoardProvider(mb)
	ctx := context.Background()

	result, err := p.Call(ctx, "sync", map[string]any{"direction": "pull"})
	if err != nil {
		t.Fatalf("Call sync pull: %v", err)
	}
	if mb.syncFromCalled {
		t.Error("SyncFromVault should not be called on pull-only")
	}
	if !mb.syncToCalled {
		t.Error("SyncToVault not called on pull")
	}

	r := result.(map[string]any)
	if r["pull"] != "complete" {
		t.Errorf("result pull = %v, want 'complete'", r["pull"])
	}
}

func TestBoardProvider_SyncBoth(t *testing.T) {
	mb := &mockBoard{}
	p := NewBoardProvider(mb)
	ctx := context.Background()

	_, err := p.Call(ctx, "sync", map[string]any{})
	if err != nil {
		t.Fatalf("Call sync both: %v", err)
	}
	if !mb.syncFromCalled {
		t.Error("SyncFromVault not called on default (both)")
	}
	if !mb.syncToCalled {
		t.Error("SyncToVault not called on default (both)")
	}
}

func TestBoardProvider_Status(t *testing.T) {
	mb := &mockBoard{
		cards: []board.BoardItem{
			{ID: "FD-001", Title: "Feature 1", Column: "planned"},
			{ID: "FD-002", Title: "Feature 2", Column: "done"},
		},
	}
	p := NewBoardProvider(mb)
	ctx := context.Background()

	// All cards.
	result, err := p.Call(ctx, "status", map[string]any{})
	if err != nil {
		t.Fatalf("Call status: %v", err)
	}
	r := result.(map[string]any)
	if r["count"] != 2 {
		t.Errorf("count = %v, want 2", r["count"])
	}

	// Filtered.
	result, err = p.Call(ctx, "status", map[string]any{"column": "planned"})
	if err != nil {
		t.Fatalf("Call status filtered: %v", err)
	}
	r = result.(map[string]any)
	if r["count"] != 1 {
		t.Errorf("filtered count = %v, want 1", r["count"])
	}
}

func TestBoardProvider_UnknownTool(t *testing.T) {
	p := NewBoardProvider(&mockBoard{})
	_, err := p.Call(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestBoardProvider_Registry(t *testing.T) {
	reg := NewProviderRegistry()
	p := NewBoardProvider(&mockBoard{})
	reg.Register(p)

	tools := reg.AllTools()
	found := 0
	for _, tool := range tools {
		if tool.Name == "forgia_board_sync" || tool.Name == "forgia_board_status" {
			found++
		}
	}
	if found != 2 {
		t.Errorf("expected 2 board tools in registry, found %d", found)
	}
}
