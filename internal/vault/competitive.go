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
	approved, err := v.GetFD(ctx, approvedID)
	if err != nil {
		return 0, fmt.Errorf("get approved FD: %w", err)
	}

	if len(approved.CompetesWith) == 0 {
		return 0, nil // no competitors
	}

	rejected := 0
	for _, competitorID := range approved.CompetesWith {
		fd, err := v.GetFD(ctx, competitorID)
		if err != nil {
			slog.WarnContext(ctx, "competitor FD not found", "id", competitorID, "error", err)
			continue
		}

		// Skip if already rejected.
		if fd.Status == FDRejected {
			continue
		}

		fd.Status = FDRejected
		fd.RejectedReason = reason
		fd.SupersededBy = approvedID

		if err := v.UpdateFD(ctx, fd); err != nil {
			return rejected, fmt.Errorf("reject FD %s: %w", competitorID, err)
		}

		slog.InfoContext(ctx, "competitor rejected",
			"rejected", competitorID, "superseded_by", approvedID)
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

	var competitors []*FD
	for _, fd := range allFDs {
		if fd.ID == fdID {
			continue
		}

		// Check competes_with.
		for _, cw := range fd.CompetesWith {
			if cw == fdID {
				competitors = append(competitors, fd)
				break
			}
		}

		// Check shared upstream_issue.
		if target.UpstreamIssue != "" && fd.UpstreamIssue == target.UpstreamIssue {
			// Avoid duplicates.
			found := false
			for _, c := range competitors {
				if c.ID == fd.ID {
					found = true
					break
				}
			}
			if !found {
				competitors = append(competitors, fd)
			}
		}
	}

	return competitors, nil
}
