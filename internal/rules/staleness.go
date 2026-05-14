package rules

import (
	"time"
)

// StalenessConfig holds thresholds for determining rule staleness.
type StalenessConfig struct {
	// MinQueryCount is the minimum number of queries required for a rule to be
	// considered active. Rules with fewer queries are stale.
	MinQueryCount float64

	// Window is the lookback duration used when evaluating usage.
	Window time.Duration
}

// DefaultStalenessConfig returns a StalenessConfig with sensible defaults.
func DefaultStalenessConfig() StalenessConfig {
	return StalenessConfig{
		MinQueryCount: 1.0,
		Window:        7 * 24 * time.Hour, // 7 days
	}
}

// RuleUsageSummary pairs a recording rule name with its observed query count.
type RuleUsageSummary struct {
	Name       string
	QueryCount float64
	LastSeen   time.Time
}

// IsStale reports whether the rule should be considered stale given cfg.
func (r RuleUsageSummary) IsStale(cfg StalenessConfig) bool {
	if r.QueryCount < cfg.MinQueryCount {
		return true
	}
	if !r.LastSeen.IsZero() && time.Since(r.LastSeen) > cfg.Window {
		return true
	}
	return false
}

// ClassifyUsage partitions a slice of RuleUsageSummary into stale and active
// sets according to cfg.
func ClassifyUsage(summaries []RuleUsageSummary, cfg StalenessConfig) (stale, active []RuleUsageSummary) {
	for _, s := range summaries {
		if s.IsStale(cfg) {
			stale = append(stale, s)
		} else {
			active = append(active, s)
		}
	}
	return stale, active
}
