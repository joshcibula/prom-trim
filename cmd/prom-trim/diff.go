package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/prom-trim/internal/rules"
)

// runDiff compares a rules file against a previously saved snapshot and prints
// a human-readable diff showing which rules were added, removed, or retained.
func runDiff(rulesFile, snapshotFile string) error {
	if rulesFile == "" {
		return fmt.Errorf("rules file path is required")
	}
	if snapshotFile == "" {
		return fmt.Errorf("snapshot file path is required")
	}

	// Load the snapshot from disk.
	snap, err := rules.LoadSnapshot(snapshotFile)
	if err != nil {
		return fmt.Errorf("loading snapshot %q: %w", snapshotFile, err)
	}

	// Compute the diff between the snapshot and the current rules file.
	result, err := rules.Diff(snap, rulesFile)
	if err != nil {
		return fmt.Errorf("computing diff: %w", err)
	}

	printDiff(result)
	return nil
}

// printDiff renders a DiffResult to stdout using a tab-aligned table.
func printDiff(result rules.DiffResult) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "Diff Summary\n")
	fmt.Fprintf(w, "------------\n")
	fmt.Fprintf(w, "Retained:\t%d\n", result.RetainedCount)
	fmt.Fprintf(w, "Removed:\t%d\n", len(result.RemovedRules))
	fmt.Fprintf(w, "Added:\t%d\n", len(result.AddedRules))
	fmt.Fprintln(w)

	if len(result.RemovedRules) > 0 {
		fmt.Fprintf(w, "Removed Rules:\n")
		for _, r := range result.RemovedRules {
			fmt.Fprintf(w, "  - %s\n", r)
		}
		fmt.Fprintln(w)
	}

	if len(result.AddedRules) > 0 {
		fmt.Fprintf(w, "Added Rules:\n")
		for _, r := range result.AddedRules {
			fmt.Fprintf(w, "  + %s\n", r)
		}
		fmt.Fprintln(w)
	}

	if len(result.RemovedRules) == 0 && len(result.AddedRules) == 0 {
		fmt.Fprintln(w, "No differences found.")
	}
}
