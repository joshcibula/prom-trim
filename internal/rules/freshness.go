package rules

import (
	"fmt"
	"sort"
	"time"

	"github.com/prometheus/prometheus/model/rulefmt"
)

// FreshnessConfig controls thresholds for freshness classification.
type FreshnessConfig struct {
	RecentDays  int // rules queried within this window are "fresh"
	StaleDays   int // rules not queried within this window are "stale"
}

// DefaultFreshnessConfig returns sensible defaults.
func DefaultFreshnessConfig() FreshnessConfig {
	return FreshnessConfig{
		RecentDays: 7,
		StaleDays:  30,
	}
}

// FreshnessEntry holds freshness metadata for a single recording rule.
type FreshnessEntry struct {
	Name       string
	Group      string
	LastSeen   time.Time
	QueryCount int64
	Tier       string // "fresh", "aging", "stale", "never-seen"
}

// String returns a human-readable summary of the entry.
func (f FreshnessEntry) String() string {
	if f.LastSeen.IsZero() {
		return fmt.Sprintf("%s [%s] never-seen queries=0", f.Name, f.Group)
	}
	return fmt.Sprintf("%s [%s] last=%s queries=%d tier=%s",
		f.Name, f.Group, f.LastSeen.Format("2006-01-02"), f.QueryCount, f.Tier)
}

// BuildFreshnessReport parses the given rules file and classifies each
// recording rule by how recently it was observed in usage data.
func BuildFreshnessReport(
	path string,
	usage map[string]UsageEntry,
	cfg FreshnessConfig,
) ([]FreshnessEntry, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("freshness: parse %s: %w", path, err)
	}

	now := time.Now().UTC()
	var entries []FreshnessEntry

	for _, g := range groups {
		for _, r := range g.Rules {
			node, ok := r.(rulefmt.RuleNode)
			if !ok {
				continue
			}
			if node.Record.Value == "" {
				continue
			}
			name := node.Record.Value
			u, found := usage[name]
			entry := FreshnessEntry{
				Name:  name,
				Group: g.Name,
			}
			if !found || u.LastSeen.IsZero() {
				entry.Tier = "never-seen"
			} else {
				entry.LastSeen = u.LastSeen
				entry.QueryCount = u.Count
				age := now.Sub(u.LastSeen)
				switch {
				case age <= time.Duration(cfg.RecentDays)*24*time.Hour:
					entry.Tier = "fresh"
				case age <= time.Duration(cfg.StaleDays)*24*time.Hour:
					entry.Tier = "aging"
				default:
					entry.Tier = "stale"
				}
			}
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}
