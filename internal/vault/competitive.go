package vault

import (
	"context"
	"fmt"
	"log/slog"
)

// RejectCompetitors finds all FDs competing with the approved FD and marks them
// as rejected with the given reason. Sets superseded_by to the approved FD's ID.
// Idempotent: already-rejected FDs are skipped.
func (v *FileVault) RejectCompetitors(ctx context.Context, approvedID, reason string) (int, error) {
	// Use FindCompetitors to catch all competitors (bidirectional competes_with + shared upstream_issue).
	competitors, err := v.FindCompetitors(ctx, approvedID)
	if err != nil {
		return 0, fmt.Errorf("find competitors: %w", err)
	}
	if len(competitors) == 0 {
		return 0, nil
	}

	rejected := 0
	for _, fd := range competitors {

		// Skip if already rejected.
		if fd.Status == FDRejected {
			continue
		}

		fd.Status = FDRejected
		fd.RejectedReason = reason
		fd.SupersededBy = approvedID

		if err := v.UpdateFD(ctx, fd); err != nil {
			return rejected, fmt.Errorf("reject FD %s: %w", fd.ID, err)
		}

		slog.InfoContext(ctx, "competitor rejected",
			"rejected", fd.ID, "superseded_by", approvedID)
		rejected++
	}

	return rejected, nil
}

// FindCompetitors returns all FDs that compete with the given FD ID.
// Searches by competes_with field and shared upstream_issue.
func (v *FileVault) FindCompetitors(ctx context.Context, fdID string) ([]*FD, error) {
	target, err := v.GetFD(ctx, fdID)
	if err != nil {
		return nil, fmt.Errorf("get FD: %w", err)
	}

	allFDs, err := v.ListFDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list FDs: %w", err)
	}

	seen := make(map[string]bool)
	var competitors []*FD

	addIfNew := func(fd *FD) {
		if fd.ID == fdID || seen[fd.ID] {
			return
		}
		seen[fd.ID] = true
		competitors = append(competitors, fd)
	}

	// Direction 1: target's own competes_with list.
	for _, id := range target.CompetesWith {
		for _, fd := range allFDs {
			if fd.ID == id {
				addIfNew(fd)
				break
			}
		}
	}

	// Direction 2: other FDs that list target in their competes_with.
	for _, fd := range allFDs {
		for _, cw := range fd.CompetesWith {
			if cw == fdID {
				addIfNew(fd)
				break
			}
		}
	}

	// Direction 3: shared upstream_issue.
	if target.UpstreamIssue != "" {
		for _, fd := range allFDs {
			if fd.UpstreamIssue == target.UpstreamIssue {
				addIfNew(fd)
			}
		}
	}

	return competitors, nil
}
