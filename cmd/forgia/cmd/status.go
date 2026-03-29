package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/Deepzima/forgia/internal/vault"
	"github.com/spf13/cobra"
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
