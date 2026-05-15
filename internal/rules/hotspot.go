package rules

import (
	"fmt"
	"sort"

	"github.com/prometheus/prometheus/model/rulefmt"
)

// HotspotEntry represents a recording rule with elevated query frequency.
type HotspotEntry struct {
	Name       string
	Group      string
	QueryCount int64
	AvgRate    float64
	Tier       string
}

func (h HotspotEntry) String() string {
	return fmt.Sprintf("%s (group=%s tier=%s count=%d rate=%.2f/s)",
		h.Name, h.Group, h.Tier, h.QueryCount, h.AvgRate)
}

// DefaultHotspotThreshold is the minimum query count to be considered a hotspot.
const DefaultHotspotThreshold int64 = 100

func classifyHotspotTier(count int64) string {
	switch {
	case count >= 1000:
		return "critical"
	case count >= 500:
		return "high"
	case count >= 100:
		return "medium"
	default:
		return "low"
	}
}

// BuildHotspotReport returns recording rules whose query count exceeds the
// given threshold, sorted descending by QueryCount.
func BuildHotspotReport(file string, usage map[string]UsageStats, threshold int64) ([]HotspotEntry, error) {
	groups, err := ParseFile(file)
	if err != nil {
		return nil, fmt.Errorf("hotspot: parse %s: %w", file, err)
	}

	var entries []HotspotEntry
	for _, g := range groups {
		for _, r := range g.Rules {
			name := ruleRecordName(rulefmt.RuleNode{RuleNode: r.RuleNode})
			if name == "" {
				continue
			}
			u, ok := usage[name]
			if !ok || u.QueryCount < threshold {
				continue
			}
			entries = append(entries, HotspotEntry{
				Name:       name,
				Group:      g.Name,
				QueryCount: u.QueryCount,
				AvgRate:    u.AvgRate,
				Tier:       classifyHotspotTier(u.QueryCount),
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].QueryCount > entries[j].QueryCount
	})
	return entries, nil
}
