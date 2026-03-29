package sandbox

import (
	"context"
	"testing"
)

func TestAppleProvider_Name(t *testing.T) {
	t.Parallel()
	p := &AppleProvider{}
	if p.Name() != "apple-container" {
		t.Errorf("expected 'apple-container', got %q", p.Name())
	}
}

func TestDockerProvider_Name(t *testing.T) {
	t.Parallel()
	p := &DockerProvider{}
	if p.Name() != "docker" {
		t.Errorf("expected 'docker', got %q", p.Name())
	}
}

func TestAppleProvider_BuildArgs(t *testing.T) {
	t.Parallel()
	p := &AppleProvider{}

	opts := RunOpts{
		Image:   "ubuntu:latest",
		Command: []string{"uname", "-a"},
		WorkDir: "/workspace",
		Mounts: []Mount{
			{Source: "/host/project", Target: "/workspace", ReadOnly: false},
			{Source: "/host/readonly", Target: "/data", ReadOnly: true},
		},
		Env: map[string]string{"FOO": "bar"},
	}

	args := p.buildArgs(opts)

	// Verify image is present.
	found := false
	for _, a := range args {
		if a == "ubuntu:latest" {
			found = true
		}
	}
	if !found {
		t.Error("image not in args")
	}

	// Verify mount with readonly.
	foundRO := false
	for _, a := range args {
		if a == "type=bind,source=/host/readonly,target=/data,readonly" {
			foundRO = true
		}
	}
	if !foundRO {
		t.Errorf("readonly mount not found in args: %v", args)
	}

	// Verify workdir.
	foundWD := false
	for i, a := range args {
		if a == "-w" && i+1 < len(args) && args[i+1] == "/workspace" {
			foundWD = true
		}
	}
	if !foundWD {
		t.Errorf("workdir not found in args: %v", args)
	}
}

func TestDockerProvider_BuildArgs(t *testing.T) {
	t.Parallel()
	p := &DockerProvider{}

	opts := RunOpts{
		Image:       "ubuntu:latest",
		Command:     []string{"bash", "-c", "echo hello"},
		WorkDir:     "/workspace",
		Mounts:      []Mount{{Source: "/host", Target: "/workspace"}},
		NetworkMode: "none",
		SeccompPath: "/tmp/seccomp.json",
	}

	args := p.buildArgs(opts)

	// Verify --rm.
	found := false
	for _, a := range args {
		if a == "--rm" {
			found = true
		}
	}
	if !found {
		t.Error("--rm not in args")
	}

	// Verify network=none.
	foundNet := false
	for _, a := range args {
		if a == "--network=none" {
			foundNet = true
		}
	}
	if !foundNet {
		t.Error("--network=none not in args")
	}

	// Verify seccomp.
	foundSec := false
	for _, a := range args {
		if a == "seccomp=/tmp/seccomp.json" {
			foundSec = true
		}
	}
	if !foundSec {
		t.Errorf("seccomp not in args: %v", args)
	}
}

func TestDockerProvider_BuildArgs_ReadOnlyMount(t *testing.T) {
	t.Parallel()
	p := &DockerProvider{}

	opts := RunOpts{
		Image:   "alpine",
		Command: []string{"ls"},
		Mounts:  []Mount{{Source: "/host", Target: "/data", ReadOnly: true}},
	}

	args := p.buildArgs(opts)
	foundRO := false
	for _, a := range args {
		if a == "/host:/data:ro" {
			foundRO = true
		}
	}
	if !foundRO {
		t.Errorf("readonly mount not found: %v", args)
	}
}

func TestResolve_None(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	p, err := Resolve(ctx, "none")
	if err != nil {
		t.Fatalf("Resolve(none) error: %v", err)
	}
	if p != nil {
		t.Error("expected nil provider for 'none'")
	}

	p2, err := Resolve(ctx, "")
	if err != nil {
		t.Fatalf("Resolve('') error: %v", err)
	}
	if p2 != nil {
		t.Error("expected nil provider for empty string")
	}
}

func TestResolve_Unknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := Resolve(ctx, "unknown-provider")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestResolve_Docker(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Docker should be available in our dev environment.
	p, err := Resolve(ctx, "docker")
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	if p.Name() != "docker" {
		t.Errorf("expected 'docker', got %q", p.Name())
	}
}

func TestResolve_AppleContainer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	p, err := Resolve(ctx, "apple-container")
	if err != nil {
		t.Skipf("apple-container not available: %v", err)
	}
	if p.Name() != "apple-container" {
		// Could be docker fallback.
		t.Logf("resolved to %q (fallback)", p.Name())
	}
}
