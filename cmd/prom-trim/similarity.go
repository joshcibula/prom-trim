package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/brancz/prom-trim/internal/rules"
)

func runSimilarity(args []string) error {
	fs := flag.NewFlagSet("similarity", flag.ContinueOnError)
	rulesFile := fs.String("rules", "", "path to rules YAML file (required)")
	minScore := fs.Float64("min-score", 0.6, "minimum Jaccard similarity score (0-1)")
	format := fs.String("format", "table", "output format: table or json")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *rulesFile == "" {
		return fmt.Errorf("--rules is required")
	}

	cfg := rules.SimilarityConfig{MinScore: *minScore}
	entries, err := rules.BuildSimilarityReport(*rulesFile, cfg)
	if err != nil {
		return fmt.Errorf("similarity report: %w", err)
	}

	switch *format {
	case "json":
		return printSimilarityJSON(entries)
	default:
		printSimilarityTable(entries)
		return nil
	}
}

func printSimilarityTable(entries []rules.SimilarityEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE_A\tRULE_B\tSCORE\tSHARED_TERMS")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%.3f\t%d\n", e.RuleA, e.RuleB, e.Score, e.SharedTerms)
	}
	w.Flush()
	if len(entries) == 0 {
		fmt.Println("no similar rule pairs found")
	}
}

func printSimilarityJSON(entries []rules.SimilarityEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
