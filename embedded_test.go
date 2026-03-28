package forgia

import (
	"io/fs"
	"testing"
)

func TestVaultTemplateFS_ContainsExpectedFiles(t *testing.T) {
	vfs := VaultTemplateFS()

	expected := []string{
		"config.toml",
		"constitution.md",
		"_dashboard.md",
		"fd/_templates/fd-template.md",
		"sdd/_templates/sdd-template.md",
		"guardrails/deny.toml",
		"dev-guide/coding-conventions.md",
	}

	for _, path := range expected {
		if _, err := fs.Stat(vfs, path); err != nil {
			t.Errorf("expected file %q in VaultTemplateFS, got error: %v", path, err)
		}
	}
}

func TestVaultTemplateFS_HasArchitectureTemplates(t *testing.T) {
	vfs := VaultTemplateFS()

	// Architecture templates added by ferruvich (#61)
	archFiles := []string{
		"architecture/system-context.yaml",
		"architecture/containers.yaml",
		"architecture/glossary.yaml",
	}

	for _, path := range archFiles {
		if _, err := fs.Stat(vfs, path); err != nil {
			t.Errorf("expected architecture template %q, got error: %v", path, err)
		}
	}
}

func TestSlashCommandFS_ContainsAllCommands(t *testing.T) {
	cmdFS := SlashCommandFS()

	expected := []string{
		"fd-new.md",
		"fd-review.md",
		"fd-sdd.md",
		"fd-close.md",
		"fd-explore.md",
		"fd-verify.md",
		"fd-status.md",
		"fd-deep.md",
		"sdd-status.md",
		"sdd-assign.md",
		"project-init.md",
	}

	for _, path := range expected {
		if _, err := fs.Stat(cmdFS, path); err != nil {
			t.Errorf("expected command %q in SlashCommandFS, got error: %v", path, err)
		}
	}
}

func TestListSlashCommands_ReturnsNames(t *testing.T) {
	names := ListSlashCommands()

	if len(names) < 11 {
		t.Fatalf("expected at least 11 slash commands, got %d: %v", len(names), names)
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	for _, want := range []string{"fd-new", "fd-review", "fd-sdd", "sdd-status", "project-init"} {
		if !nameSet[want] {
			t.Errorf("expected %q in ListSlashCommands, not found", want)
		}
	}
}

func TestReadSlashCommand_ValidName(t *testing.T) {
	content, err := ReadSlashCommand("fd-review")
	if err != nil {
		t.Fatalf("ReadSlashCommand(fd-review) failed: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("fd-review content is empty")
	}
	// Should contain review-related text
	if !contains(content, "review") && !contains(content, "Review") {
		t.Error("fd-review content doesn't contain 'review'")
	}
}

func TestReadSlashCommand_InvalidName(t *testing.T) {
	_, err := ReadSlashCommand("nonexistent-command")
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}

func contains(data []byte, substr string) bool {
	return len(data) > 0 && len(substr) > 0 && // avoid index issues
		bytesContains(data, []byte(substr))
}

func bytesContains(haystack, needle []byte) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if string(haystack[i:i+len(needle)]) == string(needle) {
			return true
		}
	}
	return false
}
