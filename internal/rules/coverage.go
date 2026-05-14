package rules

import (
	"fmt"
	"sort"

	"github.com/prometheus/prometheus/model/rulefmt"
)

// CoverageReport summarises how many recording rules have observed usage data.
type CoverageReport struct {
	Total    int
	Covered  int
	Uncovered []string
}

// Coverage returns the fraction of rules that have at least one usage entry.
func (r CoverageReport) Coverage() float64 {
	if r.Total == 0 {
		return 0
	}
	return float64(r.Covered) / float64(r.Total)
}

// String returns a human-readable one-liner.
func (r CoverageReport) String() string {
	return fmt.Sprintf("%d/%d rules covered (%.1f%%)", r.Covered, r.Total, r.Coverage()*100)
}

// BuildCoverageReport computes coverage given a parsed rule file and a usage
// map keyed by record name. Rules absent from the usage map are uncovered.
func BuildCoverageReport(
	groups []rulefmt.RuleGroup,
	usage map[string]float64,
) CoverageReport {
	var uncovered []string
	total := 0
	covered := 0

	for _, g := range groups {
		for _, r := range g.Rules {
			if r.Record.Value == "" {
				continue
			}
			total++
			if v, ok := usage[r.Record.Value]; ok && v > 0 {
				covered++
			} else {
				uncovered = append(uncovered, r.Record.Value)
			}
		}
	}

	sort.Strings(uncovered)

	return CoverageReport{
		Total:     total,
		Covered:   covered,
		Uncovered: uncovered,
	}
}
