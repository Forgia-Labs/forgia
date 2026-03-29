package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Deepzima/forgia/internal/guardrails"
	"github.com/Deepzima/forgia/internal/vault"
)

// validIDPattern matches FD/SDD identifiers: FD-001, FD-a3f2, SDD-001, etc.
// Rejects path traversal attempts like "../../etc/passwd".
var validIDPattern = regexp.MustCompile(`^[A-Z]+-[a-zA-Z0-9]{1,10}$`)

// validateID checks that an ID is safe to use in filesystem paths.
func validateID(id, kind string) error {
	if !validIDPattern.MatchString(id) {
		return fmt.Errorf("invalid %s ID: %q", kind, id)
	}
	return nil
}

// VaultProvider exposes vault read and write operations as MCP tools.
// Read: forgia_vault_fd_get, forgia_vault_fd_list, forgia_vault_sdd_get,
// forgia_vault_sdd_list, forgia_vault_status, forgia_vault_validate,
// forgia_vault_threat_model_get.
// Write: forgia_vault_fd_create, forgia_vault_fd_update,
// forgia_vault_sdd_create, forgia_vault_sdd_update.
type VaultProvider struct {
	vault    vault.Vault
	guard    *guardrails.Guardrails
	registry *ProviderRegistry
	sddMu   sync.Mutex // protects nextSDDID + callSDDCreate to prevent race conditions
}

// Compile-time interface satisfaction check.
var _ ToolProvider = (*VaultProvider)(nil)

// NewVaultProvider creates an MCP provider for vault read operations.
func NewVaultProvider(v vault.Vault, g *guardrails.Guardrails, reg *ProviderRegistry) *VaultProvider {
	return &VaultProvider{vault: v, guard: g, registry: reg}
}

func (p *VaultProvider) Name() string { return "vault" }

func (p *VaultProvider) Tools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "fd_get",
			Description: "Get a Feature Design by ID, returning frontmatter and body",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "FD identifier (e.g., FD-001)",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "fd_list",
			Description: "List Feature Designs with optional status filter",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Filter by status (e.g., planned, approved, in-progress, complete)",
					},
				},
			},
		},
		{
			Name:        "sdd_get",
			Description: "Get an SDD by FD and SDD ID, returning frontmatter and body",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"fd": map[string]any{
						"type":        "string",
						"description": "Parent FD identifier (e.g., FD-001)",
					},
					"id": map[string]any{
						"type":        "string",
						"description": "SDD identifier (e.g., SDD-001)",
					},
				},
				"required": []string{"fd", "id"},
			},
		},
		{
			Name:        "sdd_list",
			Description: "List SDDs for a Feature Design with optional status filter",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"fd": map[string]any{
						"type":        "string",
						"description": "Parent FD identifier (e.g., FD-001)",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "Filter by status (e.g., planned, in-progress, done)",
					},
				},
				"required": []string{"fd"},
			},
		},
		{
			Name:        "status",
			Description: "Get full project dashboard with FDs, SDDs, OPS tasks, and execution logs",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "validate",
			Description: "Validate an SDD or all SDDs under an FD against the template",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Path to a specific SDD file to validate",
					},
					"fd": map[string]any{
						"type":        "string",
						"description": "FD identifier — validates all SDDs under this FD",
					},
				},
			},
		},
		{
			Name:        "threat_model_get",
			Description: "Get the threat model for a Feature Design, or null if not present",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"fd": map[string]any{
						"type":        "string",
						"description": "FD identifier (e.g., FD-001)",
					},
				},
				"required": []string{"fd"},
			},
		},
		// --- write tools ---
		{
			Name:        "fd_create",
			Description: "Create a new Feature Design with auto-generated hash-based ID",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title": map[string]any{
						"type":        "string",
						"description": "Feature Design title",
					},
					"author": map[string]any{
						"type":        "string",
						"description": "Author name",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "FD status (default: planned)",
						"enum":        []string{"planned", "approved", "in-progress", "complete", "closed", "rejected", "abandoned"},
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Priority level (high, medium, low)",
					},
					"tags": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Tags for categorization",
					},
					"upstream_issue": map[string]any{
						"type":        "string",
						"description": "Link to upstream issue (e.g., GitHub issue URL)",
					},
					"body": map[string]any{
						"type":        "string",
						"description": "Markdown body content for the FD",
					},
				},
				"required": []string{"title"},
			},
		},
		{
			Name:        "fd_update",
			Description: "Update an existing Feature Design's frontmatter fields",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "FD identifier (e.g., FD-001)",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "New status",
						"enum":        []string{"planned", "approved", "in-progress", "complete", "closed", "rejected", "abandoned"},
					},
					"reviewed": map[string]any{
						"type":        "boolean",
						"description": "Whether the FD has been reviewed",
					},
					"reviewer": map[string]any{
						"type":        "string",
						"description": "Reviewer name",
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Priority level",
					},
					"tags": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Tags for categorization",
					},
					"assignee": map[string]any{
						"type":        "string",
						"description": "Assigned person",
					},
					"upstream_issue": map[string]any{
						"type":        "string",
						"description": "Link to upstream issue",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "sdd_create",
			Description: "Create a new SDD under a Feature Design with auto-generated sequential ID",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"fd": map[string]any{
						"type":        "string",
						"description": "Parent FD identifier (e.g., FD-001)",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "SDD title",
					},
					"scope": map[string]any{
						"type":        "string",
						"description": "Scope description",
					},
					"constraints": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"language":  map[string]any{"type": "string"},
							"framework": map[string]any{"type": "string"},
						},
						"description": "Language and framework constraints",
					},
					"acceptance_criteria": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Acceptance criteria as strings",
					},
					"tags": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Tags for categorization",
					},
				},
				"required": []string{"fd", "title"},
			},
		},
		{
			Name:        "sdd_update",
			Description: "Update an existing SDD's frontmatter fields",
			Namespace:   "vault",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"fd": map[string]any{
						"type":        "string",
						"description": "Parent FD identifier (e.g., FD-001)",
					},
					"id": map[string]any{
						"type":        "string",
						"description": "SDD identifier (e.g., SDD-001)",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "New status",
						"enum":        []string{"planned", "validated", "in-progress", "done", "failed", "cancelled"},
					},
					"agent": map[string]any{
						"type":        "string",
						"description": "Agent executing the SDD (e.g., claude-code)",
					},
					"assigned_to": map[string]any{
						"type":        "string",
						"description": "Person or agent assigned",
					},
					"started": map[string]any{
						"type":        "string",
						"description": "Start timestamp (ISO 8601)",
					},
					"completed": map[string]any{
						"type":        "string",
						"description": "Completion timestamp (ISO 8601)",
					},
					"tags": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Tags for categorization",
					},
				},
				"required": []string{"fd", "id"},
			},
		},
	}
}

func (p *VaultProvider) Call(ctx context.Context, tool string, params map[string]any) (any, error) {
	if params == nil {
		params = map[string]any{}
	}
	switch tool {
	case "fd_get":
		return p.callFDGet(ctx, params)
	case "fd_list":
		return p.callFDList(ctx, params)
	case "sdd_get":
		return p.callSDDGet(ctx, params)
	case "sdd_list":
		return p.callSDDList(ctx, params)
	case "status":
		return p.callStatus(ctx, params)
	case "validate":
		return p.callValidate(ctx, params)
	case "threat_model_get":
		return p.callThreatModelGet(ctx, params)
	case "fd_create":
		return p.callFDCreate(ctx, params)
	case "fd_update":
		return p.callFDUpdate(ctx, params)
	case "sdd_create":
		return p.callSDDCreate(ctx, params)
	case "sdd_update":
		return p.callSDDUpdate(ctx, params)
	default:
		return nil, fmt.Errorf("unknown vault tool: %s", tool)
	}
}

func (p *VaultProvider) Start(_ context.Context) error { return nil }
func (p *VaultProvider) Stop() error                   { return nil }
func (p *VaultProvider) Healthy() bool                 { return p.vault != nil }

// --- tool handlers ---

func (p *VaultProvider) callFDGet(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if err := validateID(id, "fd"); err != nil {
		return nil, err
	}

	fd, err := p.vault.GetFD(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get FD %s: %w", id, err)
	}

	result := map[string]any{
		"id":       fd.ID,
		"title":    fd.Title,
		"status":   string(fd.Status),
		"priority": fd.Priority,
		"author":   fd.Author,
		"created":  fd.Created,
		"tags":     fd.Tags,
	}

	if fd.FilePath != "" {
		if data, err := os.ReadFile(fd.FilePath); err == nil {
			result["body"] = extractMarkdownBody(data)
		}
	}

	return result, nil
}

func (p *VaultProvider) callFDList(ctx context.Context, params map[string]any) (any, error) {
	fds, err := p.vault.ListFDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list FDs: %w", err)
	}

	statusFilter, _ := params["status"].(string)

	var items []map[string]any
	for _, fd := range fds {
		if statusFilter != "" && string(fd.Status) != statusFilter {
			continue
		}
		items = append(items, map[string]any{
			"id":       fd.ID,
			"title":    fd.Title,
			"status":   string(fd.Status),
			"priority": fd.Priority,
			"author":   fd.Author,
		})
	}

	if items == nil {
		items = []map[string]any{}
	}

	return map[string]any{
		"count": len(items),
		"fds":   items,
	}, nil
}

func (p *VaultProvider) callSDDGet(ctx context.Context, params map[string]any) (any, error) {
	fdID, _ := params["fd"].(string)
	sddID, _ := params["id"].(string)
	if fdID == "" {
		return nil, fmt.Errorf("fd is required")
	}
	if sddID == "" {
		return nil, fmt.Errorf("id is required")
	}
	if err := validateID(fdID, "fd"); err != nil {
		return nil, err
	}
	if err := validateID(sddID, "sdd"); err != nil {
		return nil, err
	}

	sdd, err := p.vault.GetSDD(ctx, fdID, sddID)
	if err != nil {
		return nil, fmt.Errorf("get SDD %s/%s: %w", fdID, sddID, err)
	}

	result := map[string]any{
		"id":     sdd.ID,
		"fd":     sdd.FD,
		"title":  sdd.Title,
		"status": string(sdd.Status),
		"agent":  sdd.Agent,
		"tags":   sdd.Tags,
	}

	if sdd.FilePath != "" {
		if data, err := os.ReadFile(sdd.FilePath); err == nil {
			result["body"] = extractMarkdownBody(data)
		}
	}

	return result, nil
}

func (p *VaultProvider) callSDDList(ctx context.Context, params map[string]any) (any, error) {
	fdID, _ := params["fd"].(string)
	if fdID == "" {
		return nil, fmt.Errorf("fd is required")
	}
	if err := validateID(fdID, "fd"); err != nil {
		return nil, err
	}

	sdds, err := p.vault.ListSDDs(ctx, fdID)
	if err != nil {
		return nil, fmt.Errorf("list SDDs for %s: %w", fdID, err)
	}

	statusFilter, _ := params["status"].(string)

	var items []map[string]any
	for _, sdd := range sdds {
		if statusFilter != "" && string(sdd.Status) != statusFilter {
			continue
		}
		items = append(items, map[string]any{
			"id":     sdd.ID,
			"fd":     sdd.FD,
			"title":  sdd.Title,
			"status": string(sdd.Status),
		})
	}

	if items == nil {
		items = []map[string]any{}
	}

	return map[string]any{
		"count": len(items),
		"sdds":  items,
	}, nil
}

func (p *VaultProvider) callStatus(ctx context.Context, _ map[string]any) (any, error) {
	fds, err := p.vault.ListFDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list FDs: %w", err)
	}

	var fdItems []map[string]any
	for _, fd := range fds {
		sdds, _ := p.vault.ListSDDs(ctx, fd.ID)
		var sddItems []map[string]any
		for _, sdd := range sdds {
			sddItems = append(sddItems, map[string]any{
				"id":     sdd.ID,
				"title":  sdd.Title,
				"status": string(sdd.Status),
			})
		}
		if sddItems == nil {
			sddItems = []map[string]any{}
		}
		fdItems = append(fdItems, map[string]any{
			"id":       fd.ID,
			"title":    fd.Title,
			"status":   string(fd.Status),
			"priority": fd.Priority,
			"author":   fd.Author,
			"sdds":     sddItems,
		})
	}
	if fdItems == nil {
		fdItems = []map[string]any{}
	}

	dashboard := map[string]any{
		"fds": fdItems,
	}

	// OPS tasks — degrade gracefully if directory missing.
	if tasks := p.readOPSTasks(); len(tasks) > 0 {
		dashboard["ops_tasks"] = tasks
	}

	// Execution logs — degrade gracefully if directory missing.
	if logs := p.readExecLogs(); len(logs) > 0 {
		dashboard["exec_logs"] = logs
	}

	// KG enrichment: knowledge layer stats.
	if p.registry != nil {
		if stats := p.fetchKnowledgeStats(ctx); stats != nil {
			dashboard["knowledge"] = stats
		}
	}

	return dashboard, nil
}

func (p *VaultProvider) callValidate(ctx context.Context, params map[string]any) (any, error) {
	path, _ := params["path"].(string)
	fdID, _ := params["fd"].(string)

	if path == "" && fdID == "" {
		return nil, fmt.Errorf("either path or fd is required")
	}
	if fdID != "" {
		if err := validateID(fdID, "fd"); err != nil {
			return nil, err
		}
	}

	var sddFiles []string
	if fdID != "" {
		sddDir := filepath.Join(p.vault.Dir(), "sdd", fdID)
		entries, err := os.ReadDir(sddDir)
		if err != nil {
			return nil, fmt.Errorf("read SDD directory %s: %w", fdID, err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), "SDD-") && strings.HasSuffix(e.Name(), ".md") {
				sddFiles = append(sddFiles, filepath.Join(sddDir, e.Name()))
			}
		}
		if len(sddFiles) == 0 {
			return nil, fmt.Errorf("no SDD files found for %s", fdID)
		}
	} else {
		sddFiles = []string{path}
	}

	var results []map[string]any
	for _, f := range sddFiles {
		errs := p.validateSDDFile(ctx, f)
		valid := len(errs) == 0
		result := map[string]any{
			"file":  filepath.Base(f),
			"valid": valid,
		}
		if !valid {
			result["errors"] = errs
		}
		results = append(results, result)
	}

	return map[string]any{
		"files": results,
	}, nil
}

func (p *VaultProvider) callThreatModelGet(_ context.Context, params map[string]any) (any, error) {
	fdID, _ := params["fd"].(string)
	if fdID == "" {
		return nil, fmt.Errorf("fd is required")
	}
	if err := validateID(fdID, "fd"); err != nil {
		return nil, err
	}

	tmPath := filepath.Join(p.vault.Dir(), "fd", fdID+"-threat-model.md")
	data, err := os.ReadFile(tmPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{"fd": fdID, "content": nil}, nil
		}
		return nil, fmt.Errorf("read threat model for %s: %w", fdID, err)
	}

	return map[string]any{
		"fd":      fdID,
		"content": string(data),
	}, nil
}

// --- status helpers ---

func (p *VaultProvider) readOPSTasks() []map[string]any {
	opsDir := filepath.Join(p.vault.Dir(), "ops", "active")
	entries, err := os.ReadDir(opsDir)
	if err != nil {
		return nil
	}

	var tasks []map[string]any
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "OPS-") || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(opsDir, e.Name()))
		if err != nil {
			continue
		}
		fm := parseFrontmatterMap(data)
		if fm != nil {
			tasks = append(tasks, fm)
		}
	}
	return tasks
}

func (p *VaultProvider) readExecLogs() []map[string]any {
	logsDir := filepath.Join(p.vault.Dir(), "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return nil
	}

	var logs []map[string]any
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "exec-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(logsDir, e.Name()))
		if err != nil {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}
		logs = append(logs, entry)
	}

	sort.Slice(logs, func(i, j int) bool {
		si, _ := logs[i]["started"].(string)
		sj, _ := logs[j]["started"].(string)
		return si > sj
	})

	if len(logs) > 10 {
		logs = logs[:10]
	}

	return logs
}

// fetchKnowledgeStats attempts to get knowledge layer stats from the code provider.
// Returns nil if unavailable or on timeout.
func (p *VaultProvider) fetchKnowledgeStats(ctx context.Context) any {
	kgCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := p.registry.Call(kgCtx, "forgia_code_list_projects", nil)
	if err != nil {
		return nil
	}
	return result
}

// --- validate helpers ---

func (p *VaultProvider) validateSDDFile(ctx context.Context, path string) []map[string]any {
	var errs []map[string]any

	data, err := os.ReadFile(path)
	if err != nil {
		return []map[string]any{{"category": "file", "message": fmt.Sprintf("cannot read: %v", err)}}
	}

	content := string(data)

	// Frontmatter presence.
	if !strings.Contains(content, "---") {
		return []map[string]any{{"category": "frontmatter", "message": "missing YAML frontmatter delimiters (---)"}}
	}

	// Required frontmatter fields.
	for _, field := range []string{"id:", "fd:", "title:", "status:"} {
		if !hasFrontmatterField(content, field) {
			errs = append(errs, map[string]any{
				"category": "frontmatter",
				"message":  fmt.Sprintf("missing required field: %s", field),
			})
		}
	}

	// Required sections.
	for _, section := range []string{
		"## Scope",
		"## Interfaces",
		"## Constraints",
		"## Test Requirements",
		"## Acceptance Criteria",
		"## Context",
		"## Constitution Check",
		"## Work Log",
	} {
		if !strings.Contains(content, section) {
			errs = append(errs, map[string]any{
				"category": "section",
				"message":  fmt.Sprintf("missing required section: %s", section),
			})
		}
	}

	// Parent FD exists in vault.
	fdID := frontmatterFieldValue(content, "fd:")
	if fdID != "" {
		fdID = strings.Trim(fdID, "\"' ")
		if _, err := p.vault.GetFD(ctx, fdID); err != nil {
			errs = append(errs, map[string]any{
				"category": "reference",
				"message":  fmt.Sprintf("parent FD %q not found in vault", fdID),
			})
		}
	}

	// Guardrails check on context file paths.
	contextPaths := extractContextFileRefs(content)
	if p.guard != nil && len(contextPaths) > 0 {
		violations := p.guard.CheckFilePaths(ctx, contextPaths)
		for _, v := range violations {
			errs = append(errs, map[string]any{
				"category": "guardrail",
				"message":  v.Error(),
			})
		}
	}

	// KG enrichment: verify context paths via knowledge graph.
	if len(contextPaths) > 0 {
		if p.registry != nil {
			p.verifyContextPathsKG(ctx, contextPaths, &errs)
		} else {
			p.verifyContextPathsFS(contextPaths, &errs)
		}
	}

	return errs
}

// verifyContextPathsKG checks context paths via the code provider's search tool.
// Falls back to filesystem check on any error.
func (p *VaultProvider) verifyContextPathsKG(ctx context.Context, paths []string, errs *[]map[string]any) {
	kgCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	for _, path := range paths {
		result, err := p.registry.Call(kgCtx, "forgia_code_search", map[string]any{"query": path})
		if err != nil {
			// KG unavailable — fall back to filesystem for remaining paths.
			p.verifyContextPathsFS(paths, errs)
			return
		}
		if resultMap, ok := result.(map[string]any); ok {
			if count, _ := resultMap["count"].(float64); count == 0 {
				*errs = append(*errs, map[string]any{
					"category": "context",
					"message":  fmt.Sprintf("context path %q not found in codebase", path),
				})
			}
		}
	}
}

// verifyContextPathsFS checks context paths exist on the filesystem.
func (p *VaultProvider) verifyContextPathsFS(paths []string, errs *[]map[string]any) {
	projectRoot := filepath.Dir(p.vault.Dir())
	for _, path := range paths {
		fullPath := filepath.Join(projectRoot, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			*errs = append(*errs, map[string]any{
				"category": "context",
				"message":  fmt.Sprintf("context path %q not found on filesystem", path),
			})
		}
	}
}

// --- parsing helpers ---

// extractMarkdownBody returns everything after the YAML frontmatter delimiters.
func extractMarkdownBody(data []byte) string {
	content := strings.TrimSpace(string(data))
	if !strings.HasPrefix(content, "---") {
		return string(data)
	}
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return ""
	}
	body := rest[idx+4:]
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}
	return body
}

// hasFrontmatterField checks if a field exists in the YAML frontmatter section.
func hasFrontmatterField(content, field string) bool {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return false
	}
	return strings.Contains(parts[1], field)
}

// frontmatterFieldValue extracts a value from a frontmatter field.
func frontmatterFieldValue(content, field string) string {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return ""
	}
	for _, line := range strings.Split(parts[1], "\n") {
		line = strings.TrimSpace(line)
		if val, ok := strings.CutPrefix(line, field); ok {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

// extractContextFileRefs extracts file paths from the ## Context section.
func extractContextFileRefs(content string) []string {
	idx := strings.Index(content, "## Context")
	if idx < 0 {
		return nil
	}
	section := content[idx:]
	nextSection := strings.Index(section[1:], "\n## ")
	if nextSection > 0 {
		section = section[:nextSection+1]
	}

	var paths []string
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "`") {
			continue
		}
		start := strings.Index(line, "`")
		end := strings.LastIndex(line, "`")
		if start >= end {
			continue
		}
		p := line[start+1 : end]
		if !strings.HasPrefix(p, "http") && !strings.Contains(p, " ") {
			paths = append(paths, p)
		}
	}
	return paths
}

// parseFrontmatterMap parses YAML frontmatter from a markdown file into a map.
func parseFrontmatterMap(data []byte) map[string]any {
	content := strings.TrimSpace(string(data))
	if !strings.HasPrefix(content, "---") {
		return nil
	}
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return nil
	}
	fm := strings.TrimSpace(rest[:idx])

	fields := make(map[string]any)
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, "\"'")
			fields[key] = val
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}
