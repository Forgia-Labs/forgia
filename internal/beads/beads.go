// Package beads provides a resilient client for the Beads (bd) task tracker.
// Beads is ALWAYS optional — Forgia works 100% without it.
// When available, it provides fast local caching and dependency graph.
package beads

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	// ErrBeadsUnavailable means bd is not installed or circuit breaker is open.
	ErrBeadsUnavailable = errors.New("beads unavailable")

	// ErrBeadsTimeout means a bd command timed out.
	ErrBeadsTimeout = errors.New("beads call timed out")
)

// Client wraps bd CLI with timeout and graceful fallback.
type Client struct {
	available bool
	timeout   time.Duration
}

// NewClient creates a beads client, checking availability.
func NewClient() *Client {
	bc := &Client{timeout: 5 * time.Second}

	if _, err := exec.LookPath("bd"); err != nil {
		bc.available = false
		return bc
	}

	if isCircuitBreakerOpen() {
		slog.Warn("beads circuit breaker is open — using fallback",
			"fix", "mise run bd:reset")
		bc.available = false
		return bc
	}

	bc.available = true
	return bc
}

// Available returns true if beads is usable.
func (bc *Client) Available() bool {
	return bc.available
}

// Call executes a bd command with timeout and fallback.
func (bc *Client) Call(ctx context.Context, args ...string) (string, error) {
	if !bc.available {
		return "", ErrBeadsUnavailable
	}

	ctx, cancel := context.WithTimeout(ctx, bc.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bd", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			slog.Warn("beads call timed out", "args", args)
			return "", ErrBeadsTimeout
		}
		return "", fmt.Errorf("bd %v: %w", args, err)
	}

	return string(output), nil
}

// isCircuitBreakerOpen checks /tmp/ for open circuit breaker files.
func isCircuitBreakerOpen() bool {
	patterns := []string{
		"/tmp/beads-dolt-circuit-*.json",
		"/private/tmp/beads-dolt-circuit-*.json",
	}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, f := range matches {
			data, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			if strings.Contains(string(data), `"state":"open"`) {
				return true
			}
		}
	}
	return false
}
