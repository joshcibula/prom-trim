package report

import (
	"github.com/yourorg/prom-trim/internal/prometheus"
	"github.com/yourorg/prom-trim/internal/rules"
)

// FromUsage constructs a slice of RuleResults by correlating parsed rule groups
// with Prometheus usage data. A rule is considered stale when its query count
// is zero (i.e. it was not referenced within the configured lookback window).
func FromUsage(
	groups []rules.RuleGroup,
	usage map[string]prometheus.RuleUsage,
) []RuleResult {
	var results []RuleResult

	for _, g := range groups {
		for _, r := range g.Rules {
			u, ok := usage[r.Record]
			var count float64
			if ok {
				count = u.QueryCount
			}
			results = append(results, RuleResult{
				Name:       r.Record,
				Group:      g.Name,
				QueryCount: count,
				Stale:      !ok || count == 0,
			})
		}
	}

	return results
}
