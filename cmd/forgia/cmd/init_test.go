package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	forgia "github.com/Deepzima/forgia"
)

func TestCopyEmbeddedFS_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	ctx := t.Context()

	src := forgia.VaultTemplateFS()
	created, _, err := copyEmbeddedFS(ctx, src, dir)
	if err != nil {
		t.Fatalf("copyEmbeddedFS failed: %v", err)
	}
	if created == 0 {
		t.Fatal("expected at least some files created")
	}

	// Check key files exist.
	for _, f := range []string{"config.toml", "constitution.md", "guardrails/deny.toml"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}
}

func TestCopyEmbeddedFS_Idempotent(t *testing.T) {
	dir := t.TempDir()
	ctx := t.Context()

	src := forgia.VaultTemplateFS()

	// First copy.
	created1, _, err := copyEmbeddedFS(ctx, src, dir)
	if err != nil {
		t.Fatalf("first copy failed: %v", err)
	}

	// Second copy — nothing new should be created.
	created2, skipped2, err := copyEmbeddedFS(ctx, src, dir)
	if err != nil {
		t.Fatalf("second copy failed: %v", err)
	}
	if created2 != 0 {
		t.Errorf("expected 0 files created on second run, got %d", created2)
	}
	if skipped2 != created1 {
		t.Errorf("expected %d skipped on second run, got %d", created1, skipped2)
	}
}

func TestDetectStack_Go(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644)

	langs := detectStack(dir)
	found := false
	for _, l := range langs {
		if l == "go" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'go' in detected stack, got %v", langs)
	}
}

func TestDetectStack_Fallback(t *testing.T) {
	dir := t.TempDir() // empty dir
	langs := detectStack(dir)
	if len(langs) == 0 {
		t.Fatal("expected at least fallback language")
	}
}

func TestInitCreatesVault(t *testing.T) {
	dir := t.TempDir()
	ctx := t.Context()

	src := forgia.VaultTemplateFS()
	forgiaDir := filepath.Join(dir, ".forgia")
	os.MkdirAll(forgiaDir, 0o755)

	_, _, err := copyEmbeddedFS(ctx, src, forgiaDir)
	if err != nil {
		t.Fatalf("copyEmbeddedFS failed: %v", err)
	}

	// Verify vault structure.
	expectedDirs := []string{"fd", "sdd", "guardrails", "dev-guide"}
	for _, d := range expectedDirs {
		info, err := os.Stat(filepath.Join(forgiaDir, d))
		if err != nil {
			t.Errorf("expected dir %s: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s should be a directory", d)
		}
	}

	// Verify templates are from embedded FS (not minimal stubs).
	constitution, err := os.ReadFile(filepath.Join(forgiaDir, "constitution.md"))
	if err != nil {
		t.Fatalf("read constitution: %v", err)
	}
	// Embedded constitution has "Principles" section (not minimal stub).
	if len(constitution) < 100 {
		t.Error("constitution too short — expected full template, not stub")
	}

	// Verify deny.toml has real patterns.
	deny, err := os.ReadFile(filepath.Join(forgiaDir, "guardrails", "deny.toml"))
	if err != nil {
		t.Fatalf("read deny.toml: %v", err)
	}
	if !containsStr(string(deny), "[read]") {
		t.Error("deny.toml missing [read] section — expected full template")
	}
}

func TestEmbeddedFSHasDevGuide(t *testing.T) {
	vfs := forgia.VaultTemplateFS()
	entries, err := fs.ReadDir(vfs, "dev-guide/lang")
	if err != nil {
		t.Fatalf("read dev-guide/lang: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one language convention file")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findStr(s, substr))
}

func findStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
