package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runDrift(args []string) error {
	fs := flag.NewFlagSet("drift", flag.ContinueOnError)
	rulesFile := fs.String("rules", "", "path to rules YAML file (required)")
	historyFile := fs.String("history", "prom-trim-history.json", "path to history JSON file")
	format := fs.String("format", "table", "output format: table or json")
	stableThreshold := fs.Float64("stable-threshold", 10.0, "percentage delta considered stable")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *rulesFile == "" {
		return fmt.Errorf("--rules flag is required")
	}

	groups, err := rules.ParseFile(*rulesFile)
	if err != nil {
		return fmt.Errorf("parse rules: %w", err)
	}

	names := rules.AllRecordNames(groups)
	usage := make(map[string]rules.RuleUsage)
	for _, n := range names {
		usage[n] = rules.RuleUsage{RuleName: n, QueryCount: 0}
	}

	history, err := rules.LoadHistory(*historyFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load history: %w", err)
	}

	cfg := rules.DefaultDriftConfig()
	cfg.StableThresholdPct = *stableThreshold

	entries := rules.BuildDriftReport(history, usage, cfg)

	switch *format {
	case "json":
		return printDriftJSON(entries)
	default:
		printDriftTable(entries)
		return nil
	}
}

func printDriftTable(entries []rules.DriftEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tBASELINE\tCURRENT\tDELTA%\tDIRECTION")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%.1f\t%.1f\t%.1f%%\t%s\n",
			e.RuleName, e.Baseline, e.Current, e.DeltaPct, e.Direction)
	}
	if len(entries) == 0 {
		fmt.Fprintln(w, "(no drift data available)")
	}
	w.Flush()
}

func printDriftJSON(entries []rules.DriftEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
