package report

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

// Summary holds aggregated statistics for a pruning run.
type Summary struct {
	Total   int
	Stale   int
	Active  int
	DryRun  bool
	RunAt   time.Time
}

// RuleRow represents a single rule entry in the formatted output.
type RuleRow struct {
	Group     string
	Name      string
	LastSeen  string
	Stale     bool
}

// FormatTable writes a human-readable tabular report of rule rows to w.
func FormatTable(w io.Writer, rows []RuleRow, s Summary) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	if s.DryRun {
		fmt.Fprintln(tw, "[DRY RUN] No changes will be written.")
	}

	fmt.Fprintf(tw, "Run at: %s\n", s.RunAt.Format(time.RFC3339))
	fmt.Fprintf(tw, "Total: %d | Active: %d | Stale: %d\n\n", s.Total, s.Active, s.Stale)

	fmt.Fprintln(tw, "GROUP\tNAME\tLAST SEEN\tSTATUS")
	fmt.Fprintln(tw, strings.Repeat("-", 60))

	for _, r := range rows {
		status := "active"
		if r.Stale {
			status = "STALE"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.Group, r.Name, r.LastSeen, status)
	}

	return tw.Flush()
}

// FormatJSON writes a minimal JSON summary line to w (newline-delimited).
func FormatJSON(w io.Writer, s Summary) error {
	_, err := fmt.Fprintf(w,
		`{"run_at":%q,"total":%d,"active":%d,"stale":%d,"dry_run":%v}`+"\n",
		s.RunAt.Format(time.RFC3339), s.Total, s.Active, s.Stale, s.DryRun,
	)
	return err
}
