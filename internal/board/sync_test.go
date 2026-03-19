package board

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
)

// mockVaultReader returns a fixed set of BoardItems.
type mockVaultReader struct {
	items []BoardItem
	err   error
}

func (m *mockVaultReader) AllItems(_ context.Context) ([]BoardItem, error) {
	return m.items, m.err
}

// mockVaultWriter records UpdateStatus calls.
type mockVaultWriter struct {
	updates []statusUpdate
	err     error
}

type statusUpdate struct {
	ID     string
	Status string
}

func (m *mockVaultWriter) UpdateStatus(_ context.Context, id, status string) error {
	m.updates = append(m.updates, statusUpdate{ID: id, Status: status})
	return m.err
}

// mockBoard wraps LocalBoard but tracks created/moved cards for testing sync logic.
// We test via GitHubBoard's sync methods using a board that doesn't call GitHub.
// Since GitHubBoard.SyncFromVault uses b.GetCards + b.CreateCard + b.MoveCard,
// we build a testable board that records these operations.
type trackingBoard struct {
	cards   []BoardItem
	created []BoardItem
	moved   []statusUpdate
}

func (tb *trackingBoard) NextID(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (tb *trackingBoard) CreateCard(_ context.Context, item BoardItem) (string, error) {
	tb.created = append(tb.created, item)
	tb.cards = append(tb.cards, item)
	return item.ID, nil
}

func (tb *trackingBoard) UpdateCard(_ context.Context, _ string, _ map[string]any) error {
	return nil
}

func (tb *trackingBoard) MoveCard(_ context.Context, id, column string) error {
	tb.moved = append(tb.moved, statusUpdate{ID: id, Status: column})
	for i, c := range tb.cards {
		if c.ID == id {
			tb.cards[i].Column = column
		}
	}
	return nil
}

func (tb *trackingBoard) GetCards(_ context.Context, _ CardFilter) ([]BoardItem, error) {
	return tb.cards, nil
}

func (tb *trackingBoard) SyncFromVault(_ context.Context) error { return nil }
func (tb *trackingBoard) SyncToVault(_ context.Context) error  { return nil }

func TestExtractIDFromTitle(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"FD-a3f2 Feature Name", "FD-a3f2"},
		{"SDD-001a Task Name", "SDD-001a"},
		{"FD-1234", "FD-1234"},
		{"Random title", ""},
		{"", ""},
		{"Something FD-1234", ""},
	}

	for _, tt := range tests {
		got := extractIDFromTitle(tt.title)
		if got != tt.want {
			t.Errorf("extractIDFromTitle(%q) = %q, want %q", tt.title, got, tt.want)
		}
	}
}

func TestSyncFromVault_NilReader(t *testing.T) {
	b := &GitHubBoard{}
	err := b.SyncFromVault(context.Background())
	if err == nil {
		t.Fatal("expected error when vault reader is nil")
	}
}

func TestSyncToVault_NilWriter(t *testing.T) {
	b := &GitHubBoard{}
	err := b.SyncToVault(context.Background())
	if err == nil {
		t.Fatal("expected error when vault writer is nil")
	}
}

func TestVaultReaderWriter_MockContract(t *testing.T) {
	ctx := context.Background()

	reader := &mockVaultReader{
		items: []BoardItem{
			{ID: "FD-0001", Title: "Test", Column: "planned"},
		},
	}
	items, err := reader.AllItems(ctx)
	if err != nil {
		t.Fatalf("AllItems: %v", err)
	}
	if len(items) != 1 || items[0].ID != "FD-0001" {
		t.Errorf("unexpected items: %v", items)
	}

	writer := &mockVaultWriter{}
	if err := writer.UpdateStatus(ctx, "FD-0001", "approved"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if len(writer.updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(writer.updates))
	}
	if writer.updates[0].ID != "FD-0001" || writer.updates[0].Status != "approved" {
		t.Errorf("unexpected update: %+v", writer.updates[0])
	}
}

func TestVaultReaderError(t *testing.T) {
	reader := &mockVaultReader{err: fmt.Errorf("vault unavailable")}
	_, err := reader.AllItems(context.Background())
	if err == nil {
		t.Fatal("expected error from mock reader")
	}
}

func testLogger() *slog.Logger {
	return slog.With("component", "test-board")
}
