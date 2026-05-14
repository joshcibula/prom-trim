package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/org/prom-trim/internal/rules"
)

// runImpact loads a snapshot and a diff, then prints an impact report showing
// which stale rules carry the highest historical query load.
func runImpact(snapshotPath, rulesFile string, asJSON bool) error {
	snap, err := rules.LoadSnapshot(snapshotPath)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	diffResult, err := rules.Diff(snap, rulesFile)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	// Build usage map from snapshot entries.
	usage := make(map[string]int64, len(snap))
	for _, e := range snap {
		usage[e.RuleName] = e.QueryCount
	}

	// Only assess rules that were removed (stale).
	var stale []rules.DiffEntry
	for _, d := range diffResult {
		if d.Status == "removed" {
			stale = append(stale, d)
		}
	}

	report := rules.AssessImpact(stale, usage)

	if asJSON {
		return printImpactJSON(report)
	}
	return printImpactTable(report)
}

func printImpactTable(report rules.ImpactReport) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "LEVEL\tRULE\tGROUP\tQUERIES")
	fmt.Fprintln(w, "-----\t----\t-----\t-------")
	for _, e := range report {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", e.Level, e.RuleName, e.Group, e.QueryCount)
	}
	return w.Flush()
}

func printImpactJSON(report rules.ImpactReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
