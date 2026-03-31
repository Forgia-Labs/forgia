package skill

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/forgia-labs/forgia/internal/mcp"
)

// EmbeddedSkill wraps a markdown slash command from the embedded filesystem.
type EmbeddedSkill struct {
	name        string
	description string
	content     []byte
}

// NewEmbeddedSkill creates a skill from embedded markdown content.
func NewEmbeddedSkill(name string, content []byte) *EmbeddedSkill {
	// Extract first non-empty line as description.
	desc := name
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") {
			desc = line
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			break
		}
	}

	return &EmbeddedSkill{
		name:        name,
		description: desc,
		content:     content,
	}
}

func (s *EmbeddedSkill) Name() string            { return s.name }
func (s *EmbeddedSkill) Description() string      { return s.description }
func (s *EmbeddedSkill) Category() Category       { return inferCategory(s.name) }
func (s *EmbeddedSkill) Mode() Mode               { return ModeSlashCommand }
func (s *EmbeddedSkill) SlashCommandFile() string  { return s.name + ".md" }
func (s *EmbeddedSkill) MCPToolDef() *mcp.ToolDefinition { return nil }

// Content returns the raw markdown with $ARGUMENTS replaced.
func (s *EmbeddedSkill) Content(args string) string {
	content := string(s.content)
	return strings.ReplaceAll(content, "$ARGUMENTS", args)
}

// Execute is a no-op for embedded skills — they're invoked via Claude CLI, not Go logic.
func (s *EmbeddedSkill) Execute(_ context.Context, _ map[string]any) (any, error) {
	return nil, fmt.Errorf("embedded skill %q must be invoked via Claude CLI, not Execute()", s.name)
}

// LoadEmbedded populates the registry from an embedded filesystem.
// Expects .md files at root level (e.g., "fd-new.md").
func (r *Registry) LoadEmbedded(cmdFS fs.FS) error {
	return fs.WalkDir(cmdFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		content, err := fs.ReadFile(cmdFS, path)
		if err != nil {
			return fmt.Errorf("read embedded command %s: %w", path, err)
		}

		name := strings.TrimSuffix(path, ".md")
		r.Register(NewEmbeddedSkill(name, content))
		return nil
	})
}

// All returns all registered skills sorted by name.
func (r *Registry) All() []Skill {
	var result []Skill
	for _, s := range r.skills {
		result = append(result, s)
	}
	return result
}

// inferCategory guesses category from skill name prefix.
func inferCategory(name string) Category {
	switch {
	case strings.HasPrefix(name, "fd-"):
		return CategoryFD
	case strings.HasPrefix(name, "sdd-"):
		return CategorySDD
	case strings.HasPrefix(name, "arch-"):
		return CategoryArchitecture
	default:
		return CategoryFD
	}
}
