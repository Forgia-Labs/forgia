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

	// Verify known skills are listed.
	for _, skill := range []string{"fd-new", "fd-review", "fd-sdd", "sdd-assign", "arch-review", "arch-init"} {
		assertContains(t, output, skill)
	}

	assertContains(t, output, "Total:")
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
