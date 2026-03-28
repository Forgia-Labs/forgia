package cmd

import (
	"testing"
)

func TestBatchCommand_FlagsRegistered(t *testing.T) {
	t.Parallel()

	dryRunFlag := batchCmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("expected --dry-run flag to be registered on batch command")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("expected --dry-run default to be false, got %s", dryRunFlag.DefValue)
	}

	runnerFlag := batchCmd.Flags().Lookup("runner")
	if runnerFlag == nil {
		t.Fatal("expected --runner flag to be registered on batch command")
	}
	if runnerFlag.DefValue != "" {
		t.Errorf("expected --runner default to be empty, got %s", runnerFlag.DefValue)
	}
}

func TestBatchCommand_RegisteredWithRoot(t *testing.T) {
	t.Parallel()

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "batch" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected batch command to be registered with root")
	}
}

func TestBatchCommand_RequiresArgs(t *testing.T) {
	t.Parallel()

	if batchCmd.Args == nil {
		t.Fatal("expected batch command to have Args validation")
	}
	if err := batchCmd.Args(batchCmd, []string{}); err == nil {
		t.Error("expected error when no args provided")
	}
	if err := batchCmd.Args(batchCmd, []string{"FD-001"}); err != nil {
		t.Errorf("expected no error with one arg, got %v", err)
	}
}
