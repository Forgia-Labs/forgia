package skill

import (
	"fmt"

	"github.com/Deepzima/forgia/internal/mcp"
	"github.com/Deepzima/forgia/internal/vault"
)

// Compile-time check: *vault.FileVault satisfies VaultReader.
var _ VaultReader = (*vault.FileVault)(nil)

// RegisterCompositeSkills instantiates all composite skills and registers them
// in both the skill registry and the MCP provider registry.
// Called from "forgia mcp serve" startup (SDD-008).
//
// Skills registered (stubs — SDD-004 primitives + SDD-005 higher-level fill implementations):
//
//   Primitives (SDD-004):
//     1. blast-radius   — detect changes + guardrails check
//     2. context-lookup  — resolve bounded context for a path
//     3. guardrails-check — validate against deny rules
//     4. constitution-read — read project constitution
//
//   Higher-level (SDD-005):
//     5. impact-analysis — blast-radius + context-lookup combined
//     6. arch-validate   — architecture validation against contexts
//     7. design-assist   — constitution + architecture for design guidance
func RegisterCompositeSkills(reg *Registry, providerReg *mcp.ProviderRegistry, v VaultReader) error {
	if providerReg == nil {
		return fmt.Errorf("RegisterCompositeSkills: providerReg is required")
	}

	skills := []*CompositeSkill{
		// Primitives (SDD-004).
		{
			name:         "blast-radius",
			description:  "Detect changes and check guardrails impact",
			category:     CategoryDesign,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "detect_changes",
			registry:     providerReg,
			vault:        v,
		},
		{
			name:         "context-lookup",
			description:  "Resolve bounded context for a file path",
			category:     CategoryArchitecture,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "search_graph",
			registry:     providerReg,
			vault:        v,
		},
		{
			name:         "guardrails-check",
			description:  "Validate action against guardrails deny rules",
			category:     CategoryDesign,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "check_rules",
			registry:     providerReg,
			vault:        v,
		},
		{
			name:         "constitution-read",
			description:  "Read project constitution and rules",
			category:     CategoryDesign,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "read_context",
			registry:     providerReg,
			vault:        v,
		},

		// Higher-level (SDD-005).
		{
			name:         "impact-analysis",
			description:  "Analyze change impact across bounded contexts",
			category:     CategoryDesign,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "detect_changes",
			registry:     providerReg,
			vault:        v,
		},
		{
			name:         "arch-validate",
			description:  "Validate architecture against bounded contexts",
			category:     CategoryArchitecture,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "search_graph",
			registry:     providerReg,
			vault:        v,
		},
		{
			name:         "design-assist",
			description:  "Provide design guidance from constitution and architecture",
			category:     CategoryDesign,
			mode:         ModeMCPTool,
			ProviderName: "code",
			ProviderTool: "read_context",
			registry:     providerReg,
			vault:        v,
		},
	}

	for _, s := range skills {
		reg.Register(s)
	}

	return nil
}
