package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

// --- Vault tools integration tests (SDD-011) ---

// TestIntegration_VaultTools_CRUDCycle verifies the full FD+SDD lifecycle via MCP tools:
// create FD → get → list → create SDD → list SDDs → validate → update FD → verify update.
func TestIntegration_VaultTools_CRUDCycle(t *testing.T) {
	t.Parallel()

	enc, dec, cleanup := mcpServer(t)
	defer cleanup()

	id := 0
	next := func() int { id++; return id }

	// 1. Create FD.
	resp := mcpRequest(t, enc, dec, next(), "tools/call", map[string]any{
		"name": "forgia_vault_fd_create",
		"arguments": map[string]any{
			"title":  "Test CRUD Feature",
			"author": "test-agent",
		},
	})
	if resp["error"] != nil {
		t.Fatalf("fd_create error: %v", resp["error"])
	}
	createResult := extractToolResult(t, resp)
	fdID, _ := createResult["id"].(string)
	if fdID == "" {
		t.Fatal("fd_create did not return an id")
	}
	t.Logf("created FD: %s", fdID)

	// 2. Get FD — verify content.
	resp = mcpRequest(t, enc, dec, next(), "tools/call", map[string]any{
		"name":      "forgia_vault_fd_get",
		"arguments": map[string]any{"id": fdID},
	})
	if resp["error"] != nil {
		t.Fatalf("fd_get error: %v", resp["error"])
	}
	getResult := extractToolResult(t, resp)
	if getResult["title"] != "Test CRUD Feature" {
		t.Errorf("fd_get title = %v, want 'Test CRUD Feature'", getResult["title"])
	}
	if getResult["status"] != "planned" {
		t.Errorf("fd_get status = %v, want 'planned'", getResult["status"])
	}

	// 3. List FDs — verify the new one appears.
	resp = mcpRequest(t, enc, dec, next(), "tools/call", map[string]any{
		"name":      "forgia_vault_fd_list",
		"arguments": map[string]any{},
	})
	if resp["error"] != nil {
		t.Fatalf("fd_list error: %v", resp["error"])
	}
	listResult := extractToolResult(t, resp)
	fds, _ := listResult["fds"].([]any)
	found := false
	for _, item := range fds {
		fd, _ := item.(map[string]any)
		if fd["id"] == fdID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("fd_list does not contain created FD %s", fdID)
	}

	// 4. Create SDD under the FD.
	resp = mcpRequest(t, enc, dec, next(), "tools/call", map[string]any{
		"name": "forgia_vault_sdd_create",
		"arguments": map[string]any{
			"fd":    fdID,
			"title": "Test SDD Spec",
		},
	})
	if resp["error"] != nil {
		t.Fatalf("sdd_create error: %v", resp["error"])
	}
	sddResult := extractToolResult(t, resp)
	sddID, _ := sddResult["id"].(string)
	if sddID == "" {
		t.Fatal("sdd_create did not return an id")
	}
	t.Logf("created SDD: %s", sddID)

	// 5. List SDDs — verify the new one.
	resp = mcpRequest(t, enc, dec, next(), "tools/call", map[string]any{
		"name":      "forgia_vault_sdd_list",
		"arguments": map[string]any{"fd": fdID},
	})
	if resp["error"] != nil {
		t.Fatalf("sdd_list error: %v", resp["error"])
	}
	sddListResult := extractToolResult(t, resp)
	sdds, _ := sddListResult["sdds"].([]any)
	sddFound := false
	for _, item := range sdds {
		sdd, _ := item.(map[string]any)
		if sdd["id"] == sddID {
			sddFound = true
			break
		}
	}
	if !sddFound {
		t.Errorf("sdd_list does not contain created SDD %s", sddID)
	}

	// 6. Validate the SDD.
	resp = mcpRequest(t, enc, dec, next(), "tools/call", map[string]any{
		"name":      "forgia_vault_validate",
		"arguments": map[string]any{"fd": fdID},
	})
	if resp["error"] != nil {
		t.Fatalf("validate error: %v", resp["error"])
	}
	validateResult := extractToolResult(t, resp)
	files, _ := validateResult["files"].([]any)
	if len(files) == 0 {
		t.Error("validate returned no file results")
	}

	// 7. Update FD status to approved.
	resp = mcpRequest(t, enc, dec, next(), "tools/call", map[string]any{
		"name": "forgia_vault_fd_update",
		"arguments": map[string]any{
			"id":     fdID,
			"status": "approved",
		},
	})
	if resp["error"] != nil {
		t.Fatalf("fd_update error: %v", resp["error"])
	}
	updateResult := extractToolResult(t, resp)
	if updateResult["updated"] != true {
		t.Errorf("fd_update result = %v, want updated=true", updateResult)
	}

	// 8. Get FD again — verify status changed.
	resp = mcpRequest(t, enc, dec, next(), "tools/call", map[string]any{
		"name":      "forgia_vault_fd_get",
		"arguments": map[string]any{"id": fdID},
	})
	if resp["error"] != nil {
		t.Fatalf("fd_get (after update) error: %v", resp["error"])
	}
	getAfterUpdate := extractToolResult(t, resp)
	if getAfterUpdate["status"] != "approved" {
		t.Errorf("fd_get status after update = %v, want 'approved'", getAfterUpdate["status"])
	}
}

// TestIntegration_VaultTools_ToolsList verifies tools/list returns both vault tools (11)
// and composite skills (7) = 18 total tools.
func TestIntegration_VaultTools_ToolsList(t *testing.T) {
	t.Parallel()

	enc, dec, cleanup := mcpServer(t)
	defer cleanup()

	resp := mcpRequest(t, enc, dec, 1, "tools/list", nil)
	if resp["error"] != nil {
		t.Fatalf("tools/list error: %v", resp["error"])
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result: %v", resp)
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("missing tools array in result")
	}

	// Count vault tools and other tools (composite skills) separately.
	var vaultTools, otherTools int
	var toolNames []string
	for _, item := range tools {
		tool, _ := item.(map[string]any)
		name, _ := tool["name"].(string)
		toolNames = append(toolNames, name)
		if strings.HasPrefix(name, "forgia_vault_") {
			vaultTools++
		} else {
			otherTools++
		}
	}

	if vaultTools != 11 {
		t.Errorf("expected 11 vault tools, got %d; tools: %v", vaultTools, toolNames)
	}
	if otherTools != 7 {
		t.Errorf("expected 7 composite skills, got %d; tools: %v", otherTools, toolNames)
	}

	expectedTotal := 18
	if len(tools) != expectedTotal {
		t.Errorf("expected %d total tools, got %d; tools: %v", expectedTotal, len(tools), toolNames)
	}

	// Verify specific vault tool names are present.
	nameSet := make(map[string]bool)
	for _, n := range toolNames {
		nameSet[n] = true
	}
	for _, expected := range []string{
		"forgia_vault_fd_get", "forgia_vault_fd_list", "forgia_vault_fd_create", "forgia_vault_fd_update",
		"forgia_vault_sdd_get", "forgia_vault_sdd_list", "forgia_vault_sdd_create", "forgia_vault_sdd_update",
		"forgia_vault_status", "forgia_vault_validate", "forgia_vault_threat_model_get",
	} {
		if !nameSet[expected] {
			t.Errorf("missing vault tool: %s", expected)
		}
	}
}

// TestIntegration_VaultTools_WithoutKnowledgeGraph verifies MCP server starts
// and vault tools work even when no knowledge graph provider is configured.
func TestIntegration_VaultTools_WithoutKnowledgeGraph(t *testing.T) {
	t.Parallel()

	enc, dec, cleanup := mcpServer(t)
	defer cleanup()

	// The minimal vault has no codebase-memory-mcp configured.
	// Verify vault tools still work.

	// 1. Create FD — should work without KG.
	resp := mcpRequest(t, enc, dec, 1, "tools/call", map[string]any{
		"name": "forgia_vault_fd_create",
		"arguments": map[string]any{
			"title": "No KG Feature",
		},
	})
	if resp["error"] != nil {
		t.Fatalf("fd_create without KG error: %v", resp["error"])
	}
	result := extractToolResult(t, resp)
	fdID, _ := result["id"].(string)
	if fdID == "" {
		t.Fatal("fd_create without KG did not return an id")
	}

	// 2. Status — should work without KG, no knowledge field.
	resp = mcpRequest(t, enc, dec, 2, "tools/call", map[string]any{
		"name":      "forgia_vault_status",
		"arguments": map[string]any{},
	})
	if resp["error"] != nil {
		t.Fatalf("status without KG error: %v", resp["error"])
	}
	statusResult := extractToolResult(t, resp)
	if _, hasKG := statusResult["knowledge"]; hasKG {
		t.Error("knowledge field should not be present without KG provider")
	}

	// 3. List FDs — verify the created FD is there.
	resp = mcpRequest(t, enc, dec, 3, "tools/call", map[string]any{
		"name":      "forgia_vault_fd_list",
		"arguments": map[string]any{},
	})
	if resp["error"] != nil {
		t.Fatalf("fd_list without KG error: %v", resp["error"])
	}
	listResult := extractToolResult(t, resp)
	count, _ := listResult["count"].(float64)
	if count < 1 {
		t.Error("expected at least 1 FD in list")
	}
}

// extractToolResult extracts the result from a tools/call JSON-RPC response.
func extractToolResult(t *testing.T, resp map[string]any) map[string]any {
	t.Helper()

	result, ok := resp["result"].(map[string]any)
	if !ok {
		// Try to decode the result as raw JSON first.
		raw, _ := json.Marshal(resp["result"])
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err == nil {
			return m
		}
		t.Fatalf("cannot extract result from response: %v", resp)
	}
	return result
}
