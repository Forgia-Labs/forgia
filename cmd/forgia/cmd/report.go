package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgia-labs/forgia/internal/beads"
	"github.com/forgia-labs/forgia/internal/runner"
)

// writeExecReport marshals the ExecResult to JSON and writes it to logsDir.
// Returns the path written, or an error (callers should warn, not fail).
func writeExecReport(result *runner.ExecResult, logsDir string) (string, error) {
	if result == nil {
		return "", fmt.Errorf("nil result")
	}

	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return "", fmt.Errorf("create logs dir %s: %w", logsDir, err)
	}

	timestamp := result.Started.UTC().Format("2006-01-02T15-04-05Z")
	filename := fmt.Sprintf("exec-%s-%s.json", result.SDD, timestamp)
	reportPath := filepath.Join(logsDir, filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal exec report: %w", err)
	}

	if err := os.WriteFile(reportPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write exec report %s: %w", reportPath, err)
	}

	return reportPath, nil
}

// closeBeadsTask searches for a Beads task matching sddID and marks it done.
// Skips silently if Beads is unavailable or no matching task is found.
func closeBeadsTask(ctx context.Context, bc *beads.Client, sddID string) {
	if bc == nil || !bc.Available() {
		slog.InfoContext(ctx, "beads unavailable, skipping task closure", "sdd", sddID)
		return
	}

	output, err := bc.Call(ctx, "search", sddID)
	if err != nil {
		slog.WarnContext(ctx, "beads search failed", "sdd", sddID, "error", err)
		return
	}

	taskID := parseBeadsTaskID(output)
	if taskID == "" {
		slog.InfoContext(ctx, "no beads task found for SDD", "sdd", sddID)
		return
	}

	_, err = bc.Call(ctx, "update", "--status=done", taskID)
	if err != nil {
		slog.WarnContext(ctx, "beads task update failed", "sdd", sddID, "task_id", taskID, "error", err)
		return
	}

	slog.InfoContext(ctx, "beads task closed", "sdd", sddID, "task_id", taskID)
}

// parseBeadsTaskID extracts the first task ID from bd search JSON output.
// Expected format: {"id":"<task_id>", ...} (one or more JSON objects).
func parseBeadsTaskID(output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return ""
	}

	// Try to find "id" field in the first JSON object.
	type beadsTask struct {
		ID string `json:"id"`
	}

	// Handle newline-delimited JSON (one object per line).
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var task beadsTask
		if err := json.Unmarshal([]byte(line), &task); err == nil && task.ID != "" {
			return task.ID
		}
	}

	return ""
}
