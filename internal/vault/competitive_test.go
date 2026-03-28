package vault

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRejectCompetitors(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	v, err := InitVault(ctx, dir, InitOptions{ProjectName: "test"})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	fv := v.(*FileVault)

	// Create 3 competing FDs.
	fdA := &FD{ID: "FD-aaa", Title: "Approach A", Status: FDPlanned, CompetesWith: []string{"FD-bbb", "FD-ccc"}, UpstreamIssue: "test/repo#1"}
	fdB := &FD{ID: "FD-bbb", Title: "Approach B", Status: FDPlanned, CompetesWith: []string{"FD-aaa", "FD-ccc"}, UpstreamIssue: "test/repo#1"}
	fdC := &FD{ID: "FD-ccc", Title: "Approach C", Status: FDPlanned, CompetesWith: []string{"FD-aaa", "FD-bbb"}, UpstreamIssue: "test/repo#1"}

	for _, fd := range []*FD{fdA, fdB, fdC} {
		if err := v.CreateFD(ctx, fd); err != nil {
			t.Fatalf("create %s: %v", fd.ID, err)
		}
	}

	// Approve FD-aaa → reject FD-bbb and FD-ccc.
	rejected, err := fv.RejectCompetitors(ctx, "FD-aaa", "A chosen for performance reasons")
	if err != nil {
		t.Fatalf("RejectCompetitors: %v", err)
	}
	if rejected != 2 {
		t.Errorf("expected 2 rejected, got %d", rejected)
	}

	// Verify FD-bbb is rejected.
	fdBCheck, _ := v.GetFD(ctx, "FD-bbb")
	if fdBCheck.Status != FDRejected {
		t.Errorf("FD-bbb status = %q, want rejected", fdBCheck.Status)
	}
	if fdBCheck.SupersededBy != "FD-aaa" {
		t.Errorf("FD-bbb superseded_by = %q, want FD-aaa", fdBCheck.SupersededBy)
	}
	if fdBCheck.RejectedReason != "A chosen for performance reasons" {
		t.Errorf("FD-bbb rejected_reason = %q", fdBCheck.RejectedReason)
	}

	// Verify FD-ccc is rejected.
	fdCCheck, _ := v.GetFD(ctx, "FD-ccc")
	if fdCCheck.Status != FDRejected {
		t.Errorf("FD-ccc status = %q, want rejected", fdCCheck.Status)
	}
}

func TestRejectCompetitors_Idempotent(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	v, _ := InitVault(ctx, dir, InitOptions{ProjectName: "test"})
	fv := v.(*FileVault)

	fdA := &FD{ID: "FD-aaa", Title: "A", Status: FDPlanned, CompetesWith: []string{"FD-bbb"}}
	fdB := &FD{ID: "FD-bbb", Title: "B", Status: FDPlanned, CompetesWith: []string{"FD-aaa"}}
	v.CreateFD(ctx, fdA)
	v.CreateFD(ctx, fdB)

	// Reject twice — second call should be no-op.
	fv.RejectCompetitors(ctx, "FD-aaa", "reason")
	rejected2, err := fv.RejectCompetitors(ctx, "FD-aaa", "reason")
	if err != nil {
		t.Fatalf("second RejectCompetitors: %v", err)
	}
	if rejected2 != 0 {
		t.Errorf("expected 0 on second call, got %d", rejected2)
	}
}

func TestRejectCompetitors_NoCompetitors(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	v, _ := InitVault(ctx, dir, InitOptions{ProjectName: "test"})
	fv := v.(*FileVault)

	fd := &FD{ID: "FD-solo", Title: "Solo", Status: FDPlanned}
	v.CreateFD(ctx, fd)

	rejected, err := fv.RejectCompetitors(ctx, "FD-solo", "reason")
	if err != nil {
		t.Fatalf("RejectCompetitors: %v", err)
	}
	if rejected != 0 {
		t.Errorf("expected 0, got %d", rejected)
	}
}

func TestFindCompetitors(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	v, _ := InitVault(ctx, dir, InitOptions{ProjectName: "test"})
	fv := v.(*FileVault)

	fdA := &FD{ID: "FD-aaa", Title: "A", Status: FDPlanned, CompetesWith: []string{"FD-bbb"}, UpstreamIssue: "repo#1"}
	fdB := &FD{ID: "FD-bbb", Title: "B", Status: FDPlanned, CompetesWith: []string{"FD-aaa"}, UpstreamIssue: "repo#1"}
	fdC := &FD{ID: "FD-ccc", Title: "C", Status: FDPlanned, UpstreamIssue: "repo#1"} // same issue but no competes_with
	fdD := &FD{ID: "FD-ddd", Title: "D", Status: FDPlanned, UpstreamIssue: "repo#2"} // different issue

	for _, fd := range []*FD{fdA, fdB, fdC, fdD} {
		v.CreateFD(ctx, fd)
	}

	competitors, err := fv.FindCompetitors(ctx, "FD-aaa")
	if err != nil {
		t.Fatalf("FindCompetitors: %v", err)
	}

	// Should find FD-bbb (competes_with) and FD-ccc (shared upstream_issue), not FD-ddd
	if len(competitors) != 2 {
		names := make([]string, len(competitors))
		for i, c := range competitors {
			names[i] = c.ID
		}
		t.Fatalf("expected 2 competitors, got %d: %v", len(competitors), names)
	}

	// Ensure we don't have cleanup issues with test temp dirs.
	_ = os.RemoveAll(filepath.Join(dir, ".forgia"))
}
