package rules

import (
	"fmt"
	"sort"
	"time"
)

// ChurnEntry represents a rule that has changed frequently across snapshots.
type ChurnEntry struct {
	RuleName    string
	ChangeCount int
	FirstSeen   time.Time
	LastSeen    time.Time
	Status      string // "active", "removed"
}

func (e ChurnEntry) String() string {
	return fmt.Sprintf("%s: %d changes (%s)", e.RuleName, e.ChangeCount, e.Status)
}

// BuildChurnReport compares a series of history entries to identify rules
// that have been repeatedly added or removed across prune runs.
func BuildChurnReport(history []HistoryEntry, threshold int) []ChurnEntry {
	type tracker struct {
		count     int
		first     time.Time
		last      time.Time
		removed   int
	}

	seen := make(map[string]*tracker)

	for _, h := range history {
		for _, name := range h.PrunedRules {
			t, ok := seen[name]
			if !ok {
				t = &tracker{first: h.Timestamp}
				seen[name] = t
			}
			t.count++
			t.removed++
			if h.Timestamp.After(t.last) {
				t.last = h.Timestamp
			}
		}
	}

	var entries []ChurnEntry
	for name, t := range seen {
		if t.count < threshold {
			continue
		}
		status := "removed"
		entry := ChurnEntry{
			RuleName:    name,
			ChangeCount: t.count,
			FirstSeen:   t.first,
			LastSeen:    t.last,
			Status:      status,
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].ChangeCount != entries[j].ChangeCount {
			return entries[i].ChangeCount > entries[j].ChangeCount
		}
		return entries[i].RuleName < entries[j].RuleName
	})

	return entries
}
