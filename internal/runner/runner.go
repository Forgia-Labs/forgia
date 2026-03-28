// Package runner manages SDD execution via different backends
// (Claude Code host, Claude Code sandbox, OpenHands, Agent Teams).
package runner

import (
	"context"
	"time"

	"github.com/Deepzima/forgia/internal/vault"
)

// Runner executes an SDD and returns structured results.
type Runner interface {
	// Name returns the runner identifier (claude, openhands).
	Name() string

	// Execute runs an SDD and returns the result.
	Execute(ctx context.Context, sdd *vault.SDD, opts ExecOptions) (*ExecResult, error)
}

// ExecOptions configures a single execution.
type ExecOptions struct {
	PermissionMode string // default, auto, bypassPermissions
	Sandbox        string // none, native, docker
	MaxTurns       int
	UseRTK         bool
	SystemContext   string // constitution + guardrails + principles
	TaskPrompt     string // the SDD task

	// Agent Teams.
	UseTeam   bool
	LeadModel string
	Model     string // teammate model
}

// ExecResult records what happened during execution.
type ExecResult struct {
	SDD              string    `json:"sdd"`
	FD               string    `json:"fd"`
	Runner           string    `json:"runner"`
	Started          time.Time `json:"started"`
	Completed        time.Time `json:"completed"`
	DurationSecs     int       `json:"duration_seconds"`
	ExitCode         int       `json:"exit_code"`
	Status           string    `json:"status"` // success, failed, timeout
	FilesCreated     []string  `json:"files_created,omitempty"`
	FilesModified    []string  `json:"files_modified,omitempty"`
	Commits          []string  `json:"commits,omitempty"`
	TokensRaw        int       `json:"tokens_raw,omitempty"`
	TokensCompressed int       `json:"tokens_compressed,omitempty"`
	TokensSaved      int       `json:"tokens_saved,omitempty"`
	Model            string    `json:"model,omitempty"`
	TeamRole         string    `json:"team_role,omitempty"` // lead, teammate
}

// Resolve is implemented in claude.go.
