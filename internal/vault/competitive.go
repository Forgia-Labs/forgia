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
