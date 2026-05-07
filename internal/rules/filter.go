package rules

import (
	"time"

	"github.com/prometheus/prometheus/model/rulefmt"
)

// UsageMap maps a recording rule name to its last usage time.
type UsageMap map[string]time.Time

// AllRecordNames returns all recording rule names across all groups.
func AllRecordNames(groups []rulefmt.RuleGroup) []string {
	var names []string
	for _, g := range groups {
		for _, r := range g.Rules {
			if r.Record.Value != "" {
				names = append(names, r.Record.Value)
			}
		}
	}
	return names
}

// FilterStale returns a copy of the groups with stale recording rules removed.
// A rule is considered stale if it does not appear in usage or its last seen
// time is before the cutoff.
func FilterStale(groups []rulefmt.RuleGroup, usage UsageMap, cutoff time.Time) (active []rulefmt.RuleGroup, stale []string) {
	for _, g := range groups {
		var kept []rulefmt.RuleNode
		for _, r := range g.Rules {
			name := r.Record.Value
			if name == "" {
				// Not a recording rule — always keep alerting rules.
				kept = append(kept, r)
				continue
			}
			lastSeen, found := usage[name]
			if !found || lastSeen.Before(cutoff) {
				stale = append(stale, name)
			} else {
				kept = append(kept, r)
			}
		}
		if len(kept) > 0 {
			active = append(active, rulefmt.RuleGroup{
				Name:     g.Name,
				Interval: g.Interval,
				Rules:    kept,
			})
		}
	}
	return active, stale
}
