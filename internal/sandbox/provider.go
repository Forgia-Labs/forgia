// Package sandbox provides isolated execution environments for SDD runners.
// Supports multiple backends: Apple Container (VM isolation) and Docker (namespace isolation).
package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Provider creates and manages sandboxed execution environments.
type Provider interface {
	// Name returns the provider identifier ("apple-container", "docker").
	Name() string

	// Available returns true if the provider's runtime is installed and ready.
	Available(ctx context.Context) bool

	// Run executes a command inside a sandbox and returns the result.
	Run(ctx context.Context, opts RunOpts) (*RunResult, error)
}

// RunOpts configures a sandbox execution.
type RunOpts struct {
	Image       string            // OCI image (e.g., "ubuntu:latest")
	Command     []string          // command + args to run inside sandbox
	WorkDir     string            // working directory inside sandbox
	Mounts      []Mount           // volume mounts
	Env         map[string]string // environment variables
	NetworkMode string            // "none", "host", or allowlist proxy
	NetworkAllow []string         // allowed egress domains (when NetworkMode is filtered)
	SeccompPath string            // path to seccomp profile (Docker only)
}

// Mount describes a bind mount into the sandbox.
type Mount struct {
	Source   string // host path
	Target   string // container path
	ReadOnly bool
}

// RunResult holds the outcome of a sandbox execution.
type RunResult struct {
	ExitCode int
	Output   []byte // combined stdout+stderr
}

// Resolve returns the appropriate provider based on name.
// Falls back to docker if apple-container is not available.
func Resolve(ctx context.Context, name string) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "apple-container":
		p := &AppleProvider{}
		if p.Available(ctx) {
			return p, nil
		}
		// Fallback to Docker.
		d := &DockerProvider{}
		if d.Available(ctx) {
			return d, nil
		}
		return nil, fmt.Errorf("sandbox: apple-container not available, docker fallback also unavailable")
	case "docker":
		p := &DockerProvider{}
		if !p.Available(ctx) {
			return nil, fmt.Errorf("sandbox: docker not available")
		}
		return p, nil
	case "", "none":
		return nil, nil // no sandbox — run on host
	default:
		return nil, fmt.Errorf("sandbox: unknown provider %q (available: apple-container, docker, none)", name)
	}
}

// isCommandAvailable checks if a CLI tool is in PATH.
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
