package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/prom-trim/internal/rules"
)

func runRetention(args []string) error {
	fs := flag.NewFlagSet("retention", flag.ContinueOnError)
	rulesFile := fs.String("rules", "", "Path to rules YAML file (required)")
	format := fs.String("format", "table", "Output format: table or json")
	minAge := fs.Int("min-age-days", 7, "Minimum age in days before a rule is eligible")
	maxStale := fs.Int("max-stale-days", 30, "Days without usage before high-priority classification")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *rulesFile == "" {
		return fmt.Errorf("--rules flag is required")
	}

	groups, err := rules.ParseFile(*rulesFile)
	if err != nil {
		return fmt.Errorf("parsing rules: %w", err)
	}

	policy := rules.RetentionPolicy{
		MinAgeDays:   *minAge,
		MaxStaleDays: *maxStale,
	}

	// Build stale entries from all recording rules (no usage data in offline mode).
	names := rules.AllRecordNames(groups)
	stale := make([]rules.StaleEntry, 0, len(names))
	for _, n := range names {
		stale = append(stale, rules.StaleEntry{Name: n, Usage: nil})
	}

	report := rules.BuildRetentionReport(stale, policy)

	switch *format {
	case "json":
		return printRetentionJSON(report)
	default:
		printRetentionTable(report)
		return nil
	}
}

func printRetentionTable(entries []rules.RetentionEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tAGE (days)\tPRIORITY\tELIGIBLE")
	for _, e := range entries {
		eligible := "no"
		if e.Eligible {
			eligible = "yes"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", e.Name, e.AgeDays, e.Priority, eligible)
	}
	w.Flush()
}

func printRetentionJSON(entries []rules.RetentionEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
