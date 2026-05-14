package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runAliases(args []string) error {
	fs := flag.NewFlagSet("aliases", flag.ContinueOnError)
	format := fs.String("format", "table", "Output format: table or json")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: prom-trim aliases [--format=table|json] <rules-file>")
	}

	path := fs.Arg(0)
	entries, err := rules.BuildAliasReport(path)
	if err != nil {
		return fmt.Errorf("aliases: %w", err)
	}

	switch strings.ToLower(*format) {
	case "json":
		return printAliasesJSON(entries)
	default:
		return printAliasesTable(entries)
	}
}

func printAliasesTable(entries []rules.AliasEntry) error {
	if len(entries) == 0 {
		fmt.Println("No aliased recording rules found.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tUSAGE COUNT\tUSED BY")
	fmt.Fprintln(w, "----\t-----------\t-------")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%d\t%s\n", e.Name, e.UsageCount, strings.Join(e.UsedBy, ", "))
	}
	return w.Flush()
}

func printAliasesJSON(entries []rules.AliasEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
