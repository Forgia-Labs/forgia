// Package vault provides read/write access to the .forgia/ directory structure
// (FDs, SDDs, architecture, contexts, constitution, guardrails, Work Logs).
package vault

import (
	"context"
	"crypto/sha256"
	"fmt"
	"iter"
	"strconv"
	"sync/atomic"
	"time"
)

// idCounter ensures unique IDs even when called in the same nanosecond.
var idCounter atomic.Uint64

// Vault manages the .forgia/ directory structure.
//
// This interface is intentionally large (17 methods) because almost all callers
// (CLI commands, MCP server) need both read and write access. Splitting into
// VaultReader/VaultWriter is deferred until implementation reveals actual usage
// patterns — then we refactor based on real data, not speculation.
type Vault interface {
	// Init scaffolds .forgia/ in the current project.
	Init(ctx context.Context, opts InitOptions) error

	// FD operations.
	FDs() iter.Seq[*FD]
	ListFDs(ctx context.Context) ([]*FD, error)
	GetFD(ctx context.Context, id string) (*FD, error)
	CreateFD(ctx context.Context, fd *FD) error
	UpdateFD(ctx context.Context, fd *FD) error

	// SDD operations.
	SDDs(fdID string) iter.Seq[*SDD]
	ListSDDs(ctx context.Context, fdID string) ([]*SDD, error)
	GetSDD(ctx context.Context, fdID, sddID string) (*SDD, error)
	CreateSDD(ctx context.Context, sdd *SDD) error
	UpdateSDD(ctx context.Context, sdd *SDD) error

	// Architecture (C4/DDD).
	GetArchitecture(ctx context.Context) (*Architecture, error)
	ListContexts(ctx context.Context) ([]*BoundedContext, error)
	GetContext(ctx context.Context, name string) (*BoundedContext, error)

	// Read-only access — returns raw content, caller parses.
	Constitution(ctx context.Context) (string, error)
	GuardrailsRaw(ctx context.Context) ([]byte, error)

	// Render YAML → MD for human review.
	Render(ctx context.Context, yamlPath string) error
}

// Open opens an existing .forgia/ vault at the given directory.
// Returns an error if the vault doesn't exist — use Init to create one.
func Open(dir string) (Vault, error) {
	// TODO: implement — verify .forgia/ exists, return concrete implementation
	return nil, fmt.Errorf("vault.Open: not yet implemented")
}

// InitOptions configures vault scaffolding.
type InitOptions struct {
	ProjectName string
	Author      string
}

// NewFDID generates a collision-proof hash-based ID locally.
// This is the FALLBACK for offline use. When a ProjectBoard is available,
// use board.NextID() instead — the board is the central ID authority (#35).
func NewFDID(title, author string) string {
	n := idCounter.Add(1)
	h := sha256.Sum256([]byte(title + time.Now().String() + author + strconv.FormatUint(n, 10)))
	return fmt.Sprintf("FD-%x", h[:2])
}
