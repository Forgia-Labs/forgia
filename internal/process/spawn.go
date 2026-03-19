// Package process manages subprocess spawning, monitoring, and lifecycle.
// Used by runners (claude, docker) and MCP providers (subprocess).
package process

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"syscall"
	"time"
)

// Process represents a managed subprocess.
// Exported fields (Name, Cmd, Stdout, Stderr) are read-only for callers.
// Internal state (done, err, cancel) is accessed via methods (Done, Err, Kill).
type Process struct {
	Name   string
	Cmd    *exec.Cmd
	Stdout io.ReadCloser
	Stderr io.ReadCloser

	done chan struct{} // closed when process exits (safe for multiple readers)
	err  error         // exit error, valid after done is closed
}

// Spawn starts a subprocess with monitoring.
// The ctx controls the subprocess lifetime — cancelling ctx kills the process.
func Spawn(ctx context.Context, name string, args ...string) (*Process, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe %s: %w", name, err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe %s: %w", name, err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("spawn %s: %w", name, err)
	}

	p := &Process{
		Name:   name,
		Cmd:    cmd,
		Stdout: stdout,
		Stderr: stderr,
		done:   make(chan struct{}),
	}

	// Monitor in background — close channel on exit (safe for multiple readers).
	go func() {
		p.err = cmd.Wait()
		close(p.done)
	}()

	return p, nil
}

// Kill sends SIGTERM, waits 5s, then SIGKILL.
// Safe to call from multiple goroutines (done channel uses close pattern).
func (p *Process) Kill() error {
	if p.Cmd.Process == nil {
		return nil
	}
	_ = p.Cmd.Process.Signal(syscall.SIGTERM)
	select {
	case <-p.done:
		return p.err
	case <-time.After(5 * time.Second):
		return p.Cmd.Process.Kill()
	}
}

// Done returns a channel that is closed when the process exits.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// Err returns the exit error. Only valid after Done() is closed.
func (p *Process) Err() error {
	return p.err
}
