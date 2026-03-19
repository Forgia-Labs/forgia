// Package board — GitHub Projects V2 implementation.
package board

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

// GitHubBoard implements ProjectBoard using GitHub Projects V2 GraphQL API via gh CLI.
type GitHubBoard struct {
	projectID     string            // GraphQL node ID (e.g., "PVT_kwHOADx1884BSGiC")
	owner         string
	number        int
	statusFieldID string
	statusOptions map[string]string // status name → option ID

	vaultReader VaultReader // optional — nil means sync disabled
	vaultWriter VaultWriter // optional — nil means sync disabled

	logger *slog.Logger
}

// NewGitHubBoard creates a board connected to a GitHub Project.
func NewGitHubBoard(ctx context.Context, owner string, number int) (*GitHubBoard, error) {
	b := &GitHubBoard{
		owner:  owner,
		number: number,
		logger: slog.With("component", "github-board"),
	}

	if err := b.resolveProject(ctx); err != nil {
		return nil, fmt.Errorf("resolve project: %w", err)
	}

	return b, nil
}

// SetVault injects vault adapters for sync operations.
// Both reader and writer must be set for sync to work.
func (b *GitHubBoard) SetVault(reader VaultReader, writer VaultWriter) {
	b.vaultReader = reader
	b.vaultWriter = writer
}

// resolveProject fetches the project ID and status field options.
func (b *GitHubBoard) resolveProject(ctx context.Context) error {
	query := fmt.Sprintf(`{
		user(login: %q) {
			projectV2(number: %d) {
				id
				fields(first: 20) {
					nodes {
						... on ProjectV2SingleSelectField {
							id
							name
							options { id name }
						}
					}
				}
			}
		}
	}`, b.owner, b.number)

	result, err := b.graphql(ctx, query)
	if err != nil {
		return err
	}

	projectID, ok := jsonPath(result, "data", "user", "projectV2", "id")
	if !ok {
		return fmt.Errorf("project %s/%d not found", b.owner, b.number)
	}
	b.projectID = projectID.(string)

	fields, ok := jsonPath(result, "data", "user", "projectV2", "fields", "nodes")
	if !ok {
		return fmt.Errorf("could not read project fields")
	}

	b.statusOptions = make(map[string]string)
	for _, field := range fields.([]any) {
		f := field.(map[string]any)
		name, _ := f["name"].(string)
		if name == "Status" {
			b.statusFieldID, _ = f["id"].(string)
			if options, ok := f["options"].([]any); ok {
				for _, opt := range options {
					o := opt.(map[string]any)
					optName, _ := o["name"].(string)
					optID, _ := o["id"].(string)
					b.statusOptions[optName] = optID
				}
			}
		}
	}

	b.logger.InfoContext(ctx, "connected to GitHub Project",
		"owner", b.owner,
		"number", b.number,
		"project_id", b.projectID,
		"status_options", b.statusOptions,
	)

	return nil
}

// NextID generates a collision-proof ID.
// For GitHub board, this is local generation — the board confirms on CreateCard.
func (b *GitHubBoard) NextID(ctx context.Context, prefix string) (string, error) {
	// Delegate to vault.NewFDID — board doesn't generate IDs centrally yet.
	// Future: query board for existing IDs to avoid collision.
	return "", fmt.Errorf("NextID: use vault.NewFDID for now — central ID generation TBD")
}

// CreateCard adds an item to the project board.
func (b *GitHubBoard) CreateCard(ctx context.Context, item BoardItem) (string, error) {
	mutation := fmt.Sprintf(`mutation {
		addProjectV2DraftIssue(input: {
			projectId: %q
			title: %q
			body: %q
		}) {
			projectItem { id }
		}
	}`, b.projectID, item.Title, fmt.Sprintf("ID: %s\nPriority: %s\nAssignee: %s", item.ID, item.Priority, item.Assignee))

	result, err := b.graphql(ctx, mutation)
	if err != nil {
		return "", fmt.Errorf("create card %q: %w", item.Title, err)
	}

	cardID, ok := jsonPath(result, "data", "addProjectV2DraftIssue", "projectItem", "id")
	if !ok {
		return "", fmt.Errorf("create card %q: could not parse response", item.Title)
	}

	itemID := cardID.(string)

	if item.Column != "" {
		if err := b.MoveCard(ctx, itemID, item.Column); err != nil {
			b.logger.WarnContext(ctx, "created card but failed to set status",
				"card", itemID, "status", item.Column, "error", err)
		}
	}

	b.logger.InfoContext(ctx, "card created", "id", itemID, "title", item.Title)
	return itemID, nil
}

// UpdateCard updates fields on a project item.
func (b *GitHubBoard) UpdateCard(ctx context.Context, id string, fields map[string]any) error {
	if status, ok := fields["status"].(string); ok {
		return b.MoveCard(ctx, id, status)
	}
	return nil
}

// MoveCard changes the status column of a project item.
func (b *GitHubBoard) MoveCard(ctx context.Context, id string, column string) error {
	optionID, ok := b.statusOptions[column]
	if !ok {
		return fmt.Errorf("unknown status column %q, available: %v", column, b.statusOptionNames())
	}

	mutation := fmt.Sprintf(`mutation {
		updateProjectV2ItemFieldValue(input: {
			projectId: %q
			itemId: %q
			fieldId: %q
			value: { singleSelectOptionId: %q }
		}) {
			projectV2Item { id }
		}
	}`, b.projectID, id, b.statusFieldID, optionID)

	_, err := b.graphql(ctx, mutation)
	if err != nil {
		return fmt.Errorf("move card %q to %q: %w", id, column, err)
	}

	b.logger.InfoContext(ctx, "card moved", "id", id, "column", column)
	return nil
}

// GetCards returns items from the project, optionally filtered.
func (b *GitHubBoard) GetCards(ctx context.Context, filter CardFilter) ([]BoardItem, error) {
	query := fmt.Sprintf(`{
		node(id: %q) {
			... on ProjectV2 {
				items(first: 100) {
					nodes {
						id
						content {
							... on DraftIssue { title body }
							... on Issue { title number }
						}
						fieldValues(first: 10) {
							nodes {
								... on ProjectV2ItemFieldSingleSelectValue {
									name
									field { ... on ProjectV2SingleSelectField { name } }
								}
							}
						}
					}
				}
			}
		}
	}`, b.projectID)

	result, err := b.graphql(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get cards: %w", err)
	}

	items, ok := jsonPath(result, "data", "node", "items", "nodes")
	if !ok {
		return nil, nil
	}

	var cards []BoardItem
	for _, item := range items.([]any) {
		i := item.(map[string]any)
		card := BoardItem{
			ID: i["id"].(string),
		}

		if content, ok := i["content"].(map[string]any); ok {
			card.Title, _ = content["title"].(string)
		}

		if fieldValues, ok := i["fieldValues"].(map[string]any); ok {
			if nodes, ok := fieldValues["nodes"].([]any); ok {
				for _, node := range nodes {
					n := node.(map[string]any)
					if name, ok := n["name"].(string); ok {
						if field, ok := n["field"].(map[string]any); ok {
							if fieldName, _ := field["name"].(string); fieldName == "Status" {
								card.Column = name
							}
						}
					}
				}
			}
		}

		if filter.Column != "" && card.Column != filter.Column {
			continue
		}

		cards = append(cards, card)
	}

	return cards, nil
}

// SyncFromVault pushes local .forgia/ state to the board.
// Requires SetVault() to have been called with a non-nil VaultReader.
func (b *GitHubBoard) SyncFromVault(ctx context.Context) error {
	if b.vaultReader == nil {
		return fmt.Errorf("SyncFromVault: vault reader not configured (call SetVault first)")
	}

	vaultItems, err := b.vaultReader.AllItems(ctx)
	if err != nil {
		return fmt.Errorf("read vault items: %w", err)
	}

	boardCards, err := b.GetCards(ctx, CardFilter{})
	if err != nil {
		return fmt.Errorf("read board cards: %w", err)
	}

	// Index existing cards by title prefix (e.g., "FD-a3f2 Feature Name" → "FD-a3f2").
	existing := make(map[string]BoardItem, len(boardCards))
	for _, card := range boardCards {
		id := extractIDFromTitle(card.Title)
		if id != "" {
			existing[id] = card
		}
	}

	var created, moved int
	for _, item := range vaultItems {
		card, found := existing[item.ID]
		if !found {
			// New item — create card.
			item.Title = item.ID + " " + item.Title
			if _, err := b.CreateCard(ctx, item); err != nil {
				b.logger.WarnContext(ctx, "sync: failed to create card",
					"id", item.ID, "error", err)
				continue
			}
			created++
			continue
		}

		// Existing card — check if status changed.
		if item.Column != "" && item.Column != card.Column {
			if err := b.MoveCard(ctx, card.ID, item.Column); err != nil {
				b.logger.WarnContext(ctx, "sync: failed to move card",
					"id", item.ID, "from", card.Column, "to", item.Column, "error", err)
				continue
			}
			moved++
		}
	}

	b.logger.InfoContext(ctx, "SyncFromVault complete",
		"vault_items", len(vaultItems), "created", created, "moved", moved)
	return nil
}

// SyncToVault pulls board state to local .forgia/ files.
// Requires SetVault() to have been called with a non-nil VaultWriter.
func (b *GitHubBoard) SyncToVault(ctx context.Context) error {
	if b.vaultWriter == nil {
		return fmt.Errorf("SyncToVault: vault writer not configured (call SetVault first)")
	}

	boardCards, err := b.GetCards(ctx, CardFilter{})
	if err != nil {
		return fmt.Errorf("read board cards: %w", err)
	}

	var updated int
	for _, card := range boardCards {
		id := extractIDFromTitle(card.Title)
		if id == "" || card.Column == "" {
			continue
		}

		if err := b.vaultWriter.UpdateStatus(ctx, id, card.Column); err != nil {
			b.logger.WarnContext(ctx, "sync: failed to update vault item",
				"id", id, "status", card.Column, "error", err)
			continue
		}
		updated++
	}

	b.logger.InfoContext(ctx, "SyncToVault complete",
		"board_cards", len(boardCards), "updated", updated)
	return nil
}

// graphql executes a GraphQL query via gh CLI with context cancellation.
func (b *GitHubBoard) graphql(ctx context.Context, query string) (map[string]any, error) {
	cmd := exec.CommandContext(ctx, "gh", "api", "graphql", "-f", "query="+query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh api graphql: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("parse graphql response: %w", err)
	}

	if errors, ok := result["errors"]; ok {
		return nil, fmt.Errorf("graphql errors: %v", errors)
	}

	return result, nil
}

// statusOptionNames returns available status names.
func (b *GitHubBoard) statusOptionNames() []string {
	var names []string
	for name := range b.statusOptions {
		names = append(names, name)
	}
	return names
}

// extractIDFromTitle parses a Forgia ID from a card title.
// Expects format "FD-xxxx Title" or "SDD-xxxx Title".
func extractIDFromTitle(title string) string {
	parts := strings.SplitN(title, " ", 2)
	if len(parts) == 0 {
		return ""
	}
	prefix := parts[0]
	if strings.HasPrefix(prefix, "FD-") || strings.HasPrefix(prefix, "SDD-") {
		return prefix
	}
	return ""
}

// jsonPath navigates a nested map by keys.
func jsonPath(data any, keys ...string) (any, bool) {
	current := data
	for _, key := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = m[key]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// LocalBoard is the fallback when no external board is available.
// Everything stays in .forgia/ files — no sync, no API calls.
type LocalBoard struct {
	logger *slog.Logger
}

// NewLocalBoard creates a local-only board.
func NewLocalBoard() *LocalBoard {
	return &LocalBoard{
		logger: slog.With("component", "local-board"),
	}
}

func (b *LocalBoard) NextID(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("NextID: use vault.NewFDID for local board")
}

func (b *LocalBoard) CreateCard(ctx context.Context, item BoardItem) (string, error) {
	b.logger.InfoContext(ctx, "card created (no-op)", "title", item.Title)
	return item.ID, nil
}

func (b *LocalBoard) UpdateCard(_ context.Context, _ string, _ map[string]any) error {
	return nil
}

func (b *LocalBoard) MoveCard(_ context.Context, _ string, _ string) error {
	return nil
}

func (b *LocalBoard) GetCards(_ context.Context, _ CardFilter) ([]BoardItem, error) {
	return nil, nil
}

func (b *LocalBoard) SyncFromVault(_ context.Context) error {
	return nil
}

func (b *LocalBoard) SyncToVault(_ context.Context) error {
	return nil
}

// Resolve picks the right board from config.
func Resolve(ctx context.Context, provider, owner string, number int) (ProjectBoard, error) {
	switch strings.ToLower(provider) {
	case "github":
		return NewGitHubBoard(ctx, owner, number)
	case "local", "":
		return NewLocalBoard(), nil
	default:
		return nil, fmt.Errorf("unknown board provider: %s", provider)
	}
}
