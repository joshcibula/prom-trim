package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/user/prom-trim/internal/rules"
)

// runSnapshot handles the `snapshot` subcommand.
// It loads an existing snapshot and prints a summary, or saves a new one
// when --save is provided.
func runSnapshot(snapshotPath string, printOnly bool) error {
	if printOnly {
		return printSnapshot(snapshotPath)
	}
	return fmt.Errorf("snapshot: no action specified; use --print to display an existing snapshot")
}

func printSnapshot(path string) error {
	snap, err := rules.LoadSnapshot(path)
	if err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Snapshot\n")
	fmt.Fprintf(os.Stdout, "  File:    %s\n", snap.RulesFile)
	fmt.Fprintf(os.Stdout, "  Saved:   %s\n", snap.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(os.Stdout, "  Rules:   %d\n\n", len(snap.Usage))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tQUERY COUNT")
	fmt.Fprintln(w, "----\t-----------")

	for name, count := range snap.Usage {
		status := fmt.Sprintf("%d", count)
		if count == 0 {
			status = "0  (stale)"
		}
		fmt.Fprintf(w, "%s\t%s\n", name, status)
	}

	return w.Flush()
}
