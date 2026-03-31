package sandbox

import (
	"encoding/json"
	"testing"

	"github.com/forgia-labs/forgia/internal/guardrails"
)

func TestGenerateSeccompJSON_ValidOutput(t *testing.T) {
	t.Parallel()

	g := &guardrails.Guardrails{
		Read: guardrails.DenyList{Patterns: []string{
			"**/.env", "**/*.pem", "!**/.env.example",
		}},
		Write: guardrails.DenyList{Patterns: []string{
			"**/.env", ".forgia/constitution.md",
		}},
		Execute: guardrails.DenyList{Patterns: []string{
			"pass show*", "gpg --export-secret*",
		}},
	}

	data, err := GenerateSeccompJSON(g)
	if err != nil {
		t.Fatalf("GenerateSeccompJSON: %v", err)
	}

	// Verify valid JSON.
	var profile SeccompProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Default action should be ALLOW.
	if profile.DefaultAction != "SCMP_ACT_ALLOW" {
		t.Errorf("expected SCMP_ACT_ALLOW, got %q", profile.DefaultAction)
	}

	// Should have syscall rules.
	if len(profile.Syscalls) == 0 {
		t.Error("expected syscall rules")
	}

	// ptrace should be blocked.
	foundPtrace := false
	for _, rule := range profile.Syscalls {
		for _, name := range rule.Names {
			if name == "ptrace" {
				foundPtrace = true
				if rule.Action != "SCMP_ACT_ERRNO" {
					t.Errorf("ptrace action = %q, want SCMP_ACT_ERRNO", rule.Action)
				}
			}
		}
	}
	if !foundPtrace {
		t.Error("ptrace not blocked")
	}

	// BlockedPaths should contain deny patterns (without ! exceptions).
	if len(profile.BlockedPaths) != 4 {
		t.Errorf("expected 4 blocked paths, got %d: %v", len(profile.BlockedPaths), profile.BlockedPaths)
	}
}

func TestGenerateSeccompJSON_EmptyGuardrails(t *testing.T) {
	t.Parallel()

	g := &guardrails.Guardrails{}
	data, err := GenerateSeccompJSON(g)
	if err != nil {
		t.Fatalf("GenerateSeccompJSON: %v", err)
	}

	var profile SeccompProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Should still have base syscall rules.
	if len(profile.Syscalls) == 0 {
		t.Error("expected base syscall rules even with empty guardrails")
	}
	if len(profile.BlockedPaths) != 0 {
		t.Errorf("expected 0 blocked paths, got %d", len(profile.BlockedPaths))
	}
}

func TestWorkspaceMounts(t *testing.T) {
	t.Parallel()

	mounts := WorkspaceMounts("/home/user/project")
	if len(mounts) != 1 {
		t.Fatalf("expected 1 mount, got %d", len(mounts))
	}
	if mounts[0].Source != "/home/user/project" {
		t.Errorf("source = %q", mounts[0].Source)
	}
	if mounts[0].Target != "/workspace" {
		t.Errorf("target = %q", mounts[0].Target)
	}
	if mounts[0].ReadOnly {
		t.Error("workspace should be read-write")
	}
}

func TestExcludedHostPaths(t *testing.T) {
	t.Parallel()

	paths := ExcludedHostPaths()
	if len(paths) == 0 {
		t.Fatal("expected excluded paths")
	}

	// ~/.ssh must be excluded.
	found := false
	for _, p := range paths {
		if p == "~/.ssh" {
			found = true
		}
	}
	if !found {
		t.Error("~/.ssh not in excluded paths")
	}
}
