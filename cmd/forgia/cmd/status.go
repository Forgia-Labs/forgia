package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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

	for _, fd := range fds {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			fd.ID, fd.Title, fd.Status, fd.Priority, fd.Author)

		// List SDDs for this FD.
		sdds, err := v.ListSDDs(ctx, fd.ID)
		if err != nil {
			slog.WarnContext(ctx, "failed to list SDDs", "fd", fd.ID, "error", err)
			continue
		}
		for _, sdd := range sdds {
			fmt.Fprintf(w, "  └ %s\t%s\t%s\t\t\n",
				sdd.ID, sdd.Title, sdd.Status)
		}
	}

	w.Flush()
	return nil
}
