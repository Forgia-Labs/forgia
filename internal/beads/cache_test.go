package beads

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// --- Unavailable fallback tests ---

func TestCacheCardMapping_Unavailable(t *testing.T) {
	bc := &Client{available: false}
	err := bc.CacheCardMapping(context.Background(), CardMapping{VaultID: "FD-001", CardID: "card-1", Column: "planned"})
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

// --- Client.Call() guard tests ---

func TestCall_Unavailable(t *testing.T) {
	bc := &Client{available: false}
	_, err := bc.Call(context.Background(), "get", "key")
	if !errors.Is(err, ErrBeadsUnavailable) {
		t.Errorf("expected ErrBeadsUnavailable, got: %v", err)
	}
}

func TestCall_ContextCancelled(t *testing.T) {
	bc := &Client{available: false}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := bc.Call(ctx, "get", "key")
	if !errors.Is(err, ErrBeadsUnavailable) {
		t.Errorf("expected ErrBeadsUnavailable (guard hit first), got: %v", err)
	}
}

func TestAvailable_False(t *testing.T) {
	bc := &Client{available: false}
	if bc.Available() {
		t.Error("expected Available() = false")
	}
}

func TestAvailable_True(t *testing.T) {
	bc := &Client{available: true}
	if !bc.Available() {
		t.Error("expected Available() = true")
	}
}

// --- GetCardMapping parsing tests ---

func TestGetCardMapping_ValidJSON(t *testing.T) {
	input := `{"vault_id":"FD-001","card_id":"PVTI_abc","column":"approved"}`
	output := trimOutput(input)

	var mapping CardMapping
	if err := json.Unmarshal([]byte(output), &mapping); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if mapping.VaultID != "FD-001" {
		t.Errorf("VaultID = %q, want 'FD-001'", mapping.VaultID)
	}
	if mapping.CardID != "PVTI_abc" {
		t.Errorf("CardID = %q, want 'PVTI_abc'", mapping.CardID)
	}
	if mapping.Column != "approved" {
		t.Errorf("Column = %q, want 'approved'", mapping.Column)
	}
}

func TestGetCardMapping_CorruptedJSON(t *testing.T) {
	inputs := []string{
		"not json at all",
		"{broken",
		"null",
		`{"vault_id": 123}`,
	}
	for _, input := range inputs {
		var mapping CardMapping
		_ = json.Unmarshal([]byte(input), &mapping)
		// Must not panic — either error or partial parse.
	}
}

func TestGetCardMapping_EmptyOutput(t *testing.T) {
	outputs := []string{"", "  ", "\n", "\r\n", "  \n  "}
	for _, output := range outputs {
		trimmed := trimOutput(output)
		if trimmed != "" {
			t.Errorf("trimOutput(%q) = %q, want empty", output, trimmed)
		}
	}
}

func TestGetCardMapping_OutputWithWhitespace(t *testing.T) {
	input := `  {"vault_id":"FD-002","card_id":"card-2","column":"done"}  ` + "\n"
	trimmed := trimOutput(input)

	var mapping CardMapping
	if err := json.Unmarshal([]byte(trimmed), &mapping); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if mapping.VaultID != "FD-002" {
		t.Errorf("VaultID = %q, want 'FD-002'", mapping.VaultID)
	}
}

// --- ListCardMappings parsing tests ---

func TestListCardMappings_MultipleEntries(t *testing.T) {
	output := `{"vault_id":"FD-001","card_id":"c1","column":"planned"}
{"vault_id":"FD-002","card_id":"c2","column":"done"}
{"vault_id":"SDD-001a","card_id":"c3","column":"in-progress"}`

	lines := splitLines(output)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	var mappings []CardMapping
	for _, line := range lines {
		var m CardMapping
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("parse line %q: %v", line, err)
			continue
		}
		mappings = append(mappings, m)
	}
	if len(mappings) != 3 {
		t.Fatalf("expected 3 mappings, got %d", len(mappings))
	}
	if mappings[2].VaultID != "SDD-001a" {
		t.Errorf("third mapping VaultID = %q, want 'SDD-001a'", mappings[2].VaultID)
	}
}

func TestListCardMappings_MixedCorruptedAndValid(t *testing.T) {
	output := `{"vault_id":"FD-001","card_id":"c1","column":"planned"}
not valid json
{"vault_id":"FD-002","card_id":"c2","column":"done"}
{broken
{"vault_id":"FD-003","card_id":"c3","column":"approved"}`

	lines := splitLines(output)
	var valid int
	for _, line := range lines {
		var m CardMapping
		if err := json.Unmarshal([]byte(line), &m); err == nil && m.VaultID != "" {
			valid++
		}
	}
	if valid != 3 {
		t.Errorf("expected 3 valid mappings from mixed output, got %d", valid)
	}
}

func TestListCardMappings_EmptyOutput(t *testing.T) {
	lines := splitLines("")
	if len(lines) != 0 {
		t.Errorf("expected 0 lines from empty output, got %d", len(lines))
	}
}

// --- trimOutput edge cases ---

func TestTrimOutput(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"  hello  ", "hello"},
		{"\nhello\n", "hello"},
		{"", ""},
		{"no-trim", "no-trim"},
		{"\t\ttabs\t\t", "tabs"},
		{"\r\nwindows\r\n", "windows"},
		{"   ", ""},
		{"\n\n\n", ""},
		{" a ", "a"},
	}
	for _, tt := range tests {
		got := trimOutput(tt.input)
		if got != tt.want {
			t.Errorf("trimOutput(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- splitLines edge cases ---

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"normal", "line1\nline2\nline3", 3},
		{"trailing newline", "line1\nline2\n", 2},
		{"empty lines skipped", "line1\n\n\nline2", 2},
		{"single line", "only", 1},
		{"empty", "", 0},
		{"only newlines", "\n\n\n", 0},
		{"whitespace lines", "  \n  \n  ", 0},
		{"mixed whitespace", "  data  \n  \n  more  ", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if len(got) != tt.want {
				t.Errorf("splitLines(%q) = %d lines %v, want %d", tt.input, len(got), got, tt.want)
			}
		})
	}
}

func TestSplitLines_WindowsLineEndings(t *testing.T) {
	input := "line1\r\nline2\r\nline3"
	lines := splitLines(input)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines with \\r\\n, got %d: %v", len(lines), lines)
	}
	for _, line := range lines {
		if len(line) > 0 && (line[0] == '\r' || line[len(line)-1] == '\r') {
			t.Errorf("line still contains \\r: %q", line)
		}
	}
}

// --- CardMapping JSON round-trip ---

func TestCardMapping_MarshalRoundTrip(t *testing.T) {
	original := CardMapping{
		VaultID: "FD-test",
		CardID:  "PVTI_12345",
		Column:  "in-progress",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var got CardMapping
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if got != original {
		t.Errorf("round-trip failed: got %+v, want %+v", got, original)
	}
}

// --- Circuit breaker ---

func TestCircuitBreakerOpen_NoFiles(t *testing.T) {
	// Without circuit breaker files in /tmp, should return false.
	// Skip if real breaker files exist on test machine.
	if isCircuitBreakerOpen() {
		t.Skip("circuit breaker files found on test machine")
	}
}
