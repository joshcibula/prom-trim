package prometheus

import (
	"context"
	"fmt"
	"time"
)

// RuleUsage holds usage statistics for a single recording rule.
type RuleUsage struct {
	RuleName   string
	QueryCount float64
	LastSeen   time.Time
}

// FetchRuleUsage queries Prometheus for the number of times each recording
// rule metric has been queried over the given lookback duration.
func (c *Client) FetchRuleUsage(ctx context.Context, ruleNames []string, lookback time.Duration) ([]RuleUsage, error) {
	if len(ruleNames) == 0 {
		return nil, nil
	}

	usages := make([]RuleUsage, 0, len(ruleNames))

	for _, name := range ruleNames {
		count, lastSeen, err := c.queryRuleMetric(ctx, name, lookback)
		if err != nil {
			return nil, fmt.Errorf("querying usage for rule %q: %w", name, err)
		}
		usages = append(usages, RuleUsage{
			RuleName:   name,
			QueryCount: count,
			LastSeen:   lastSeen,
		})
	}

	return usages, nil
}

// queryRuleMetric fetches the query count and last-seen time for a single rule
// metric using the Prometheus HTTP API.
func (c *Client) queryRuleMetric(ctx context.Context, ruleName string, lookback time.Duration) (float64, time.Time, error) {
	startTime := time.Now().Add(-lookback)

	// Query total number of samples scraped for this metric over the lookback window.
	countExpr := fmt.Sprintf(`count_over_time(%s[%s])`, ruleName, durationToPromQL(lookback))
	countResult, err := c.QueryInstant(ctx, countExpr)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("count query: %w", err)
	}

	var queryCount float64
	var lastSeen time.Time

	if len(countResult) > 0 {
		queryCount = countResult[0].Value
		lastSeen = countResult[0].Timestamp
	} else {
		// Metric was never seen in the window.
		lastSeen = startTime
	}

	return queryCount, lastSeen, nil
}

// durationToPromQL converts a Go duration to a PromQL duration string.
func durationToPromQL(d time.Duration) string {
	h := int(d.Hours())
	if h > 0 && d == time.Duration(h)*time.Hour {
		return fmt.Sprintf("%dh", h)
	}
	m := int(d.Minutes())
	if m > 0 && d == time.Duration(m)*time.Minute {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
