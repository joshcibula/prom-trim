package rules

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// VelocityEntry represents the rate-of-change in usage for a single recording rule.
type VelocityEntry struct {
	RuleName  string
	EarlyAvg  float64
	LateAvg   float64
	Delta     float64 // LateAvg - EarlyAvg
	Trend     string  // "rising", "falling", "stable"
}

func (v VelocityEntry) String() string {
	return fmt.Sprintf("%s: delta=%.2f trend=%s", v.RuleName, v.Delta, v.Trend)
}

// DefaultVelocityConfig returns sensible defaults for velocity analysis.
func DefaultVelocityConfig() VelocityConfig {
	return VelocityConfig{
		RisingThreshold:  0.10,
		FallingThreshold: -0.10,
		WindowSplit:      0.5,
	}
}

// VelocityConfig controls how velocity trends are classified.
type VelocityConfig struct {
	RisingThreshold  float64
	FallingThreshold float64
	WindowSplit      float64 // fraction of history used as "early" window
}

// BuildVelocityReport computes per-rule usage velocity from historical snapshots.
func BuildVelocityReport(history []HistoryEntry, cfg VelocityConfig) []VelocityEntry {
	if len(history) == 0 {
		return nil
	}

	// Group history entries by rule name.
	byRule := make(map[string][]HistoryEntry)
	for _, h := range history {
		for _, name := range h.Pruned {
			byRule[name] = append(byRule[name], h)
		}
	}

	// For each rule, split timeline and compute averages.
	var entries []VelocityEntry
	for name, rows := range byRule {
		if len(rows) < 2 {
			continue
		}
		sort.Slice(rows, func(i, j int) bool {
			return time.Time(rows[i].At).Before(time.Time(rows[j].At))
		})
		split := int(math.Ceil(float64(len(rows)) * cfg.WindowSplit))
		earlyAvg := avgCount(rows[:split])
		lateAvg := avgCount(rows[split:])
		delta := lateAvg - earlyAvg
		trend := classifyVelocityTrend(delta, cfg)
		entries = append(entries, VelocityEntry{
			RuleName: name,
			EarlyAvg: earlyAvg,
			LateAvg:  lateAvg,
			Delta:    delta,
			Trend:    trend,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Delta < entries[j].Delta
	})
	return entries
}

func classifyVelocityTrend(delta float64, cfg VelocityConfig) string {
	switch {
	case delta >= cfg.RisingThreshold:
		return "rising"
	case delta <= cfg.FallingThreshold:
		return "falling"
	default:
		return "stable"
	}
}

func avgCount(rows []HistoryEntry) float64 {
	if len(rows) == 0 {
		return 0
	}
	sum := 0
	for _, r := range rows {
		sum += len(r.Pruned)
	}
	return float64(sum) / float64(len(rows))
}
