package runner

import (
	"encoding/json"
	"testing"
	"time"
)

func TestExecResult_JSONContainsRequiredFields(t *testing.T) {
	t.Parallel()

	result := &ExecResult{
		SDD:          "SDD-001",
		FD:           "FD-004",
		File:         ".forgia/sdd/FD-004/SDD-001-status-enrichment.md",
		Runner:       "claude",
		Started:      time.Date(2026, 3, 28, 15, 30, 0, 0, time.UTC),
		Completed:    time.Date(2026, 3, 28, 15, 35, 0, 0, time.UTC),
		DurationSecs: 300,
		ExitCode:     0,
		Status:       "success",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Parse into generic map to verify all required fields are present.
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	requiredFields := []string{"sdd", "fd", "file", "runner", "started", "completed", "duration_seconds", "exit_code", "status"}
	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			t.Errorf("missing required field %q in JSON output", field)
		}
	}
}

func TestExecResult_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := &ExecResult{
		SDD:          "SDD-003",
		FD:           "FD-004",
		File:         ".forgia/sdd/FD-004/SDD-003.md",
		Runner:       "claude",
		Started:      time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC),
		Completed:    time.Date(2026, 3, 28, 10, 5, 0, 0, time.UTC),
		DurationSecs: 300,
		ExitCode:     0,
		Status:       "success",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed ExecResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if parsed.SDD != original.SDD {
		t.Errorf("SDD = %q, want %q", parsed.SDD, original.SDD)
	}
	if parsed.FD != original.FD {
		t.Errorf("FD = %q, want %q", parsed.FD, original.FD)
	}
	if parsed.File != original.File {
		t.Errorf("File = %q, want %q", parsed.File, original.File)
	}
	if parsed.Runner != original.Runner {
		t.Errorf("Runner = %q, want %q", parsed.Runner, original.Runner)
	}
	if parsed.DurationSecs != original.DurationSecs {
		t.Errorf("DurationSecs = %d, want %d", parsed.DurationSecs, original.DurationSecs)
	}
	if parsed.ExitCode != original.ExitCode {
		t.Errorf("ExitCode = %d, want %d", parsed.ExitCode, original.ExitCode)
	}
	if parsed.Status != original.Status {
		t.Errorf("Status = %q, want %q", parsed.Status, original.Status)
	}
}

func TestExecResult_FileFieldInJSON(t *testing.T) {
	t.Parallel()

	result := &ExecResult{
		SDD:    "SDD-001",
		File:   ".forgia/sdd/FD-004/SDD-001.md",
		Status: "success",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	file, ok := m["file"]
	if !ok {
		t.Fatal("file field missing from JSON")
	}
	if file != ".forgia/sdd/FD-004/SDD-001.md" {
		t.Errorf("file = %q, want %q", file, ".forgia/sdd/FD-004/SDD-001.md")
	}
}
