package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Non-existent dir → defaults.
	cfg, err := LoadConfig(context.Background(), "/nonexistent")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Runner.Default != "claude" {
		t.Errorf("Runner.Default = %q, want 'claude'", cfg.Runner.Default)
	}
	if cfg.Runner.Claude.MaxTurns != 200 {
		t.Errorf("Claude.MaxTurns = %d, want 200", cfg.Runner.Claude.MaxTurns)
	}
}

func TestLoadConfig_TeamConfig(t *testing.T) {
	dir := t.TempDir()
	tomlContent := `
[runner]
default = "openhands"

[runner.claude]
max_turns = 500

[board]
provider = "github"

[board.github]
owner = "Deepzima"
project_number = 5

[beads]
enabled = true
auto_create_tasks = true
`
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(context.Background(), dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Runner.Default != "openhands" {
		t.Errorf("Runner.Default = %q, want 'openhands'", cfg.Runner.Default)
	}
	if cfg.Runner.Claude.MaxTurns != 500 {
		t.Errorf("Claude.MaxTurns = %d, want 500", cfg.Runner.Claude.MaxTurns)
	}
	if cfg.Board.Provider != "github" {
		t.Errorf("Board.Provider = %q, want 'github'", cfg.Board.Provider)
	}
	if cfg.Board.GitHub.Owner != "Deepzima" {
		t.Errorf("Board.GitHub.Owner = %q, want 'Deepzima'", cfg.Board.GitHub.Owner)
	}
	if cfg.Board.GitHub.ProjectNumber != 5 {
		t.Errorf("Board.GitHub.ProjectNumber = %d, want 5", cfg.Board.GitHub.ProjectNumber)
	}
	if !cfg.Beads.Enabled {
		t.Error("Beads.Enabled should be true")
	}
}

func TestLoadConfig_LocalOverride(t *testing.T) {
	dir := t.TempDir()

	// Team config.
	team := `
[runner]
default = "claude"

[runner.claude]
max_turns = 200
`
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(team), 0o644); err != nil {
		t.Fatal(err)
	}

	// Local override.
	local := `
[runner.claude]
max_turns = 50
`
	if err := os.WriteFile(filepath.Join(dir, "config.local.toml"), []byte(local), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(context.Background(), dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	// Local override wins.
	if cfg.Runner.Claude.MaxTurns != 50 {
		t.Errorf("Claude.MaxTurns = %d, want 50 (local override)", cfg.Runner.Claude.MaxTurns)
	}
	// Team value preserved for non-overridden fields.
	if cfg.Runner.Default != "claude" {
		t.Errorf("Runner.Default = %q, want 'claude' (from team config)", cfg.Runner.Default)
	}
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("not valid [[[toml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(context.Background(), dir)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}
