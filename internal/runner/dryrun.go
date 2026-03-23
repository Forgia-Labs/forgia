package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Deepzima/forgia/internal/vault"
)

// Feasibility represents the dry-run verdict.
type Feasibility string

const (
	FeasibilityGo      Feasibility = "GO"
	FeasibilityNoGo    Feasibility = "NO-GO"
	FeasibilityCaution Feasibility = "CAUTION"
)

// DryRunOptions configures a dry-run invocation.
type DryRunOptions struct {
	SlashCommandPath string // path to sdd-dry-run.md
	SystemContext    string // constitution + guardrails + principles
	MaxTurns         int
}

// DryRunResult holds the parsed dry-run report.
type DryRunResult struct {
	Feasibility Feasibility
	Confidence  string
	Complexity  string
	ReportText  string
	Blockers    []string
	Warnings    []string
}

// feasibilityRe matches "Feasibility: GO", "Feasibility: NO-GO", "Feasibility: CAUTION".
var feasibilityRe = regexp.MustCompile(`(?i)Feasibility:\s*(GO|NO-GO|CAUTION)`)

// DryRun runs a dry-run simulation for the given SDD via Claude CLI.
func DryRun(ctx context.Context, sdd *vault.SDD, opts DryRunOptions) (*DryRunResult, error) {
	if _, err := exec.LookPath("claude"); err != nil {
		return nil, fmt.Errorf("'claude' CLI non trovato: installa Claude Code prima")
	}

	prompt := fmt.Sprintf(
		"Run /sdd-dry-run on %s. Read the slash command instructions from %s and follow them exactly.",
		sdd.FilePath, opts.SlashCommandPath,
	)

	args := []string{"-p", prompt}
	if opts.SystemContext != "" {
		args = append([]string{"--append-system-prompt", opts.SystemContext}, args...)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if err != nil {
		return nil, fmt.Errorf("claude invocation failed: %w\nstderr: %s", err, stderr.String())
	}

	result := ParseDryRunReport(output)
	return result, nil
}

// ParseDryRunReport extracts structured data from a dry-run report text.
func ParseDryRunReport(report string) *DryRunResult {
	result := &DryRunResult{
		ReportText:  report,
		Feasibility: FeasibilityNoGo, // default to NO-GO if unparseable
	}

	if match := feasibilityRe.FindStringSubmatch(report); len(match) > 1 {
		switch strings.ToUpper(match[1]) {
		case "GO":
			result.Feasibility = FeasibilityGo
		case "NO-GO":
			result.Feasibility = FeasibilityNoGo
		case "CAUTION":
			result.Feasibility = FeasibilityCaution
		}
	}

	// Parse confidence
	confRe := regexp.MustCompile(`(?i)Confidence:\s*(\d+%?)`)
	if match := confRe.FindStringSubmatch(report); len(match) > 1 {
		result.Confidence = match[1]
	}

	// Parse complexity
	compRe := regexp.MustCompile(`(?i)Complexity:\s*(low|medium|high|very-high)`)
	if match := compRe.FindStringSubmatch(report); len(match) > 1 {
		result.Complexity = match[1]
	}

	// Parse blockers (lines starting with "x " under BLOCKERS)
	for _, line := range strings.Split(report, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "x ") || strings.HasPrefix(trimmed, "✗ ") {
			result.Blockers = append(result.Blockers, trimmed)
		}
		if strings.HasPrefix(trimmed, "! ") || strings.HasPrefix(trimmed, "⚠ ") {
			result.Warnings = append(result.Warnings, trimmed)
		}
	}

	return result
}

// IsGo returns true if execution should proceed (GO or CAUTION).
func (r *DryRunResult) IsGo() bool {
	return r.Feasibility == FeasibilityGo || r.Feasibility == FeasibilityCaution
}
