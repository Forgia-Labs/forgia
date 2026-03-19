// Package config handles Forgia configuration loading and management.
// Loads config.toml (team) + config.local.toml (personal override).
package config

import "context"

// Config is the root configuration structure.
type Config struct {
	Runner    RunnerConfig    `toml:"runner"`
	Board     BoardConfig     `toml:"board"`
	MCP       MCPConfig       `toml:"mcp"`
	Beads     BeadsConfig     `toml:"beads"`
	Watcher   WatcherConfig   `toml:"watcher"`
	Knowledge KnowledgeConfig `toml:"knowledge"`
	Docker    DockerConfig    `toml:"docker"`
}

// RunnerConfig selects and configures execution backends.
type RunnerConfig struct {
	Default     string             `toml:"default"` // claude, openhands
	Parallel    bool               `toml:"parallel"`
	MaxParallel int                `toml:"max_parallel"`
	Claude      ClaudeRunnerConfig `toml:"claude"`
	OpenHands   OpenHandsConfig    `toml:"openhands"`
}

// ClaudeRunnerConfig configures the Claude Code runner.
type ClaudeRunnerConfig struct {
	PermissionMode  string `toml:"permission_mode"` // default, auto, bypassPermissions
	MaxTurns        int    `toml:"max_turns"`
	Sandbox         string `toml:"sandbox"` // none, native, docker
	UseRTK          bool   `toml:"use_rtk"`
	RTKTrackSavings bool   `toml:"rtk_track_savings"`

	// Agent Teams.
	TeamEnabled      bool   `toml:"team_enabled"`
	TeamLeadModel    string `toml:"team_lead_model"`
	TeamDefaultModel string `toml:"team_default_model"`
	TeamMaxTeammates int    `toml:"team_max_teammates"`
	TeamSandbox      string `toml:"team_sandbox"`
}

// OpenHandsConfig configures the OpenHands runner.
type OpenHandsConfig struct {
	Image          string `toml:"image"`
	Model          string `toml:"model"`
	WorkspaceMount string `toml:"workspace_mount"`
	MaxIterations  int    `toml:"max_iterations"`
	UIPort         int    `toml:"ui_port"`
}

// BoardConfig selects the project board backend.
type BoardConfig struct {
	Provider string            `toml:"provider"` // github, gitlab, local
	GitHub   GitHubBoardConfig `toml:"github"`
	GitLab   GitLabBoardConfig `toml:"gitlab"`
}

// GitHubBoardConfig for GitHub Projects.
type GitHubBoardConfig struct {
	ProjectNumber int    `toml:"project_number"`
	Owner         string `toml:"owner"`
}

// GitLabBoardConfig for GitLab Boards.
type GitLabBoardConfig struct {
	ProjectID int `toml:"project_id"`
	BoardID   int `toml:"board_id"`
}

// MCPConfig configures MCP provider chaining.
type MCPConfig struct {
	Providers map[string]MCPProviderConfig `toml:"providers"`
}

// MCPProviderConfig defines an MCP subprocess provider.
type MCPProviderConfig struct {
	Command        string   `toml:"command"`
	Args           []string `toml:"args"`
	Lazy           bool     `toml:"lazy"`
	RestartOnCrash bool     `toml:"restart_on_crash"`
	HealthCheck    string   `toml:"health_check"`
	ExposeTools    []string `toml:"expose_tools"`
	Namespace      string   `toml:"namespace"`
}

// BeadsConfig for local cache integration.
type BeadsConfig struct {
	Enabled              bool `toml:"enabled"`
	AutoCreateTasks      bool `toml:"auto_create_tasks"`
	AutoCloseTasks       bool `toml:"auto_close_tasks"`
	DependencyAwareBatch bool `toml:"dependency_aware_batch"`
}

// WatcherConfig for forgia watch.
type WatcherConfig struct {
	Enabled  bool   `toml:"enabled"`
	Path     string `toml:"path"`
	Debounce int    `toml:"debounce"` // seconds
}

// KnowledgeConfig for Tier 3 knowledge layer.
type KnowledgeConfig struct {
	Provider  string `toml:"provider"` // codebase-memory-mcp
	AutoIndex bool   `toml:"auto_index"`
	AutoSync  bool   `toml:"auto_sync"`
}

// DockerConfig for sandbox container management via Engine API.
type DockerConfig struct {
	Host           string   `toml:"host"` // unix:///var/run/docker.sock, tcp://, ssh://
	SandboxImage   string   `toml:"sandbox_image"`
	NetworkAllow   []string `toml:"sandbox_network_allow"`
	WorkspaceMount string   `toml:"sandbox_workspace_mount"`
	ExtraMounts    []string `toml:"sandbox_extra_mounts"`
}

// LoadConfig reads config.toml + config.local.toml with merge.
// Priority: CLI flags > env vars > config.local.toml > config.toml > defaults.
func LoadConfig(ctx context.Context, vaultDir string) (*Config, error) {
	// TODO: implement TOML loading + merge
	_ = ctx // will be used for cancellation during file reads
	return DefaultConfig(), nil
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Runner: RunnerConfig{
			Default:     "claude",
			MaxParallel: 3,
			Claude: ClaudeRunnerConfig{
				PermissionMode:   "default",
				MaxTurns:         200,
				Sandbox:          "none",
				TeamEnabled:      false,
				TeamLeadModel:    "sonnet",
				TeamDefaultModel: "sonnet",
				TeamMaxTeammates: 5,
			},
		},
		Watcher: WatcherConfig{
			Path:     ".forgia/sdd",
			Debounce: 5,
		},
	}
}
