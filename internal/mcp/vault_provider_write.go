package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/forgia-labs/forgia/internal/vault"
)

// --- write tool handlers ---

func (p *VaultProvider) callFDCreate(ctx context.Context, params map[string]any) (any, error) {
	title, _ := params["title"].(string)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	author, _ := params["author"].(string)
	statusStr, _ := params["status"].(string)
	if statusStr == "" {
		statusStr = string(vault.FDPlanned)
	}
	if !isValidFDStatus(statusStr) {
		return nil, fmt.Errorf("invalid status %q: must be planned, approved, in-progress, complete, closed, rejected, or abandoned", statusStr)
	}

	priority, _ := params["priority"].(string)
	tagsRaw, _ := params["tags"].([]any)
	upstreamIssue, _ := params["upstream_issue"].(string)

	id := vault.NewFDID(title, author)

	fd := &vault.FD{
		ID:            id,
		Title:         title,
		Status:        vault.FDStatus(statusStr),
		Priority:      priority,
		Author:        author,
		Created:       time.Now().Format("2006-01-02"),
		Tags:          toStringSlice(tagsRaw),
		UpstreamIssue: upstreamIssue,
	}

	// Check guardrails before writing.
	targetPath := filepath.Join(".forgia", "fd", id+".md")
	if p.guard != nil {
		if violations := p.guard.CheckWritePaths(ctx, []string{targetPath}); len(violations) > 0 {
			return nil, fmt.Errorf("write denied: %s", violations[0].Error())
		}
	}

	// KG enrichment (optional, 3s timeout, graceful degradation).
	var enrichment map[string]any
	if p.registry != nil {
		enrichment = p.enrichFDWithKG(ctx, title)
	}

	if err := p.vault.CreateFD(ctx, fd); err != nil {
		return nil, fmt.Errorf("create FD: %w", err)
	}

	slog.InfoContext(ctx, "fd_create", "tool", "forgia_vault_fd_create", "path", targetPath)

	result := map[string]any{
		"id":   id,
		"path": targetPath,
	}
	if enrichment != nil {
		result["enrichment"] = enrichment
	}
	return result, nil
}

func (p *VaultProvider) callFDUpdate(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if err := validateID(id, "fd"); err != nil {
		return nil, err
	}

	// At least one field to update.
	updateFields := []string{"status", "reviewed", "reviewer", "priority", "tags", "assignee", "upstream_issue"}
	hasUpdate := false
	for _, field := range updateFields {
		if _, ok := params[field]; ok {
			hasUpdate = true
			break
		}
	}
	if !hasUpdate {
		return nil, fmt.Errorf("at least one field to update is required")
	}

	fd, err := p.vault.GetFD(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get FD %s: %w", id, err)
	}

	// Apply updates.
	if statusStr, ok := params["status"].(string); ok {
		if !isValidFDStatus(statusStr) {
			return nil, fmt.Errorf("invalid status %q", statusStr)
		}
		fd.Status = vault.FDStatus(statusStr)
	}
	if reviewed, ok := params["reviewed"].(bool); ok {
		fd.Reviewed = reviewed
	}
	if reviewer, ok := params["reviewer"].(string); ok {
		fd.Reviewer = reviewer
	}
	if priority, ok := params["priority"].(string); ok {
		fd.Priority = priority
	}
	if tagsRaw, ok := params["tags"].([]any); ok {
		fd.Tags = toStringSlice(tagsRaw)
	}
	if assignee, ok := params["assignee"].(string); ok {
		fd.Assignee = assignee
	}
	if upstreamIssue, ok := params["upstream_issue"].(string); ok {
		fd.UpstreamIssue = upstreamIssue
	}

	targetPath := filepath.Join(".forgia", "fd", id+".md")
	if p.guard != nil {
		if violations := p.guard.CheckWritePaths(ctx, []string{targetPath}); len(violations) > 0 {
			return nil, fmt.Errorf("write denied: %s", violations[0].Error())
		}
	}

	if err := p.vault.UpdateFD(ctx, fd); err != nil {
		return nil, fmt.Errorf("update FD: %w", err)
	}

	slog.InfoContext(ctx, "fd_update", "tool", "forgia_vault_fd_update", "path", targetPath)

	return map[string]any{"updated": true}, nil
}

func (p *VaultProvider) callSDDCreate(ctx context.Context, params map[string]any) (any, error) {
	// Lock to prevent race between nextSDDID read and CreateSDD write.
	p.sddMu.Lock()
	defer p.sddMu.Unlock()

	fdID, _ := params["fd"].(string)
	if fdID == "" {
		return nil, fmt.Errorf("fd is required")
	}
	if err := validateID(fdID, "fd"); err != nil {
		return nil, err
	}
	title, _ := params["title"].(string)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Verify parent FD exists.
	if _, err := p.vault.GetFD(ctx, fdID); err != nil {
		return nil, fmt.Errorf("parent FD %s not found: %w", fdID, err)
	}

	// Generate sequential ID.
	id, err := p.nextSDDID(ctx, fdID)
	if err != nil {
		return nil, fmt.Errorf("generate SDD ID: %w", err)
	}

	scope, _ := params["scope"].(string)

	sdd := &vault.SDD{
		ID:      id,
		FD:      fdID,
		Title:   title,
		Status:  vault.SDDPlanned,
		Created: time.Now().Format("2006-01-02"),
		Scope:   scope,
	}

	if tagsRaw, ok := params["tags"].([]any); ok {
		sdd.Tags = toStringSlice(tagsRaw)
	}
	if constraintsMap, ok := params["constraints"].(map[string]any); ok {
		if lang, ok := constraintsMap["language"].(string); ok {
			sdd.Constraints.Language = lang
		}
		if fw, ok := constraintsMap["framework"].(string); ok {
			sdd.Constraints.Framework = fw
		}
	}
	if criteriaRaw, ok := params["acceptance_criteria"].([]any); ok {
		for _, c := range criteriaRaw {
			if cStr, ok := c.(string); ok {
				sdd.Criteria = append(sdd.Criteria, vault.SDDCriterion{Criterion: cStr})
			}
		}
	}

	targetPath := filepath.Join(".forgia", "sdd", fdID, id+".md")
	if p.guard != nil {
		if violations := p.guard.CheckWritePaths(ctx, []string{targetPath}); len(violations) > 0 {
			return nil, fmt.Errorf("write denied: %s", violations[0].Error())
		}
	}

	// KG enrichment (optional, 3s timeout, graceful degradation).
	var enrichment map[string]any
	if p.registry != nil {
		enrichment = p.enrichSDDWithKG(ctx, fdID, title)
	}

	if err := p.vault.CreateSDD(ctx, sdd); err != nil {
		return nil, fmt.Errorf("create SDD: %w", err)
	}

	slog.InfoContext(ctx, "sdd_create", "tool", "forgia_vault_sdd_create", "path", targetPath)

	result := map[string]any{
		"id":   id,
		"path": targetPath,
	}
	if enrichment != nil {
		result["enrichment"] = enrichment
	}
	return result, nil
}

func (p *VaultProvider) callSDDUpdate(ctx context.Context, params map[string]any) (any, error) {
	fdID, _ := params["fd"].(string)
	if fdID == "" {
		return nil, fmt.Errorf("fd is required")
	}
	if err := validateID(fdID, "fd"); err != nil {
		return nil, err
	}
	id, _ := params["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if err := validateID(id, "sdd"); err != nil {
		return nil, err
	}

	// At least one field to update.
	updateFields := []string{"status", "agent", "assigned_to", "started", "completed", "tags"}
	hasUpdate := false
	for _, field := range updateFields {
		if _, ok := params[field]; ok {
			hasUpdate = true
			break
		}
	}
	if !hasUpdate {
		return nil, fmt.Errorf("at least one field to update is required")
	}

	sdd, err := p.vault.GetSDD(ctx, fdID, id)
	if err != nil {
		return nil, fmt.Errorf("get SDD %s/%s: %w", fdID, id, err)
	}

	// Apply updates.
	if statusStr, ok := params["status"].(string); ok {
		if !isValidSDDStatus(statusStr) {
			return nil, fmt.Errorf("invalid status %q", statusStr)
		}
		sdd.Status = vault.SDDStatus(statusStr)
	}
	if agent, ok := params["agent"].(string); ok {
		sdd.Agent = agent
	}
	if assignedTo, ok := params["assigned_to"].(string); ok {
		sdd.AssignedTo = assignedTo
	}
	if started, ok := params["started"].(string); ok {
		sdd.Started = started
	}
	if completed, ok := params["completed"].(string); ok {
		sdd.Completed = completed
	}
	if tagsRaw, ok := params["tags"].([]any); ok {
		sdd.Tags = toStringSlice(tagsRaw)
	}

	targetPath := filepath.Join(".forgia", "sdd", fdID, id+".md")
	if p.guard != nil {
		if violations := p.guard.CheckWritePaths(ctx, []string{targetPath}); len(violations) > 0 {
			return nil, fmt.Errorf("write denied: %s", violations[0].Error())
		}
	}

	if err := p.vault.UpdateSDD(ctx, sdd); err != nil {
		return nil, fmt.Errorf("update SDD: %w", err)
	}

	slog.InfoContext(ctx, "sdd_update", "tool", "forgia_vault_sdd_update", "path", targetPath)

	return map[string]any{"updated": true}, nil
}

// --- validation helpers ---

func isValidFDStatus(s string) bool {
	switch vault.FDStatus(s) {
	case vault.FDPlanned, vault.FDApproved, vault.FDInProgress,
		vault.FDComplete, vault.FDClosed, vault.FDRejected, vault.FDAbandoned:
		return true
	}
	return false
}

func isValidSDDStatus(s string) bool {
	switch vault.SDDStatus(s) {
	case vault.SDDPlanned, vault.SDDValidated, vault.SDDInProgress,
		vault.SDDDone, vault.SDDFailed, vault.SDDCancelled:
		return true
	}
	return false
}

// nextSDDID generates the next sequential SDD-NNN ID for an FD.
func (p *VaultProvider) nextSDDID(ctx context.Context, fdID string) (string, error) {
	sdds, err := p.vault.ListSDDs(ctx, fdID)
	if err != nil {
		return "", err
	}

	maxNum := 0
	for _, sdd := range sdds {
		if !strings.HasPrefix(sdd.ID, "SDD-") {
			continue
		}
		numStr := strings.TrimPrefix(sdd.ID, "SDD-")
		// Handle IDs like "SDD-010-vault-write-tools" by taking only the numeric part.
		parts := strings.SplitN(numStr, "-", 2)
		if n, err := strconv.Atoi(parts[0]); err == nil && n > maxNum {
			maxNum = n
		}
	}

	return fmt.Sprintf("SDD-%03d", maxNum+1), nil
}

// toStringSlice converts []any to []string for MCP parameter arrays.
func toStringSlice(raw []any) []string {
	if raw == nil {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// --- KG enrichment helpers ---

// enrichFDWithKG calls knowledge graph providers to enrich FD creation.
// Attempts architecture detection and code search. Returns nil if KG unavailable.
func (p *VaultProvider) enrichFDWithKG(ctx context.Context, title string) map[string]any {
	kgCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	enrichment := make(map[string]any)

	// Architecture auto-detection.
	if result, err := p.registry.Call(kgCtx, "forgia_code_arch_init", nil); err == nil {
		enrichment["architecture"] = result
	}

	// Code search for interface identification.
	if result, err := p.registry.Call(kgCtx, "forgia_code_search", map[string]any{"query": title}); err == nil {
		enrichment["interfaces"] = result
	}

	if len(enrichment) == 0 {
		return nil
	}
	return enrichment
}

// enrichSDDWithKG calls knowledge graph providers to enrich SDD creation.
// Attempts blast radius assessment and context path search. Returns nil if KG unavailable.
func (p *VaultProvider) enrichSDDWithKG(ctx context.Context, _ string, title string) map[string]any {
	kgCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	enrichment := make(map[string]any)

	// Blast radius assessment.
	if result, err := p.registry.Call(kgCtx, "forgia_code_blast_radius", map[string]any{"scope": title}); err == nil {
		enrichment["blast_radius"] = result
	}

	// Context path search.
	if result, err := p.registry.Call(kgCtx, "forgia_code_search", map[string]any{"query": title}); err == nil {
		enrichment["context_paths"] = result
	}

	if len(enrichment) == 0 {
		return nil
	}
	return enrichment
}
