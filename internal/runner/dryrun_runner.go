package runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/Deepzima/forgia/internal/vault"
)

// Compile-time interface check.
var _ Runner = (*DryRunRunner)(nil)

// DryRunRunner wraps the dry-run simulation as a Runner interface.
type DryRunRunner struct {
	logger *slog.Logger
}

// NewDryRunRunner creates a dry-run runner.
func NewDryRunRunner() *DryRunRunner {
	return &DryRunRunner{
		logger: slog.With("runner", "dry-run"),
	}
}

// Name returns "dry-run".
func (r *DryRunRunner) Name() string {
	return "dry-run"
}

// Execute simulates execution and returns a stub result.
// For full dry-run analysis, use the DryRun() function directly.
func (r *DryRunRunner) Execute(ctx context.Context, sdd *vault.SDD, opts ExecOptions) (*ExecResult, error) {
	r.logger.InfoContext(ctx, "dry-run simulation", "sdd", sdd.ID)

	return &ExecResult{
		SDD:          sdd.ID,
		FD:           sdd.FD,
		Runner:       "dry-run",
		Started:      time.Now(),
		Completed:    time.Now(),
		DurationSecs: 0,
		ExitCode:     0,
		Status:       "dry-run",
	}, nil
}
