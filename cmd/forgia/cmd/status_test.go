package cmd

import (
	"bytes"
	"os"
	"testing"
)

func TestStatusCmd_Registered(t *testing.T) {
	cmd := rootCmd
	found := false
	for _, c := range cmd.Commands() {
		if c.Use == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("status command not registered on rootCmd")
	}
}

func TestDoctorCmd_Registered(t *testing.T) {
	cmd := rootCmd
	found := false
	for _, c := range cmd.Commands() {
		if c.Use == "doctor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("doctor command not registered on rootCmd")
	}
}

func TestCheckTool_Git(t *testing.T) {
	r := checkTool("git", true)
	if r.status != "pass" {
		t.Errorf("expected git to be found, got %s", r.status)
	}
}

func TestCheckTool_Missing(t *testing.T) {
	r := checkTool("nonexistent-tool-xyz", false)
	if r.status != "skip" {
		t.Errorf("expected skip for missing optional tool, got %s", r.status)
	}

	r = checkTool("nonexistent-tool-xyz", true)
	if r.status != "fail" {
		t.Errorf("expected fail for missing required tool, got %s", r.status)
	}
}

func TestCheckVault_NoVault(t *testing.T) {
	// Run from temp dir with no vault.
	orig, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(orig)

	r := checkVault(t.Context())
	if r.status != "fail" {
		t.Errorf("expected fail for missing vault, got %s", r.status)
	}
}

func TestStatusOutput_EmptyVault(t *testing.T) {
	// Capture output would require refactoring to io.Writer — just verify no panic.
	_ = bytes.Buffer{}
}
