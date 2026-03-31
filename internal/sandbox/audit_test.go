package sandbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAuditLogger_WriteAndRead(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	logger, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}

	// Write entries.
	entries := []AuditEntry{
		{Timestamp: time.Now(), Command: "go test ./...", ExitCode: 0, DurationMs: 3421, OutputBytes: 12340},
		{Timestamp: time.Now(), Command: "git status", ExitCode: 0, DurationMs: 50, OutputBytes: 500},
		{Timestamp: time.Now(), Command: "rm -rf /", ExitCode: 1, DurationMs: 1, OutputBytes: 100},
	}

	for _, e := range entries {
		if err := logger.Log(e); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}
	logger.Close()

	// Read and verify.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	var parsed AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Command != "go test ./..." {
		t.Errorf("command = %q", parsed.Command)
	}
	if parsed.ExitCode != 0 {
		t.Errorf("exit_code = %d", parsed.ExitCode)
	}
}

func TestAuditLogger_Append(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	// Write first batch.
	l1, _ := NewAuditLogger(path)
	l1.Log(AuditEntry{Command: "first"})
	l1.Close()

	// Append second batch.
	l2, _ := NewAuditLogger(path)
	l2.Log(AuditEntry{Command: "second"})
	l2.Close()

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (appended), got %d", len(lines))
	}
}
