package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgia-labs/forgia/internal/guardrails"
	"github.com/forgia-labs/forgia/internal/vault"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <sdd-file | FD-NNN>",
	Short: "Validate an SDD before execution",
	Long:  "Checks frontmatter, required sections, guardrails, and boundaries.",
	Args:  cobra.ExactArgs(1),
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

// ValidationError represents a single validation failure.
type ValidationError struct {
	Category string // frontmatter, section, guardrail, boundary
	Message  string
}

func (e ValidationError) String() string {
	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

func runValidate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	target := args[0]

	v, err := vault.Open(".")
	if err != nil {
		return fmt.Errorf("vault: %w", err)
	}

	// Load guardrails.
	guardrailsData, err := v.GuardrailsRaw(ctx)
	if err != nil {
		slog.WarnContext(ctx, "guardrails not available", "error", err)
	}
	g, _ := guardrails.Parse(guardrailsData)
	if g == nil {
		g = &guardrails.Guardrails{}
	}

	// Determine mode: single file or FD directory.
	var sddFiles []string
	if strings.HasPrefix(target, "FD-") {
		// FD directory mode.
		sddDir := filepath.Join(v.Dir(), "sdd", target)
		entries, err := os.ReadDir(sddDir)
		if err != nil {
			return fmt.Errorf("read SDD directory %s: %w", target, err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && strings.HasPrefix(e.Name(), "SDD-") {
				sddFiles = append(sddFiles, filepath.Join(sddDir, e.Name()))
			}
		}
		if len(sddFiles) == 0 {
			return fmt.Errorf("no SDD files found in %s", sddDir)
		}
	} else {
		// Single file mode.
		sddFiles = []string{target}
	}

	// Validate each SDD.
	totalErrors := 0
	for _, f := range sddFiles {
		errors := validateSDD(ctx, v, g, f)
		name := filepath.Base(f)
		if len(errors) == 0 {
			fmt.Printf("✓ %s — valid\n", name)
		} else {
			fmt.Printf("✗ %s — %d error(s):\n", name, len(errors))
			for _, e := range errors {
				fmt.Printf("    %s\n", e)
			}
			totalErrors += len(errors)
		}
	}

	if totalErrors > 0 {
		return fmt.Errorf("validation failed: %d error(s) in %d file(s)", totalErrors, len(sddFiles))
	}
	return nil
}

// validateSDD checks a single SDD file and returns all errors found.
func validateSDD(ctx context.Context, v vault.Vault, g *guardrails.Guardrails, path string) []ValidationError {
	var errors []ValidationError

	data, err := os.ReadFile(path)
	if err != nil {
		return []ValidationError{{Category: "file", Message: fmt.Sprintf("cannot read: %v", err)}}
	}

	content := string(data)

	// Check frontmatter.
	if !strings.Contains(content, "---") {
		errors = append(errors, ValidationError{"frontmatter", "missing YAML frontmatter delimiters (---)"})
		return errors // Can't proceed without frontmatter.
	}

	requiredFields := []string{"id:", "fd:", "title:", "status:"}
	for _, field := range requiredFields {
		if !containsFrontmatterField(content, field) {
			errors = append(errors, ValidationError{"frontmatter", fmt.Sprintf("missing required field: %s", field)})
		}
	}

	// Check required sections.
	requiredSections := []string{
		"## Scope",
		"## Interfaces",
		"## Constraints",
		"## Test Requirements",
		"## Acceptance Criteria",
		"## Context",
		"## Constitution Check",
		"## Work Log",
	}
	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			errors = append(errors, ValidationError{"section", fmt.Sprintf("missing required section: %s", section)})
		}
	}

	// Check parent FD exists.
	fdID := extractFrontmatterValue(content, "fd:")
	if fdID != "" {
		fdID = strings.Trim(fdID, "\"' ")
		fds, _ := v.ListFDs(ctx)
		found := false
		for _, fd := range fds {
			if fd.ID == fdID {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, ValidationError{"reference", fmt.Sprintf("parent FD %q not found in vault", fdID)})
		}
	}

	// Guardrails check on context files.
	if g != nil {
		contextPaths := extractContextPaths(content)
		if len(contextPaths) > 0 {
			violations := g.CheckFilePaths(ctx, contextPaths)
			for _, v := range violations {
				errors = append(errors, ValidationError{"guardrail", v.Error()})
			}
		}
	}

	return errors
}

// containsFrontmatterField checks if a field exists in the frontmatter section.
func containsFrontmatterField(content, field string) bool {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return false
	}
	frontmatter := parts[1]
	return strings.Contains(frontmatter, field)
}

// extractFrontmatterValue extracts a value from a frontmatter field.
func extractFrontmatterValue(content, field string) string {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return ""
	}
	for _, line := range strings.Split(parts[1], "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, field) {
			return strings.TrimSpace(strings.TrimPrefix(line, field))
		}
	}
	return ""
}

// extractContextPaths extracts file paths from the Context section.
func extractContextPaths(content string) []string {
	idx := strings.Index(content, "## Context")
	if idx < 0 {
		return nil
	}
	section := content[idx:]
	// Find next section.
	nextSection := strings.Index(section[1:], "\n## ")
	if nextSection > 0 {
		section = section[:nextSection+1]
	}

	var paths []string
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		// Extract paths from "- [ ] `path/to/file`" format.
		if strings.Contains(line, "`") {
			start := strings.Index(line, "`")
			end := strings.LastIndex(line, "`")
			if start < end {
				p := line[start+1 : end]
				if !strings.HasPrefix(p, "http") && !strings.Contains(p, " ") {
					paths = append(paths, p)
				}
			}
		}
	}
	return paths
}
