package report

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// RuleResult holds the outcome of evaluating a single recording rule.
type RuleResult struct {
	Name      string
	Group     string
	LastSeen  time.Time
	QueryCount float64
	Stale     bool
}

// Summary holds aggregated report statistics.
type Summary struct {
	Total  int
	Stale  int
	Kept   int
	DryRun bool
}

// Write renders a human-readable report of rule results to w.
func Write(w io.Writer, results []RuleResult, summary Summary) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "RULE\tGROUP\tQUERY COUNT\tSTATUS")
	fmt.Fprintln(tw, "----\t-----\t-----------\t------")

	for _, r := range results {
		status := "active"
		if r.Stale {
			status = "STALE"
		}
		fmt.Fprintf(tw, "%s\t%s\t%.0f\t%s\n", r.Name, r.Group, r.QueryCount, status)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flushing tabwriter: %w", err)
	}

	fmt.Fprintln(w)
	if summary.DryRun {
		fmt.Fprintln(w, "[dry-run] no changes written")
	}
	fmt.Fprintf(w, "Total: %d  Stale: %d  Kept: %d\n", summary.Total, summary.Stale, summary.Kept)
	return nil
}

// BuildSummary derives a Summary from a slice of RuleResults.
func BuildSummary(results []RuleResult, dryRun bool) Summary {
	s := Summary{Total: len(results), DryRun: dryRun}
	for _, r := range results {
		if r.Stale {
			s.Stale++
		} else {
			s.Kept++
		}
	}
	return s
}
