package vault

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileVault implements Vault by reading/writing .forgia/ files on disk.
type FileVault struct {
	dir    string // absolute path to .forgia/
	logger *slog.Logger
}

// Open opens an existing .forgia/ vault at the given directory.
// Returns an error if the vault doesn't exist — use InitVault to create one.
func Open(dir string) (Vault, error) {
	forgiaDir := filepath.Join(dir, ".forgia")
	info, err := os.Stat(forgiaDir)
	if err != nil {
		return nil, fmt.Errorf("vault not found at %s: %w (run 'forgia init' first)", forgiaDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", forgiaDir)
	}

	return &FileVault{
		dir:    forgiaDir,
		logger: slog.With("component", "vault"),
	}, nil
}

// InitVault creates and initializes a new .forgia/ vault at the given directory.
// Use this when the vault doesn't exist yet. Use Open() for existing vaults.
func InitVault(ctx context.Context, dir string, opts InitOptions) (Vault, error) {
	forgiaDir := filepath.Join(dir, ".forgia")
	if err := os.MkdirAll(forgiaDir, 0o755); err != nil {
		return nil, fmt.Errorf("create vault dir: %w", err)
	}

	fv := &FileVault{
		dir:    forgiaDir,
		logger: slog.With("component", "vault"),
	}
	if err := fv.Init(ctx, opts); err != nil {
		return nil, err
	}
	return fv, nil
}

// Init scaffolds .forgia/ in the given directory.
func (v *FileVault) Init(ctx context.Context, opts InitOptions) error {
	dirs := []string{
		"fd", "fd/_templates",
		"sdd", "sdd/_templates",
		"ops", "ops/_templates",
		"architecture", "architecture/contexts",
		"dev-guide", "dev-guide/principles", "dev-guide/lang",
		"guardrails",
		"logs",
	}

	for _, d := range dirs {
		path := filepath.Join(v.dir, d)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", d, err)
		}
	}

	// Create minimal constitution.
	constitutionPath := filepath.Join(v.dir, "constitution.md")
	if _, err := os.Stat(constitutionPath); os.IsNotExist(err) {
		content := fmt.Sprintf("# %s — Constitution\n\nProject standards and rules.\n", opts.ProjectName)
		if err := os.WriteFile(constitutionPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write constitution: %w", err)
		}
	}

	// Create minimal guardrails.
	denyPath := filepath.Join(v.dir, "guardrails", "deny.toml")
	if _, err := os.Stat(denyPath); os.IsNotExist(err) {
		if err := os.WriteFile(denyPath, []byte("# Guardrails — deny rules\n"), 0o644); err != nil {
			return fmt.Errorf("write deny.toml: %w", err)
		}
	}

	v.logger.InfoContext(ctx, "vault initialized", "dir", v.dir, "project", opts.ProjectName)
	return nil
}

// FDs returns an iterator over all FD files.
func (v *FileVault) FDs() iter.Seq[*FD] {
	return func(yield func(*FD) bool) {
		fdDir := filepath.Join(v.dir, "fd")
		entries, err := os.ReadDir(fdDir)
		if err != nil {
			v.logger.Warn("failed to read fd directory", "dir", fdDir, "error", err)
			return
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasPrefix(e.Name(), "FD-") || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			fd, err := v.parseFDFile(filepath.Join(fdDir, e.Name()))
			if err != nil {
				v.logger.Warn("failed to parse FD file", "file", e.Name(), "error", err)
				continue
			}
			if !yield(fd) {
				return
			}
		}
	}
}

// ListFDs returns all FDs as a slice.
func (v *FileVault) ListFDs(_ context.Context) ([]*FD, error) {
	var fds []*FD
	for fd := range v.FDs() {
		fds = append(fds, fd)
	}
	return fds, nil
}

// GetFD returns a single FD by ID.
func (v *FileVault) GetFD(_ context.Context, id string) (*FD, error) {
	for fd := range v.FDs() {
		if fd.ID == id {
			return fd, nil
		}
	}
	return nil, fmt.Errorf("FD %q not found", id)
}

// CreateFD writes a new FD file.
func (v *FileVault) CreateFD(_ context.Context, fd *FD) error {
	if err := validateID(fd.ID); err != nil {
		return fmt.Errorf("invalid FD: %w", err)
	}
	path := filepath.Join(v.dir, "fd", fd.ID+".md")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("FD %q already exists at %s", fd.ID, path)
	}

	// Create the SDD subdirectory for this FD.
	sddDir := filepath.Join(v.dir, "sdd", fd.ID)
	if err := os.MkdirAll(sddDir, 0o755); err != nil {
		return fmt.Errorf("create SDD dir: %w", err)
	}

	return v.writeFrontmatterFile(path, fd, fmt.Sprintf("# %s: %s\n", fd.ID, fd.Title))
}

// UpdateFD overwrites an existing FD file, preserving the markdown body.
func (v *FileVault) UpdateFD(_ context.Context, fd *FD) error {
	path := fd.FilePath
	if path == "" {
		path = filepath.Join(v.dir, "fd", fd.ID+".md")
	}

	_, body, err := v.readFrontmatterFile(path)
	if err != nil {
		return fmt.Errorf("read existing FD %q: %w", fd.ID, err)
	}

	return v.writeFrontmatterFile(path, fd, body)
}

// SDDs returns an iterator over all SDDs for a given FD.
func (v *FileVault) SDDs(fdID string) iter.Seq[*SDD] {
	return func(yield func(*SDD) bool) {
		sddDir := filepath.Join(v.dir, "sdd", fdID)
		entries, err := os.ReadDir(sddDir)
		if err != nil {
			if !os.IsNotExist(err) {
				v.logger.Warn("failed to read sdd directory", "dir", sddDir, "error", err)
			}
			return
		}
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), "_") {
				continue
			}
			path := filepath.Join(sddDir, e.Name())
			sdd, err := v.parseSDDFile(path)
			if err != nil {
				v.logger.Warn("failed to parse SDD file", "file", e.Name(), "error", err)
				continue
			}
			if !yield(sdd) {
				return
			}
		}
	}
}

// ListSDDs returns all SDDs for a given FD.
func (v *FileVault) ListSDDs(_ context.Context, fdID string) ([]*SDD, error) {
	var sdds []*SDD
	for sdd := range v.SDDs(fdID) {
		sdds = append(sdds, sdd)
	}
	return sdds, nil
}

// GetSDD returns a single SDD by FD ID and SDD ID.
func (v *FileVault) GetSDD(_ context.Context, fdID, sddID string) (*SDD, error) {
	for sdd := range v.SDDs(fdID) {
		if sdd.ID == sddID {
			return sdd, nil
		}
	}
	return nil, fmt.Errorf("SDD %q not found in FD %q", sddID, fdID)
}

// CreateSDD writes a new SDD file.
func (v *FileVault) CreateSDD(_ context.Context, sdd *SDD) error {
	if err := validateID(sdd.ID); err != nil {
		return fmt.Errorf("invalid SDD: %w", err)
	}
	if err := validateID(sdd.FD); err != nil {
		return fmt.Errorf("invalid SDD parent FD: %w", err)
	}

	sddDir := filepath.Join(v.dir, "sdd", sdd.FD)
	if err := os.MkdirAll(sddDir, 0o755); err != nil {
		return fmt.Errorf("create SDD dir: %w", err)
	}

	path := filepath.Join(sddDir, sdd.ID+".md")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("SDD %q already exists at %s", sdd.ID, path)
	}

	return v.writeFrontmatterFile(path, sdd, fmt.Sprintf("# %s: %s\n\n> Parent FD: [[%s]]\n", sdd.ID, sdd.Title, sdd.FD))
}

// UpdateSDD overwrites an existing SDD file, preserving the markdown body.
func (v *FileVault) UpdateSDD(_ context.Context, sdd *SDD) error {
	path := sdd.FilePath
	if path == "" {
		path = filepath.Join(v.dir, "sdd", sdd.FD, sdd.ID+".md")
	}

	_, body, err := v.readFrontmatterFile(path)
	if err != nil {
		return fmt.Errorf("read existing SDD %q: %w", sdd.ID, err)
	}

	return v.writeFrontmatterFile(path, sdd, body)
}

// GetArchitecture reads the architecture definition.
func (v *FileVault) GetArchitecture(_ context.Context) (*Architecture, error) {
	path := filepath.Join(v.dir, "architecture", "system.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read architecture: %w", err)
	}

	var arch Architecture
	if err := yaml.Unmarshal(data, &arch); err != nil {
		return nil, fmt.Errorf("parse architecture: %w", err)
	}
	return &arch, nil
}

// ListContexts returns all bounded contexts.
func (v *FileVault) ListContexts(_ context.Context) ([]*BoundedContext, error) {
	ctxDir := filepath.Join(v.dir, "architecture", "contexts")
	entries, err := os.ReadDir(ctxDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read contexts dir: %w", err)
	}

	var contexts []*BoundedContext
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(ctxDir, e.Name()))
		if err != nil {
			continue
		}
		var bc BoundedContext
		if err := yaml.Unmarshal(data, &bc); err != nil {
			continue
		}
		bc.FilePath = filepath.Join(ctxDir, e.Name())
		contexts = append(contexts, &bc)
	}
	return contexts, nil
}

// GetContext returns a single bounded context by name.
func (v *FileVault) GetContext(ctx context.Context, name string) (*BoundedContext, error) {
	contexts, err := v.ListContexts(ctx)
	if err != nil {
		return nil, err
	}
	for _, bc := range contexts {
		if bc.Name == name {
			return bc, nil
		}
	}
	return nil, fmt.Errorf("bounded context %q not found", name)
}

// Constitution returns the raw constitution content.
func (v *FileVault) Constitution(_ context.Context) (string, error) {
	data, err := os.ReadFile(filepath.Join(v.dir, "constitution.md"))
	if err != nil {
		return "", fmt.Errorf("read constitution: %w", err)
	}
	return string(data), nil
}

// GuardrailsRaw returns the raw deny.toml content.
func (v *FileVault) GuardrailsRaw(_ context.Context) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(v.dir, "guardrails", "deny.toml"))
	if err != nil {
		return nil, fmt.Errorf("read guardrails: %w", err)
	}
	return data, nil
}

// Render converts a YAML file to markdown for human review.
func (v *FileVault) Render(_ context.Context, yamlPath string) error {
	// TODO: implement full YAML → MD rendering
	return fmt.Errorf("Render: not yet implemented for %s", yamlPath)
}

// --- validation helpers ---

// validateID checks that an ID is safe for use as a filename.
// Rejects path traversal attempts, absolute paths, and invalid characters.
func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("ID is required")
	}
	if strings.Contains(id, "..") {
		return fmt.Errorf("ID %q contains path traversal", id)
	}
	if strings.ContainsAny(id, "/\\") {
		return fmt.Errorf("ID %q contains path separator", id)
	}
	if filepath.IsAbs(id) {
		return fmt.Errorf("ID %q is an absolute path", id)
	}
	return nil
}

// --- frontmatter parsing helpers ---

// splitFrontmatter splits a "---\nyaml\n---\nbody" document.
func splitFrontmatter(data []byte) (frontmatter, body []byte, err error) {
	const sep = "---"
	content := string(data)

	// Must start with "---".
	if !strings.HasPrefix(strings.TrimSpace(content), sep) {
		return nil, data, fmt.Errorf("no frontmatter found")
	}

	// Find second "---".
	trimmed := strings.TrimSpace(content)
	rest := trimmed[len(sep):]
	idx := strings.Index(rest, "\n"+sep)
	if idx < 0 {
		return nil, data, fmt.Errorf("unclosed frontmatter")
	}

	fm := strings.TrimSpace(rest[:idx])
	// Body starts after the second "---" + newline.
	afterSep := rest[idx+1+len(sep):]
	if len(afterSep) > 0 && afterSep[0] == '\n' {
		afterSep = afterSep[1:]
	}

	return []byte(fm), []byte(afterSep), nil
}

// parseFDFile reads an FD markdown file with YAML frontmatter.
func (v *FileVault) parseFDFile(path string) (*FD, error) {
	fm, _, err := v.readFrontmatterFile(path)
	if err != nil {
		return nil, err
	}

	var fd FD
	if err := yaml.Unmarshal(fm, &fd); err != nil {
		return nil, fmt.Errorf("parse FD frontmatter at %s: %w", path, err)
	}
	fd.FilePath = path
	return &fd, nil
}

// parseSDDFile reads an SDD file (markdown with frontmatter or full YAML).
func (v *FileVault) parseSDDFile(path string) (*SDD, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sdd SDD
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		// Full YAML format — parse the "meta" wrapper.
		var wrapper struct {
			Meta SDD `yaml:"meta"`
		}
		if err := yaml.Unmarshal(data, &wrapper); err != nil {
			return nil, fmt.Errorf("parse SDD YAML at %s: %w", path, err)
		}
		sdd = wrapper.Meta
	} else {
		// Markdown with frontmatter.
		fm, _, fmErr := splitFrontmatter(data)
		if fmErr != nil {
			return nil, fmt.Errorf("parse SDD frontmatter at %s: %w", path, fmErr)
		}
		if err := yaml.Unmarshal(fm, &sdd); err != nil {
			return nil, fmt.Errorf("parse SDD frontmatter at %s: %w", path, err)
		}
	}

	sdd.FilePath = path
	return &sdd, nil
}

// readFrontmatterFile reads a file and splits frontmatter from body.
func (v *FileVault) readFrontmatterFile(path string) (frontmatter []byte, body string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	fm, bodyBytes, err := splitFrontmatter(data)
	if err != nil {
		return nil, "", err
	}
	return fm, string(bodyBytes), nil
}

// writeFrontmatterFile writes a YAML frontmatter + markdown body file.
func (v *FileVault) writeFrontmatterFile(path string, frontmatterObj any, body string) error {
	var buf bytes.Buffer

	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(frontmatterObj); err != nil {
		return fmt.Errorf("marshal frontmatter: %w", err)
	}
	enc.Close()
	buf.WriteString("---\n")

	if body != "" {
		if !strings.HasPrefix(body, "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString(body)
	}

	return os.WriteFile(path, buf.Bytes(), 0o644)
}
