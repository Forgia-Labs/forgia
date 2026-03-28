package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Deepzima/forgia/internal/beads"
	"github.com/Deepzima/forgia/internal/vault"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show FD and SDD dashboard",
	Long:  "Displays all Feature Designs and their SDDs with current status.",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	v, err := vault.Open(".")
	if err != nil {
		return fmt.Errorf("vault: %w", err)
	}

	return printDashboard(ctx, v)
}

func printDashboard(ctx context.Context, v vault.Vault) error {
	fds, err := v.ListFDs(ctx)
	if err != nil {
		return fmt.Errorf("list FDs: %w", err)
	}

	if len(fds) == 0 {
		fmt.Println("No Feature Designs found.")
		fmt.Println("Create one with: forgia skill fd-new")
		return nil
	}

	fmt.Println("=== Forgia Dashboard ===")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tTITLE\tSTATUS\tPRIORITY\tAUTHOR\n")
	fmt.Fprintf(w, "──\t─────\t──────\t────────\t──────\n")

	// Group competing FDs together.
	// Uses upstream_issue when available, otherwise builds a stable key
	// from the sorted set of competing FD IDs.
	groups := make(map[string][]*vault.FD)
	grouped := make(map[string]bool)
	var ungrouped []*vault.FD

	for _, fd := range fds {
		isCompetitive := len(fd.CompetesWith) > 0 || (fd.UpstreamIssue != "" && countByUpstream(fds, fd.UpstreamIssue) > 1)
		if isCompetitive {
			key := fd.UpstreamIssue
			if key == "" {
				// Build stable key from sorted competing IDs.
				ids := append(append([]string{}, fd.CompetesWith...), fd.ID)
				sort.Strings(ids)
				key = "competing:" + strings.Join(ids, ",")
			}
			groups[key] = append(groups[key], fd)
			grouped[fd.ID] = true
		}
	}

	for _, fd := range fds {
		if !grouped[fd.ID] {
			ungrouped = append(ungrouped, fd)
		}
	}

	// Print ungrouped FDs first.
	for _, fd := range ungrouped {
		printFDRow(ctx, w, v, fd, "")
	}

	// Print competitive groups (sorted keys for deterministic output).
	sortedKeys := make([]string, 0, len(groups))
	for k := range groups {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		group := groups[key]
		if len(group) > 1 {
			label := key
			if strings.HasPrefix(label, "competing:") {
				label = "linked FDs"
			}
			fmt.Fprintf(w, "\t── competing for %s ──\t\t\t\n", label)
		}
		for _, fd := range group {
			prefix := ""
			if fd.Status == vault.FDRejected {
				prefix = "[rejected] "
			}
			printFDRow(ctx, w, v, fd, prefix)
		}
	}

	w.Flush()

	// New enrichment sections — each degrades gracefully.
	printOPSTasks(ctx, v.Dir())
	printExecLogs(ctx, v.Dir())
	printBeadsReady(ctx)
	printKnowledgeStats(ctx)

	return nil
}

// countByUpstream counts FDs sharing the same upstream_issue.
func countByUpstream(fds []*vault.FD, issue string) int {
	n := 0
	for _, fd := range fds {
		if fd.UpstreamIssue == issue {
			n++
		}
	}
	return n
}

func printFDRow(ctx context.Context, w *tabwriter.Writer, v vault.Vault, fd *vault.FD, prefix string) {
	fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\t%s\n",
		prefix, fd.ID, fd.Title, fd.Status, fd.Priority, fd.Author)

	sdds, err := v.ListSDDs(ctx, fd.ID)
	if err != nil {
		slog.WarnContext(ctx, "failed to list SDDs", "fd", fd.ID, "error", err)
		return
	}
	for _, sdd := range sdds {
		fmt.Fprintf(w, "  └ %s\t%s\t%s\t\t\n",
			sdd.ID, sdd.Title, sdd.Status)
	}
}

// OPSFrontmatter holds parsed frontmatter from an OPS markdown file.
type OPSFrontmatter struct {
	ID       string `yaml:"id"`
	Title    string `yaml:"title"`
	Priority string `yaml:"priority"`
}

// parseOPSFile reads an OPS markdown file and extracts YAML frontmatter.
func parseOPSFile(data []byte) (OPSFrontmatter, error) {
	var ops OPSFrontmatter

	content := strings.TrimSpace(string(data))
	if !strings.HasPrefix(content, "---") {
		return ops, fmt.Errorf("no frontmatter found")
	}

	rest := content[3:] // skip opening "---"
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return ops, fmt.Errorf("unclosed frontmatter")
	}

	fm := strings.TrimSpace(rest[:idx])
	if err := yaml.Unmarshal([]byte(fm), &ops); err != nil {
		return ops, fmt.Errorf("parse frontmatter: %w", err)
	}
	return ops, nil
}

// printOPSTasks reads .forgia/ops/active/OPS-*.md and displays a table.
func printOPSTasks(ctx context.Context, vaultDir string) {
	opsDir := filepath.Join(vaultDir, "ops", "active")
	entries, err := os.ReadDir(opsDir)
	if err != nil {
		// Directory doesn't exist or can't be read — skip silently.
		return
	}

	var tasks []OPSFrontmatter
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "OPS-") || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(opsDir, e.Name()))
		if err != nil {
			slog.WarnContext(ctx, "failed to read OPS file", "file", e.Name(), "error", err)
			continue
		}
		ops, err := parseOPSFile(data)
		if err != nil {
			slog.WarnContext(ctx, "failed to parse OPS file", "file", e.Name(), "error", err)
			continue
		}
		tasks = append(tasks, ops)
	}

	if len(tasks) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("--- Active Tasks ---")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tTITLE\tPRIORITY\n")
	fmt.Fprintf(w, "──\t─────\t────────\n")
	for _, t := range tasks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", t.ID, t.Title, t.Priority)
	}
	w.Flush()
}

// ExecLog holds parsed fields from an execution JSON report.
type ExecLog struct {
	SDD             string `json:"sdd"`
	FD              string `json:"fd"`
	Runner          string `json:"runner"`
	Started         string `json:"started"`
	Completed       string `json:"completed"`
	DurationSeconds int    `json:"duration_seconds"`
	ExitCode        int    `json:"exit_code"`
	Status          string `json:"status"`
}

// parseExecLog parses a JSON execution report.
func parseExecLog(data []byte) (ExecLog, error) {
	var log ExecLog
	if err := json.Unmarshal(data, &log); err != nil {
		return log, fmt.Errorf("parse exec log: %w", err)
	}
	return log, nil
}

// printExecLogs reads .forgia/logs/exec-*.json and displays a summary table.
func printExecLogs(ctx context.Context, vaultDir string) {
	logsDir := filepath.Join(vaultDir, "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return
	}

	var logs []ExecLog
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "exec-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(logsDir, e.Name()))
		if err != nil {
			slog.WarnContext(ctx, "failed to read exec log", "file", e.Name(), "error", err)
			continue
		}
		log, err := parseExecLog(data)
		if err != nil {
			slog.WarnContext(ctx, "failed to parse exec log", "file", e.Name(), "error", err)
			continue
		}
		logs = append(logs, log)
	}

	if len(logs) == 0 {
		return
	}

	// Sort by started timestamp descending (most recent first).
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Started > logs[j].Started
	})

	// Show at most 10 entries.
	if len(logs) > 10 {
		logs = logs[:10]
	}

	fmt.Println()
	fmt.Println("--- Last Executions ---")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "SDD\tSTATUS\tDURATION\n")
	fmt.Fprintf(w, "───\t──────\t────────\n")
	for _, l := range logs {
		sddLabel := l.SDD
		if sddLabel == "" {
			sddLabel = "(unknown)"
		}
		mins := l.DurationSeconds / 60
		secs := l.DurationSeconds % 60
		fmt.Fprintf(w, "%s\t[%s]\t%dm%ds\n", sddLabel, l.Status, mins, secs)
	}
	w.Flush()
}

// printBeadsReady calls `bd ready` via beads.Client and displays the output.
func printBeadsReady(ctx context.Context) {
	bc := beads.NewClient()
	if !bc.Available() {
		return
	}

	output, err := bc.Call(ctx, "ready")
	if err != nil {
		slog.WarnContext(ctx, "beads ready call failed", "error", err)
		return
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return
	}

	fmt.Println()
	fmt.Println("--- Beads (ready tasks) ---")
	fmt.Println(output)
}

// KnowledgeProject holds fields from codebase-memory-mcp list_projects output.
type KnowledgeProject struct {
	NodeCount   int    `json:"node_count"`
	EdgeCount   int    `json:"edge_count"`
	LastIndexed string `json:"last_indexed"`
}

// formatRelativeTime converts a duration to a human-readable relative string.
func formatRelativeTime(d time.Duration) string {
	secs := max(int(d.Seconds()), 0)
	switch {
	case secs < 60:
		return fmt.Sprintf("%ds ago", secs)
	case secs < 3600:
		return fmt.Sprintf("%dm ago", secs/60)
	case secs < 86400:
		return fmt.Sprintf("%dh ago", secs/3600)
	default:
		return fmt.Sprintf("%dd ago", secs/86400)
	}
}

// printKnowledgeStats calls codebase-memory-mcp and displays index statistics.
func printKnowledgeStats(ctx context.Context) {
	if _, err := exec.LookPath("codebase-memory-mcp"); err != nil {
		return
	}

	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "codebase-memory-mcp", "cli", "list_projects")
	output, err := cmd.Output()
	if err != nil {
		slog.WarnContext(ctx, "knowledge layer stats failed", "error", err)
		return
	}

	var projects []KnowledgeProject
	if err := json.Unmarshal(output, &projects); err != nil {
		slog.WarnContext(ctx, "parse knowledge stats", "error", err)
		return
	}
	if len(projects) == 0 {
		return
	}

	p := projects[0]
	fmt.Println()
	fmt.Println("--- Knowledge Layer ---")

	msg := fmt.Sprintf("%d symbols  %d edges", p.NodeCount, p.EdgeCount)
	if p.LastIndexed != "" {
		t, err := time.Parse(time.RFC3339, p.LastIndexed)
		if err == nil {
			msg += fmt.Sprintf("  synced %s", formatRelativeTime(time.Since(t)))
		}
	}
	fmt.Printf("  Tier 3: codebase-memory-mcp  %s\n", msg)
}
