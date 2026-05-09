package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

// runHistory prints the run history stored in the given file.
func runHistory(historyPath string) error {
	h, err := rules.LoadHistory(historyPath)
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}
	if len(h.Entries) == 0 {
		fmt.Println("No history entries found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tFILE\tTOTAL\tSTALE\tDRY-RUN\tPRUNED")
	for _, e := range h.Entries {
		dryRun := "no"
		if e.DryRun {
			dryRun = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%d\n",
			e.Timestamp.Format("2006-01-02 15:04:05"),
			e.RulesFile,
			e.TotalRules,
			e.StaleRules,
			dryRun,
			len(e.Pruned),
		)
	}
	return w.Flush()
}
