package rules

import (
	"fmt"
	"sort"
)

// GroupStat holds aggregate statistics for a single rule group.
type GroupStat struct {
	GroupName  string
	RuleCount  int
	StaleCount int
	Coverage   float64 // percentage of rules with usage data
}

// String returns a human-readable summary of the group stat.
func (g GroupStat) String() string {
	return fmt.Sprintf("group=%s rules=%d stale=%d coverage=%.1f%%",
		g.GroupName, g.RuleCount, g.StaleCount, g.Coverage)
}

// BuildGroupStats computes per-group statistics given a parsed rule file
// and a usage map (record name -> query count).
func BuildGroupStats(file string, usage map[string]float64) ([]GroupStat, error) {
	groups, err := ParseFile(file)
	if err != nil {
		return nil, fmt.Errorf("groupstats: parse %s: %w", file, err)
	}

	var stats []GroupStat
	for _, g := range groups {
		stat := GroupStat{GroupName: g.Name}
		covered := 0
		for _, r := range g.Rules {
			name := ruleRecordName(r)
			if name == "" {
				continue
			}
			stat.RuleCount++
			count, ok := usage[name]
			if ok {
				covered++
			}
			if !ok || count == 0 {
				stat.StaleCount++
			}
		}
		if stat.RuleCount > 0 {
			stat.Coverage = float64(covered) / float64(stat.RuleCount) * 100
		}
		stats = append(stats, stat)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].GroupName < stats[j].GroupName
	})
	return stats, nil
}
