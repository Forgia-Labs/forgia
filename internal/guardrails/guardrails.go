// Package guardrails implements validation and safety checks for SDD execution.
// Reads deny.toml and enforces read/execute/write deny patterns.
// Supports three runtime modes: Careful (warn on destructive), Freeze (restrict writes),
// Guard (careful + freeze combined).
package guardrails

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Mode controls guardrail enforcement level during SDD execution.
type Mode int

const (
	// ModeOff disables runtime guardrails (deny.toml still applies).
	ModeOff Mode = iota
	// ModeCareful warns before destructive commands.
	ModeCareful
	// ModeFreeze restricts file writes to SDD boundaries.
	ModeFreeze
	// ModeGuard combines Careful + Freeze.
	ModeGuard
)

// String returns the mode name.
func (m Mode) String() string {
	switch m {
	case ModeCareful:
		return "careful"
	case ModeFreeze:
		return "freeze"
	case ModeGuard:
		return "guard"
	default:
		return "off"
	}
}

// ParseMode converts a string to a Mode.
func ParseMode(s string) Mode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "careful":
		return ModeCareful
	case "freeze":
		return ModeFreeze
	case "guard":
		return ModeGuard
	default:
		return ModeOff
	}
}

// Guardrails holds the deny lists from deny.toml.
type Guardrails struct {
	Read    DenyList `toml:"read"`
	Execute DenyList `toml:"execute"`
	Write   DenyList `toml:"write"`
}

// DenyList contains glob patterns that are denied.
// Patterns prefixed with "!" are allow-list exceptions.
type DenyList struct {
	Patterns []string `toml:"patterns"`
}

// Violation records a guardrail breach.
type Violation struct {
	Type    string // read, execute, write, boundary, secret, destructive
	Pattern string // the pattern that matched
	Target  string // the file or command that violated
}

// Error implements the error interface for Violation.
func (v Violation) Error() string {
	return fmt.Sprintf("guardrail violation [%s]: %q matched pattern %q", v.Type, v.Target, v.Pattern)
}

// destructiveCommands are commands that ModeCareful warns about.
var destructiveCommands = []string{
	"rm -rf",
	"rm -r",
	"DROP TABLE",
	"DROP DATABASE",
	"TRUNCATE",
	"DELETE FROM",
	"git push --force",
	"git push -f",
	"git reset --hard",
	"git clean -f",
	"docker rm",
	"docker rmi",
	"docker system prune",
	"kubectl delete",
	"kill -9",
	"pkill",
	"chmod 777",
}

// secretPatterns are regexes for common API keys and credentials.
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),                          // AWS access key
	regexp.MustCompile(`sk-ant-[a-zA-Z0-9\-_]{20,}`),                // Anthropic API key
	regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`),                       // OpenAI API key
	regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),                       // GitHub personal access token
	regexp.MustCompile(`gho_[a-zA-Z0-9]{36}`),                       // GitHub OAuth token
	regexp.MustCompile(`github_pat_[a-zA-Z0-9_]{22,}`),              // GitHub fine-grained PAT
	regexp.MustCompile(`glpat-[a-zA-Z0-9\-_]{20,}`),                 // GitLab PAT
	regexp.MustCompile(`xoxb-[0-9]{10,}-[0-9]{10,}-[a-zA-Z0-9]+`),  // Slack bot token
	regexp.MustCompile(`xoxp-[0-9]{10,}-[0-9]{10,}-[a-zA-Z0-9]+`),  // Slack user token
	regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`), // Private keys
	regexp.MustCompile(`(?i)password\s*[:=]\s*["'][^"']{8,}["']`),    // Hardcoded passwords
	regexp.MustCompile(`(?i)api[_-]?key\s*[:=]\s*["'][a-zA-Z0-9\-_]{16,}["']`), // Generic API keys
}

// Parse reads deny.toml from raw bytes.
func Parse(data []byte) (*Guardrails, error) {
	if len(data) == 0 {
		return &Guardrails{}, nil
	}
	var g Guardrails
	if err := toml.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("guardrails: parse deny.toml: %w", err)
	}
	return &g, nil
}

// CheckFilePaths returns violations for file paths that match [write] or [read] deny patterns.
func (g *Guardrails) CheckFilePaths(ctx context.Context, paths []string) []Violation {
	var violations []Violation
	for _, p := range paths {
		if v := matchDenyList(p, &g.Read, "read"); v != nil {
			violations = append(violations, *v)
		}
		if v := matchDenyList(p, &g.Write, "write"); v != nil {
			violations = append(violations, *v)
		}
	}
	return violations
}

// CheckWritePaths returns violations only for [write] deny patterns.
func (g *Guardrails) CheckWritePaths(ctx context.Context, paths []string) []Violation {
	var violations []Violation
	for _, p := range paths {
		if v := matchDenyList(p, &g.Write, "write"); v != nil {
			violations = append(violations, *v)
		}
	}
	return violations
}

// CheckReadPaths returns violations only for [read] deny patterns.
func (g *Guardrails) CheckReadPaths(ctx context.Context, paths []string) []Violation {
	var violations []Violation
	for _, p := range paths {
		if v := matchDenyList(p, &g.Read, "read"); v != nil {
			violations = append(violations, *v)
		}
	}
	return violations
}

// CheckCommand returns a violation if the command matches [execute] patterns.
func (g *Guardrails) CheckCommand(ctx context.Context, cmd string) *Violation {
	cmd = strings.TrimSpace(cmd)
	for _, pattern := range g.Execute.Patterns {
		if strings.HasPrefix(pattern, "!") {
			continue
		}
		// Execute patterns use prefix/glob matching.
		// "pass show*" matches "pass show secret/email"
		if matchExecutePattern(cmd, pattern) {
			return &Violation{Type: "execute", Pattern: pattern, Target: cmd}
		}
	}
	return nil
}

// CheckDestructive returns a violation if the command is destructive (for ModeCareful/ModeGuard).
func (g *Guardrails) CheckDestructive(_ context.Context, cmd string) *Violation {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, d := range destructiveCommands {
		if strings.Contains(lower, strings.ToLower(d)) {
			return &Violation{Type: "destructive", Pattern: d, Target: cmd}
		}
	}
	return nil
}

// CheckBoundaries validates that paths are within SDD write_dirs and not in forbidden_dirs.
// Used in ModeFreeze and ModeGuard.
func (g *Guardrails) CheckBoundaries(_ context.Context, paths []string, writeDirs, forbiddenDirs []string) []Violation {
	var violations []Violation

	for _, p := range paths {
		clean := filepath.Clean(p)

		// Check forbidden dirs first.
		for _, forbidden := range forbiddenDirs {
			matched, _ := filepath.Match(forbidden, clean)
			if matched || strings.HasPrefix(clean, filepath.Clean(forbidden)) {
				violations = append(violations, Violation{
					Type:    "boundary",
					Pattern: forbidden,
					Target:  p,
				})
			}
		}

		// If write_dirs is specified, path must be within at least one.
		if len(writeDirs) > 0 {
			allowed := false
			for _, dir := range writeDirs {
				if strings.HasPrefix(clean, filepath.Clean(dir)) {
					allowed = true
					break
				}
			}
			if !allowed {
				violations = append(violations, Violation{
					Type:    "boundary",
					Pattern: fmt.Sprintf("not in write_dirs: %v", writeDirs),
					Target:  p,
				})
			}
		}
	}
	return violations
}

// ScanForSecrets checks file contents for API key / credential patterns.
func (g *Guardrails) ScanForSecrets(ctx context.Context, files []string) []Violation {
	var violations []Violation
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			slog.WarnContext(ctx, "guardrails: cannot read file for secret scan", "file", f, "error", err)
			continue
		}
		content := string(data)
		for _, re := range secretPatterns {
			if loc := re.FindStringIndex(content); loc != nil {
				// Don't include the actual secret in the violation.
				violations = append(violations, Violation{
					Type:    "secret",
					Pattern: re.String(),
					Target:  f,
				})
				break // One finding per file is enough.
			}
		}
	}
	return violations
}

// Enforce runs all applicable checks for the given mode and returns all violations.
// This is the main entry point for runtime guardrails during SDD execution.
func (g *Guardrails) Enforce(ctx context.Context, mode Mode, opts EnforceOpts) []Violation {
	var all []Violation

	// deny.toml checks always apply (regardless of mode).
	if len(opts.FilePaths) > 0 {
		all = append(all, g.CheckFilePaths(ctx, opts.FilePaths)...)
	}
	if opts.Command != "" {
		if v := g.CheckCommand(ctx, opts.Command); v != nil {
			all = append(all, *v)
		}
	}
	if len(opts.ScanFiles) > 0 {
		all = append(all, g.ScanForSecrets(ctx, opts.ScanFiles)...)
	}

	// Mode-specific checks.
	if mode == ModeCareful || mode == ModeGuard {
		if opts.Command != "" {
			if v := g.CheckDestructive(ctx, opts.Command); v != nil {
				all = append(all, *v)
			}
		}
	}
	if mode == ModeFreeze || mode == ModeGuard {
		if len(opts.FilePaths) > 0 {
			all = append(all, g.CheckBoundaries(ctx, opts.FilePaths, opts.WriteDirs, opts.ForbiddenDirs)...)
		}
	}

	return all
}

// EnforceOpts holds the inputs for an Enforce call.
type EnforceOpts struct {
	Command      string   // command being executed (for execute + destructive checks)
	FilePaths    []string // file paths being accessed (for read/write + boundary checks)
	ScanFiles    []string // files to scan for secrets
	WriteDirs    []string // allowed write directories (from SDD boundaries)
	ForbiddenDirs []string // forbidden directories (from SDD boundaries)
}

// matchDenyList checks a path against a deny list, respecting ! exceptions.
func matchDenyList(path string, dl *DenyList, violationType string) *Violation {
	denied := false
	matchedPattern := ""

	for _, pattern := range dl.Patterns {
		// Exception pattern: "!**/.env.example" undoes a previous deny.
		if strings.HasPrefix(pattern, "!") {
			exception := pattern[1:]
			if matchGlob(path, exception) {
				denied = false
			}
			continue
		}

		if matchGlob(path, pattern) {
			denied = true
			matchedPattern = pattern
		}
	}

	if denied {
		return &Violation{Type: violationType, Pattern: matchedPattern, Target: path}
	}
	return nil
}

// matchGlob matches a path against a glob pattern.
// Supports ** for recursive directory matching.
func matchGlob(path, pattern string) bool {
	// Normalize separators.
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	// Handle ** patterns by converting to a more flexible match.
	if strings.Contains(pattern, "**") {
		// "**/*.pem" should match "foo/bar/baz.pem" and "baz.pem"
		// Convert ** to regex-like matching.
		return matchDoubleStarGlob(path, pattern)
	}

	// Simple filepath.Match for non-** patterns.
	matched, _ := filepath.Match(pattern, path)
	if matched {
		return true
	}

	// Also try matching against just the filename.
	matched, _ = filepath.Match(pattern, filepath.Base(path))
	return matched
}

// matchDoubleStarGlob handles ** glob patterns.
func matchDoubleStarGlob(path, pattern string) bool {
	// Split pattern on **
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		// Multiple ** — fall back to simple check.
		return strings.Contains(path, strings.ReplaceAll(pattern, "**", ""))
	}

	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")

	// If no prefix, just match the suffix against the path and all subpaths.
	if prefix == "" {
		// "**/*.pem" — check if any path segment + suffix matches.
		if suffix == "" {
			return true
		}
		// Match suffix against the full path and the basename.
		matched, _ := filepath.Match(suffix, filepath.ToSlash(path))
		if matched {
			return true
		}
		matched, _ = filepath.Match(suffix, filepath.Base(path))
		if matched {
			return true
		}
		// Check each possible subpath.
		segments := strings.Split(path, "/")
		for i := range segments {
			subpath := strings.Join(segments[i:], "/")
			matched, _ = filepath.Match(suffix, subpath)
			if matched {
				return true
			}
		}
		return false
	}

	// "foo/**/*.pem" — prefix must match, then suffix against remainder.
	if !strings.HasPrefix(path, prefix+"/") && path != prefix {
		return false
	}
	remainder := strings.TrimPrefix(path, prefix+"/")
	if suffix == "" {
		return true
	}
	matched, _ := filepath.Match(suffix, remainder)
	if matched {
		return true
	}
	matched, _ = filepath.Match(suffix, filepath.Base(remainder))
	return matched
}

// matchExecutePattern matches a command against an execute deny pattern.
// Patterns ending in * are prefix matches. Others are exact.
func matchExecutePattern(cmd, pattern string) bool {
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(cmd, prefix)
	}
	return cmd == pattern
}
