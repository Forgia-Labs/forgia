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

	// globCache holds pre-compiled regexes for ** glob patterns (populated by Parse).
	globCache map[string]*regexp.Regexp
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

// Parse reads deny.toml from raw bytes and pre-compiles glob regexes.
func Parse(data []byte) (*Guardrails, error) {
	if len(data) == 0 {
		return &Guardrails{globCache: make(map[string]*regexp.Regexp)}, nil
	}
	var g Guardrails
	if err := toml.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("guardrails: parse deny.toml: %w", err)
	}
	// Pre-compile glob regexes for all ** patterns.
	g.globCache = make(map[string]*regexp.Regexp)
	for _, patterns := range [][]string{g.Read.Patterns, g.Write.Patterns} {
		for _, p := range patterns {
			clean := strings.TrimPrefix(p, "!")
			if strings.Contains(clean, "**") {
				g.globCache[clean] = globToRegex(filepath.ToSlash(clean))
			}
		}
	}
	return &g, nil
}

// CheckFilePaths returns violations for file paths that match [write] or [read] deny patterns.
func (g *Guardrails) CheckFilePaths(ctx context.Context, paths []string) []Violation {
	var violations []Violation
	for _, p := range paths {
		if v := matchDenyList(p, &g.Read, "read", g.globCache); v != nil {
			violations = append(violations, *v)
		}
		if v := matchDenyList(p, &g.Write, "write", g.globCache); v != nil {
			violations = append(violations, *v)
		}
	}
	return violations
}

// CheckWritePaths returns violations only for [write] deny patterns.
func (g *Guardrails) CheckWritePaths(ctx context.Context, paths []string) []Violation {
	var violations []Violation
	for _, p := range paths {
		if v := matchDenyList(p, &g.Write, "write", g.globCache); v != nil {
			violations = append(violations, *v)
		}
	}
	return violations
}

// CheckReadPaths returns violations only for [read] deny patterns.
func (g *Guardrails) CheckReadPaths(ctx context.Context, paths []string) []Violation {
	var violations []Violation
	for _, p := range paths {
		if v := matchDenyList(p, &g.Read, "read", g.globCache); v != nil {
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
			cleanForbidden := filepath.Clean(forbidden)
			matched, _ := filepath.Match(cleanForbidden, clean)
			// Append separator to prevent "src" matching "src2/" (sibling bypass).
			if matched || clean == cleanForbidden || strings.HasPrefix(clean, cleanForbidden+string(filepath.Separator)) {
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
				cleanDir := filepath.Clean(dir)
				// Append separator to prevent "src" matching "src2/" (sibling bypass).
				if clean == cleanDir || strings.HasPrefix(clean, cleanDir+string(filepath.Separator)) {
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
	const maxScanSize = 1 << 20 // 1 MB — skip large/binary files
	var violations []Violation
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			slog.WarnContext(ctx, "guardrails: cannot stat file for secret scan", "file", f, "error", err)
			continue
		}
		if info.Size() > maxScanSize {
			slog.InfoContext(ctx, "guardrails: skipping large file for secret scan", "file", f, "size", info.Size())
			continue
		}
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
// cache is a map of pre-compiled regexes for ** patterns (may be nil).
func matchDenyList(path string, dl *DenyList, violationType string, cache map[string]*regexp.Regexp) *Violation {
	denied := false
	matchedPattern := ""

	for _, pattern := range dl.Patterns {
		// Exception pattern: "!**/.env.example" undoes a previous deny.
		if strings.HasPrefix(pattern, "!") {
			exception := pattern[1:]
			if matchGlob(path, exception, cache) {
				denied = false
			}
			continue
		}

		if matchGlob(path, pattern, cache) {
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
// Supports ** for recursive directory matching. cache is optional (may be nil).
func matchGlob(path, pattern string, cache map[string]*regexp.Regexp) bool {
	// Normalize separators.
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	// Handle ** patterns — use cached regex if available.
	if strings.Contains(pattern, "**") {
		if cache != nil {
			if re, ok := cache[pattern]; ok {
				return re.MatchString(path)
			}
		}
		// Fallback: compile on the fly (for patterns not in cache).
		return globToRegex(pattern).MatchString(path)
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

// globToRegex converts a glob pattern with ** to a compiled regex.
// ** matches zero or more path segments, * matches within a segment.
func globToRegex(pattern string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")

	i := 0
	for i < len(pattern) {
		if i+1 < len(pattern) && pattern[i] == '*' && pattern[i+1] == '*' {
			// ** matches any number of path segments (including zero).
			// Consume trailing slash if present.
			b.WriteString(".*")
			i += 2
			if i < len(pattern) && pattern[i] == '/' {
				i++
			}
			continue
		}
		switch pattern[i] {
		case '*':
			b.WriteString("[^/]*") // * matches within a single segment
		case '?':
			b.WriteString("[^/]")
		case '.':
			b.WriteString(`\.`)
		case '/':
			b.WriteString("/")
		default:
			b.WriteString(regexp.QuoteMeta(string(pattern[i])))
		}
		i++
	}

	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		// Fallback: literal substring check.
		return regexp.MustCompile(regexp.QuoteMeta(pattern))
	}
	return re
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
