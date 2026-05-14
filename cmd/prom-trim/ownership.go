package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/prometheus-community/prom-trim/internal/rules"
)

func runOwnership(args []string) error {
	fs := flag.NewFlagSet("ownership", flag.ContinueOnError)
	format := fs.String("format", "table", "Output format: table or json")
	byTeam := fs.Bool("by-team", false, "Group output by team")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: ownership [flags] <rules-file>")
	}
	path := fs.Arg(0)

	entries, err := rules.ExtractOwnership(path)
	if err != nil {
		return fmt.Errorf("extract ownership: %w", err)
	}

	if *byTeam {
		index := rules.BuildOwnerIndex(entries)
		if *format == "json" {
			return printOwnerIndexJSON(index)
		}
		return printOwnerIndexTable(index)
	}

	if *format == "json" {
		return printOwnershipJSON(entries)
	}
	return printOwnershipTable(entries)
}

func printOwnershipTable(entries []rules.OwnershipEntry) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tGROUP\tOWNER\tTEAM")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Rule, e.Group, e.Owner, e.Team)
	}
	return w.Flush()
}

func printOwnershipJSON(entries []rules.OwnershipEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

func printOwnerIndexTable(index map[string][]rules.OwnershipEntry) error {
	teams := make([]string, 0, len(index))
	for k := range index {
		teams = append(teams, k)
	}
	sort.Strings(teams)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TEAM\tRULE\tGROUP\tOWNER")
	for _, team := range teams {
		for _, e := range index[team] {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", team, e.Rule, e.Group, e.Owner)
		}
	}
	return w.Flush()
}

func printOwnerIndexJSON(index map[string][]rules.OwnershipEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(index)
}
