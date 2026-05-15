package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/your-org/prom-trim/internal/rules"
)

func runAging(args []string) error {
	fs := flag.NewFlagSet("aging", flag.ContinueOnError)
	file := fs.String("file", "", "path to rules YAML file (required)")
	format := fs.String("format", "table", "output format: table or json")
	warnDays := fs.Int("warn-days", 14, "days before a rule is considered warn-level")
	staleDays := fs.Int("stale-days", 30, "days before a rule is considered high-risk")
	minCount := fs.Int64("min-count", 5, "minimum query count to avoid high risk")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *file == "" {
		return fmt.Errorf("--file is required")
	}

	cfg := rules.AgingConfig{
		StaleAfterDays: *staleDays,
		WarnAfterDays:  *warnDays,
		MinQueryCount:  *minCount,
	}

	// No live Prometheus query here; build empty usage map (offline mode).
	usage := map[string]rules.UsageStats{}

	entries, err := rules.BuildAgingReport(*file, usage, cfg)
	if err != nil {
		return err
	}

	switch *format {
	case "json":
		return printAgingJSON(entries)
	default:
		printAgingTable(entries)
		return nil
	}
}

func printAgingTable(entries []rules.AgingEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tGROUP\tAGE (days)\tLAST SEEN\tQUERIES\tRISK")
	for _, e := range entries {
		ls := "never"
		if !e.LastSeen.IsZero() {
			ls = e.LastSeen.Format(time.DateOnly)
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%d\t%s\n",
			e.RuleName, e.GroupName, e.AgeDays, ls, e.QueryCount, e.Risk)
	}
	w.Flush()
}

func printAgingJSON(entries []rules.AgingEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
