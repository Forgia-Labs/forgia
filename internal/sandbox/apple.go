package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

// Compile-time interface check.
var _ Provider = (*AppleProvider)(nil)

// AppleProvider runs commands in Apple Container lightweight VMs.
// Provides kernel-level isolation (separate Linux kernel per execution).
// Requires macOS 26+ with Apple Silicon and `container` CLI installed.
type AppleProvider struct {
	logger *slog.Logger
}

// Name returns "apple-container".
func (p *AppleProvider) Name() string {
	return "apple-container"
}

// Available checks if the container CLI is installed and the system service is running.
func (p *AppleProvider) Available(ctx context.Context) bool {
	if !isCommandAvailable("container") {
		return false
	}
	// Check if system service is running.
	cmd := exec.CommandContext(ctx, "container", "system", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "running")
}

// Run executes a command inside an Apple Container VM.
func (p *AppleProvider) Run(ctx context.Context, opts RunOpts) (*RunResult, error) {
	if p.logger == nil {
		p.logger = slog.With("sandbox", "apple-container")
	}

	args := p.buildArgs(opts)

	p.logger.InfoContext(ctx, "running in Apple Container",
		"image", opts.Image, "command", opts.Command)

	cmd := exec.CommandContext(ctx, "container", args...)
	output, err := cmd.CombinedOutput()

	result := &RunResult{
		Output: output,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return result, fmt.Errorf("apple-container run: %w", err)
		}
	}

	return result, nil
}

// buildArgs constructs the `container run` argument list.
func (p *AppleProvider) buildArgs(opts RunOpts) []string {
	args := []string{"run"}

	// Volume mounts.
	for _, m := range opts.Mounts {
		mountSpec := fmt.Sprintf("type=bind,source=%s,target=%s", m.Source, m.Target)
		if m.ReadOnly {
			mountSpec += ",readonly"
		}
		args = append(args, "--mount", mountSpec)
	}

	// Environment variables.
	for k, v := range opts.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Working directory.
	if opts.WorkDir != "" {
		args = append(args, "-w", opts.WorkDir)
	}

	// Image.
	args = append(args, opts.Image)

	// Command.
	args = append(args, opts.Command...)

	return args
}
