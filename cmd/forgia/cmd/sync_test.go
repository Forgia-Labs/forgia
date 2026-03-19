package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// resetCmd resets global rootCmd state between tests.
// Cobra caches help flag state, so we need to reset it.
func resetCmd(args []string) *bytes.Buffer {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	// Reset help flag cache from previous --help calls.
	syncCmd.Flags().Set("help", "false")
	return buf
}

func TestSyncCommandRegistered(t *testing.T) {
	buf := resetCmd([]string{"help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "sync") {
		t.Error("expected 'sync' in help output")
	}
}

func TestSyncCommandHelp(t *testing.T) {
	buf := resetCmd([]string{"sync", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "push") {
		t.Error("expected '--push' in sync help")
	}
	if !strings.Contains(got, "pull") {
		t.Error("expected '--pull' in sync help")
	}
}

func TestSyncCommandNoVault(t *testing.T) {
	_ = resetCmd([]string{"sync"})

	// Change to a temp dir with no .forgia/.
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no .forgia/ exists")
	}
	if !strings.Contains(err.Error(), "vault") {
		t.Errorf("expected vault-related error, got: %v", err)
	}
}
