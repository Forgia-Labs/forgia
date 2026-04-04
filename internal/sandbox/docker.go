package sandbox

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// Compile-time interface check.
var _ Provider = (*DockerProvider)(nil)

// DockerProvider runs commands in Docker containers.
// Provides namespace-level isolation with optional seccomp enforcement.
type DockerProvider struct {
	logger *slog.Logger
}

// Name returns "docker".
func (p *DockerProvider) Name() string {
	return "docker"
}

// Available checks if docker CLI is installed and daemon is running.
func (p *DockerProvider) Available(ctx context.Context) bool {
	if !isCommandAvailable("docker") {
		return false
	}
	cmd := exec.CommandContext(ctx, "docker", "info")
	return cmd.Run() == nil
}

// Run executes a command inside a Docker container.
func (p *DockerProvider) Run(ctx context.Context, opts RunOpts) (*RunResult, error) {
	if p.logger == nil {
		p.logger = slog.With("sandbox", "docker")
	}

	args := p.buildArgs(opts)

	p.logger.InfoContext(ctx, "running in Docker",
		"image", opts.Image, "command", opts.Command)

	cmd := exec.CommandContext(ctx, "docker", args...)
	// Pass stdout/stderr to terminal so user sees Claude's progress.
	// Stdin is nil (closed) — Claude in -p mode doesn't need interactive input.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	result := &RunResult{}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return result, fmt.Errorf("docker run: %w", err)
		}
	}

	return result, nil
}

// buildArgs constructs the `docker run` argument list.
func (p *DockerProvider) buildArgs(opts RunOpts) []string {
	args := []string{"run", "--rm"}

	// Volume mounts.
	for _, m := range opts.Mounts {
		mountStr := fmt.Sprintf("%s:%s", m.Source, m.Target)
		if m.ReadOnly {
			mountStr += ":ro"
		}
		args = append(args, "-v", mountStr)
	}

	// Environment variables.
	for k, v := range opts.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Working directory.
	if opts.WorkDir != "" {
		args = append(args, "-w", opts.WorkDir)
	}

	// Network mode.
	switch opts.NetworkMode {
	case "none":
		args = append(args, "--network=none")
	case "host":
		args = append(args, "--network=host")
	case "", "bridge":
		// Docker default — no flag needed.
	}

	// Seccomp profile (Docker-only).
	if opts.SeccompPath != "" {
		args = append(args, "--security-opt", fmt.Sprintf("seccomp=%s", opts.SeccompPath))
	}

	// Image.
	args = append(args, opts.Image)

	// Command.
	args = append(args, opts.Command...)

	return args
}
