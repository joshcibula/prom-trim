package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runRedundancy(args []string) error {
	fs := flag.NewFlagSet("redundancy", flag.ContinueOnError)
	file := fs.String("file", "", "Path to rules YAML file (required)")
	threshold := fs.Float64("threshold", 0.5, "Jaccard similarity threshold (0.0–1.0)")
	format := fs.String("format", "table", "Output format: table or json")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *file == "" {
		return fmt.Errorf("--file is required")
	}

	entries, err := rules.BuildRedundancyReport(*file, *threshold)
	if err != nil {
		return err
	}

	switch *format {
	case "json":
		return printRedundancyJSON(entries)
	default:
		return printRedundancyTable(entries)
	}
}

func printRedundancyTable(entries []rules.RedundancyEntry) error {
	if len(entries) == 0 {
		fmt.Println("No redundant rules found.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tGROUP\tOVERLAPS WITH")
	for _, e := range entries {
		overlaps := ""
		for i, o := range e.OverlapsBy {
			if i > 0 {
				overlaps += ", "
			}
			overlaps += o
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.Rule, e.Group, overlaps)
	}
	return w.Flush()
}

func printRedundancyJSON(entries []rules.RedundancyEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
