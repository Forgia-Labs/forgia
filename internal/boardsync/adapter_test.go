package boardsync

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/forgia-labs/forgia/internal/vault"

	"gopkg.in/yaml.v3"
)

func TestAllItems(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	v, err := vault.InitVault(ctx, dir, vault.InitOptions{ProjectName: "test", Author: "tester"})
	if err != nil {
		t.Fatal(err)
	}

	// Create FD + SDD.
	fd := &vault.FD{ID: "FD-t001", Title: "Test Feature", Status: vault.FDApproved, Priority: "high", Author: "tester"}
	if err := v.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}
	sdd := &vault.SDD{ID: "SDD-001a", FD: "FD-t001", Title: "Login API", Status: vault.SDDPlanned}
	if err := v.CreateSDD(ctx, sdd); err != nil {
		t.Fatal(err)
	}

	adapter := NewVaultAdapter(v)
	items, err := adapter.AllItems(ctx)
	if err != nil {
		t.Fatalf("AllItems: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items (1 FD + 1 SDD), got %d", len(items))
	}

	// First item is FD.
	if items[0].ID != "FD-t001" {
		t.Errorf("item[0].ID = %q, want FD-t001", items[0].ID)
	}
	if items[0].Column != "approved" {
		t.Errorf("item[0].Column = %q, want 'approved'", items[0].Column)
	}
	if items[0].Priority != "high" {
		t.Errorf("item[0].Priority = %q, want 'high'", items[0].Priority)
	}

	// Second item is SDD.
	if items[1].ID != "SDD-001a" {
		t.Errorf("item[1].ID = %q, want SDD-001a", items[1].ID)
	}
	if items[1].Column != "planned" {
		t.Errorf("item[1].Column = %q, want 'planned'", items[1].Column)
	}
	if items[1].Fields["fd"] != "FD-t001" {
		t.Errorf("item[1].Fields[fd] = %q, want 'FD-t001'", items[1].Fields["fd"])
	}
}

func TestUpdateStatus_FD(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	v, err := vault.InitVault(ctx, dir, vault.InitOptions{ProjectName: "test", Author: "tester"})
	if err != nil {
		t.Fatal(err)
	}

	fd := &vault.FD{ID: "FD-u001", Title: "Update Test", Status: vault.FDPlanned, Author: "tester"}
	if err := v.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}

	adapter := NewVaultAdapter(v)
	if err := adapter.UpdateStatus(ctx, "FD-u001", "approved"); err != nil {
		t.Fatalf("UpdateStatus FD: %v", err)
	}

	got, err := v.GetFD(ctx, "FD-u001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != vault.FDApproved {
		t.Errorf("FD status = %q, want 'approved'", got.Status)
	}
}

func TestUpdateStatus_SDD(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	v, err := vault.InitVault(ctx, dir, vault.InitOptions{ProjectName: "test", Author: "tester"})
	if err != nil {
		t.Fatal(err)
	}

	fd := &vault.FD{ID: "FD-u002", Title: "Parent", Status: vault.FDApproved, Author: "tester"}
	if err := v.CreateFD(ctx, fd); err != nil {
		t.Fatal(err)
	}
	sdd := &vault.SDD{ID: "SDD-u001", FD: "FD-u002", Title: "Task", Status: vault.SDDPlanned}
	if err := v.CreateSDD(ctx, sdd); err != nil {
		t.Fatal(err)
	}

	adapter := NewVaultAdapter(v)
	if err := adapter.UpdateStatus(ctx, "SDD-u001", "in-progress"); err != nil {
		t.Fatalf("UpdateStatus SDD: %v", err)
	}

	got, err := v.GetSDD(ctx, "FD-u002", "SDD-u001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != vault.SDDInProgress {
		t.Errorf("SDD status = %q, want 'in-progress'", got.Status)
	}
}

func TestAllItems_CompetitiveFDs(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	v, err := vault.InitVault(ctx, dir, vault.InitOptions{ProjectName: "test", Author: "tester"})
	if err != nil {
		t.Fatal(err)
	}

	// Create two competing FDs.
	fdA := &vault.FD{ID: "FD-comp1", Title: "Option A", Status: vault.FDPlanned, Author: "eng1", CompetesWith: []string{"FD-comp2"}}
	fdB := &vault.FD{ID: "FD-comp2", Title: "Option B", Status: vault.FDPlanned, Author: "eng2", CompetesWith: []string{"FD-comp1"}, UpstreamIssue: "#99"}
	if err := v.CreateFD(ctx, fdA); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateFD(ctx, fdB); err != nil {
		t.Fatal(err)
	}

	adapter := NewVaultAdapter(v)
	items, err := adapter.AllItems(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Find FD-comp1.
	var comp1, comp2 *struct{ Fields map[string]string }
	for i := range items {
		if items[i].ID == "FD-comp1" {
			comp1 = &struct{ Fields map[string]string }{items[i].Fields}
		}
		if items[i].ID == "FD-comp2" {
			comp2 = &struct{ Fields map[string]string }{items[i].Fields}
		}
	}
	if comp1 == nil || comp2 == nil {
		t.Fatal("missing competitive FDs in items")
	}
	if comp1.Fields["competes_with"] != "FD-comp2" {
		t.Errorf("FD-comp1 competes_with = %q, want 'FD-comp2'", comp1.Fields["competes_with"])
	}
	if comp2.Fields["competes_with"] != "FD-comp1" {
		t.Errorf("FD-comp2 competes_with = %q, want 'FD-comp1'", comp2.Fields["competes_with"])
	}
	if comp2.Fields["upstream_issue"] != "#99" {
		t.Errorf("FD-comp2 upstream_issue = %q, want '#99'", comp2.Fields["upstream_issue"])
	}
}

func TestAllItems_ArchitectureContexts(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	v, err := vault.InitVault(ctx, dir, vault.InitOptions{ProjectName: "test", Author: "tester"})
	if err != nil {
		t.Fatal(err)
	}

	// Create a bounded context YAML file.
	bc := vault.BoundedContext{
		Name:        "trading-core",
		Description: "Core trading engine",
		Status: vault.ContextStatus{
			FDCreated:   true,
			FDCompleted: false,
		},
	}
	ctxDir := filepath.Join(dir, ".forgia", "architecture", "contexts")
	data, err := yaml.Marshal(&bc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ctxDir, "trading-core.yaml"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewVaultAdapter(v)
	items, err := adapter.AllItems(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Should have the context as a board item.
	var found bool
	for _, item := range items {
		if item.ID == "CTX-trading-core" {
			found = true
			if item.Column != "in-progress" {
				t.Errorf("context column = %q, want 'in-progress'", item.Column)
			}
			if item.Fields["type"] != "bounded-context" {
				t.Errorf("context type = %q, want 'bounded-context'", item.Fields["type"])
			}
			hasArchLabel := false
			for _, l := range item.Labels {
				if l == "architecture" {
					hasArchLabel = true
				}
			}
			if !hasArchLabel {
				t.Error("context missing 'architecture' label")
			}
		}
	}
	if !found {
		t.Error("bounded context not found in AllItems")
	}
}

func TestUpdateStatus_NotFound(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	v, err := vault.InitVault(ctx, dir, vault.InitOptions{ProjectName: "test", Author: "tester"})
	if err != nil {
		t.Fatal(err)
	}

	adapter := NewVaultAdapter(v)
	if err := adapter.UpdateStatus(ctx, "FD-nonexistent", "approved"); err == nil {
		t.Fatal("expected error for non-existent FD")
	}
	if err := adapter.UpdateStatus(ctx, "SDD-nonexistent", "done"); err == nil {
		t.Fatal("expected error for non-existent SDD")
	}
}
