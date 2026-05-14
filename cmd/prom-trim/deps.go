package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runDeps(args []string) error {
	fs := flag.NewFlagSet("deps", flag.ContinueOnError)
	format := fs.String("format", "table", "Output format: table or json")
	target := fs.String("rule", "", "Show dependents of a specific rule name")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: prom-trim deps [flags] <rules-file>")
	}
	path := fs.Arg(0)

	graph, edges, err := rules.BuildDependencyGraph(path)
	if err != nil {
		return fmt.Errorf("deps: %w", err)
	}

	if *target != "" {
		deps := rules.Dependents(graph, *target)
		fmt.Printf("Dependents of %q:\n", *target)
		if len(deps) == 0 {
			fmt.Println("  (none)")
			return nil
		}
		for _, d := range deps {
			fmt.Printf("  %s\n", d)
		}
		return nil
	}

	switch *format {
	case "json":
		return printDepsJSON(edges)
	default:
		printDepsTable(graph, edges)
		return nil
	}
}

func printDepsTable(graph rules.DependencyGraph, edges []rules.DependencyEdge) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tDEPENDS ON")
	fmt.Fprintln(w, "----\t----------")
	if len(edges) == 0 {
		fmt.Fprintln(w, "(no inter-rule dependencies found)")
	} else {
		for _, e := range edges {
			fmt.Fprintf(w, "%s\t%s\n", e.From, e.To)
		}
	}
	w.Flush()
	fmt.Printf("\nTotal rules: %d, Total edges: %d\n", len(graph), len(edges))
}

func printDepsJSON(edges []rules.DependencyEdge) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(edges)
}
