package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runOrphans(args []string) error {
	fs := flag.NewFlagSet("orphans", flag.ContinueOnError)
	format := fs.String("format", "table", "Output format: table or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: prom-trim orphans [--format=table|json] <rules-file>")
	}
	path := fs.Arg(0)

	orphans, err := rules.BuildOrphansReport(path)
	if err != nil {
		return fmt.Errorf("orphans: %w", err)
	}

	switch *format {
	case "json":
		return printOrphansJSON(orphans)
	default:
		return printOrphansTable(orphans)
	}
}

func printOrphansTable(orphans []rules.OrphanEntry) error {
	if len(orphans) == 0 {
		fmt.Println("No orphaned recording rules found.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "GROUP\tRECORD\tEXPR")
	fmt.Fprintln(w, "-----\t------\t----")
	for _, e := range orphans {
		expr := e.Expr
		if len(expr) > 60 {
			expr = expr[:57] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.Group, e.Record, expr)
	}
	return w.Flush()
}

func printOrphansJSON(orphans []rules.OrphanEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(orphans)
}
