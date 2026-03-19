package vault

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestVault creates a temp .forgia/ directory with sample files.
func setupTestVault(t *testing.T) (string, *FileVault) {
	t.Helper()
	dir := t.TempDir()
	forgiaDir := filepath.Join(dir, ".forgia")

	// Create a vault via Init.
	fv := &FileVault{
		dir:    forgiaDir,
		logger: slog.With("component", "vault-test"),
	}
	if err := os.MkdirAll(forgiaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := fv.Init(ctx, InitOptions{ProjectName: "test-project", Author: "tester"}); err != nil {
		t.Fatal(err)
	}

	return dir, fv
}

func TestOpen(t *testing.T) {
	dir, _ := setupTestVault(t)

	v, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil vault")
	}
}

func TestInitVault(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	v, err := InitVault(ctx, dir, InitOptions{ProjectName: "new-project", Author: "tester"})
	if err != nil {
		t.Fatalf("InitVault: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil vault")
	}

	// Should be openable after init.
	v2, err := Open(dir)
	if err != nil {
		t.Fatalf("Open after InitVault: %v", err)
	}
	content, err := v2.Constitution(ctx)
	if err != nil {
		t.Fatalf("Constitution: %v", err)
	}
	if content == "" {
		t.Error("constitution is empty after InitVault")
	}
}

func TestOpenNotFound(t *testing.T) {
	_, err := Open(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing .forgia/")
	}
}

func TestInit(t *testing.T) {
	dir, _ := setupTestVault(t)
	forgiaDir := filepath.Join(dir, ".forgia")

	// Verify directories exist.
	for _, sub := range []string{"fd", "sdd", "ops", "architecture", "guardrails", "dev-guide"} {
		path := filepath.Join(forgiaDir, sub)
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Errorf("expected directory %s to exist", sub)
		}
	}

	// Verify constitution.
	data, err := os.ReadFile(filepath.Join(forgiaDir, "constitution.md"))
	if err != nil {
		t.Fatalf("constitution not created: %v", err)
	}
	if len(data) == 0 {
		t.Error("constitution is empty")
	}

	// Verify guardrails.
	if _, err := os.Stat(filepath.Join(forgiaDir, "guardrails", "deny.toml")); err != nil {
		t.Errorf("deny.toml not created: %v", err)
	}
}

func TestCreateAndGetFD(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{
		ID:       "FD-a1b2",
		Title:    "Test Feature",
		Status:   FDPlanned,
		Priority: "high",
		Author:   "tester",
		Created:  "2026-03-19",
	}

	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatalf("CreateFD: %v", err)
	}

	// Verify file exists.
	path := filepath.Join(fv.dir, "fd", "FD-a1b2.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("FD file not created: %v", err)
	}

	// Verify SDD subdirectory created.
	sddDir := filepath.Join(fv.dir, "sdd", "FD-a1b2")
	if info, err := os.Stat(sddDir); err != nil || !info.IsDir() {
		t.Error("SDD subdirectory not created for FD")
	}

	// Read it back.
	got, err := fv.GetFD(ctx, "FD-a1b2")
	if err != nil {
		t.Fatalf("GetFD: %v", err)
	}
	if got.ID != "FD-a1b2" {
		t.Errorf("ID = %q, want FD-a1b2", got.ID)
	}
	if got.Title != "Test Feature" {
		t.Errorf("Title = %q, want 'Test Feature'", got.Title)
	}
	if got.Status != FDPlanned {
		t.Errorf("Status = %q, want 'planned'", got.Status)
	}
	if got.Priority != "high" {
		t.Errorf("Priority = %q, want 'high'", got.Priority)
	}
}

func TestCreateFDDuplicate(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{ID: "FD-dup1", Title: "Dup", Status: FDPlanned, Author: "x"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}
	if err := fv.CreateFD(ctx, fd); err == nil {
		t.Fatal("expected error for duplicate FD")
	}
}

func TestListFDs(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	// Create 3 FDs.
	for _, id := range []string{"FD-0001", "FD-0002", "FD-0003"} {
		fd := &FD{ID: id, Title: "Feature " + id, Status: FDPlanned, Author: "tester"}
		if err := fv.CreateFD(ctx, fd); err != nil {
			t.Fatal(err)
		}
	}

	fds, err := fv.ListFDs(ctx)
	if err != nil {
		t.Fatalf("ListFDs: %v", err)
	}
	if len(fds) != 3 {
		t.Errorf("expected 3 FDs, got %d", len(fds))
	}
}

func TestUpdateFD(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{ID: "FD-upd1", Title: "Original", Status: FDPlanned, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	// Update status.
	fd.Status = FDApproved
	fd.Reviewer = "reviewer1"
	if err := fv.UpdateFD(ctx, fd); err != nil {
		t.Fatalf("UpdateFD: %v", err)
	}

	got, err := fv.GetFD(ctx, "FD-upd1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != FDApproved {
		t.Errorf("Status = %q, want 'approved'", got.Status)
	}
	if got.Reviewer != "reviewer1" {
		t.Errorf("Reviewer = %q, want 'reviewer1'", got.Reviewer)
	}
}

func TestFDsIterator(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	for _, id := range []string{"FD-it01", "FD-it02"} {
		fd := &FD{ID: id, Title: "Iter " + id, Status: FDPlanned, Author: "tester"}
		if err := fv.CreateFD(ctx, fd); err != nil {
			t.Fatal(err)
		}
	}

	var count int
	for range fv.FDs() {
		count++
	}
	if count != 2 {
		t.Errorf("FDs iterator: expected 2, got %d", count)
	}
}

func TestCreateAndGetSDD(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	// Create parent FD first.
	fd := &FD{ID: "FD-s001", Title: "Parent", Status: FDApproved, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	sdd := &SDD{
		ID:     "SDD-001a",
		FD:     "FD-s001",
		Title:  "Login API",
		Status: SDDPlanned,
		Scope:  "Implement login endpoint",
		Constraints: SDDConstraints{
			Language:  "go",
			Framework: "stdlib",
		},
	}

	if err := fv.CreateSDD(ctx, sdd); err != nil {
		t.Fatalf("CreateSDD: %v", err)
	}

	got, err := fv.GetSDD(ctx, "FD-s001", "SDD-001a")
	if err != nil {
		t.Fatalf("GetSDD: %v", err)
	}
	if got.ID != "SDD-001a" {
		t.Errorf("ID = %q, want 'SDD-001a'", got.ID)
	}
	if got.FD != "FD-s001" {
		t.Errorf("FD = %q, want 'FD-s001'", got.FD)
	}
	if got.Status != SDDPlanned {
		t.Errorf("Status = %q, want 'planned'", got.Status)
	}
}

func TestListSDDs(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{ID: "FD-ls01", Title: "Parent", Status: FDApproved, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"SDD-001a", "SDD-001b"} {
		sdd := &SDD{ID: id, FD: "FD-ls01", Title: "Task " + id, Status: SDDPlanned}
		if err := fv.CreateSDD(ctx, sdd); err != nil {
			t.Fatal(err)
		}
	}

	sdds, err := fv.ListSDDs(ctx, "FD-ls01")
	if err != nil {
		t.Fatal(err)
	}
	if len(sdds) != 2 {
		t.Errorf("expected 2 SDDs, got %d", len(sdds))
	}
}

func TestUpdateSDD(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{ID: "FD-us01", Title: "Parent", Status: FDApproved, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	sdd := &SDD{ID: "SDD-upd1", FD: "FD-us01", Title: "Original", Status: SDDPlanned}
	if err := fv.CreateSDD(ctx, sdd); err != nil {
		t.Fatal(err)
	}

	sdd.Status = SDDInProgress
	sdd.Agent = "claude-code"
	if err := fv.UpdateSDD(ctx, sdd); err != nil {
		t.Fatalf("UpdateSDD: %v", err)
	}

	got, err := fv.GetSDD(ctx, "FD-us01", "SDD-upd1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != SDDInProgress {
		t.Errorf("Status = %q, want 'in-progress'", got.Status)
	}
	if got.Agent != "claude-code" {
		t.Errorf("Agent = %q, want 'claude-code'", got.Agent)
	}
}

func TestConstitution(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	content, err := fv.Constitution(ctx)
	if err != nil {
		t.Fatalf("Constitution: %v", err)
	}
	if content == "" {
		t.Error("constitution is empty")
	}
}

func TestGuardrailsRaw(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	data, err := fv.GuardrailsRaw(ctx)
	if err != nil {
		t.Fatalf("GuardrailsRaw: %v", err)
	}
	if len(data) == 0 {
		t.Error("guardrails is empty")
	}
}

func TestGetFDNotFound(t *testing.T) {
	_, fv := setupTestVault(t)
	_, err := fv.GetFD(context.Background(), "FD-nonexistent")
	if err == nil {
		t.Fatal("expected error for missing FD")
	}
}

func TestGetSDDNotFound(t *testing.T) {
	_, fv := setupTestVault(t)
	_, err := fv.GetSDD(context.Background(), "FD-nope", "SDD-nope")
	if err == nil {
		t.Fatal("expected error for missing SDD")
	}
}

func TestSplitFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantFM  string
		wantErr bool
	}{
		{
			name:   "valid",
			input:  "---\nid: test\n---\n# Body\n",
			wantFM: "id: test",
		},
		{
			name:    "no frontmatter",
			input:   "# Just markdown\n",
			wantErr: true,
		},
		{
			name:    "unclosed",
			input:   "---\nid: test\n# No closing\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, _, err := splitFrontmatter([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(fm) != tt.wantFM {
				t.Errorf("frontmatter = %q, want %q", string(fm), tt.wantFM)
			}
		})
	}
}

func TestListContextsEmpty(t *testing.T) {
	_, fv := setupTestVault(t)
	contexts, err := fv.ListContexts(context.Background())
	if err != nil {
		t.Fatalf("ListContexts: %v", err)
	}
	if len(contexts) != 0 {
		t.Errorf("expected 0 contexts, got %d", len(contexts))
	}
}

// --- Security tests ---

func TestCreateFD_PathTraversal(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	attacks := []string{
		"../../../etc/passwd",
		"FD-../evil",
		"FD-..%2f..%2fetc",
		"/etc/shadow",
		"FD-test/../../escape",
		"FD-test\\..\\..\\escape",
	}

	for _, id := range attacks {
		fd := &FD{ID: id, Title: "Evil", Status: FDPlanned, Author: "attacker"}
		err := fv.CreateFD(ctx, fd)
		if err == nil {
			t.Errorf("CreateFD should reject path traversal ID %q", id)
		}
	}
}

func TestCreateSDD_PathTraversal(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	// Valid FD first.
	fd := &FD{ID: "FD-safe", Title: "Safe", Status: FDPlanned, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	// Attack via SDD ID.
	sdd := &SDD{ID: "../../../evil", FD: "FD-safe", Title: "Evil", Status: SDDPlanned}
	if err := fv.CreateSDD(ctx, sdd); err == nil {
		t.Error("CreateSDD should reject path traversal in SDD ID")
	}

	// Attack via FD reference.
	sdd2 := &SDD{ID: "SDD-ok", FD: "../../etc", Title: "Evil", Status: SDDPlanned}
	if err := fv.CreateSDD(ctx, sdd2); err == nil {
		t.Error("CreateSDD should reject path traversal in FD reference")
	}
}

func TestCreateFD_EmptyID(t *testing.T) {
	_, fv := setupTestVault(t)
	fd := &FD{ID: "", Title: "No ID", Status: FDPlanned, Author: "tester"}
	if err := fv.CreateFD(context.Background(), fd); err == nil {
		t.Error("CreateFD should reject empty ID")
	}
}

// --- Data integrity tests ---

func TestUpdateFD_PreservesMarkdownBody(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{ID: "FD-body", Title: "Body Test", Status: FDPlanned, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	// Manually append extra markdown to the file.
	path := filepath.Join(fv.dir, "fd", "FD-body.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	extraBody := "\n## Custom Section\n\nThis should survive updates.\n\n### Sub-section\n\n- bullet 1\n- bullet 2\n"
	if err := os.WriteFile(path, append(data, []byte(extraBody)...), 0o644); err != nil {
		t.Fatal(err)
	}

	// Update frontmatter only.
	fd.Status = FDApproved
	fd.Reviewer = "reviewer1"
	if err := fv.UpdateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	// Read back and verify body is preserved.
	updatedData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(updatedData)
	if !strings.Contains(content, "## Custom Section") {
		t.Error("markdown body lost after UpdateFD")
	}
	if !strings.Contains(content, "- bullet 1") {
		t.Error("markdown bullets lost after UpdateFD")
	}

	// Verify frontmatter was updated.
	got, err := fv.GetFD(ctx, "FD-body")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != FDApproved {
		t.Errorf("Status = %q, want 'approved'", got.Status)
	}
}

func TestUpdateSDD_PreservesMarkdownBody(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{ID: "FD-sbody", Title: "Parent", Status: FDApproved, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}
	sdd := &SDD{ID: "SDD-body", FD: "FD-sbody", Title: "Body Test", Status: SDDPlanned}
	if err := fv.CreateSDD(ctx, sdd); err != nil {
		t.Fatal(err)
	}

	// Append custom markdown.
	path := filepath.Join(fv.dir, "sdd", "FD-sbody", "SDD-body.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, []byte("\n## Implementation Notes\n\nKeep this.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	// Update status.
	sdd.Status = SDDDone
	sdd.Agent = "claude-code"
	if err := fv.UpdateSDD(ctx, sdd); err != nil {
		t.Fatal(err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "## Implementation Notes") {
		t.Error("markdown body lost after UpdateSDD")
	}
}

// --- Corrupted file tests ---

func TestFDs_SkipsCorruptedFiles(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	// Create a valid FD.
	fd := &FD{ID: "FD-good", Title: "Good", Status: FDPlanned, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	// Create a corrupted FD file (invalid YAML frontmatter).
	corruptPath := filepath.Join(fv.dir, "fd", "FD-corrupt.md")
	if err := os.WriteFile(corruptPath, []byte("---\n: invalid: yaml: [[\n---\n# Broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create an empty file.
	emptyPath := filepath.Join(fv.dir, "fd", "FD-empty.md")
	if err := os.WriteFile(emptyPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a binary file.
	binPath := filepath.Join(fv.dir, "fd", "FD-binary.md")
	if err := os.WriteFile(binPath, []byte{0x00, 0xFF, 0xFE, 0xFD}, 0o644); err != nil {
		t.Fatal(err)
	}

	// ListFDs should still return the valid one, skipping corrupted.
	fds, err := fv.ListFDs(ctx)
	if err != nil {
		t.Fatalf("ListFDs: %v", err)
	}
	if len(fds) != 1 {
		t.Errorf("expected 1 valid FD, got %d", len(fds))
	}
	if len(fds) > 0 && fds[0].ID != "FD-good" {
		t.Errorf("expected FD-good, got %q", fds[0].ID)
	}
}

func TestSDDs_SkipsCorruptedFiles(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{ID: "FD-cs01", Title: "Parent", Status: FDApproved, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	// Valid SDD.
	sdd := &SDD{ID: "SDD-good", FD: "FD-cs01", Title: "Good", Status: SDDPlanned}
	if err := fv.CreateSDD(ctx, sdd); err != nil {
		t.Fatal(err)
	}

	// Corrupted SDD.
	corruptPath := filepath.Join(fv.dir, "sdd", "FD-cs01", "SDD-corrupt.md")
	if err := os.WriteFile(corruptPath, []byte("not even frontmatter"), 0o644); err != nil {
		t.Fatal(err)
	}

	sdds, err := fv.ListSDDs(ctx, "FD-cs01")
	if err != nil {
		t.Fatal(err)
	}
	if len(sdds) != 1 {
		t.Errorf("expected 1 valid SDD, got %d", len(sdds))
	}
}

// --- SDD YAML format test ---

func TestParseSDDYamlFormat(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{ID: "FD-yaml", Title: "YAML Parent", Status: FDApproved, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	// Write a .yaml format SDD (full YAML with meta wrapper).
	yamlContent := `meta:
  id: "SDD-yaml1"
  fd: "FD-yaml"
  title: "YAML Format Task"
  status: planned
  agent: "claude-code"
  complexity: "high"

scope: |
  Build something in YAML format.

constraints:
  language: "go"
  framework: "stdlib"
`
	yamlPath := filepath.Join(fv.dir, "sdd", "FD-yaml", "SDD-yaml1.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	sdd, err := fv.GetSDD(ctx, "FD-yaml", "SDD-yaml1")
	if err != nil {
		t.Fatalf("GetSDD (yaml format): %v", err)
	}
	if sdd.ID != "SDD-yaml1" {
		t.Errorf("ID = %q, want 'SDD-yaml1'", sdd.ID)
	}
	if sdd.Agent != "claude-code" {
		t.Errorf("Agent = %q, want 'claude-code'", sdd.Agent)
	}
	if sdd.Complexity != "high" {
		t.Errorf("Complexity = %q, want 'high'", sdd.Complexity)
	}
}

// --- Round-trip fidelity test ---

func TestFDRoundTrip_AllFields(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	original := &FD{
		ID:            "FD-rt01",
		Title:         "Round Trip Test",
		Status:        FDInProgress,
		Priority:      "critical",
		Effort:        "high",
		Impact:        "high",
		Author:        "engineer1",
		Assignee:      "engineer2",
		Created:       "2026-03-19",
		Reviewed:      true,
		Reviewer:      "lead",
		Tags:          []string{"security", "backend"},
		UpstreamIssue: "#42",
		CompetesWith:  []string{"FD-rt02"},
	}

	if err := fv.CreateFD(ctx, original); err != nil {
		t.Fatal(err)
	}

	got, err := fv.GetFD(ctx, "FD-rt01")
	if err != nil {
		t.Fatal(err)
	}

	if got.Title != original.Title {
		t.Errorf("Title = %q, want %q", got.Title, original.Title)
	}
	if got.Status != original.Status {
		t.Errorf("Status = %q, want %q", got.Status, original.Status)
	}
	if got.Priority != original.Priority {
		t.Errorf("Priority = %q, want %q", got.Priority, original.Priority)
	}
	if got.Effort != original.Effort {
		t.Errorf("Effort = %q, want %q", got.Effort, original.Effort)
	}
	if got.Author != original.Author {
		t.Errorf("Author = %q, want %q", got.Author, original.Author)
	}
	if got.Assignee != original.Assignee {
		t.Errorf("Assignee = %q, want %q", got.Assignee, original.Assignee)
	}
	if !got.Reviewed {
		t.Error("Reviewed should be true")
	}
	if got.Reviewer != original.Reviewer {
		t.Errorf("Reviewer = %q, want %q", got.Reviewer, original.Reviewer)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "security" {
		t.Errorf("Tags = %v, want [security backend]", got.Tags)
	}
	if got.UpstreamIssue != "#42" {
		t.Errorf("UpstreamIssue = %q, want '#42'", got.UpstreamIssue)
	}
	if len(got.CompetesWith) != 1 || got.CompetesWith[0] != "FD-rt02" {
		t.Errorf("CompetesWith = %v, want [FD-rt02]", got.CompetesWith)
	}
}

func TestSDDRoundTrip_AllFields(t *testing.T) {
	_, fv := setupTestVault(t)
	ctx := context.Background()

	fd := &FD{ID: "FD-srt1", Title: "Parent", Status: FDApproved, Author: "tester"}
	if err := fv.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	original := &SDD{
		ID:         "SDD-rt01",
		FD:         "FD-srt1",
		Title:      "Round Trip SDD",
		Status:     SDDInProgress,
		Agent:      "claude-code",
		AssignedTo: "engineer1",
		Created:    "2026-03-19",
		Tags:       []string{"api", "auth"},
		Complexity: "high",
		Scope:      "Build the login endpoint",
		Constraints: SDDConstraints{
			Language:     "go",
			Framework:    "stdlib",
			Dependencies: []string{"jwt-go"},
			Patterns:     []string{"repository"},
		},
		Boundaries: SDDBoundaries{
			WriteDirs:     []string{"internal/auth/"},
			ForbiddenDirs: []string{"cmd/"},
			MaxFiles:      10,
		},
		TestReqs: []SDDTestReq{
			{Type: "unit", What: "auth handler", Coverage: "80%"},
		},
		Criteria: []SDDCriterion{
			{Criterion: "login returns JWT", Met: false},
		},
	}

	if err := fv.CreateSDD(ctx, original); err != nil {
		t.Fatal(err)
	}

	got, err := fv.GetSDD(ctx, "FD-srt1", "SDD-rt01")
	if err != nil {
		t.Fatal(err)
	}

	if got.Agent != "claude-code" {
		t.Errorf("Agent = %q, want 'claude-code'", got.Agent)
	}
	if got.Complexity != "high" {
		t.Errorf("Complexity = %q, want 'high'", got.Complexity)
	}
	if got.Constraints.Language != "go" {
		t.Errorf("Language = %q, want 'go'", got.Constraints.Language)
	}
	if len(got.Constraints.Dependencies) != 1 || got.Constraints.Dependencies[0] != "jwt-go" {
		t.Errorf("Dependencies = %v, want [jwt-go]", got.Constraints.Dependencies)
	}
	if len(got.Boundaries.WriteDirs) != 1 || got.Boundaries.WriteDirs[0] != "internal/auth/" {
		t.Errorf("WriteDirs = %v, want [internal/auth/]", got.Boundaries.WriteDirs)
	}
	if got.Boundaries.MaxFiles != 10 {
		t.Errorf("MaxFiles = %d, want 10", got.Boundaries.MaxFiles)
	}
	if len(got.TestReqs) != 1 || got.TestReqs[0].Coverage != "80%" {
		t.Errorf("TestReqs = %v, unexpected", got.TestReqs)
	}
	if len(got.Criteria) != 1 || got.Criteria[0].Met {
		t.Errorf("Criteria = %v, unexpected", got.Criteria)
	}
}

// --- Edge case tests ---

func TestListFDs_EmptyVault(t *testing.T) {
	_, fv := setupTestVault(t)
	fds, err := fv.ListFDs(context.Background())
	if err != nil {
		t.Fatalf("ListFDs: %v", err)
	}
	if fds != nil && len(fds) != 0 {
		t.Errorf("expected nil or empty slice, got %d FDs", len(fds))
	}
}

func TestListSDDs_NonexistentFD(t *testing.T) {
	_, fv := setupTestVault(t)
	sdds, err := fv.ListSDDs(context.Background(), "FD-doesnt-exist")
	if err != nil {
		t.Fatalf("ListSDDs: %v", err)
	}
	if sdds != nil && len(sdds) != 0 {
		t.Errorf("expected nil or empty slice, got %d SDDs", len(sdds))
	}
}

func TestValidateID(t *testing.T) {
	valid := []string{"FD-a1b2", "SDD-001a", "FD-test-feature", "SDD-with-dashes"}
	for _, id := range valid {
		if err := validateID(id); err != nil {
			t.Errorf("validateID(%q) unexpected error: %v", id, err)
		}
	}

	invalid := []string{
		"",
		"..",
		"../evil",
		"FD-../bad",
		"FD-test/sub",
		"FD-test\\sub",
		"/absolute",
	}
	for _, id := range invalid {
		if err := validateID(id); err == nil {
			t.Errorf("validateID(%q) should return error", id)
		}
	}
}
