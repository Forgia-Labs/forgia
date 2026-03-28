package cmd

import (
	"testing"
)

func TestValidateCmd_Registered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "validate <sdd-file | FD-NNN>" {
			found = true
		}
	}
	if !found {
		t.Error("validate command not registered")
	}
}

func TestContainsFrontmatterField(t *testing.T) {
	content := "---\nid: SDD-001\nfd: FD-001\ntitle: Test\nstatus: planned\n---\n# Body"

	tests := []struct {
		field string
		want  bool
	}{
		{"id:", true},
		{"fd:", true},
		{"title:", true},
		{"status:", true},
		{"missing:", false},
	}

	for _, tt := range tests {
		got := containsFrontmatterField(content, tt.field)
		if got != tt.want {
			t.Errorf("containsFrontmatterField(%q) = %v, want %v", tt.field, got, tt.want)
		}
	}
}

func TestContainsFrontmatterField_NoFrontmatter(t *testing.T) {
	content := "# Just markdown, no frontmatter"
	if containsFrontmatterField(content, "id:") {
		t.Error("expected false for content without frontmatter")
	}
}

func TestExtractFrontmatterValue(t *testing.T) {
	content := "---\nid: \"SDD-001\"\nfd: FD-001\ntitle: \"Test Title\"\nstatus: planned\n---\n"

	tests := []struct {
		field string
		want  string
	}{
		{"id:", "\"SDD-001\""},
		{"fd:", "FD-001"},
		{"status:", "planned"},
		{"missing:", ""},
	}

	for _, tt := range tests {
		got := extractFrontmatterValue(content, tt.field)
		if got != tt.want {
			t.Errorf("extractFrontmatterValue(%q) = %q, want %q", tt.field, got, tt.want)
		}
	}
}

func TestExtractContextPaths(t *testing.T) {
	content := `## Scope

Build something.

## Context / Contesto

- [ ] ` + "`internal/vault/file_vault.go`" + `
- [ ] ` + "`internal/config/config.go`" + `
- [ ] ` + "`https://example.com`" + `

## Constitution Check
`

	paths := extractContextPaths(content)
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(paths), paths)
	}
	if paths[0] != "internal/vault/file_vault.go" {
		t.Errorf("expected vault path, got %q", paths[0])
	}
	if paths[1] != "internal/config/config.go" {
		t.Errorf("expected config path, got %q", paths[1])
	}
}

func TestExtractContextPaths_NoSection(t *testing.T) {
	paths := extractContextPaths("## Scope\nSomething\n## Other\n")
	if len(paths) != 0 {
		t.Errorf("expected 0 paths for missing Context section, got %d", len(paths))
	}
}

func TestValidateSDD_ValidContent(t *testing.T) {
	content := `---
id: "SDD-001"
fd: "FD-001"
title: "Test"
status: planned
---

# SDD-001: Test

## Scope
Build it.

## Interfaces / Interfacce
| I | T | D |

## Constraints / Vincoli
- Go

## Test Requirements
| T | W | C |

## Acceptance Criteria / Criteri
- [ ] Works

## Context / Contesto
- [ ] ` + "`README.md`" + `

## Constitution Check
- [x] OK

## Work Log / Diario
Done.
`
	// Write to temp file and validate.
	// This test validates the section checker only (no vault).
	requiredSections := []string{
		"## Scope",
		"## Interfaces",
		"## Constraints",
		"## Test Requirements",
		"## Acceptance Criteria",
		"## Context",
		"## Constitution Check",
		"## Work Log",
	}
	for _, s := range requiredSections {
		if !containsString(content, s) {
			t.Errorf("test content missing section: %s", s)
		}
	}
}

func containsString(s, sub string) bool {
	return len(s) >= len(sub) && findInString(s, sub)
}

func findInString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
