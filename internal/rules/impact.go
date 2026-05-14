package rules

import (
	"fmt"
	"sort"
)

// ImpactLevel categorises how significant removing a rule would be.
type ImpactLevel string

const (
	ImpactHigh   ImpactLevel = "high"
	ImpactMedium ImpactLevel = "medium"
	ImpactLow    ImpactLevel = "low"
)

// ImpactEntry describes the estimated removal impact for a single rule.
type ImpactEntry struct {
	RuleName    string
	Group       string
	QueryCount  int64
	LastSeenAge string
	Level       ImpactLevel
}

// ImpactReport is an ordered list of impact entries.
type ImpactReport []ImpactEntry

// AssessImpact scores each stale rule by its historical query count so that
// operators can prioritise which rules to remove first.
func AssessImpact(stale []DiffEntry, usage map[string]int64) ImpactReport {
	report := make(ImpactReport, 0, len(stale))

	for _, e := range stale {
		count := usage[e.RuleName]
		level := classifyImpact(count)
		report = append(report, ImpactEntry{
			RuleName:   e.RuleName,
			Group:      e.Group,
			QueryCount: count,
			Level:      level,
		})
	}

	sort.Slice(report, func(i, j int) bool {
		return report[i].QueryCount > report[j].QueryCount
	})

	return report
}

func classifyImpact(count int64) ImpactLevel {
	switch {
	case count >= 100:
		return ImpactHigh
	case count >= 10:
		return ImpactMedium
	default:
		return ImpactLow
	}
}

// String returns a human-readable summary line for an ImpactEntry.
func (e ImpactEntry) String() string {
	return fmt.Sprintf("[%s] %s (group: %s, queries: %d)", e.Level, e.RuleName, e.Group, e.QueryCount)
}
