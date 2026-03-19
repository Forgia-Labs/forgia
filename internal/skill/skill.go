// Package skill manages the unified skill system.
//
// A skill is any capability Forgia exposes — to humans (slash commands),
// to agents (MCP tools), or both. Skills come in three types:
//
//   - SlashCommand: user types /fd-new → Claude reads .md, calls Go logic
//   - MCPTool: agent calls forgia_blast_radius → Go logic directly
//   - Both: /arch-review (human) and forgia_arch_review (agent) share the same Go logic
//
// Skills can be "native" (implemented in Go) or "composite" (wrapping an
// external MCP provider like codebase-memory-mcp with added Forgia logic).
package skill

import (
	"context"

	"github.com/Deepzima/forgia/internal/mcp"
)

// Category groups skills by domain.
type Category string

const (
	CategoryFD           Category = "fd"
	CategorySDD          Category = "sdd"
	CategoryArchitecture Category = "architecture"
	CategoryDesign       Category = "design"
	CategoryKnowledge    Category = "knowledge" // codebase-memory-mcp derived
)

// Mode describes how a skill is invoked.
type Mode string

const (
	ModeSlashCommand Mode = "slash_command" // user types /fd-new
	ModeMCPTool      Mode = "mcp_tool"     // agent calls forgia_fd_new
	ModeBoth         Mode = "both"         // available as both
)

// Skill is a capability Forgia exposes.
type Skill interface {
	// Metadata.
	Name() string            // "fd-new", "blast-radius"
	Description() string     // human-readable description
	Category() Category      // fd, sdd, architecture, design, knowledge
	Mode() Mode              // slash_command, mcp_tool, both

	// Slash command support (nil if Mode == MCPTool).
	SlashCommandFile() string // relative path: "modules/claude-commands/fd-new.md"

	// MCP tool support (nil if Mode == SlashCommand).
	MCPToolDef() *mcp.ToolDefinition

	// Execute runs the skill's Go logic.
	Execute(ctx context.Context, params map[string]any) (any, error)
}

// CompositeSkill wraps an external MCP provider tool with Forgia logic.
// Example: forgia_blast_radius = codebase-memory detect_changes + guardrails check.
type CompositeSkill struct {
	name        string
	description string
	category    Category
	mode        Mode

	// The underlying provider tool to call.
	ProviderName string // "codebase-memory"
	ProviderTool string // "detect_changes"

	// Pre/post processing hooks.
	PreProcess  func(ctx context.Context, params map[string]any) (map[string]any, error)
	PostProcess func(ctx context.Context, result any) (any, error)

	// Provider registry for calling the underlying tool.
	registry *mcp.ProviderRegistry
}

// Name returns the skill name.
func (s *CompositeSkill) Name() string { return s.name }

// Description returns the skill description.
func (s *CompositeSkill) Description() string { return s.description }

// Category returns the skill category.
func (s *CompositeSkill) Category() Category { return s.category }

// Mode returns the skill invocation mode.
func (s *CompositeSkill) Mode() Mode { return s.mode }

// SlashCommandFile returns empty — composite skills are MCP-only.
func (s *CompositeSkill) SlashCommandFile() string { return "" }

// MCPToolDef returns the MCP tool definition.
// NOTE: Namespace is set to the skill name here (not "forgia") because this definition
// is returned directly to the MCP client, NOT passed through ProviderRegistry.AllTools()
// which would double-prefix it as "forgia_forgia_<tool>".
func (s *CompositeSkill) MCPToolDef() *mcp.ToolDefinition {
	return &mcp.ToolDefinition{
		Name:        s.name,
		Description: s.description,
	}
}

// Execute calls the underlying provider tool with pre/post processing.
func (s *CompositeSkill) Execute(ctx context.Context, params map[string]any) (any, error) {
	// Pre-process (e.g., add guardrails context, filter params).
	if s.PreProcess != nil {
		var err error
		params, err = s.PreProcess(ctx, params)
		if err != nil {
			return nil, err
		}
	}

	// Call underlying provider.
	result, err := s.registry.Call(ctx, mcp.ToolName(s.ProviderName, s.ProviderTool), params)
	if err != nil {
		return nil, err
	}

	// Post-process (e.g., enrich with bounded context info, check guardrails).
	if s.PostProcess != nil {
		return s.PostProcess(ctx, result)
	}

	return result, nil
}

// Registry manages all registered skills.
type Registry struct {
	skills map[string]Skill
}

// NewRegistry creates an empty skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]Skill),
	}
}

// Register adds a skill to the registry.
func (r *Registry) Register(s Skill) {
	r.skills[s.Name()] = s
}

// Get returns a skill by name.
func (r *Registry) Get(name string) (Skill, bool) {
	s, ok := r.skills[name]
	return s, ok
}

// ByCategory returns all skills in a category.
func (r *Registry) ByCategory(cat Category) []Skill {
	var result []Skill
	for _, s := range r.skills {
		if s.Category() == cat {
			result = append(result, s)
		}
	}
	return result
}

// SlashCommands returns all skills that have a slash command.
func (r *Registry) SlashCommands() []Skill {
	var result []Skill
	for _, s := range r.skills {
		if s.Mode() == ModeSlashCommand || s.Mode() == ModeBoth {
			result = append(result, s)
		}
	}
	return result
}

// MCPTools returns all skills that have an MCP tool.
func (r *Registry) MCPTools() []Skill {
	var result []Skill
	for _, s := range r.skills {
		if s.Mode() == ModeMCPTool || s.Mode() == ModeBoth {
			result = append(result, s)
		}
	}
	return result
}

// MCPToolDefs returns MCP tool definitions for all MCP-capable skills.
// Used by the MCP server to advertise available tools.
func (r *Registry) MCPToolDefs() []mcp.ToolDefinition {
	var defs []mcp.ToolDefinition
	for _, s := range r.MCPTools() {
		if def := s.MCPToolDef(); def != nil {
			defs = append(defs, *def)
		}
	}
	return defs
}
