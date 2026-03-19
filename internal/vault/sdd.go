package vault

import (
	"log/slog"
	"time"
)

// SDDStatus represents the lifecycle state of an SDD.
type SDDStatus string

const (
	SDDPlanned    SDDStatus = "planned"
	SDDValidated  SDDStatus = "validated"
	SDDInProgress SDDStatus = "in-progress"
	SDDDone       SDDStatus = "done"
	SDDFailed     SDDStatus = "failed"
)

// SDD represents a Spec-Driven Development execution spec.
type SDD struct {
	ID         string    `yaml:"id"`
	FD         string    `yaml:"fd"`
	Title      string    `yaml:"title"`
	Status     SDDStatus `yaml:"status"`
	Agent      string    `yaml:"agent,omitempty"`
	AssignedTo string    `yaml:"assigned_to,omitempty"`
	Created    string    `yaml:"created"`
	Started    string    `yaml:"started,omitempty"`
	Completed  string    `yaml:"completed,omitempty"`
	Tags       []string  `yaml:"tags,omitempty"`
	BdTaskID   string    `yaml:"bd_task_id,omitempty"`
	Complexity string    `yaml:"complexity,omitempty"` // low, medium, high → model routing

	Scope       string           `yaml:"scope"`
	Interfaces  []SDDInterface   `yaml:"interfaces,omitempty"`
	Constraints SDDConstraints   `yaml:"constraints"`
	Boundaries  SDDBoundaries    `yaml:"boundaries,omitempty"`
	TestReqs    []SDDTestReq     `yaml:"test_requirements,omitempty"`
	Criteria    []SDDCriterion   `yaml:"acceptance_criteria,omitempty"`
	Context     []string         `yaml:"context,omitempty"`
	WorkLog     *WorkLog         `yaml:"work_log,omitempty"`

	// Internal.
	FilePath string `yaml:"-"`
}

// SDDInterface defines an input/output contract.
type SDDInterface struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Contract    string `yaml:"contract,omitempty"`
}

// SDDConstraints captures language, framework, patterns.
type SDDConstraints struct {
	Language     string   `yaml:"language"`
	Framework    string   `yaml:"framework,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
	Patterns     []string `yaml:"patterns,omitempty"`
}

// SDDBoundaries defines file ownership for Agent Teams.
type SDDBoundaries struct {
	WriteDirs    []string `yaml:"write_dirs,omitempty"`
	ForbiddenDirs []string `yaml:"forbidden_dirs,omitempty"`
	MaxFiles     int      `yaml:"max_files_created,omitempty"`
	MaxFileSize  int      `yaml:"max_file_size_kb,omitempty"`
	BannedImports []string `yaml:"banned_imports,omitempty"`
}

// SDDTestReq specifies a test requirement.
type SDDTestReq struct {
	Type     string `yaml:"type"`     // unit, integration, e2e
	What     string `yaml:"what"`
	Coverage string `yaml:"coverage,omitempty"`
}

// SDDCriterion is a single acceptance criterion.
type SDDCriterion struct {
	Criterion string `yaml:"criterion"`
	Met       bool   `yaml:"met"`
}

// WorkLog records execution details (mandatory).
type WorkLog struct {
	Executor     string        `yaml:"executor"`
	StartedAt    time.Time     `yaml:"started,omitempty"`
	CompletedAt  time.Time     `yaml:"completed,omitempty"`
	Duration     string        `yaml:"duration,omitempty"`
	Decisions    []Decision    `yaml:"decisions,omitempty"`
	Output       WorkLogOutput `yaml:"output"`
	Retrospective Retrospective `yaml:"retrospective"`
}

// Decision records a choice made during execution.
type Decision struct {
	What string `yaml:"what"`
	Why  string `yaml:"why"`
}

// WorkLogOutput records what was produced.
type WorkLogOutput struct {
	Commits      []string `yaml:"commits,omitempty"`
	PR           string   `yaml:"pr,omitempty"`
	FilesChanged []string `yaml:"files_changed,omitempty"`
}

// Retrospective records lessons learned.
type Retrospective struct {
	Worked      string `yaml:"worked,omitempty"`
	DidntWork   string `yaml:"didnt_work,omitempty"`
	Suggestions string `yaml:"suggestions,omitempty"`
}

// LogValue controls slog output for SDD.
func (sdd SDD) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", sdd.ID),
		slog.String("fd", sdd.FD),
		slog.String("status", string(sdd.Status)),
		slog.String("complexity", sdd.Complexity),
	)
}
