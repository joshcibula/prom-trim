package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runHeatmap(args []string) error {
	fs := flag.NewFlagSet("heatmap", flag.ContinueOnError)
	rulesFile := fs.String("rules", "", "Path to rules YAML file (required)")
	historyFile := fs.String("history", "history.json", "Path to history JSON file")
	buckets := fs.Int("buckets", 7, "Number of time buckets to display")
	format := fs.String("format", "table", "Output format: table or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *rulesFile == "" {
		return fmt.Errorf("--rules is required")
	}

	groups, err := rules.ParseFile(*rulesFile)
	if err != nil {
		return fmt.Errorf("parse rules: %w", err)
	}
	names := rules.AllRecordNames(groups)

	history, err := rules.LoadHistory(*historyFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load history: %w", err)
	}

	// Build usageByPeriod from history entries.
	usageByPeriod := make(map[string]map[string]int64)
	for _, entry := range history {
		period := entry.Timestamp.Format("2006-01-02")
		if usageByPeriod[period] == nil {
			usageByPeriod[period] = make(map[string]int64)
		}
		for rule, count := range entry.Counts {
			usageByPeriod[period][rule] += count
		}
	}

	cfg := rules.HeatmapConfig{BucketCount: *buckets, BucketSize: 24 * 3600 * 1e9}
	report := rules.BuildHeatmapReport(names, usageByPeriod, cfg)

	switch *format {
	case "json":
		return printHeatmapJSON(report)
	default:
		printHeatmapTable(report)
		return nil
	}
}

func printHeatmapTable(entries []rules.HeatmapEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tPEAK\tAVERAGE\tBUCKETS")
	for _, e := range entries {
		periods := make([]string, 0, len(e.Buckets))
		for _, b := range e.Buckets {
			periods = append(periods, fmt.Sprintf("%s:%d", b.Period, b.Count))
		}
		fmt.Fprintf(w, "%s\t%d\t%.1f\t%v\n", e.RuleName, e.Peak, e.Average, periods)
	}
	w.Flush()
}

func printHeatmapJSON(entries []rules.HeatmapEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
