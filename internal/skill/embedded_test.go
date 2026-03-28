package skill

import (
	"testing"
	"testing/fstest"
)

func TestNewEmbeddedSkill(t *testing.T) {
	content := []byte("Perform mandatory review of a Feature Design.\n\n## Instructions\n\nGiven FD: $ARGUMENTS")
	s := NewEmbeddedSkill("fd-review", content)

	if s.Name() != "fd-review" {
		t.Errorf("expected name 'fd-review', got %q", s.Name())
	}
	if s.Category() != CategoryFD {
		t.Errorf("expected category FD, got %q", s.Category())
	}
	if s.Mode() != ModeSlashCommand {
		t.Errorf("expected mode SlashCommand, got %q", s.Mode())
	}
	if s.MCPToolDef() != nil {
		t.Error("expected nil MCPToolDef for embedded skill")
	}
}

func TestEmbeddedSkill_Content(t *testing.T) {
	content := []byte("Given FD: $ARGUMENTS\nDo the review.")
	s := NewEmbeddedSkill("fd-review", content)

	result := s.Content("FD-001")
	if result != "Given FD: FD-001\nDo the review." {
		t.Errorf("unexpected content: %q", result)
	}
}

func TestEmbeddedSkill_ContentNoArgs(t *testing.T) {
	content := []byte("List all FDs.")
	s := NewEmbeddedSkill("fd-status", content)

	result := s.Content("")
	if result != "List all FDs." {
		t.Errorf("unexpected content: %q", result)
	}
}

func TestRegistry_LoadEmbedded(t *testing.T) {
	mockFS := fstest.MapFS{
		"fd-new.md":      {Data: []byte("Create a new FD.\n$ARGUMENTS")},
		"fd-review.md":   {Data: []byte("Review an FD.\n$ARGUMENTS")},
		"sdd-status.md":  {Data: []byte("Show SDD status.")},
		"project-init.md": {Data: []byte("Initialize vault.")},
	}

	reg := NewRegistry()
	if err := reg.LoadEmbedded(mockFS); err != nil {
		t.Fatalf("LoadEmbedded failed: %v", err)
	}

	if len(reg.All()) != 4 {
		t.Fatalf("expected 4 skills, got %d", len(reg.All()))
	}

	s, ok := reg.Get("fd-new")
	if !ok {
		t.Fatal("expected fd-new in registry")
	}
	if s.Category() != CategoryFD {
		t.Errorf("expected FD category, got %q", s.Category())
	}

	s2, ok := reg.Get("sdd-status")
	if !ok {
		t.Fatal("expected sdd-status in registry")
	}
	if s2.Category() != CategorySDD {
		t.Errorf("expected SDD category, got %q", s2.Category())
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	reg := NewRegistry()
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Error("expected false for unknown skill")
	}
}

func TestInferCategory(t *testing.T) {
	tests := []struct {
		name string
		want Category
	}{
		{"fd-new", CategoryFD},
		{"fd-review", CategoryFD},
		{"sdd-status", CategorySDD},
		{"sdd-assign", CategorySDD},
		{"arch-init", CategoryArchitecture},
		{"arch-review", CategoryArchitecture},
		{"project-init", CategoryFD}, // fallback
	}
	for _, tt := range tests {
		got := inferCategory(tt.name)
		if got != tt.want {
			t.Errorf("inferCategory(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
