package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prometheus/prometheus/model/rulefmt"
)

// NamespaceStat holds aggregated statistics for a rule namespace (group name prefix).
type NamespaceStat struct {
	Namespace  string
	GroupCount int
	RuleCount  int
	StaleCount int
}

func (n NamespaceStat) String() string {
	return fmt.Sprintf("namespace=%s groups=%d rules=%d stale=%d",
		n.Namespace, n.GroupCount, n.RuleCount, n.StaleCount)
}

// ExtractNamespace derives a namespace from a group name by taking the portion
// before the first '/' or ':' separator. Falls back to the full name.
func ExtractNamespace(groupName string) string {
	for _, sep := range []string{"/", ":"} {
		if idx := strings.Index(groupName, sep); idx > 0 {
			return groupName[:idx]
		}
	}
	return groupName
}

// BuildNamespaceStats aggregates rule and staleness counts grouped by namespace.
// staleNames is the set of record names classified as stale.
func BuildNamespaceStats(groups []rulefmt.RuleGroup, staleNames map[string]bool) []NamespaceStat {
	type entry struct {
		groupsSeen map[string]bool
		ruleCount  int
		staleCount int
	}

	agg := map[string]*entry{}

	for _, g := range groups {
		ns := ExtractNamespace(g.Name)
		if _, ok := agg[ns]; !ok {
			agg[ns] = &entry{groupsSeen: map[string]bool{}}
		}
		e := agg[ns]
		e.groupsSeen[g.Name] = true

		for _, r := range g.Rules {
			if r.Record.Value == "" {
				continue
			}
			e.ruleCount++
			if staleNames[r.Record.Value] {
				e.staleCount++
			}
		}
	}

	stats := make([]NamespaceStat, 0, len(agg))
	for ns, e := range agg {
		stats = append(stats, NamespaceStat{
			Namespace:  ns,
			GroupCount: len(e.groupsSeen),
			RuleCount:  e.ruleCount,
			StaleCount: e.staleCount,
		})
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Namespace < stats[j].Namespace
	})
	return stats
}
