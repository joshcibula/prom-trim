package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/your-org/prom-trim/internal/rules"
)

func runAnnotations(args []string) error {
	fs := flag.NewFlagSet("annotations", flag.ContinueOnError)
	format := fs.String("format", "table", "Output format: table or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: prom-trim annotations [--format=table|json] <rules-file>")
	}
	path := fs.Arg(0)

	summaries, err := rules.ExtractAnnotations(path)
	if err != nil {
		return fmt.Errorf("extract annotations: %w", err)
	}

	switch *format {
	case "json":
		return printAnnotationsJSON(summaries)
	default:
		return printAnnotationsTable(summaries)
	}
}

func printAnnotationsTable(summaries []rules.AnnotationSummary) error {
	keys := rules.AnnotationKeys(summaries)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	fmt.Fprint(w, "GROUP\tRECORD")
	for _, k := range keys {
		fmt.Fprintf(w, "\t%s", k)
	}
	fmt.Fprintln(w)

	for _, s := range summaries {
		fmt.Fprintf(w, "%s\t%s", s.Group, s.Record)
		for _, k := range keys {
			fmt.Fprintf(w, "\t%s", s.Annotations[k])
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func printAnnotationsJSON(summaries []rules.AnnotationSummary) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(summaries)
}
