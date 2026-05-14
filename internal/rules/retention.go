package rules

import (
	"fmt"
	"sort"
	"time"
)

// RetentionPolicy defines thresholds for classifying rule retention priority.
type RetentionPolicy struct {
	// MinAgeDays is the minimum age (in days) before a rule is eligible for removal.
	MinAgeDays int
	// MaxStaleDays is the number of days without usage after which a rule is high-priority for removal.
	MaxStaleDays int
}

// DefaultRetentionPolicy returns a sensible default retention policy.
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		MinAgeDays:   7,
		MaxStaleDays: 30,
	}
}

// RetentionEntry describes the retention classification of a single recording rule.
type RetentionEntry struct {
	Name      string
	LastSeen  time.Time
	AgeDays   int
	Priority  string // "low", "medium", "high"
	Eligible  bool
}

func (r RetentionEntry) String() string {
	return fmt.Sprintf("%s | age=%dd | priority=%s | eligible=%v", r.Name, r.AgeDays, r.Priority, r.Eligible)
}

// BuildRetentionReport classifies each stale rule entry according to the given policy.
func BuildRetentionReport(stale []StaleEntry, policy RetentionPolicy) []RetentionEntry {
	now := time.Now()
	result := make([]RetentionEntry, 0, len(stale))

	for _, s := range stale {
		var lastSeen time.Time
		if s.Usage != nil {
			lastSeen = s.Usage.LastSeen
		}

		var ageDays int
		if !lastSeen.IsZero() {
			ageDays = int(now.Sub(lastSeen).Hours() / 24)
		}

		eligible := ageDays >= policy.MinAgeDays || lastSeen.IsZero()

		priority := classifyRetentionPriority(ageDays, lastSeen.IsZero(), policy)

		result = append(result, RetentionEntry{
			Name:     s.Name,
			LastSeen: lastSeen,
			AgeDays:  ageDays,
			Priority: priority,
			Eligible: eligible,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].AgeDays > result[j].AgeDays
	})

	return result
}

func classifyRetentionPriority(ageDays int, neverSeen bool, policy RetentionPolicy) string {
	if neverSeen {
		return "high"
	}
	if ageDays >= policy.MaxStaleDays {
		return "high"
	}
	if ageDays >= policy.MinAgeDays {
		return "medium"
	}
	return "low"
}
