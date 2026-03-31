package skill

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Deepzima/forgia/internal/guardrails"
	"github.com/Deepzima/forgia/internal/mcp"
	"github.com/Deepzima/forgia/internal/vault"
	"gopkg.in/yaml.v3"
)

// Compile-time check: *vault.FileVault satisfies VaultReader.
var _ VaultReader = (*vault.FileVault)(nil)

// RegisterCompositeSkills instantiates all composite skills and registers them
// in both the skill registry and the MCP provider registry.
// Called from "forgia mcp serve" startup (SDD-008).
//
// Skills registered:
//
//	Primitives (SDD-004):
//	  1. arch_init    — get_architecture() + YAML mapping + merge-write
//	  2. blast_radius — detect_changes() + risk labels + context annotations
//	  3. trace_calls  — trace_call_path() + context annotations
//	  4. search_code  — search_graph() + guardrails filtering
//
//	Higher-level (SDD-005):
//	  5. security_scan  — security pattern search + guardrails gap analysis
//	  6. arch_coherence — call path vs documented architecture drift detection
//	  7. context_map    — symbol-to-context mapping
func RegisterCompositeSkills(reg *Registry, providerReg *mcp.ProviderRegistry, v VaultReader) error {
	if providerReg == nil {
		return fmt.Errorf("RegisterCompositeSkills: providerReg is required")
	}

	skills := []*CompositeSkill{
		// --- Primitives (SDD-004) ---

		// 1. arch_init: wraps get_architecture(), maps output to Architecture YAML,
		// merges with existing .forgia/architecture/system.yaml (never overwrites).
		{
			name:         "arch_init",
			description:  "Initialize architecture from codebase knowledge graph",
			category:     CategoryArchitecture,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "get_architecture",
			registry:     providerReg,
			vault:        v,
			PostProcess: func(ctx context.Context, result any) (any, error) {
				incoming := mapToArchitecture(result)

				// Merge with existing architecture if present.
				if v != nil {
					existing, err := v.GetArchitecture(ctx)
					if err == nil && existing != nil {
						incoming = mergeArchitecture(existing, incoming)
					}
				}

				// Write YAML if vault exposes Dir() (FileVault does, test mocks optionally do).
				if dg, ok := v.(interface{ Dir() string }); ok {
					if err := writeArchitectureYAML(dg.Dir(), incoming); err != nil {
						return nil, fmt.Errorf("arch_init: write architecture: %w", err)
					}
				}

				return incoming, nil
			},
		},

		// 2. blast_radius: wraps detect_changes(), adds risk labels (High/Medium/Low)
		// based on dependent count, annotates with bounded context names.
		{
			name:         "blast_radius",
			description:  "Detect changes and assess blast radius with risk labels",
			category:     CategoryDesign,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "detect_changes",
			registry:     providerReg,
			vault:        v,
			PostProcess: func(ctx context.Context, result any) (any, error) {
				resultMap, ok := result.(map[string]any)
				if !ok {
					return result, nil
				}

				addRiskLabels(resultMap)

				if v != nil {
					contexts, err := v.ListContexts(ctx)
					if err == nil {
						annotateWithContexts(resultMap, contexts)
					}
				}

				return resultMap, nil
			},
		},

		// 3. trace_calls: wraps trace_call_path(), annotates call path nodes
		// with bounded context from .forgia/contexts/.
		{
			name:         "trace_calls",
			description:  "Trace call paths with bounded context annotations",
			category:     CategoryKnowledge,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "trace_call_path",
			registry:     providerReg,
			vault:        v,
			PostProcess: func(ctx context.Context, result any) (any, error) {
				resultMap, ok := result.(map[string]any)
				if !ok {
					return result, nil
				}

				if v != nil {
					contexts, err := v.ListContexts(ctx)
					if err == nil {
						annotateCallPathWithContexts(resultMap, contexts)
					}
				}

				return resultMap, nil
			},
		},

		// 4. search_code: wraps search_graph(), filters results through
		// guardrails.CheckFilePaths() — strips paths matching [read] deny patterns.
		{
			name:         "search_code",
			description:  "Search code graph with guardrails filtering",
			category:     CategoryKnowledge,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "search_graph",
			registry:     providerReg,
			vault:        v,
			PostProcess: func(ctx context.Context, result any) (any, error) {
				if v == nil {
					return result, nil
				}

				raw, err := v.GuardrailsRaw(ctx)
				if err != nil {
					slog.WarnContext(ctx, "search_code: failed to read guardrails, returning unfiltered", "error", err)
					return result, nil
				}
				g, err := guardrails.Parse(raw)
				if err != nil {
					slog.WarnContext(ctx, "search_code: failed to parse guardrails, returning unfiltered", "error", err)
					return result, nil
				}

				return filterResultPaths(ctx, result, g), nil
			},
		},

		// --- Higher-level (SDD-005) ---
		securityScanSkill(providerReg, v),
		archCoherenceSkill(providerReg, v),
		contextMapSkill(providerReg, v),
	}

	for _, s := range skills {
		reg.Register(s)
	}

	return nil
}

// --- PostProcess helpers ---

// mapToArchitecture converts a provider result (typically map[string]any) to vault.Architecture.
func mapToArchitecture(result any) *vault.Architecture {
	arch := &vault.Architecture{}

	m, ok := result.(map[string]any)
	if !ok {
		return arch
	}

	if sc, ok := m["system_context"].(map[string]any); ok {
		arch.SystemContext = &vault.SystemContext{
			Name:        stringFromMap(sc, "name"),
			Description: stringFromMap(sc, "description"),
		}
	}

	if containers, ok := m["containers"].([]any); ok {
		items := make([]vault.Container, 0, len(containers))
		for _, c := range containers {
			if cm, ok := c.(map[string]any); ok {
				items = append(items, vault.Container{
					Name:       stringFromMap(cm, "name"),
					Technology: stringFromMap(cm, "technology"),
					Purpose:    stringFromMap(cm, "purpose"),
					Context:    stringFromMap(cm, "context"),
				})
			}
		}
		if len(items) > 0 {
			arch.Containers = &vault.Containers{Items: items}
		}
	}

	if decisions, ok := m["technology_decisions"].([]any); ok {
		items := make([]vault.TechDecision, 0, len(decisions))
		for _, d := range decisions {
			if dm, ok := d.(map[string]any); ok {
				items = append(items, vault.TechDecision{
					Area:      stringFromMap(dm, "area"),
					Choice:    stringFromMap(dm, "choice"),
					Rationale: stringFromMap(dm, "rationale"),
				})
			}
		}
		if len(items) > 0 {
			arch.TechnologyDecisions = &vault.TechnologyDecisions{Decisions: items}
		}
	}

	return arch
}

// stringFromMap safely extracts a string from a map.
func stringFromMap(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// mergeArchitecture merges incoming architecture data into existing, preserving existing values.
// New containers/decisions are appended; existing ones (matched by name/area) are updated.
func mergeArchitecture(existing, incoming *vault.Architecture) *vault.Architecture {
	if existing == nil {
		return incoming
	}
	if incoming == nil {
		return existing
	}

	merged := *existing

	// Merge SystemContext.
	if incoming.SystemContext != nil {
		if merged.SystemContext == nil {
			merged.SystemContext = incoming.SystemContext
		} else {
			if incoming.SystemContext.Name != "" {
				merged.SystemContext.Name = incoming.SystemContext.Name
			}
			if incoming.SystemContext.Description != "" {
				merged.SystemContext.Description = incoming.SystemContext.Description
			}
		}
	}

	// Merge Containers by name.
	if incoming.Containers != nil && len(incoming.Containers.Items) > 0 {
		if merged.Containers == nil {
			merged.Containers = incoming.Containers
		} else {
			byName := make(map[string]int, len(merged.Containers.Items))
			for i, c := range merged.Containers.Items {
				byName[c.Name] = i
			}
			for _, c := range incoming.Containers.Items {
				if idx, ok := byName[c.Name]; ok {
					merged.Containers.Items[idx] = c
				} else {
					merged.Containers.Items = append(merged.Containers.Items, c)
				}
			}
		}
	}

	// Merge TechnologyDecisions by area.
	if incoming.TechnologyDecisions != nil && len(incoming.TechnologyDecisions.Decisions) > 0 {
		if merged.TechnologyDecisions == nil {
			merged.TechnologyDecisions = incoming.TechnologyDecisions
		} else {
			byArea := make(map[string]int, len(merged.TechnologyDecisions.Decisions))
			for i, d := range merged.TechnologyDecisions.Decisions {
				byArea[d.Area] = i
			}
			for _, d := range incoming.TechnologyDecisions.Decisions {
				if idx, ok := byArea[d.Area]; ok {
					merged.TechnologyDecisions.Decisions[idx] = d
				} else {
					merged.TechnologyDecisions.Decisions = append(merged.TechnologyDecisions.Decisions, d)
				}
			}
		}
	}

	return &merged
}

// writeArchitectureYAML writes the architecture to <vaultDir>/architecture/system.yaml.
func writeArchitectureYAML(vaultDir string, arch *vault.Architecture) error {
	archDir := filepath.Join(vaultDir, "architecture")
	if err := os.MkdirAll(archDir, 0o755); err != nil {
		return fmt.Errorf("create architecture dir: %w", err)
	}

	data, err := yaml.Marshal(arch)
	if err != nil {
		return fmt.Errorf("marshal architecture: %w", err)
	}

	return os.WriteFile(filepath.Join(archDir, "system.yaml"), data, 0o644)
}

// addRiskLabels adds High/Medium/Low risk labels to detect_changes results
// based on the number of dependents for each changed item.
func addRiskLabels(result map[string]any) {
	changes, ok := result["changes"].([]any)
	if !ok {
		return
	}

	for _, change := range changes {
		cm, ok := change.(map[string]any)
		if !ok {
			continue
		}

		dependents, _ := cm["dependents"].([]any)
		count := len(dependents)

		switch {
		case count >= 5:
			cm["risk"] = "High"
		case count >= 2:
			cm["risk"] = "Medium"
		default:
			cm["risk"] = "Low"
		}
	}
}

// annotateWithContexts adds bounded context names to change results.
func annotateWithContexts(result map[string]any, contexts []*vault.BoundedContext) {
	changes, ok := result["changes"].([]any)
	if !ok {
		return
	}

	for _, change := range changes {
		cm, ok := change.(map[string]any)
		if !ok {
			continue
		}

		path, _ := cm["path"].(string)
		if path != "" {
			if ctx := findContextForPath(path, contexts); ctx != "" {
				cm["context"] = ctx
			}
		}
	}
}

// annotateCallPathWithContexts annotates call path nodes with bounded context.
func annotateCallPathWithContexts(result map[string]any, contexts []*vault.BoundedContext) {
	callPath, ok := result["call_path"].([]any)
	if !ok {
		return
	}

	for _, node := range callPath {
		nm, ok := node.(map[string]any)
		if !ok {
			continue
		}

		file, _ := nm["file"].(string)
		if file != "" {
			if ctx := findContextForPath(file, contexts); ctx != "" {
				nm["context"] = ctx
			}
		}
	}
}

// findContextForPath matches a file path to a bounded context by checking
// if the path contains any context's responsibility keywords.
func findContextForPath(path string, contexts []*vault.BoundedContext) string {
	for _, bc := range contexts {
		for _, resp := range bc.Responsibilities {
			if strings.Contains(path, resp) {
				return bc.Name
			}
		}
	}
	return ""
}

// filterResultPaths removes file paths from search results that match guardrails deny patterns.
func filterResultPaths(ctx context.Context, result any, g *guardrails.Guardrails) any {
	resultMap, ok := result.(map[string]any)
	if !ok {
		return result
	}

	results, ok := resultMap["results"].([]any)
	if !ok {
		return result
	}

	var filtered []any
	for _, item := range results {
		im, ok := item.(map[string]any)
		if !ok {
			filtered = append(filtered, item)
			continue
		}

		path, _ := im["path"].(string)
		if path == "" {
			path, _ = im["file"].(string)
		}
		if path == "" {
			filtered = append(filtered, item)
			continue
		}

		violations := g.CheckFilePaths(ctx, []string{path})
		if len(violations) == 0 {
			filtered = append(filtered, item)
		}
	}

	resultMap["results"] = filtered
	return resultMap
}

// --- Higher-level skill constructors (SDD-005) ---

// securityPatternQueries defines the search queries for security-relevant code patterns.
var securityPatternQueries = []string{
	"auth", "authentication", "authorization",
	"validate", "validation", "sanitize",
	"crypto", "encrypt", "decrypt", "hash",
	"secret", "password", "credential", "token",
	"injection", "xss", "csrf",
}

// securityScanSkill creates the security_scan higher-level skill.
// Searches codebase for security patterns (auth, validation, crypto, secrets,
// input sanitization), then cross-references findings against deny.toml —
// identifies gaps where code handles secrets but guardrails don't protect the files.
// Returns file paths and pattern names ONLY, never actual secret values (SDD-005).
func securityScanSkill(providerReg *mcp.ProviderRegistry, v VaultReader) *CompositeSkill {
	return &CompositeSkill{
		name:         "security_scan",
		description:  "Security pattern search + guardrails gap analysis",
		category:     CategoryKnowledge,
		mode:         ModeMCPTool,
		ProviderName: "code",
		ProviderTool: "search_graph",
		registry:     providerReg,
		vault:        v,
		graceful:     true,
		PreProcess:   securityScanPreProcess,
		PostProcess: func(ctx context.Context, result any) (any, error) {
			return securityScanPostProcess(ctx, result, v)
		},
	}
}

// securityScanPreProcess builds security-focused search queries.
// If no query is provided, uses all security pattern keywords.
func securityScanPreProcess(_ context.Context, params map[string]any) (map[string]any, error) {
	if params == nil {
		params = make(map[string]any)
	}
	if _, ok := params["query"]; !ok {
		params["query"] = strings.Join(securityPatternQueries, " ")
	}
	return params, nil
}

// securityScanPostProcess cross-references search results with deny.toml patterns.
// Returns file paths and pattern names ONLY — never actual secret content.
// Identifies guardrail gaps: files that handle security-sensitive code but aren't
// protected by deny.toml [read] patterns.
func securityScanPostProcess(ctx context.Context, result any, v VaultReader) (any, error) {
	resultMap, ok := result.(map[string]any)
	if !ok {
		return result, nil
	}

	results, ok := resultMap["results"].([]any)
	if !ok {
		return result, nil
	}

	// Sanitize output: extract only path and name (no content/values).
	var findings []map[string]any
	var filePaths []string
	for _, r := range results {
		rm, ok := r.(map[string]any)
		if !ok {
			continue
		}
		path, _ := rm["path"].(string)
		name, _ := rm["name"].(string)
		finding := map[string]any{"path": path}
		if name != "" {
			finding["name"] = name
		}
		findings = append(findings, finding)
		if path != "" {
			filePaths = append(filePaths, path)
		}
	}

	// Cross-reference with deny.toml to identify gaps.
	var gaps []map[string]any
	if v != nil {
		raw, err := v.GuardrailsRaw(ctx)
		if err == nil {
			g, gErr := guardrails.Parse(raw)
			if gErr == nil {
				for _, path := range filePaths {
					violations := g.CheckReadPaths(ctx, []string{path})
					if len(violations) == 0 {
						gaps = append(gaps, map[string]any{
							"path":   path,
							"status": "unprotected",
							"detail": "handles security patterns but not in deny.toml read list",
						})
					}
				}
			} else {
				slog.WarnContext(ctx, "security_scan: failed to parse guardrails, skipping gap analysis", "error", gErr)
			}
		} else {
			slog.WarnContext(ctx, "security_scan: failed to read guardrails, skipping gap analysis", "error", err)
		}
	}

	return map[string]any{
		"findings":       findings,
		"guardrail_gaps": gaps,
	}, nil
}

// archCoherenceSkill creates the arch_coherence higher-level skill.
// Traces call paths and compares against documented architecture — flags drift
// where components call each other but aren't documented as connected (SDD-005).
func archCoherenceSkill(providerReg *mcp.ProviderRegistry, v VaultReader) *CompositeSkill {
	return &CompositeSkill{
		name:         "arch_coherence",
		description:  "Call path vs documented architecture drift detection",
		category:     CategoryArchitecture,
		mode:         ModeMCPTool,
		ProviderName: "code",
		ProviderTool: "trace_call_path",
		registry:     providerReg,
		vault:        v,
		graceful:     true,
		PostProcess: func(ctx context.Context, result any) (any, error) {
			return archCoherencePostProcess(ctx, result, v)
		},
	}
}

// archCoherencePostProcess compares actual call paths against documented architecture.
// Flags undocumented component dependencies (drift).
// Degrades gracefully: if architecture or contexts are unavailable, returns raw
// results with a warning instead of failing.
func archCoherencePostProcess(ctx context.Context, result any, v VaultReader) (any, error) {
	resultMap, ok := result.(map[string]any)
	if !ok {
		return result, nil
	}

	callPath, ok := resultMap["call_path"].([]any)
	if !ok {
		return result, nil
	}

	if v == nil {
		resultMap["drift"] = []any{}
		resultMap["_warning"] = "vault not available, drift detection skipped"
		return resultMap, nil
	}

	arch, err := v.GetArchitecture(ctx)
	if err != nil || arch == nil {
		slog.WarnContext(ctx, "arch_coherence: architecture not available, skipping drift detection", "error", err)
		resultMap["drift"] = []any{}
		resultMap["_warning"] = "architecture not initialized, drift detection skipped"
		return resultMap, nil
	}

	contexts, err := v.ListContexts(ctx)
	if err != nil {
		slog.WarnContext(ctx, "arch_coherence: contexts not available, skipping drift detection", "error", err)
		resultMap["drift"] = []any{}
		resultMap["_warning"] = "contexts not available, drift detection skipped"
		return resultMap, nil
	}

	// Build documented dependency set from bounded contexts.
	documented := buildDocumentedDeps(contexts)

	// Detect undocumented connections in call path.
	var drift []map[string]any
	for i := 0; i < len(callPath)-1; i++ {
		from, ok1 := callPath[i].(map[string]any)
		to, ok2 := callPath[i+1].(map[string]any)
		if !ok1 || !ok2 {
			continue
		}

		fromFile, _ := from["file"].(string)
		toFile, _ := to["file"].(string)

		fromCtx := findContextForPath(fromFile, contexts)
		toCtx := findContextForPath(toFile, contexts)

		if fromCtx != "" && toCtx != "" && fromCtx != toCtx {
			key := fromCtx + " → " + toCtx
			if !documented[key] {
				drift = append(drift, map[string]any{
					"from":        fromCtx,
					"to":          toCtx,
					"source_file": fromFile,
					"target_file": toFile,
					"status":      "undocumented",
				})
			}
		}
	}

	resultMap["drift"] = drift
	return resultMap, nil
}

// buildDocumentedDeps builds a set of documented dependency edges from bounded contexts.
// Format: "context_a → context_b".
func buildDocumentedDeps(contexts []*vault.BoundedContext) map[string]bool {
	deps := make(map[string]bool)
	for _, c := range contexts {
		for _, dep := range c.Dependencies.DependsOn {
			deps[c.Name+" → "+dep] = true
		}
		for _, dep := range c.Dependencies.DependedBy {
			deps[dep+" → "+c.Name] = true
		}
	}
	return deps
}

// contextMapSkill creates the context_map higher-level skill.
// Searches code and maps symbols/changes to their bounded context,
// identifying cross-context boundaries (SDD-005).
func contextMapSkill(providerReg *mcp.ProviderRegistry, v VaultReader) *CompositeSkill {
	return &CompositeSkill{
		name:         "context_map",
		description:  "Symbol-to-context mapping",
		category:     CategoryKnowledge,
		mode:         ModeMCPTool,
		ProviderName: "code",
		ProviderTool: "search_graph",
		registry:     providerReg,
		vault:        v,
		graceful:     true,
		PostProcess: func(ctx context.Context, result any) (any, error) {
			return contextMapPostProcess(ctx, result, v)
		},
	}
}

// contextMapPostProcess maps code symbols to bounded contexts and identifies
// cross-context boundary crossings.
// Degrades gracefully: if contexts are unavailable, returns raw results with a warning.
func contextMapPostProcess(ctx context.Context, result any, v VaultReader) (any, error) {
	resultMap, ok := result.(map[string]any)
	if !ok {
		return result, nil
	}

	results, ok := resultMap["results"].([]any)
	if !ok {
		return result, nil
	}

	if v == nil {
		resultMap["_warning"] = "vault not available, context mapping skipped"
		return resultMap, nil
	}

	contexts, err := v.ListContexts(ctx)
	if err != nil {
		slog.WarnContext(ctx, "context_map: contexts not available, skipping mapping", "error", err)
		resultMap["_warning"] = "contexts not available, mapping skipped"
		return resultMap, nil
	}

	// Map each result to its bounded context.
	var mappings []map[string]any
	contextNames := make(map[string]bool)

	for _, r := range results {
		rm, ok := r.(map[string]any)
		if !ok {
			continue
		}

		path, _ := rm["path"].(string)
		name, _ := rm["name"].(string)
		ctxName := findContextForPath(path, contexts)

		m := map[string]any{
			"path":   path,
			"symbol": name,
		}
		if ctxName != "" {
			m["context"] = ctxName
			contextNames[ctxName] = true
		}
		mappings = append(mappings, m)
	}

	// Detect boundary crossings: adjacent results in different contexts.
	var crossings []map[string]any
	for i := 0; i < len(mappings)-1; i++ {
		fromCtx, _ := mappings[i]["context"].(string)
		toCtx, _ := mappings[i+1]["context"].(string)
		if fromCtx != "" && toCtx != "" && fromCtx != toCtx {
			crossings = append(crossings, map[string]any{
				"from_context": fromCtx,
				"to_context":   toCtx,
				"from_path":    mappings[i]["path"],
				"to_path":      mappings[i+1]["path"],
			})
		}
	}

	// Convert context set to sorted list.
	contextList := make([]string, 0, len(contextNames))
	for name := range contextNames {
		contextList = append(contextList, name)
	}
	sort.Strings(contextList)

	return map[string]any{
		"mappings":  mappings,
		"contexts":  contextList,
		"crossings": crossings,
	}, nil
}
