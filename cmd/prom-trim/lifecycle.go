package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/prom-trim/internal/rules"
)

func runLifecycle(args []string) error {
	fs := flag.NewFlagSet("lifecycle", flag.ContinueOnError)
	rulesFile := fs.String("rules", "", "path to rules YAML file (required)")
	usageFile := fs.String("usage", "", "path to usage snapshot JSON file")
	format := fs.String("format", "table", "output format: table or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *rulesFile == "" {
		return fmt.Errorf("lifecycle: --rules is required")
	}

	var usage map[string]rules.UsageStats
	if *usageFile != "" {
		snap, err := rules.LoadSnapshot(*usageFile)
		if err != nil {
			return fmt.Errorf("lifecycle: load snapshot: %w", err)
		}
		usage = snap
	}

	cfg := rules.DefaultStalenessConfig()
	entries, err := rules.BuildLifecycleReport(*rulesFile, usage, cfg)
	if err != nil {
		return err
	}

	switch *format {
	case "json":
		return printLifecycleJSON(entries)
	default:
		printLifecycleTable(entries)
		return nil
	}
}

func printLifecycleTable(entries []rules.LifecycleEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tSTAGE\tTRANSITION\tREASON")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			e.Rule, e.Stage, e.Transition.Format("2006-01-02"), e.Reason)
	}
	w.Flush()
}

func printLifecycleJSON(entries []rules.LifecycleEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
