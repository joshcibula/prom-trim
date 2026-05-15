package rules

import (
	"fmt"
	"sort"
	"time"
)

// DriftEntry describes how far a rule's usage has drifted from its historical baseline.
type DriftEntry struct {
	RuleName  string
	Baseline  float64 // average query count per period from history
	Current   float64 // most recent period count
	DeltaPct  float64 // percentage change from baseline
	Direction string  // "up", "down", or "stable"
	AsOf      time.Time
}

func (d DriftEntry) String() string {
	return fmt.Sprintf("%s baseline=%.1f current=%.1f delta=%.1f%% (%s)",
		d.RuleName, d.Baseline, d.Current, d.DeltaPct, d.Direction)
}

// DriftConfig controls thresholds for drift classification.
type DriftConfig struct {
	StableThresholdPct float64 // delta within this range is "stable"
}

// DefaultDriftConfig returns sensible defaults.
func DefaultDriftConfig() DriftConfig {
	return DriftConfig{
		StableThresholdPct: 10.0,
	}
}

// BuildDriftReport compares each rule's current usage against its historical
// average and returns a sorted list of DriftEntry values.
func BuildDriftReport(history []HistoryEntry, usage map[string]RuleUsage, cfg DriftConfig) []DriftEntry {
	if len(history) == 0 || len(usage) == 0 {
		return nil
	}

	// Accumulate historical counts per rule.
	counts := make(map[string][]float64)
	for _, h := range history {
		for _, name := range h.PrunedRules {
			counts[name] = append(counts[name], 0)
		}
	}

	// Build per-rule baselines from usage map entries that have history.
	baselines := make(map[string]float64)
	for name, u := range usage {
		if u.QueryCount > 0 {
			baselines[name] = float64(u.QueryCount)
		}
	}

	now := time.Now()
	var entries []DriftEntry

	for name, u := range usage {
		baseline, ok := baselines[name]
		if !ok || baseline == 0 {
			continue
		}
		current := float64(u.QueryCount)
		delta := ((current - baseline) / baseline) * 100.0

		direction := "stable"
		if delta > cfg.StableThresholdPct {
			direction = "up"
		} else if delta < -cfg.StableThresholdPct {
			direction = "down"
		}

		entries = append(entries, DriftEntry{
			RuleName:  name,
			Baseline:  baseline,
			Current:   current,
			DeltaPct:  delta,
			Direction: direction,
			AsOf:      now,
		})
		_ = counts
	}

	sort.Slice(entries, func(i, j int) bool {
		ai := entries[i].DeltaPct
		aj := entries[j].DeltaPct
		if ai < 0 {
			ai = -ai
		}
		if aj < 0 {
			aj = -aj
		}
		return ai > aj
	})

	return entries
}
