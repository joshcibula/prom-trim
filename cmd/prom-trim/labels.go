package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/yourorg/prom-trim/internal/rules"
)

func runLabels(rulesFile, format string, indexMode bool) error {
	summaries, err := rules.ExtractLabels(rulesFile)
	if err != nil {
		return fmt.Errorf("extracting labels: %w", err)
	}

	if len(summaries) == 0 {
		fmt.Println("no recording rules found")
		return nil
	}

	if indexMode {
		idx := rules.BuildLabelIndex(summaries)
		if format == "json" {
			return printLabelIndexJSON(idx)
		}
		return printLabelIndexTable(idx)
	}

	if format == "json" {
		return printLabelSummariesJSON(summaries)
	}
	return printLabelSummariesTable(summaries)
}

func printLabelSummariesTable(summaries []rules.LabelSummary) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RULE\tLABELS")
	for _, s := range summaries {
		labels := "{}"
		if len(s.Labels) > 0 {
			pairs := make([]string, 0, len(s.Labels))
			for k, v := range s.Labels {
				pairs = append(pairs, k+"="+v)
			}
			labels = fmt.Sprintf("%v", pairs)
		}
		fmt.Fprintf(w, "%s\t%s\n", s.RuleName, labels)
	}
	return w.Flush()
}

func printLabelSummariesJSON(summaries []rules.LabelSummary) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(summaries)
}

func printLabelIndexTable(idx rules.LabelIndex) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "LABEL KEY\tRULES")
	keys := make([]string, 0, len(idx))
	for k := range idx {
		keys = append(keys, k)
	}
	for _, k := range keys {
		fmt.Fprintf(w, "%s\t%v\n", k, idx[k])
	}
	return w.Flush()
}

func printLabelIndexJSON(idx rules.LabelIndex) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(idx)
}
