package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath holds the path to the compiled forgia binary.
var binaryPath string

func TestMain(m *testing.M) {
	root := findProjectRoot()

	dir, err := os.MkdirTemp("", "forgia-e2e-build-*")
	if err != nil {
		panic("create temp build dir: " + err.Error())
	}

	binaryPath = filepath.Join(dir, "forgia")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/forgia/")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		panic("build failed: " + string(out))
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic("getwd: " + err.Error())
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("cannot find project root (no go.mod)")
		}
		dir = parent
	}
}

// runForgia executes the forgia binary in the given directory and returns
// combined stdout+stderr output and the exit code.
func runForgia(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	// Suppress beads/knowledge noise by clearing env hints.
	cmd.Env = append(os.Environ(), "HOME="+dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(out), exitErr.ExitCode()
		}
		t.Fatalf("exec %v failed: %v", args, err)
	}
	return string(out), 0
}

// initVault runs forgia init and fails the test on error.
func initVault(t *testing.T, dir string) {
	t.Helper()
	_, code := runForgia(t, dir, "init")
	if code != 0 {
		t.Fatalf("forgia init failed with exit code %d", code)
	}
}

func assertContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("output does not contain %q\noutput:\n%s", substr, truncate(output, 500))
	}
}

func assertExitCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("exit code = %d, want %d", got, want)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("expected directory %s: %v", path, err)
		return
	}
	if !info.IsDir() {
		t.Errorf("%s is not a directory", path)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("expected file %s: %v", path, err)
		return
	}
	if info.IsDir() {
		t.Errorf("%s is a directory, expected file", path)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}

// writeFD creates a fixture FD file in the vault.
func writeFD(t *testing.T, dir, id, title, status string) {
	t.Helper()
	fdDir := filepath.Join(dir, ".forgia", "fd")
	os.MkdirAll(fdDir, 0o755)
	content := "---\nid: " + id + "\ntitle: \"" + title + "\"\nstatus: " + status + "\n---\n# " + id + "\n"
	if err := os.WriteFile(filepath.Join(fdDir, id+"-test.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write FD fixture: %v", err)
	}
}

// writeSDD creates a fixture SDD file in the vault.
func writeSDD(t *testing.T, dir, fdID, sddID, title, status string) {
	t.Helper()
	sddDir := filepath.Join(dir, ".forgia", "sdd", fdID)
	os.MkdirAll(sddDir, 0o755)
	content := "---\nid: " + sddID + "\nfd: " + fdID + "\ntitle: \"" + title + "\"\nstatus: " + status + "\n---\n# " + sddID + "\n"
	if err := os.WriteFile(filepath.Join(sddDir, sddID+"-test.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write SDD fixture: %v", err)
	}
}

// writeValidSDD creates a complete, valid SDD file that passes validation.
func writeValidSDD(t *testing.T, dir, fdID string) string {
	t.Helper()
	sddDir := filepath.Join(dir, ".forgia", "sdd", fdID)
	os.MkdirAll(sddDir, 0o755)
	path := filepath.Join(sddDir, "SDD-001-valid.md")
	content := `---
id: "SDD-001"
fd: "` + fdID + `"
title: "Valid SDD"
status: planned
---

# SDD-001: Valid SDD

> Parent FD: [[` + fdID + `]]

## Scope
Build it.

## Interfaces / Interfacce
| Interface | Type | Description |
|-----------|------|-------------|
| API | REST | Main endpoint |

## Constraints / Vincoli
- Language: Go

## Test Requirements
| Type | What | Coverage |
|------|------|----------|
| Unit | Core | 80% |

## Acceptance Criteria / Criteri di Accettazione
- [ ] Works correctly

## Context / Contesto
- [ ] ` + "`README.md`" + `

## Constitution Check
- [x] Respects code standards

## Work Log / Diario di Lavoro
Pending.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write valid SDD: %v", err)
	}
	return path
}

// writeInvalidSDD creates an SDD missing required sections.
func writeInvalidSDD(t *testing.T, dir, fdID string) string {
	t.Helper()
	sddDir := filepath.Join(dir, ".forgia", "sdd", fdID)
	os.MkdirAll(sddDir, 0o755)
	path := filepath.Join(sddDir, "SDD-002-invalid.md")
	content := `---
id: "SDD-002"
fd: "` + fdID + `"
title: "Invalid SDD"
status: planned
---

# SDD-002: Missing sections

## Scope
Only scope, nothing else.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write invalid SDD: %v", err)
	}
	return path
}

func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("read %s: %v", path, err)
		return
	}
	if !strings.Contains(string(data), substr) {
		t.Errorf("file %s does not contain %q", filepath.Base(path), substr)
	}
}

func assertFileNotOverwritten(t *testing.T, path, marker string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("read %s: %v", path, err)
		return
	}
	if string(data) != marker {
		t.Errorf("file %s was overwritten: got %q, want %q", filepath.Base(path), string(data), marker)
	}
}

// writeFDWithAuthor creates a fixture FD with author field.
func writeFDWithAuthor(t *testing.T, dir, id, title, status, author string) {
	t.Helper()
	fdDir := filepath.Join(dir, ".forgia", "fd")
	os.MkdirAll(fdDir, 0o755)
	content := "---\nid: " + id + "\ntitle: \"" + title + "\"\nstatus: " + status + "\nauthor: \"" + author + "\"\n---\n# " + id + "\n"
	if err := os.WriteFile(filepath.Join(fdDir, id+"-test.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write FD fixture: %v", err)
	}
}

// setupMinimalVault creates a bare .forgia/ directory with required
// subdirectories so that vault.Open(".") succeeds and vault CRUD
// operations work. Used by MCP E2E tests that spawn forgia mcp serve.
func setupMinimalVault(t *testing.T, dir string) {
	t.Helper()
	forgiaDir := filepath.Join(dir, ".forgia")
	for _, sub := range []string{"", "fd", "sdd"} {
		if err := os.MkdirAll(filepath.Join(forgiaDir, sub), 0o755); err != nil {
			t.Fatalf("create .forgia/%s: %v", sub, err)
		}
	}
}

// --- E2E Tests ---

func TestE2E_Init(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	output, code := runForgia(t, dir, "init")
	assertExitCode(t, code, 0)
	assertContains(t, output, "Vault scaffolded")

	// Verify directory structure.
	expectedDirs := []string{
		".forgia",
		".forgia/fd",
		".forgia/sdd",
		".forgia/ops",
		".forgia/dev-guide",
		".forgia/dev-guide/principles",
		".forgia/dev-guide/lang",
		".forgia/guardrails",
		".forgia/architecture",
		".forgia/contexts",
		".forgia/learnings",
	}
	for _, d := range expectedDirs {
		assertDirExists(t, filepath.Join(dir, d))
	}

	// Verify key files.
	expectedFiles := []string{
		".forgia/constitution.md",
		".forgia/config.toml",
		".forgia/_dashboard.md",
		".forgia/guardrails/deny.toml",
		".forgia/fd/_templates/fd-template.md",
		".forgia/sdd/_templates/sdd-template.md",
		".forgia/ops/_templates/ops-task.md",
		".forgia/dev-guide/lang/shell.md",
		".forgia/dev-guide/principles/clean-code.md",
		".forgia/dev-guide/principles/solid.md",
		".forgia/dev-guide/principles/design-patterns.md",
		".forgia/architecture/system-context.yaml",
		".forgia/architecture/containers.yaml",
		".forgia/architecture/technology-decisions.yaml",
		".forgia/architecture/quality-attributes.yaml",
		".forgia/architecture/constraints.yaml",
		".forgia/architecture/glossary.yaml",
		".forgia/contexts/_template.yaml",
		".forgia/learnings/_template.yaml",
	}
	for _, f := range expectedFiles {
		assertFileExists(t, filepath.Join(dir, f))
	}
}

func TestE2E_Init_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	initVault(t, dir)

	// Modify constitution — re-init should NOT overwrite.
	constitutionPath := filepath.Join(dir, ".forgia", "constitution.md")
	os.WriteFile(constitutionPath, []byte("# Custom rules"), 0o644)

	// Second init.
	_, code := runForgia(t, dir, "init")
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(constitutionPath)
	if err != nil {
		t.Fatalf("read constitution: %v", err)
	}
	if string(data) != "# Custom rules" {
		t.Errorf("constitution was overwritten on re-init: got %q", string(data))
	}
}

func TestE2E_Init_LanguageDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		marker string
		lang   string
	}{
		{"Go", "go.mod", "go"},
		{"Rust", "Cargo.toml", "rust"},
		{"Python", "pyproject.toml", "python"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, tt.marker), []byte(""), 0o644)

			output, code := runForgia(t, dir, "init")
			assertExitCode(t, code, 0)
			assertContains(t, output, tt.lang)
		})
	}
}

func TestE2E_Status_Empty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	output, code := runForgia(t, dir, "status")
	assertExitCode(t, code, 0)
	assertContains(t, output, "No Feature Designs found")
}

func TestE2E_Status_WithFDs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	writeFD(t, dir, "FD-001", "Test Feature", "planned")
	writeSDD(t, dir, "FD-001", "SDD-001", "API Module", "planned")

	output, code := runForgia(t, dir, "status")
	assertExitCode(t, code, 0)
	assertContains(t, output, "Forgia Dashboard")
	assertContains(t, output, "FD-001")
	assertContains(t, output, "Test Feature")
	assertContains(t, output, "planned")
	assertContains(t, output, "SDD-001")
	assertContains(t, output, "API Module")
}

func TestE2E_Status_NoVault(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	_, code := runForgia(t, dir, "status")
	if code == 0 {
		t.Error("expected non-zero exit for status without vault")
	}
}

func TestE2E_Doctor(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	output, code := runForgia(t, dir, "doctor")
	// Doctor always exits 0.
	assertExitCode(t, code, 0)
	assertContains(t, output, "Forgia Doctor")
	assertContains(t, output, "vault")
	assertContains(t, output, "config")
	assertContains(t, output, "guardrails")
}

func TestE2E_Validate_Valid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	fdID := "FD-001"
	writeFD(t, dir, fdID, "Test Feature", "planned")
	sddPath := writeValidSDD(t, dir, fdID)

	output, code := runForgia(t, dir, "validate", sddPath)
	assertExitCode(t, code, 0)
	assertContains(t, output, "valid")
}

func TestE2E_Validate_Invalid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	fdID := "FD-001"
	writeFD(t, dir, fdID, "Test Feature", "planned")
	sddPath := writeInvalidSDD(t, dir, fdID)

	output, code := runForgia(t, dir, "validate", sddPath)
	if code == 0 {
		t.Error("expected non-zero exit for invalid SDD")
	}
	assertContains(t, output, "error")
}

func TestE2E_Validate_MissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	_, code := runForgia(t, dir, "validate", "nonexistent.md")
	if code == 0 {
		t.Error("expected non-zero exit for missing file")
	}
}

func TestE2E_Skills(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	output, code := runForgia(t, dir, "skills")
	assertExitCode(t, code, 0)
	assertContains(t, output, "Available Skills")
	assertContains(t, output, "MODE")

	// Verify known embedded skills are listed.
	for _, skill := range []string{"fd-new", "fd-review", "fd-sdd", "sdd-assign", "arch-review"} {
		assertContains(t, output, skill)
	}

	assertContains(t, output, "Slash Command")
	assertContains(t, output, "Total:")
}

func TestE2E_Skills_WithVault(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	output, code := runForgia(t, dir, "skills")
	assertExitCode(t, code, 0)

	// With vault + config (which includes knowledge.provider = "codebase-memory-mcp"),
	// composite skills should also be listed.
	assertContains(t, output, "MCP Tool")
	for _, skill := range []string{
		"blast_radius", "arch_init", "trace_calls", "search_code",
		"security_scan", "arch_coherence", "context_map",
	} {
		assertContains(t, output, skill)
	}

	// Should have 26 total (19 embedded + 7 composite).
	assertContains(t, output, "Total: 26 skills")
}

func TestE2E_Version(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	output, code := runForgia(t, dir, "version")
	assertExitCode(t, code, 0)
	assertContains(t, output, "forgia")
}

func TestE2E_ExecDryRun(t *testing.T) {
	t.Skip("requires claude CLI for dry-run execution")
}

func TestE2E_BatchDryRun(t *testing.T) {
	t.Skip("requires claude CLI for batch dry-run execution")
}

func TestE2E_ExecMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	_, code := runForgia(t, dir, "exec", "nonexistent.md")
	if code == 0 {
		t.Error("expected non-zero exit for exec with missing file")
	}
}

func TestE2E_NoArgs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	output, code := runForgia(t, dir)
	// Cobra root command with SilenceUsage — may exit 0 and show help.
	_ = code
	assertContains(t, output, "forgia")
}

func TestE2E_UnknownCommand(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	_, code := runForgia(t, dir, "foobar")
	if code == 0 {
		t.Error("expected non-zero exit for unknown command")
	}
}

func TestE2E_TemplateContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	// Constitution has required sections.
	constitution, err := os.ReadFile(filepath.Join(dir, ".forgia", "constitution.md"))
	if err != nil {
		t.Fatalf("read constitution: %v", err)
	}
	for _, section := range []string{"Principles", "Code Standards", "Commit Conventions"} {
		if !strings.Contains(string(constitution), section) {
			t.Errorf("constitution missing section %q", section)
		}
	}

	// FD template has required sections.
	fdTpl, err := os.ReadFile(filepath.Join(dir, ".forgia", "fd", "_templates", "fd-template.md"))
	if err != nil {
		t.Fatalf("read fd template: %v", err)
	}
	for _, section := range []string{"Problem", "Architecture", "Interfaces", "SDD Previsti"} {
		if !strings.Contains(string(fdTpl), section) {
			t.Errorf("fd template missing section %q", section)
		}
	}

	// SDD template has required sections.
	sddTpl, err := os.ReadFile(filepath.Join(dir, ".forgia", "sdd", "_templates", "sdd-template.md"))
	if err != nil {
		t.Fatalf("read sdd template: %v", err)
	}
	for _, section := range []string{"Scope", "Interfaces", "Constraints", "Test Requirements", "Acceptance Criteria", "Work Log"} {
		if !strings.Contains(string(sddTpl), section) {
			t.Errorf("sdd template missing section %q", section)
		}
	}

	// Config.toml has required sections.
	configContent, err := os.ReadFile(filepath.Join(dir, ".forgia", "config.toml"))
	if err != nil {
		t.Fatalf("read config.toml: %v", err)
	}
	for _, section := range []string{"[runner]", "default"} {
		if !strings.Contains(string(configContent), section) {
			t.Errorf("config.toml missing %q", section)
		}
	}
}

// --- Init: architecture idempotency (from e2e.sh) ---

func TestE2E_Init_ArchitectureIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	// Modify an architecture file — re-init should NOT overwrite.
	archPath := filepath.Join(dir, ".forgia", "architecture", "system-context.yaml")
	os.WriteFile(archPath, []byte("# Custom content"), 0o644)

	initVault(t, dir) // second init
	assertFileNotOverwritten(t, archPath, "# Custom content")
}

// --- Init: .gitignore (from e2e-multi-eng.sh) ---

func TestE2E_Init_GitIgnore(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	giPath := filepath.Join(dir, ".forgia", ".gitignore")
	assertFileExists(t, giPath)
	assertFileContains(t, giPath, "logs/")
	assertFileContains(t, giPath, "run/")
	assertFileContains(t, giPath, ".beads/")
	assertFileContains(t, giPath, "*.pid")
}

func TestE2E_Init_GitIgnoreIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	giPath := filepath.Join(dir, ".forgia", ".gitignore")
	os.WriteFile(giPath, []byte("# custom"), 0o644)

	initVault(t, dir)
	assertFileNotOverwritten(t, giPath, "# custom")
}

// --- Init: CODEOWNERS (from e2e-multi-eng.sh) ---

func TestE2E_Init_NoCODEOWNERS_WithoutGithubDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	coPath := filepath.Join(dir, ".github", "CODEOWNERS")
	if _, err := os.Stat(coPath); err == nil {
		t.Error("CODEOWNERS should not be created when .github/ is missing")
	}
}

func TestE2E_Init_CODEOWNERS_WithGithubDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".github"), 0o755)

	initVault(t, dir)

	coPath := filepath.Join(dir, ".github", "CODEOWNERS")
	assertFileExists(t, coPath)
	assertFileContains(t, coPath, "constitution.md")
	assertFileContains(t, coPath, "guardrails/")
	assertFileContains(t, coPath, "architecture/")
}

func TestE2E_Init_CODEOWNERSIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".github"), 0o755)

	initVault(t, dir)

	coPath := filepath.Join(dir, ".github", "CODEOWNERS")
	os.WriteFile(coPath, []byte("# custom owners"), 0o644)

	// Remove .forgia to force re-init, keep .github/CODEOWNERS
	os.RemoveAll(filepath.Join(dir, ".forgia"))
	initVault(t, dir)

	assertFileNotOverwritten(t, coPath, "# custom owners")
}

// --- Init: guardrails content (from e2e-guardrails.sh) ---

func TestE2E_Init_GuardrailsContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	deny := filepath.Join(dir, ".forgia", "guardrails", "deny.toml")
	assertFileExists(t, deny)

	// Structure.
	assertFileContains(t, deny, "[read]")
	assertFileContains(t, deny, "[execute]")
	assertFileContains(t, deny, "[write]")

	// Read patterns.
	assertFileContains(t, deny, ".env")
	assertFileContains(t, deny, "*.pem")
	assertFileContains(t, deny, "*.key")
	assertFileContains(t, deny, ".aws/credentials")
	assertFileContains(t, deny, ".ssh/id_")
	assertFileContains(t, deny, ".gnupg/")
	assertFileContains(t, deny, ".password-store/")
	assertFileContains(t, deny, "kubeconfig")
	assertFileContains(t, deny, ".tfstate")
	assertFileContains(t, deny, ".env.example") // negation pattern

	// Execute patterns.
	assertFileContains(t, deny, "cat ~/.ssh/id_")
	assertFileContains(t, deny, "gpg --export-secret")
	assertFileContains(t, deny, "pass show")

	// Write patterns.
	assertFileContains(t, deny, "constitution.md")
	assertFileContains(t, deny, "config.toml")
	assertFileContains(t, deny, "deny.toml")

	// Ignore file.
	ignore := filepath.Join(dir, ".forgia", "guardrails", "ignore")
	assertFileExists(t, ignore)
	assertFileContains(t, ignore, ".env")
	assertFileContains(t, ignore, "node_modules/")
	assertFileContains(t, ignore, ".DS_Store")
}

func TestE2E_Init_GuardrailsIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	denyPath := filepath.Join(dir, ".forgia", "guardrails", "deny.toml")
	os.WriteFile(denyPath, []byte("# Custom deny"), 0o644)
	ignorePath := filepath.Join(dir, ".forgia", "guardrails", "ignore")
	os.WriteFile(ignorePath, []byte("# Custom ignore"), 0o644)

	initVault(t, dir)

	assertFileNotOverwritten(t, denyPath, "# Custom deny")
	assertFileNotOverwritten(t, ignorePath, "# Custom ignore")
}

// --- Init: config.toml sections (from e2e-beads-autospec.sh) ---

func TestE2E_Init_ConfigSections(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	cfg := filepath.Join(dir, ".forgia", "config.toml")
	assertFileContains(t, cfg, "[runner]")
	assertFileContains(t, cfg, "[runner.claude]")
	assertFileContains(t, cfg, "[runner.openhands]")
	assertFileContains(t, cfg, "[watcher]")
	assertFileContains(t, cfg, "[beads]")
	assertFileContains(t, cfg, "[knowledge]")
	assertFileContains(t, cfg, "auto_approve")
	assertFileContains(t, cfg, "max_turns")
	assertFileContains(t, cfg, "debounce")
	assertFileContains(t, cfg, `provider = "codebase-memory-mcp"`)
	assertFileContains(t, cfg, "auto_index")
}

// --- Init: constitution security (from e2e-guardrails.sh) ---

func TestE2E_Init_ConstitutionSecurity(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	c := filepath.Join(dir, ".forgia", "constitution.md")
	assertFileContains(t, c, "Security")
	assertFileContains(t, c, "guardrails")
	assertFileContains(t, c, "deny.toml")
	assertFileContains(t, c, "fail-closed")
}

// --- Init: FD template Mermaid (from e2e-beads-autospec.sh) ---

func TestE2E_Init_FDTemplateMermaid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	fd := filepath.Join(dir, ".forgia", "fd", "_templates", "fd-template.md")
	assertFileContains(t, fd, "Integration Context")
	assertFileContains(t, fd, "Data Flow")
	assertFileContains(t, fd, "sequenceDiagram")
	assertFileContains(t, fd, "flowchart")
	assertFileContains(t, fd, "OBBLIGATORIO")
	assertFileContains(t, fd, "author:")
	assertFileContains(t, fd, "{{AUTHOR}}")
	assertFileContains(t, fd, "mermaid")
	assertFileContains(t, fd, "Verification")
}

// --- Init: YAML SDD template (from e2e-beads-autospec.sh) ---

func TestE2E_Init_YAMLSDDTemplate(t *testing.T) {
	t.Parallel()

	// Check the source template in modules/ (not the scaffolded one).
	root := findProjectRoot()
	yamlPath := filepath.Join(root, "modules", "vault-template", "sdd", "_templates", "sdd-template.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Skip("YAML SDD template not found in modules/")
	}

	for _, section := range []string{
		"meta:", "scope:", "interfaces:", "constraints:", "best_practices:",
		"test_requirements:", "acceptance_criteria:", "context:", "constitution_check:",
		"work_log:", "bd_task_id:", "retrospective:",
	} {
		assertFileContains(t, yamlPath, section)
	}
}

// --- Status: author display (from e2e-multi-eng.sh) ---

func TestE2E_Status_Author(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	writeFDWithAuthor(t, dir, "FD-001", "Auth Feature", "planned", "ferruvich")

	output, code := runForgia(t, dir, "status")
	assertExitCode(t, code, 0)
	assertContains(t, output, "ferruvich")
}

func TestE2E_Status_NoAuthorGraceful(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	writeFD(t, dir, "FD-002", "No Author Feature", "planned")

	output, code := runForgia(t, dir, "status")
	assertExitCode(t, code, 0)
	assertContains(t, output, "FD-002")
	assertContains(t, output, "No Author Feature")
}

// --- Validate: FD directory batch (from e2e-beads-autospec.sh) ---

func TestE2E_Validate_FDDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	fdID := "FD-TEST"
	writeFD(t, dir, fdID, "Test", "planned")
	writeValidSDD(t, dir, fdID)

	// Valid FD directory.
	_, code := runForgia(t, dir, "validate", fdID)
	assertExitCode(t, code, 0)

	// Add an invalid SDD — batch should fail.
	writeInvalidSDD(t, dir, fdID)
	_, code = runForgia(t, dir, "validate", fdID)
	if code == 0 {
		t.Error("expected non-zero exit for FD directory with invalid SDD")
	}
}

// --- Batch: missing FD directory (from e2e-beads-autospec.sh) ---

func TestE2E_Batch_MissingFD(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	_, code := runForgia(t, dir, "batch", "FD-NOPE")
	if code == 0 {
		t.Error("expected non-zero exit for batch on nonexistent FD")
	}
}

// --- Doctor: missing guardrails (from e2e-guardrails.sh) ---

func TestE2E_Doctor_MissingGuardrails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	os.Remove(filepath.Join(dir, ".forgia", "guardrails", "deny.toml"))

	output, code := runForgia(t, dir, "doctor")
	assertExitCode(t, code, 0) // doctor always exits 0
	// Should report issue about missing guardrails.
	assertContains(t, output, "guardrails")
}

// --- Doctor: checks fswatch, yq, beads (from e2e-beads-autospec.sh) ---

func TestE2E_Doctor_ToolChecks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	output, code := runForgia(t, dir, "doctor")
	assertExitCode(t, code, 0)
	assertContains(t, output, "fswatch")
	assertContains(t, output, "yq")
}

// --- Runner: default is claude (from e2e.sh) ---

func TestE2E_Init_DefaultRunnerClaude(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	cfg, err := os.ReadFile(filepath.Join(dir, ".forgia", "config.toml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	assertContains(t, string(cfg), `default = "claude"`)
}

// --- Slash command content: fd-review (from e2e-guardrails.sh, e2e-beads-autospec.sh) ---

func TestE2E_SlashCommandContent(t *testing.T) {
	t.Parallel()
	root := findProjectRoot()

	// fd-review checks security and guardrails.
	reviewPath := filepath.Join(root, "modules", "claude-commands", "fd-review.md")
	assertFileContains(t, reviewPath, "Security")
	assertFileContains(t, reviewPath, "guardrails")
	assertFileContains(t, reviewPath, "plaintext")
	assertFileContains(t, reviewPath, "Integration Context")
	assertFileContains(t, reviewPath, "Data Flow")

	// fd-sdd loads guardrails.
	sddPath := filepath.Join(root, "modules", "claude-commands", "fd-sdd.md")
	assertFileContains(t, sddPath, "guardrails")
	assertFileContains(t, sddPath, "deny.toml")
	assertFileContains(t, sddPath, "Beads")
}

// --- Knowledge: config and graceful skip (from e2e-knowledge-config.sh, e2e-codebase-memory.sh) ---
// Note: tests requiring mock codebase-memory-mcp in PATH are covered by
// unit tests in internal/knowledge/. These E2E tests verify config presence only.

func TestE2E_Init_KnowledgeConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initVault(t, dir)

	cfg := filepath.Join(dir, ".forgia", "config.toml")
	assertFileContains(t, cfg, "[knowledge]")
	assertFileContains(t, cfg, `provider = "codebase-memory-mcp"`)
	assertFileContains(t, cfg, "auto_index = true")
	assertFileContains(t, cfg, "auto_sync = true")
}

// --- MCP: vault tools exposed (SDD-011) ---

func TestE2E_MCPServe_VaultToolsExposed(t *testing.T) {
	t.Parallel()

	enc, dec, cleanup := mcpServer(t)
	defer cleanup()

	resp := mcpRequest(t, enc, dec, 1, "tools/list", nil)
	if resp["error"] != nil {
		t.Fatalf("tools/list error: %v", resp["error"])
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result: %v", resp)
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("missing tools array")
	}

	toolNames := make(map[string]bool)
	for _, item := range tools {
		tool, _ := item.(map[string]any)
		name, _ := tool["name"].(string)
		toolNames[name] = true
	}

	// Verify all 11 vault tools are present.
	vaultTools := []string{
		"forgia_vault_fd_get", "forgia_vault_fd_list",
		"forgia_vault_fd_create", "forgia_vault_fd_update",
		"forgia_vault_sdd_get", "forgia_vault_sdd_list",
		"forgia_vault_sdd_create", "forgia_vault_sdd_update",
		"forgia_vault_status", "forgia_vault_validate",
		"forgia_vault_threat_model_get",
	}
	for _, name := range vaultTools {
		if !toolNames[name] {
			t.Errorf("vault tool %s not found in tools/list", name)
		}
	}

	// Total should be at least 18 (7 composite + 11 vault).
	if len(tools) < 18 {
		t.Errorf("expected at least 18 tools, got %d", len(tools))
	}
}
