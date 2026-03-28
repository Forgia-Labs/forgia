package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Deepzima/forgia/internal/beads"
	"github.com/Deepzima/forgia/internal/runner"
)

// --- writeExecReport tests ---

func TestWriteExecReport_WritesValidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logsDir := filepath.Join(dir, "logs")

	started := time.Date(2026, 3, 28, 15, 30, 0, 0, time.UTC)
	completed := started.Add(5 * time.Minute)

	result := &runner.ExecResult{
		SDD:          "SDD-001",
		FD:           "FD-004",
		File:         ".forgia/sdd/FD-004/SDD-001-status-enrichment.md",
		Runner:       "claude",
		Started:      started,
		Completed:    completed,
		DurationSecs: 300,
		ExitCode:     0,
		Status:       "success",
	}

	reportPath, err := writeExecReport(result, logsDir)
	if err != nil {
		t.Fatalf("writeExecReport() error: %v", err)
	}

	// Verify file exists.
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		t.Fatalf("report file not created: %s", reportPath)
	}

	// Verify valid JSON.
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	var parsed runner.ExecResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify all required fields.
	if parsed.SDD != "SDD-001" {
		t.Errorf("SDD = %q, want %q", parsed.SDD, "SDD-001")
	}
	if parsed.FD != "FD-004" {
		t.Errorf("FD = %q, want %q", parsed.FD, "FD-004")
	}
	if parsed.File != ".forgia/sdd/FD-004/SDD-001-status-enrichment.md" {
		t.Errorf("File = %q, want SDD file path", parsed.File)
	}
	if parsed.Runner != "claude" {
		t.Errorf("Runner = %q, want %q", parsed.Runner, "claude")
	}
	if parsed.DurationSecs != 300 {
		t.Errorf("DurationSecs = %d, want 300", parsed.DurationSecs)
	}
	if parsed.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", parsed.ExitCode)
	}
	if parsed.Status != "success" {
		t.Errorf("Status = %q, want %q", parsed.Status, "success")
	}
}

func TestWriteExecReport_FileNaming(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logsDir := filepath.Join(dir, "logs")

	started := time.Date(2026, 3, 28, 15, 30, 0, 0, time.UTC)
	result := &runner.ExecResult{
		SDD:     "SDD-003",
		FD:      "FD-004",
		Runner:  "claude",
		Started: started,
		Status:  "success",
	}

	reportPath, err := writeExecReport(result, logsDir)
	if err != nil {
		t.Fatalf("writeExecReport() error: %v", err)
	}

	expectedName := "exec-SDD-003-2026-03-28T15-30-00Z.json"
	if filepath.Base(reportPath) != expectedName {
		t.Errorf("filename = %q, want %q", filepath.Base(reportPath), expectedName)
	}
}

func TestWriteExecReport_CreatesLogsDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logsDir := filepath.Join(dir, "nested", "logs")

	result := &runner.ExecResult{
		SDD:     "SDD-001",
		Started: time.Now(),
		Status:  "success",
	}

	_, err := writeExecReport(result, logsDir)
	if err != nil {
		t.Fatalf("writeExecReport() error: %v", err)
	}

	info, err := os.Stat(logsDir)
	if err != nil {
		t.Fatalf("logs dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected logs path to be a directory")
	}
}

func TestWriteExecReport_NilResult(t *testing.T) {
	t.Parallel()

	_, err := writeExecReport(nil, t.TempDir())
	if err == nil {
		t.Error("expected error for nil result")
	}
}

func TestWriteExecReport_FailedExecution(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := &runner.ExecResult{
		SDD:      "SDD-002",
		FD:       "FD-004",
		Runner:   "claude",
		Started:  time.Now().UTC(),
		ExitCode: 1,
		Status:   "failed",
	}

	reportPath, err := writeExecReport(result, dir)
	if err != nil {
		t.Fatalf("writeExecReport() error: %v", err)
	}

	data, _ := os.ReadFile(reportPath)
	var parsed runner.ExecResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed.Status != "failed" {
		t.Errorf("Status = %q, want %q", parsed.Status, "failed")
	}
	if parsed.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", parsed.ExitCode)
	}
}

// --- parseBeadsTaskID tests ---

func TestParseBeadsTaskID_ValidJSON(t *testing.T) {
	t.Parallel()

	output := `{"id":"task-123","title":"SDD-001","status":"open"}`
	got := parseBeadsTaskID(output)
	if got != "task-123" {
		t.Errorf("parseBeadsTaskID() = %q, want %q", got, "task-123")
	}
}

func TestParseBeadsTaskID_MultipleLines(t *testing.T) {
	t.Parallel()

	output := `{"id":"task-1","title":"SDD-001"}
{"id":"task-2","title":"SDD-002"}`
	got := parseBeadsTaskID(output)
	if got != "task-1" {
		t.Errorf("parseBeadsTaskID() = %q, want first ID %q", got, "task-1")
	}
}

func TestParseBeadsTaskID_EmptyOutput(t *testing.T) {
	t.Parallel()

	got := parseBeadsTaskID("")
	if got != "" {
		t.Errorf("parseBeadsTaskID(\"\") = %q, want empty", got)
	}
}

func TestParseBeadsTaskID_NoIDField(t *testing.T) {
	t.Parallel()

	got := parseBeadsTaskID(`{"title":"SDD-001","status":"open"}`)
	if got != "" {
		t.Errorf("parseBeadsTaskID() = %q, want empty (no id field)", got)
	}
}

func TestParseBeadsTaskID_InvalidJSON(t *testing.T) {
	t.Parallel()

	got := parseBeadsTaskID("not json at all")
	if got != "" {
		t.Errorf("parseBeadsTaskID() = %q, want empty for invalid JSON", got)
	}
}

func TestParseBeadsTaskID_WhitespaceOutput(t *testing.T) {
	t.Parallel()

	got := parseBeadsTaskID("  \n  \n  ")
	if got != "" {
		t.Errorf("parseBeadsTaskID() = %q, want empty for whitespace", got)
	}
}

// --- closeBeadsTask tests ---

func TestCloseBeadsTask_UnavailableClient(t *testing.T) {
	t.Parallel()

	bc := &beads.Client{}
	// Client with available=false (zero value) — should skip silently.
	closeBeadsTask(context.Background(), bc, "SDD-001")
	// No panic, no error — just a silent skip.
}

func TestCloseBeadsTask_NilClient(t *testing.T) {
	t.Parallel()

	// Should not panic with nil client.
	closeBeadsTask(context.Background(), nil, "SDD-001")
}
