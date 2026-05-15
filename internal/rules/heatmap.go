package rules

import (
	"fmt"
	"sort"
	"time"
)

// HeatmapBucket represents a time-bucketed usage count for a single rule.
type HeatmapBucket struct {
	Period string
	Count  int64
}

// HeatmapEntry holds the heatmap data for a single recording rule.
type HeatmapEntry struct {
	RuleName string
	Buckets  []HeatmapBucket
	Peak     int64
	Average  float64
}

func (e HeatmapEntry) String() string {
	return fmt.Sprintf("%s peak=%d avg=%.1f", e.RuleName, e.Peak, e.Average)
}

// HeatmapConfig controls how heatmap buckets are generated.
type HeatmapConfig struct {
	BucketSize time.Duration
	BucketCount int
}

// DefaultHeatmapConfig returns sensible defaults: 7 daily buckets.
func DefaultHeatmapConfig() HeatmapConfig {
	return HeatmapConfig{
		BucketSize:  24 * time.Hour,
		BucketCount: 7,
	}
}

// BuildHeatmapReport constructs a per-rule usage heatmap from snapshot history.
// usageByPeriod maps period-label -> ruleName -> count.
func BuildHeatmapReport(
	rules []string,
	usageByPeriod map[string]map[string]int64,
	cfg HeatmapConfig,
) []HeatmapEntry {
	if len(rules) == 0 || len(usageByPeriod) == 0 {
		return nil
	}

	// Collect sorted period labels.
	periods := make([]string, 0, len(usageByPeriod))
	for p := range usageByPeriod {
		periods = append(periods, p)
	}
	sort.Strings(periods)
	if len(periods) > cfg.BucketCount {
		periods = periods[len(periods)-cfg.BucketCount:]
	}

	entries := make([]HeatmapEntry, 0, len(rules))
	for _, rule := range rules {
		buckets := make([]HeatmapBucket, 0, len(periods))
		var total int64
		var peak int64
		for _, p := range periods {
			count := usageByPeriod[p][rule]
			buckets = append(buckets, HeatmapBucket{Period: p, Count: count})
			total += count
			if count > peak {
				peak = count
			}
		}
		var avg float64
		if len(periods) > 0 {
			avg = float64(total) / float64(len(periods))
		}
		entries = append(entries, HeatmapEntry{
			RuleName: rule,
			Buckets:  buckets,
			Peak:     peak,
			Average:  avg,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RuleName < entries[j].RuleName
	})
	return entries
}
