package guardrails

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

var ctx = context.Background()

// --- Parse ---

func TestParseValidDenyToml(t *testing.T) {
	data := []byte(`
[read]
patterns = ["**/.env", "**/*.pem"]

[execute]
patterns = ["pass show*", "gpg --export-secret*"]

[write]
patterns = ["**/.env", ".forgia/constitution.md"]
`)
	g, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(g.Read.Patterns) != 2 {
		t.Errorf("expected 2 read patterns, got %d", len(g.Read.Patterns))
	}
	if len(g.Execute.Patterns) != 2 {
		t.Errorf("expected 2 execute patterns, got %d", len(g.Execute.Patterns))
	}
	if len(g.Write.Patterns) != 2 {
		t.Errorf("expected 2 write patterns, got %d", len(g.Write.Patterns))
	}
}

func TestParseEmpty(t *testing.T) {
	g, err := Parse([]byte{})
	if err != nil {
		t.Fatalf("Parse empty failed: %v", err)
	}
	if len(g.Read.Patterns) != 0 {
		t.Errorf("expected 0 read patterns, got %d", len(g.Read.Patterns))
	}
}

func TestParseInvalidToml(t *testing.T) {
	_, err := Parse([]byte(`this is not valid toml [[[`))
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

// --- CheckFilePaths ---

func TestCheckFilePathsDenied(t *testing.T) {
	g := &Guardrails{
		Read:  DenyList{Patterns: []string{"**/.env", "**/*.pem"}},
		Write: DenyList{Patterns: []string{"**/.env", ".forgia/constitution.md"}},
	}

	violations := g.CheckFilePaths(ctx, []string{
		"config/.env",
		"src/main.go",
		"certs/server.pem",
	})

	// .env matches read + write (2), server.pem matches read (1) = 3
	if len(violations) != 3 {
		t.Fatalf("expected 3 violations, got %d: %v", len(violations), violations)
	}
}

func TestCheckFilePathsAllowed(t *testing.T) {
	g := &Guardrails{
		Read:  DenyList{Patterns: []string{"**/.env"}},
		Write: DenyList{Patterns: []string{"**/.env"}},
	}

	violations := g.CheckFilePaths(ctx, []string{"src/main.go", "README.md"})
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(violations))
	}
}

func TestCheckFilePathsException(t *testing.T) {
	g := &Guardrails{
		Read: DenyList{Patterns: []string{"**/.env", "**/.env.*", "!**/.env.example"}},
	}

	// .env.example should be allowed (exception), .env.local should be denied
	violations := g.CheckReadPaths(ctx, []string{".env.example", ".env.local"})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if violations[0].Target != ".env.local" {
		t.Errorf("expected violation on .env.local, got %s", violations[0].Target)
	}
}

func TestCheckFilePathsConstitution(t *testing.T) {
	g := &Guardrails{
		Write: DenyList{Patterns: []string{".forgia/constitution.md", ".forgia/config.toml"}},
	}

	violations := g.CheckWritePaths(ctx, []string{".forgia/constitution.md"})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
}

// --- CheckCommand ---

func TestCheckCommandDenied(t *testing.T) {
	g := &Guardrails{
		Execute: DenyList{Patterns: []string{"pass show*", "gpg --export-secret*", "cat ~/.ssh/id_*"}},
	}

	tests := []struct {
		cmd    string
		denied bool
	}{
		{"pass show secret/email", true},
		{"pass show", true},
		{"gpg --export-secret-keys", true},
		{"cat ~/.ssh/id_rsa", true},
		{"cat ~/.ssh/id_ed25519", true},
		{"ls -la", false},
		{"go build ./...", false},
		{"git push origin main", false},
	}

	for _, tt := range tests {
		v := g.CheckCommand(ctx, tt.cmd)
		if tt.denied && v == nil {
			t.Errorf("expected %q to be denied", tt.cmd)
		}
		if !tt.denied && v != nil {
			t.Errorf("expected %q to be allowed, got violation: %v", tt.cmd, v)
		}
	}
}

// --- CheckDestructive ---

func TestCheckDestructive(t *testing.T) {
	g := &Guardrails{}

	tests := []struct {
		cmd         string
		destructive bool
	}{
		{"rm -rf /tmp/test", true},
		{"rm -r build/", true},
		{"git push --force", true},
		{"git push -f origin main", true},
		{"git reset --hard HEAD~1", true},
		{"docker rm container123", true},
		{"kubectl delete pod foo", true},
		{"chmod 777 /tmp/test", true},
		{"git push origin main", false},
		{"go test ./...", false},
		{"rm single-file.txt", false},
		{"docker ps", false},
	}

	for _, tt := range tests {
		v := g.CheckDestructive(ctx, tt.cmd)
		if tt.destructive && v == nil {
			t.Errorf("expected %q to be destructive", tt.cmd)
		}
		if !tt.destructive && v != nil {
			t.Errorf("expected %q to be safe, got: %v", tt.cmd, v)
		}
	}
}

// --- CheckBoundaries ---

func TestCheckBoundariesAllowed(t *testing.T) {
	g := &Guardrails{}

	violations := g.CheckBoundaries(ctx,
		[]string{"src/main.go", "src/lib/utils.go"},
		[]string{"src/"},
		nil,
	)
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations, got %d: %v", len(violations), violations)
	}
}

func TestCheckBoundariesOutsideWriteDirs(t *testing.T) {
	g := &Guardrails{}

	violations := g.CheckBoundaries(ctx,
		[]string{"src/main.go", "docs/README.md"},
		[]string{"src/"},
		nil,
	)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if violations[0].Target != "docs/README.md" {
		t.Errorf("expected violation on docs/README.md, got %s", violations[0].Target)
	}
}

func TestCheckBoundariesForbiddenDir(t *testing.T) {
	g := &Guardrails{}

	violations := g.CheckBoundaries(ctx,
		[]string{"src/main.go", ".forgia/config.toml"},
		nil, // no write_dirs restriction
		[]string{".forgia/"},
	)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if violations[0].Target != ".forgia/config.toml" {
		t.Errorf("expected violation on .forgia/config.toml, got %s", violations[0].Target)
	}
}

func TestCheckBoundariesSiblingBypass(t *testing.T) {
	g := &Guardrails{}

	// "src2/main.go" must NOT be allowed when write_dirs is ["src/"]
	violations := g.CheckBoundaries(ctx,
		[]string{"src/main.go", "src2/main.go"},
		[]string{"src/"},
		nil,
	)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation (src2 not allowed), got %d: %v", len(violations), violations)
	}
	if violations[0].Target != "src2/main.go" {
		t.Errorf("expected violation on src2/main.go, got %s", violations[0].Target)
	}

	// Same for forbidden: ".forgia" must not match ".forgia2/"
	violations = g.CheckBoundaries(ctx,
		[]string{".forgia/config.toml", ".forgia2/data.txt"},
		nil,
		[]string{".forgia"},
	)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation (.forgia only), got %d: %v", len(violations), violations)
	}
	if violations[0].Target != ".forgia/config.toml" {
		t.Errorf("expected violation on .forgia/config.toml, got %s", violations[0].Target)
	}
}

func TestCheckBoundariesNoRestrictions(t *testing.T) {
	g := &Guardrails{}

	violations := g.CheckBoundaries(ctx,
		[]string{"anywhere/file.go"},
		nil,
		nil,
	)
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations with no restrictions, got %d", len(violations))
	}
}

// --- ScanForSecrets ---

func TestScanForSecretsFindsKeys(t *testing.T) {
	g := &Guardrails{}
	dir := t.TempDir()

	// File with AWS key
	awsFile := filepath.Join(dir, "config.py")
	os.WriteFile(awsFile, []byte(`aws_key = "AKIAIOSFODNN7EXAMPLE"`), 0o644)

	// File with OpenAI key
	openaiFile := filepath.Join(dir, "app.js")
	os.WriteFile(openaiFile, []byte(`const key = "sk-abcdefghijklmnopqrstuvwxyz123456"`), 0o644)

	// Clean file
	cleanFile := filepath.Join(dir, "main.go")
	os.WriteFile(cleanFile, []byte(`package main; func main() {}`), 0o644)

	violations := g.ScanForSecrets(ctx, []string{awsFile, openaiFile, cleanFile})
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %v", len(violations), violations)
	}

	targets := map[string]bool{}
	for _, v := range violations {
		targets[v.Target] = true
		if v.Type != "secret" {
			t.Errorf("expected type 'secret', got %q", v.Type)
		}
	}
	if !targets[awsFile] {
		t.Error("expected AWS key file to be flagged")
	}
	if !targets[openaiFile] {
		t.Error("expected OpenAI key file to be flagged")
	}
}

func TestScanForSecretsPrivateKey(t *testing.T) {
	g := &Guardrails{}
	dir := t.TempDir()

	keyFile := filepath.Join(dir, "key.pem")
	os.WriteFile(keyFile, []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----"), 0o644)

	violations := g.ScanForSecrets(ctx, []string{keyFile})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
}

func TestScanForSecretsCleanFile(t *testing.T) {
	g := &Guardrails{}
	dir := t.TempDir()

	cleanFile := filepath.Join(dir, "main.go")
	os.WriteFile(cleanFile, []byte(`package main; func main() { println("hello") }`), 0o644)

	violations := g.ScanForSecrets(ctx, []string{cleanFile})
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(violations))
	}
}

func TestScanForSecretsMissingFile(t *testing.T) {
	g := &Guardrails{}
	// Should not panic on missing file, just skip with warning.
	violations := g.ScanForSecrets(ctx, []string{"/nonexistent/file.go"})
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations for missing file, got %d", len(violations))
	}
}

// --- Enforce ---

func TestEnforceGuardMode(t *testing.T) {
	g := &Guardrails{
		Write:   DenyList{Patterns: []string{"**/.env"}},
		Execute: DenyList{Patterns: []string{"pass show*"}},
	}

	violations := g.Enforce(ctx, ModeGuard, EnforceOpts{
		Command:      "rm -rf /tmp",
		FilePaths:    []string{"src/main.go", "docs/readme.md"},
		WriteDirs:    []string{"src/"},
		ForbiddenDirs: []string{},
	})

	// Expected: "rm -rf" is destructive (guard mode) + docs/readme.md outside write_dirs (freeze)
	found := map[string]bool{}
	for _, v := range violations {
		found[v.Type] = true
	}
	if !found["destructive"] {
		t.Error("expected destructive violation for rm -rf")
	}
	if !found["boundary"] {
		t.Error("expected boundary violation for docs/readme.md")
	}
}

func TestEnforceCarefulModeNoFreeze(t *testing.T) {
	g := &Guardrails{}

	violations := g.Enforce(ctx, ModeCareful, EnforceOpts{
		Command:   "rm -rf /tmp",
		FilePaths: []string{"anywhere/file.go"},
		WriteDirs: []string{"src/"}, // careful mode ignores write_dirs
	})

	// Only destructive, no boundary check.
	for _, v := range violations {
		if v.Type == "boundary" {
			t.Error("careful mode should not check boundaries")
		}
	}
}

func TestEnforceOffMode(t *testing.T) {
	g := &Guardrails{
		Execute: DenyList{Patterns: []string{"pass show*"}},
	}

	violations := g.Enforce(ctx, ModeOff, EnforceOpts{
		Command:   "rm -rf /tmp",
		FilePaths: []string{"anywhere/file.go"},
		WriteDirs: []string{"src/"},
	})

	// deny.toml still applies ("pass show" would be caught), but rm -rf is not in execute patterns.
	// No destructive check (mode off), no boundary check (mode off).
	for _, v := range violations {
		if v.Type == "destructive" || v.Type == "boundary" {
			t.Errorf("mode off should not check %s", v.Type)
		}
	}
}

// --- Mode ---

func TestParseMode(t *testing.T) {
	tests := []struct {
		input    string
		expected Mode
	}{
		{"careful", ModeCareful},
		{"Careful", ModeCareful},
		{"freeze", ModeFreeze},
		{"guard", ModeGuard},
		{"GUARD", ModeGuard},
		{"off", ModeOff},
		{"", ModeOff},
		{"unknown", ModeOff},
	}
	for _, tt := range tests {
		got := ParseMode(tt.input)
		if got != tt.expected {
			t.Errorf("ParseMode(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestModeString(t *testing.T) {
	if ModeGuard.String() != "guard" {
		t.Errorf("expected 'guard', got %q", ModeGuard.String())
	}
	if ModeOff.String() != "off" {
		t.Errorf("expected 'off', got %q", ModeOff.String())
	}
}

// --- Glob matching ---

func TestMatchGlobDoublestar(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		match   bool
	}{
		// Single ** prefix
		{"config/.env", "**/.env", true},
		{"deep/nested/.env", "**/.env", true},
		{".env", "**/.env", true},
		{"certs/server.pem", "**/*.pem", true},
		{"deep/nested/cert.pem", "**/*.pem", true},
		{"main.go", "**/*.pem", false},
		{".env.example", "**/.env.example", true},
		{".env.local", "**/.env.*", true},
		{".ssh/id_rsa", "**/.ssh/id_*", true},

		// Multiple ** (e.g. "**/.azure/**")
		{".azure/config", "**/.azure/**", true},
		{"home/.azure/credentials", "**/.azure/**", true},
		{"deep/nested/.azure/tokens/access.json", "**/.azure/**", true},
		{"flask-aws-utils/config.py", "**/.azure/**", false},
		{".gcloud/key.json", "**/.gcloud/**", true},

		// Prefix ** + suffix
		{".aws/credentials", "**/.aws/credentials", true},
		{"home/.aws/credentials", "**/.aws/credentials", true},
		{".aws/config", "**/.aws/credentials", false},
	}

	for _, tt := range tests {
		got := matchGlob(tt.path, tt.pattern, nil)
		if got != tt.match {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.match)
		}
	}
}

// --- Full deny.toml integration ---

func TestFullDenyTomlIntegration(t *testing.T) {
	data := []byte(`
[read]
patterns = [
  "**/.env",
  "**/.env.*",
  "!**/.env.example",
  "**/*.pem",
  "**/.ssh/id_*",
]

[execute]
patterns = [
  "pass show*",
  "gpg --export-secret*",
  "cat ~/.ssh/id_*",
]

[write]
patterns = [
  "**/.env",
  ".forgia/constitution.md",
  "**/*.pem",
]
`)
	g, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Read check
	readV := g.CheckReadPaths(ctx, []string{".env", ".env.example", ".env.local", "server.pem", "main.go"})
	// .env (denied), .env.example (exception = allowed), .env.local (denied), server.pem (denied), main.go (allowed)
	if len(readV) != 3 {
		t.Errorf("expected 3 read violations, got %d: %v", len(readV), readV)
	}

	// Write check
	writeV := g.CheckWritePaths(ctx, []string{".forgia/constitution.md", "src/main.go"})
	if len(writeV) != 1 {
		t.Errorf("expected 1 write violation, got %d: %v", len(writeV), writeV)
	}

	// Execute check
	if v := g.CheckCommand(ctx, "pass show secret/email"); v == nil {
		t.Error("expected 'pass show' to be denied")
	}
	if v := g.CheckCommand(ctx, "go build ./..."); v != nil {
		t.Error("expected 'go build' to be allowed")
	}
}
