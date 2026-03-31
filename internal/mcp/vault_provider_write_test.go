package mcp

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/forgia-labs/forgia/internal/guardrails"
	"github.com/forgia-labs/forgia/internal/vault"
)

// --- fd_create tests ---

func TestVaultProvider_FDCreate_Valid(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	result, err := p.Call(context.Background(), "fd_create", map[string]any{
		"title":  "New Feature",
		"author": "alice",
		"status": "planned",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	id, _ := r["id"].(string)
	if !strings.HasPrefix(id, "FD-") {
		t.Errorf("id %q doesn't start with FD-", id)
	}
	path, _ := r["path"].(string)
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("path %q doesn't end with .md", path)
	}
	if len(mv.createdFDs) != 1 {
		t.Fatalf("expected 1 created FD, got %d", len(mv.createdFDs))
	}
	if mv.createdFDs[0].Title != "New Feature" {
		t.Errorf("title = %q, want 'New Feature'", mv.createdFDs[0].Title)
	}
	if mv.createdFDs[0].Author != "alice" {
		t.Errorf("author = %q, want 'alice'", mv.createdFDs[0].Author)
	}
	if mv.createdFDs[0].Status != vault.FDPlanned {
		t.Errorf("status = %s, want planned", mv.createdFDs[0].Status)
	}
}

func TestVaultProvider_FDCreate_EmptyTitle(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "fd_create", map[string]any{})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention title: %v", err)
	}
}

func TestVaultProvider_FDCreate_InvalidStatus(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "fd_create", map[string]any{
		"title": "Bad Status", "status": "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("error should mention invalid status: %v", err)
	}
}

func TestVaultProvider_FDCreate_AutoID(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	r1, err1 := p.Call(context.Background(), "fd_create", map[string]any{"title": "Alpha"})
	r2, err2 := p.Call(context.Background(), "fd_create", map[string]any{"title": "Beta"})
	if err1 != nil || err2 != nil {
		t.Fatalf("errors: %v, %v", err1, err2)
	}

	id1 := r1.(map[string]any)["id"].(string)
	id2 := r2.(map[string]any)["id"].(string)
	if id1 == id2 {
		t.Errorf("IDs should be unique: %s == %s", id1, id2)
	}
	if len(mv.createdFDs) != 2 {
		t.Errorf("expected 2 created FDs, got %d", len(mv.createdFDs))
	}
}

func TestVaultProvider_FDCreate_DefaultStatus(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	_, err := p.Call(context.Background(), "fd_create", map[string]any{
		"title": "No Status",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mv.createdFDs[0].Status != vault.FDPlanned {
		t.Errorf("default status = %s, want planned", mv.createdFDs[0].Status)
	}
}

// --- fd_update tests ---

func TestVaultProvider_FDUpdate_Status(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	result, err := p.Call(context.Background(), "fd_update", map[string]any{
		"id": "FD-001", "status": "complete",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["updated"] != true {
		t.Error("expected updated=true")
	}
	if len(mv.updatedFDs) != 1 {
		t.Fatalf("expected 1 updated FD, got %d", len(mv.updatedFDs))
	}
	if mv.updatedFDs[0].Status != vault.FDComplete {
		t.Errorf("status = %s, want complete", mv.updatedFDs[0].Status)
	}
}

func TestVaultProvider_FDUpdate_MultipleFields(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	_, err := p.Call(context.Background(), "fd_update", map[string]any{
		"id": "FD-001", "reviewed": true, "reviewer": "claude",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mv.updatedFDs[0].Reviewed {
		t.Error("expected reviewed=true")
	}
	if mv.updatedFDs[0].Reviewer != "claude" {
		t.Errorf("reviewer = %q, want 'claude'", mv.updatedFDs[0].Reviewer)
	}
}

func TestVaultProvider_FDUpdate_MissingID(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "fd_update", map[string]any{
		"status": "approved",
	})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestVaultProvider_FDUpdate_NoFields(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "fd_update", map[string]any{
		"id": "FD-001",
	})
	if err == nil {
		t.Fatal("expected error for no update fields")
	}
}

func TestVaultProvider_FDUpdate_InvalidStatus(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "fd_update", map[string]any{
		"id": "FD-001", "status": "bogus",
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestVaultProvider_FDUpdate_NotFound(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "fd_update", map[string]any{
		"id": "FD-999", "status": "approved",
	})
	if err == nil {
		t.Fatal("expected error for missing FD")
	}
}

// --- sdd_create tests ---

func TestVaultProvider_SDDCreate_Valid(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	result, err := p.Call(context.Background(), "sdd_create", map[string]any{
		"fd": "FD-001", "title": "New Spec",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	id := r["id"].(string)
	// FD-001 has SDD-001 and SDD-002, so next is SDD-003.
	if id != "SDD-003" {
		t.Errorf("id = %s, want SDD-003", id)
	}
	path := r["path"].(string)
	if !strings.Contains(path, "FD-001") {
		t.Errorf("path %q should contain FD-001", path)
	}
	if len(mv.createdSDDs) != 1 {
		t.Fatalf("expected 1 created SDD, got %d", len(mv.createdSDDs))
	}
	if mv.createdSDDs[0].FD != "FD-001" {
		t.Errorf("fd = %s, want FD-001", mv.createdSDDs[0].FD)
	}
	if mv.createdSDDs[0].Title != "New Spec" {
		t.Errorf("title = %q, want 'New Spec'", mv.createdSDDs[0].Title)
	}
}

func TestVaultProvider_SDDCreate_EmptyFD(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	// FD-002 has no SDDs, so next is SDD-001.
	result, err := p.Call(context.Background(), "sdd_create", map[string]any{
		"fd": "FD-002", "title": "First Spec",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	id := result.(map[string]any)["id"].(string)
	if id != "SDD-001" {
		t.Errorf("id = %s, want SDD-001 for empty FD", id)
	}
}

func TestVaultProvider_SDDCreate_MissingFD(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "sdd_create", map[string]any{
		"fd": "FD-999", "title": "Ghost Spec",
	})
	if err == nil {
		t.Fatal("expected error for missing parent FD")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found: %v", err)
	}
}

func TestVaultProvider_SDDCreate_EmptyTitle(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "sdd_create", map[string]any{
		"fd": "FD-001",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestVaultProvider_SDDCreate_WithConstraints(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	_, err := p.Call(context.Background(), "sdd_create", map[string]any{
		"fd":    "FD-001",
		"title": "Constrained Spec",
		"constraints": map[string]any{
			"language":  "Go",
			"framework": "Cobra",
		},
		"acceptance_criteria": []any{"It compiles", "Tests pass"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sdd := mv.createdSDDs[0]
	if sdd.Constraints.Language != "Go" {
		t.Errorf("language = %q, want Go", sdd.Constraints.Language)
	}
	if sdd.Constraints.Framework != "Cobra" {
		t.Errorf("framework = %q, want Cobra", sdd.Constraints.Framework)
	}
	if len(sdd.Criteria) != 2 {
		t.Errorf("criteria count = %d, want 2", len(sdd.Criteria))
	}
}

// --- sdd_update tests ---

func TestVaultProvider_SDDUpdate_AgentStatus(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	result, err := p.Call(context.Background(), "sdd_update", map[string]any{
		"fd": "FD-001", "id": "SDD-001",
		"status": "done", "agent": "claude-code",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if r["updated"] != true {
		t.Error("expected updated=true")
	}
	if len(mv.updatedSDDs) != 1 {
		t.Fatalf("expected 1 updated SDD, got %d", len(mv.updatedSDDs))
	}
	if mv.updatedSDDs[0].Status != vault.SDDDone {
		t.Errorf("status = %s, want done", mv.updatedSDDs[0].Status)
	}
	if mv.updatedSDDs[0].Agent != "claude-code" {
		t.Errorf("agent = %s, want claude-code", mv.updatedSDDs[0].Agent)
	}
}

func TestVaultProvider_SDDUpdate_MissingID(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "sdd_update", map[string]any{
		"fd": "FD-001", "status": "done",
	})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestVaultProvider_SDDUpdate_MissingFD(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "sdd_update", map[string]any{
		"id": "SDD-001", "status": "done",
	})
	if err == nil {
		t.Fatal("expected error for missing fd")
	}
}

func TestVaultProvider_SDDUpdate_NoFields(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "sdd_update", map[string]any{
		"fd": "FD-001", "id": "SDD-001",
	})
	if err == nil {
		t.Fatal("expected error for no update fields")
	}
}

func TestVaultProvider_SDDUpdate_InvalidStatus(t *testing.T) {
	t.Parallel()
	p := NewVaultProvider(newTestVault(t), nil, nil)

	_, err := p.Call(context.Background(), "sdd_update", map[string]any{
		"fd": "FD-001", "id": "SDD-001", "status": "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

// --- guardrails tests ---

func TestVaultProvider_FDCreate_GuardrailDenied(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	g, err := guardrails.Parse([]byte("[write]\npatterns = [\".forgia/fd/*\"]\n"))
	if err != nil {
		t.Fatal(err)
	}
	p := NewVaultProvider(mv, g, nil)

	_, err = p.Call(context.Background(), "fd_create", map[string]any{
		"title": "Denied Feature",
	})
	if err == nil {
		t.Fatal("expected error from guardrails")
	}
	if !strings.Contains(err.Error(), "write denied") {
		t.Errorf("expected guardrails error, got: %v", err)
	}
	if len(mv.createdFDs) != 0 {
		t.Error("FD should not have been created when guardrails deny")
	}
}

func TestVaultProvider_SDDCreate_GuardrailDenied(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	g, err := guardrails.Parse([]byte("[write]\npatterns = [\".forgia/sdd/**/*\"]\n"))
	if err != nil {
		t.Fatal(err)
	}
	p := NewVaultProvider(mv, g, nil)

	_, err = p.Call(context.Background(), "sdd_create", map[string]any{
		"fd": "FD-001", "title": "Denied Spec",
	})
	if err == nil {
		t.Fatal("expected error from guardrails")
	}
	if !strings.Contains(err.Error(), "write denied") {
		t.Errorf("expected guardrails error, got: %v", err)
	}
}

// --- KG enrichment tests ---

func TestVaultProvider_FDCreate_WithKG(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	reg := NewProviderRegistry()
	reg.Register(&mockCodeProvider{
		callFn: func(_ context.Context, tool string, _ map[string]any) (any, error) {
			switch tool {
			case "arch_init":
				return map[string]any{"language": "Go", "framework": "Cobra"}, nil
			case "search":
				return map[string]any{"components": []string{"internal/mcp", "internal/vault"}}, nil
			}
			return nil, fmt.Errorf("unknown tool")
		},
	})
	p := NewVaultProvider(mv, nil, reg)

	result, err := p.Call(context.Background(), "fd_create", map[string]any{
		"title": "KG Feature",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	enrichment, ok := r["enrichment"].(map[string]any)
	if !ok {
		t.Fatal("expected enrichment in result")
	}
	if _, ok := enrichment["architecture"]; !ok {
		t.Error("expected architecture in enrichment")
	}
	if _, ok := enrichment["interfaces"]; !ok {
		t.Error("expected interfaces in enrichment")
	}
}

func TestVaultProvider_SDDCreate_WithKG(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	reg := NewProviderRegistry()
	reg.Register(&mockCodeProvider{
		callFn: func(_ context.Context, tool string, _ map[string]any) (any, error) {
			switch tool {
			case "blast_radius":
				return map[string]any{"impact": "low", "files": 5}, nil
			case "search":
				return map[string]any{"paths": []string{"internal/mcp/vault_provider.go"}}, nil
			}
			return nil, fmt.Errorf("unknown tool")
		},
	})
	p := NewVaultProvider(mv, nil, reg)

	result, err := p.Call(context.Background(), "sdd_create", map[string]any{
		"fd": "FD-001", "title": "KG Spec",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	enrichment, ok := r["enrichment"].(map[string]any)
	if !ok {
		t.Fatal("expected enrichment in result")
	}
	if _, ok := enrichment["blast_radius"]; !ok {
		t.Error("expected blast_radius in enrichment")
	}
	if _, ok := enrichment["context_paths"]; !ok {
		t.Error("expected context_paths in enrichment")
	}
}

func TestVaultProvider_FDCreate_WithoutKG(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	result, err := p.Call(context.Background(), "fd_create", map[string]any{
		"title": "Simple Feature",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if _, ok := r["enrichment"]; ok {
		t.Error("expected no enrichment without KG")
	}
	if len(mv.createdFDs) != 1 {
		t.Error("FD should still be created without KG")
	}
}

func TestVaultProvider_SDDCreate_WithoutKG(t *testing.T) {
	t.Parallel()
	mv := newTestVault(t)
	p := NewVaultProvider(mv, nil, nil)

	result, err := p.Call(context.Background(), "sdd_create", map[string]any{
		"fd": "FD-001", "title": "Simple Spec",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := result.(map[string]any)
	if _, ok := r["enrichment"]; ok {
		t.Error("expected no enrichment without KG")
	}
}
