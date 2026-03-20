package beads

import (
	"context"
	"testing"
)

func TestCacheCardMapping_Unavailable(t *testing.T) {
	bc := &Client{available: false}
	ctx := context.Background()

	// Should not error when Beads unavailable — silent fallback.
	err := bc.CacheCardMapping(ctx, CardMapping{VaultID: "FD-001", CardID: "card-1", Column: "planned"})
	if err != nil {
		t.Errorf("expected nil error for unavailable beads, got: %v", err)
	}
}

func TestGetCardMapping_Unavailable(t *testing.T) {
	bc := &Client{available: false}

	m, err := bc.GetCardMapping(context.Background(), "FD-001")
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if m != nil {
		t.Error("expected nil mapping when beads unavailable")
	}
}

func TestListCardMappings_Unavailable(t *testing.T) {
	bc := &Client{available: false}

	mappings, err := bc.ListCardMappings(context.Background())
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if mappings != nil {
		t.Error("expected nil mappings when beads unavailable")
	}
}

func TestTrimOutput(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"  hello  ", "hello"},
		{"\nhello\n", "hello"},
		{"", ""},
		{"no-trim", "no-trim"},
	}
	for _, tt := range tests {
		got := trimOutput(tt.input)
		if got != tt.want {
			t.Errorf("trimOutput(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSplitLines(t *testing.T) {
	input := "line1\nline2\n\nline3\n"
	got := splitLines(input)
	if len(got) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(got), got)
	}
}
