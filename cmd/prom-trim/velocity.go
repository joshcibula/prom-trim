package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runVelocity(args []string) error {
	fs := flag.NewFlagSet("velocity", flag.ContinueOnError)
	historyFile := fs.String("history", "history.json", "Path to history JSON file")
	format := fs.String("format", "table", "Output format: table or json")
	risingThreshold := fs.Float64("rising", 0.10, "Delta threshold to classify as rising")
	fallingThreshold := fs.Float64("falling", -0.10, "Delta threshold to classify as falling")
	if err := fs.Parse(args); err != nil {
		return err
	}

	history, err := rules.LoadHistory(*historyFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("loading history: %w", err)
	}

	cfg := rules.DefaultVelocityConfig()
	cfg.RisingThreshold = *risingThreshold
	cfg.FallingThreshold = *fallingThreshold

	report := rules.BuildVelocityReport(history, cfg)

	switch *format {
	case "json":
		return printVelocityJSON(report)
	default:
		printVelocityTable(report)
		return nil
	}
}

func printVelocityTable(entries []rules.VelocityEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tEARLY_AVG\tLATE_AVG\tDELTA\tTREND")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%.2f\t%s\n",
			e.RuleName, e.EarlyAvg, e.LateAvg, e.Delta, e.Trend)
	}
	w.Flush()
	if len(entries) == 0 {
		fmt.Println("(no velocity data)")
	}
}

func printVelocityJSON(entries []rules.VelocityEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
