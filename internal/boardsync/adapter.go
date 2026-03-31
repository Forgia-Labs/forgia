// Package boardsync adapts vault types to board interfaces.
// This package exists to avoid circular imports between board and vault.
package boardsync

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/forgia-labs/forgia/internal/board"
	"github.com/forgia-labs/forgia/internal/vault"
)

// Compile-time interface satisfaction checks.
var (
	_ board.VaultReader = (*VaultAdapter)(nil)
	_ board.VaultWriter = (*VaultAdapter)(nil)
)

// VaultAdapter implements board.VaultReader and board.VaultWriter using a vault.Vault.
type VaultAdapter struct {
	vault vault.Vault
}

// NewVaultAdapter creates an adapter that bridges vault and board.
func NewVaultAdapter(v vault.Vault) *VaultAdapter {
	return &VaultAdapter{vault: v}
}

// AllItems returns all FDs and SDDs as BoardItems.
func (a *VaultAdapter) AllItems(ctx context.Context) ([]board.BoardItem, error) {
	fds, err := a.vault.ListFDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list FDs: %w", err)
	}

	var items []board.BoardItem
	for _, fd := range fds {
		fields := map[string]string{}
		if fd.UpstreamIssue != "" {
			fields["upstream_issue"] = fd.UpstreamIssue
		}
		// Competitive FD linking — store competes_with as comma-separated field.
		if len(fd.CompetesWith) > 0 {
			fields["competes_with"] = strings.Join(fd.CompetesWith, ",")
		}

		items = append(items, board.BoardItem{
			ID:       fd.ID,
			Title:    fd.Title,
			Column:   string(fd.Status),
			Assignee: fd.Author,
			Priority: fd.Priority,
			Labels:   fd.Tags,
			Fields:   fields,
		})

		// Include SDDs under this FD.
		sdds, err := a.vault.ListSDDs(ctx, fd.ID)
		if err != nil {
			slog.Warn("skip SDDs during sync", "fd", fd.ID, "error", err)
			continue
		}
		for _, sdd := range sdds {
			items = append(items, board.BoardItem{
				ID:       sdd.ID,
				Title:    sdd.Title,
				Column:   string(sdd.Status),
				Assignee: sdd.AssignedTo,
				Labels:   sdd.Tags,
				Fields: map[string]string{
					"fd":         sdd.FD,
					"complexity": sdd.Complexity,
				},
			})
		}
	}

	// Include architecture bounded contexts.
	contexts, err := a.vault.ListContexts(ctx)
	if err != nil {
		slog.Warn("skip bounded contexts during sync", "error", err)
		return items, nil
	}
	for _, bc := range contexts {
		column := "planned"
		if bc.Status.FDCompleted {
			column = "complete"
		} else if bc.Status.FDCreated {
			column = "in-progress"
		}
		items = append(items, board.BoardItem{
			ID:     "CTX-" + bc.Name,
			Title:  bc.Name,
			Column: column,
			Labels: []string{"architecture", "bounded-context"},
			Fields: map[string]string{
				"type":        "bounded-context",
				"description": bc.Description,
			},
		})
	}

	return items, nil
}

// UpdateStatus updates a vault item's status by ID.
// Detects FD vs SDD by ID prefix.
func (a *VaultAdapter) UpdateStatus(ctx context.Context, id, status string) error {
	if len(id) > 4 && id[:4] == "SDD-" {
		return a.updateSDDStatus(ctx, id, status)
	}
	return a.updateFDStatus(ctx, id, status)
}

// validFDStatuses is the set of valid FD status values.
var validFDStatuses = map[vault.FDStatus]bool{
	vault.FDPlanned:    true,
	vault.FDApproved:   true,
	vault.FDInProgress: true,
	vault.FDComplete:   true,
	vault.FDClosed:     true,
	vault.FDRejected:   true,
	vault.FDAbandoned:  true,
}

// validSDDStatuses is the set of valid SDD status values.
var validSDDStatuses = map[vault.SDDStatus]bool{
	vault.SDDPlanned:    true,
	vault.SDDValidated:  true,
	vault.SDDInProgress: true,
	vault.SDDDone:       true,
	vault.SDDFailed:     true,
	vault.SDDCancelled:  true,
}

func (a *VaultAdapter) updateFDStatus(ctx context.Context, id, status string) error {
	fdStatus := vault.FDStatus(status)
	if !validFDStatuses[fdStatus] {
		return fmt.Errorf("invalid FD status %q from board (valid: planned, approved, in-progress, complete, closed, rejected, abandoned)", status)
	}
	fd, err := a.vault.GetFD(ctx, id)
	if err != nil {
		return fmt.Errorf("get FD %s: %w", id, err)
	}
	fd.Status = fdStatus
	if err := a.vault.UpdateFD(ctx, fd); err != nil {
		return fmt.Errorf("update FD %s status: %w", id, err)
	}
	return nil
}

func (a *VaultAdapter) updateSDDStatus(ctx context.Context, id, status string) error {
	sddStatus := vault.SDDStatus(status)
	if !validSDDStatuses[sddStatus] {
		return fmt.Errorf("invalid SDD status %q from board (valid: planned, validated, in-progress, done, failed, cancelled)", status)
	}
	// SDD lookup requires the parent FD ID. Walk all FDs to find it.
	fds, err := a.vault.ListFDs(ctx)
	if err != nil {
		return fmt.Errorf("list FDs for SDD lookup: %w", err)
	}
	for _, fd := range fds {
		sdd, err := a.vault.GetSDD(ctx, fd.ID, id)
		if err != nil {
			continue
		}
		sdd.Status = sddStatus
		return a.vault.UpdateSDD(ctx, sdd)
	}
	return fmt.Errorf("SDD %q not found in any FD", id)
}
