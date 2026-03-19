// Package guardrails implements validation and safety checks for SDD execution.
// Reads deny.toml and enforces read/execute/write deny patterns.
package guardrails

import "context"

// Guardrails holds the deny lists from deny.toml.
type Guardrails struct {
	Read    DenyList `toml:"read"`
	Execute DenyList `toml:"execute"`
	Write   DenyList `toml:"write"`
}

// DenyList contains glob patterns that are denied.
type DenyList struct {
	Patterns []string `toml:"patterns"`
}

// Violation records a guardrail breach.
type Violation struct {
	Type    string // read, execute, write
	Pattern string // the pattern that matched
	Target  string // the file or command that violated
}

// CheckFilePaths returns violations for file paths that match [write] or [read] glob patterns.
// This checks paths only — use ScanForSecrets to check file contents.
func (g *Guardrails) CheckFilePaths(ctx context.Context, paths []string) []Violation {
	// TODO: implement glob matching
	return nil
}

// CheckCommand returns a violation if the command matches [execute] patterns.
func (g *Guardrails) CheckCommand(ctx context.Context, cmd string) *Violation {
	// TODO: implement pattern matching
	return nil
}

// ScanForSecrets checks file contents for API key / credential patterns.
func (g *Guardrails) ScanForSecrets(ctx context.Context, files []string) []Violation {
	// TODO: implement regex scanning (AKIA, sk-ant-, ghp_, etc.)
	return nil
}

// Parse reads deny.toml from raw bytes.
// Vault provides the raw bytes via GuardrailsRaw(), this package parses them.
func Parse(data []byte) (*Guardrails, error) {
	// TODO: implement TOML parsing
	return &Guardrails{}, nil
}
