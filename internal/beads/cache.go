package beads

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// CardMapping stores the relationship between a vault item and a board card.
type CardMapping struct {
	VaultID string `json:"vault_id"` // FD-xxxx or SDD-xxxx
	CardID  string `json:"card_id"`  // board-side ID
	Column  string `json:"column"`   // last known status
}

// CacheCardMapping stores a vault→board card mapping in Beads.
// Falls back silently if Beads is unavailable.
func (bc *Client) CacheCardMapping(ctx context.Context, mapping CardMapping) error {
	if !bc.available {
		return nil // silent fallback — cache is optional
	}

	data, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("marshal card mapping: %w", err)
	}

	_, err = bc.Call(ctx, "set", "forgia:card:"+mapping.VaultID, string(data))
	if err != nil {
		slog.Warn("failed to cache card mapping in beads", "vault_id", mapping.VaultID, "error", err)
		return nil // non-fatal
	}
	return nil
}

// GetCardMapping retrieves a cached card mapping from Beads.
// Returns nil, nil if not found or Beads unavailable.
func (bc *Client) GetCardMapping(ctx context.Context, vaultID string) (*CardMapping, error) {
	if !bc.available {
		return nil, nil
	}

	output, err := bc.Call(ctx, "get", "forgia:card:"+vaultID)
	if err != nil {
		return nil, nil // not found or unavailable — not an error
	}

	output = trimOutput(output)
	if output == "" {
		return nil, nil
	}

	var mapping CardMapping
	if err := json.Unmarshal([]byte(output), &mapping); err != nil {
		return nil, nil // corrupted cache entry — treat as miss
	}
	return &mapping, nil
}

// ListCardMappings returns all cached card mappings.
// Returns nil if Beads unavailable.
func (bc *Client) ListCardMappings(ctx context.Context) ([]CardMapping, error) {
	if !bc.available {
		return nil, nil
	}

	output, err := bc.Call(ctx, "list", "forgia:card:")
	if err != nil {
		return nil, nil
	}

	output = trimOutput(output)
	if output == "" {
		return nil, nil
	}

	var mappings []CardMapping
	for _, line := range splitLines(output) {
		var m CardMapping
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		mappings = append(mappings, m)
	}
	return mappings, nil
}

// trimOutput removes leading/trailing whitespace.
func trimOutput(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n' || s[0] == '\r' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// splitLines splits output by newlines, skipping empty lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := trimOutput(s[start:i])
			if line != "" {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		line := trimOutput(s[start:])
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
