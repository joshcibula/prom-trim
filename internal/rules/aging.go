package rules

import (
	"fmt"
	"sort"
	"time"

	"github.com/prometheus/prometheus/model/rulefmt"
)

// AgingEntry describes how long a recording rule has existed without meaningful use.
type AgingEntry struct {
	RuleName  string
	GroupName string
	AgeDays   int
	LastSeen  time.Time
	QueryCount int64
	Risk      string
}

func (a AgingEntry) String() string {
	return fmt.Sprintf("%s (group=%s age=%dd risk=%s)", a.RuleName, a.GroupName, a.AgeDays, a.Risk)
}

// AgingConfig controls thresholds for aging classification.
type AgingConfig struct {
	StaleAfterDays  int
	WarnAfterDays   int
	MinQueryCount   int64
}

// DefaultAgingConfig returns sensible defaults.
func DefaultAgingConfig() AgingConfig {
	return AgingConfig{
		StaleAfterDays: 30,
		WarnAfterDays:  14,
		MinQueryCount:  5,
	}
}

func classifyAgingRisk(ageDays int, count int64, cfg AgingConfig) string {
	switch {
	case ageDays >= cfg.StaleAfterDays && count < cfg.MinQueryCount:
		return "high"
	case ageDays >= cfg.WarnAfterDays:
		return "medium"
	default:
		return "low"
	}
}

// BuildAgingReport returns aging entries for all recording rules in file,
// enriched with usage data from usageMap (rule name -> count, last seen).
func BuildAgingReport(file string, usageMap map[string]UsageStats, cfg AgingConfig) ([]AgingEntry, error) {
	groups, err := ParseFile(file)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var entries []AgingEntry

	for _, g := range groups {
		for _, r := range g.Rules {
			node := rulefmt.RuleNode(r)
			name := ruleRecordName(node)
			if name == "" {
				continue
			}
			us := usageMap[name]
			var ageDays int
			if !us.LastSeen.IsZero() {
				ageDays = int(now.Sub(us.LastSeen).Hours() / 24)
			} else {
				ageDays = cfg.StaleAfterDays // treat never-seen as maximally aged
			}
			entries = append(entries, AgingEntry{
				RuleName:   name,
				GroupName:  g.Name,
				AgeDays:    ageDays,
				LastSeen:   us.LastSeen,
				QueryCount: us.Count,
				Risk:       classifyAgingRisk(ageDays, us.Count, cfg),
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].AgeDays > entries[j].AgeDays
	})
	return entries, nil
}

// UsageStats holds minimal usage information for aging analysis.
type UsageStats struct {
	Count    int64
	LastSeen time.Time
}
