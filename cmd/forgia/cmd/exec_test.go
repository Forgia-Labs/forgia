package cmd

import (
	"testing"
)

func TestExecCommand_FlagsRegistered(t *testing.T) {
	t.Parallel()

	dryRunFlag := execCmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("expected --dry-run flag to be registered on exec command")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("expected --dry-run default to be false, got %s", dryRunFlag.DefValue)
	}

	runnerFlag := execCmd.Flags().Lookup("runner")
	if runnerFlag == nil {
		t.Fatal("expected --runner flag to be registered on exec command")
	}
	if runnerFlag.DefValue != "" {
		t.Errorf("expected --runner default to be empty, got %s", runnerFlag.DefValue)
	}
}

func TestExecCommand_RegisteredWithRoot(t *testing.T) {
	t.Parallel()

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "exec" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected exec command to be registered with root")
	}
}

func TestExecCommand_RequiresArgs(t *testing.T) {
	t.Parallel()

	if execCmd.Args == nil {
		t.Fatal("expected exec command to have Args validation")
	}
	if err := execCmd.Args(execCmd, []string{}); err == nil {
		t.Error("expected error when no args provided")
	}
	if err := execCmd.Args(execCmd, []string{"file.md"}); err != nil {
		t.Errorf("expected no error with one arg, got %v", err)
	}
}
