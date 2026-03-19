// Package boardsync adapts vault types to board interfaces.
// This package exists to avoid circular imports between board and vault.
package boardsync

import (
	"context"
	"fmt"

	"github.com/Deepzima/forgia/internal/board"
	"github.com/Deepzima/forgia/internal/vault"
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
		items = append(items, board.BoardItem{
			ID:       fd.ID,
			Title:    fd.Title,
			Column:   string(fd.Status),
			Assignee: fd.Author,
			Priority: fd.Priority,
			Labels:   fd.Tags,
		})

		// Include SDDs under this FD.
		sdds, err := a.vault.ListSDDs(ctx, fd.ID)
		if err != nil {
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

func (a *VaultAdapter) updateFDStatus(ctx context.Context, id, status string) error {
	fd, err := a.vault.GetFD(ctx, id)
	if err != nil {
		return err
	}
	fd.Status = vault.FDStatus(status)
	return a.vault.UpdateFD(ctx, fd)
}

func (a *VaultAdapter) updateSDDStatus(ctx context.Context, id, status string) error {
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
		sdd.Status = vault.SDDStatus(status)
		return a.vault.UpdateSDD(ctx, sdd)
	}
	return fmt.Errorf("SDD %q not found in any FD", id)
}
