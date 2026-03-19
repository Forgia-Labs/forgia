package vault

import (
	"log/slog"
	"time"
)

// FDStatus represents the lifecycle state of a Feature Design.
type FDStatus string

const (
	FDPlanned    FDStatus = "planned"
	FDApproved   FDStatus = "approved"
	FDInProgress FDStatus = "in-progress"
	FDComplete   FDStatus = "complete"
	FDClosed     FDStatus = "closed"
	FDRejected   FDStatus = "rejected"
	FDAbandoned  FDStatus = "abandoned"
)

// FD represents a Feature Design.
type FD struct {
	ID             string   `yaml:"id"`
	Title          string   `yaml:"title"`
	Status         FDStatus `yaml:"status"`
	Priority       string   `yaml:"priority"`
	Effort         string   `yaml:"effort"`
	Impact         string   `yaml:"impact"`
	Author         string   `yaml:"author"`
	Assignee       string   `yaml:"assignee,omitempty"`
	Created        string   `yaml:"created"`
	Reviewed       bool     `yaml:"reviewed"`
	Reviewer       string   `yaml:"reviewer,omitempty"`
	Tags           []string `yaml:"tags,omitempty"`
	UpstreamIssue  string   `yaml:"upstream_issue,omitempty"`
	CompetesWith   []string `yaml:"competes_with,omitempty"`
	RejectedReason string   `yaml:"rejected_reason,omitempty"`
	SupersededBy   string   `yaml:"superseded_by,omitempty"`

	// Execution tracking.
	ExecProfile *ExecProfile `yaml:"exec_profile,omitempty"`
	DesignedBy  *DesignedBy  `yaml:"designed_by,omitempty"`
	ExecutedBy  *ExecutedBy  `yaml:"executed_by,omitempty"`

	// Internal — not serialized.
	FilePath string `yaml:"-"`
}

// LogValue controls how FD appears in slog output.
// Only metadata is logged — never content.
func (fd FD) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", fd.ID),
		slog.String("status", string(fd.Status)),
		slog.String("author", fd.Author),
	)
}

// ExecProfile describes how this FD should be executed by Agent Teams.
type ExecProfile struct {
	LeadModel       string            `yaml:"lead_model"`
	DefaultModel    string            `yaml:"default_model"`
	TeammateModels  map[string]string `yaml:"teammate_models,omitempty"` // SDD ID → model
	MaxTeammates    int               `yaml:"max_teammates"`
	Sandbox         string            `yaml:"sandbox"`
	EstimatedTokens int               `yaml:"estimated_tokens,omitempty"`
}

// DesignedBy records who designed this FD.
type DesignedBy struct {
	Human string `yaml:"human"`
	Model string `yaml:"model"` // e.g., "claude-opus"
	Mode  string `yaml:"mode"`  // "host" or "sandbox"
}

// ExecutedBy records how this FD was executed.
type ExecutedBy struct {
	Method    string     `yaml:"method"` // "agent-team", "sequential", "manual"
	LeadModel string    `yaml:"lead_model,omitempty"`
	StartedAt time.Time `yaml:"started_at,omitempty"`
	Duration  string    `yaml:"duration,omitempty"`
}
