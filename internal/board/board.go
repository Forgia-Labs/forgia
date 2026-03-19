// Package board provides a generic project board interface
// (GitHub Projects, GitLab Boards, local fallback).
// The board is the central authority for FD/SDD IDs and status.
package board

import "context"

// ProjectBoard abstracts project management backends.
type ProjectBoard interface {
	// NextID generates a collision-proof ID for a new FD.
	// This is the PRIMARY source for IDs when online.
	// Offline fallback: vault.NewFDID() generates locally.
	NextID(ctx context.Context, prefix string) (string, error)

	// Card CRUD.
	CreateCard(ctx context.Context, item BoardItem) (string, error)
	UpdateCard(ctx context.Context, id string, fields map[string]any) error
	MoveCard(ctx context.Context, id string, column string) error
	GetCards(ctx context.Context, filter CardFilter) ([]BoardItem, error)

	// Sync between vault and board.
	SyncFromVault(ctx context.Context) error // push local state → board
	SyncToVault(ctx context.Context) error   // pull board state → local
}

// BoardItem represents a card on the project board.
type BoardItem struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Column   string            `json:"column"`
	Assignee string            `json:"assignee,omitempty"`
	Priority string            `json:"priority,omitempty"`
	Labels   []string          `json:"labels,omitempty"`
	Fields   map[string]string `json:"fields,omitempty"` // custom fields (duration, tokens)
}

// CardFilter for querying cards.
type CardFilter struct {
	Column   string
	Assignee string
	Label    string
}

// VaultReader provides vault items as board items for sync.
// Implemented by the vault adapter — board never imports vault types.
type VaultReader interface {
	// AllItems returns all FDs and SDDs mapped to BoardItems.
	AllItems(ctx context.Context) ([]BoardItem, error)
}

// VaultWriter updates vault files from board state.
// Implemented by the vault adapter — board never imports vault types.
type VaultWriter interface {
	// UpdateStatus updates a vault item's status by ID.
	UpdateStatus(ctx context.Context, id, status string) error
}
