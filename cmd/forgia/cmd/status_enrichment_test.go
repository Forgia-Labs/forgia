package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseOPSFile_ValidFrontmatter(t *testing.T) {
	t.Parallel()

	data := []byte(`---
id: "OPS-001"
title: "Setup CI pipeline"
priority: "high"
---

# OPS-001: Setup CI pipeline

Some body content here.
`)
	ops, err := parseOPSFile(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ops.ID != "OPS-001" {
		t.Errorf("expected ID OPS-001, got %q", ops.ID)
	}
	if ops.Title != "Setup CI pipeline" {
		t.Errorf("expected title 'Setup CI pipeline', got %q", ops.Title)
	}
	if ops.Priority != "high" {
		t.Errorf("expected priority 'high', got %q", ops.Priority)
	}
}

func TestParseOPSFile_NoFrontmatter(t *testing.T) {
	t.Parallel()

	data := []byte("# Just a plain markdown file\n\nNo frontmatter here.\n")
	_, err := parseOPSFile(data)
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestParseOPSFile_UnclosedFrontmatter(t *testing.T) {
	t.Parallel()

	data := []byte("---\nid: OPS-001\ntitle: test\n")
	_, err := parseOPSFile(data)
	if err == nil {
		t.Error("expected error for unclosed frontmatter")
	}
}

func TestParseOPSFile_MinimalFields(t *testing.T) {
	t.Parallel()

	data := []byte("---\nid: OPS-002\n---\n")
	ops, err := parseOPSFile(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ops.ID != "OPS-002" {
		t.Errorf("expected ID OPS-002, got %q", ops.ID)
	}
	if ops.Title != "" {
		t.Errorf("expected empty title, got %q", ops.Title)
	}
}

func TestParseExecLog_Valid(t *testing.T) {
	t.Parallel()

	data := []byte(`{
  "sdd": "SDD-001",
  "fd": "FD-004",
  "runner": "claude",
  "started": "2026-03-28T22:31:54.105216+01:00",
  "completed": "2026-03-28T22:36:23.72396+01:00",
  "duration_seconds": 269,
  "exit_code": 0,
  "status": "success"
}`)
	log, err := parseExecLog(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.SDD != "SDD-001" {
		t.Errorf("expected SDD SDD-001, got %q", log.SDD)
	}
	if log.DurationSeconds != 269 {
		t.Errorf("expected duration 269, got %d", log.DurationSeconds)
	}
	if log.Status != "success" {
		t.Errorf("expected status success, got %q", log.Status)
	}
	if log.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", log.ExitCode)
	}
}

func TestParseExecLog_EmptySDD(t *testing.T) {
	t.Parallel()

	data := []byte(`{"sdd":"","status":"success","duration_seconds":10}`)
	log, err := parseExecLog(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.SDD != "" {
		t.Errorf("expected empty SDD, got %q", log.SDD)
	}
}

func TestParseExecLog_InvalidJSON(t *testing.T) {
	t.Parallel()

	data := []byte(`not json at all`)
	_, err := parseExecLog(data)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFormatRelativeTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"seconds", 30 * time.Second, "30s ago"},
		{"one_minute", 90 * time.Second, "1m ago"},
		{"minutes", 5 * time.Minute, "5m ago"},
		{"one_hour", 90 * time.Minute, "1h ago"},
		{"hours", 5 * time.Hour, "5h ago"},
		{"one_day", 36 * time.Hour, "1d ago"},
		{"days", 72 * time.Hour, "3d ago"},
		{"zero", 0, "0s ago"},
		{"negative", -5 * time.Second, "0s ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatRelativeTime(tt.duration)
			if got != tt.want {
				t.Errorf("formatRelativeTime(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestPrintOPSTasks_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No ops/active dir — should not panic.
	printOPSTasks(t.Context(), dir)
}

func TestPrintExecLogs_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No logs dir — should not panic.
	printExecLogs(t.Context(), dir)
}

func TestPrintOPSTasks_WithFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	opsDir := filepath.Join(dir, "ops", "active")
	if err := os.MkdirAll(opsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("---\nid: OPS-001\ntitle: Test task\npriority: high\n---\n# OPS-001\n")
	if err := os.WriteFile(filepath.Join(opsDir, "OPS-001.md"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	// Should not panic; output goes to stdout.
	printOPSTasks(t.Context(), dir)
}

func TestPrintExecLogs_WithFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logsDir := filepath.Join(dir, "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte(`{"sdd":"SDD-001","status":"success","duration_seconds":120,"started":"2026-03-28T10:00:00Z"}`)
	if err := os.WriteFile(filepath.Join(logsDir, "exec-SDD-001-2026-03-28.json"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	// Should not panic; output goes to stdout.
	printExecLogs(t.Context(), dir)
}

func TestPrintBeadsReady_Graceful(t *testing.T) {
	t.Parallel()

	// Should not panic even if bd is not installed.
	printBeadsReady(t.Context())
}

func TestPrintKnowledgeStats_Graceful(t *testing.T) {
	t.Parallel()

	// Should not panic even if codebase-memory-mcp is not installed.
	printKnowledgeStats(t.Context())
}
