package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/andremissaglia/prom-trim/internal/rules"
)

func runNamespaces(args []string) error {
	fs := flag.NewFlagSet("namespaces", flag.ContinueOnError)
	format := fs.String("format", "table", "Output format: table or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: prom-trim namespaces [--format=table|json] <rules-file>")
	}
	rulesFile := fs.Arg(0)

	groups, err := rules.ParseFile(rulesFile)
	if err != nil {
		return fmt.Errorf("parse rules: %w", err)
	}

	// Build stale set from snapshot if available, otherwise empty.
	staleNames := map[string]bool{}
	snap, snapErr := rules.LoadSnapshot(rulesFile + ".snapshot.json")
	if snapErr == nil {
		for _, name := range snap.StaleRules {
			staleNames[name] = true
		}
	}

	stats := rules.BuildNamespaceStats(groups, staleNames)

	switch *format {
	case "json":
		return printNamespacesJSON(stats)
	default:
		printNamespacesTable(stats)
		return nil
	}
}

func printNamespacesTable(stats []rules.NamespaceStat) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tGROUPS\tRULES\tSTALE")
	for _, s := range stats {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", s.Namespace, s.GroupCount, s.RuleCount, s.StaleCount)
	}
	w.Flush()
}

func printNamespacesJSON(stats []rules.NamespaceStat) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(stats)
}
