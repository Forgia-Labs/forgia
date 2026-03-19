package board

import (
	"context"
	"os/exec"
	"testing"
)

func TestResolveLocal(t *testing.T) {
	ctx := context.Background()
	b, err := Resolve(ctx, "local", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil board")
	}

	// Local board operations are no-ops.
	id, err := b.CreateCard(ctx, BoardItem{ID: "test-1", Title: "Test"})
	if err != nil {
		t.Fatalf("CreateCard: %v", err)
	}
	if id != "test-1" {
		t.Errorf("expected id 'test-1', got %q", id)
	}

	if err := b.MoveCard(ctx, "test-1", "Done"); err != nil {
		t.Fatalf("MoveCard: %v", err)
	}

	cards, err := b.GetCards(ctx, CardFilter{})
	if err != nil {
		t.Fatalf("GetCards: %v", err)
	}
	if len(cards) != 0 {
		t.Errorf("expected 0 cards from local board, got %d", len(cards))
	}
}

func TestResolveUnknownProvider(t *testing.T) {
	_, err := Resolve(context.Background(), "jira", "", 0)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestResolveGitHubRequiresGH(t *testing.T) {
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not available")
	}
	// This will fail because "nobody" is not a real owner with project 999.
	_, err := Resolve(context.Background(), "github", "nobody", 999)
	if err == nil {
		t.Fatal("expected error for non-existent project")
	}
}

func TestJsonPath(t *testing.T) {
	data := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "found",
			},
		},
	}

	val, ok := jsonPath(data, "a", "b", "c")
	if !ok || val != "found" {
		t.Errorf("expected 'found', got %v (ok=%v)", val, ok)
	}

	_, ok = jsonPath(data, "a", "x", "c")
	if ok {
		t.Error("expected not found for missing key")
	}
}
