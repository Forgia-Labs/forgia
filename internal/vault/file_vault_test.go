package vault

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
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
