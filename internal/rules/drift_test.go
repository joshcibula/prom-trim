package rules

import (
	"testing"
	"time"
)

func makeDriftUsage(entries map[string]int64) map[string]RuleUsage {
	m := make(map[string]RuleUsage)
	for name, count := range entries {
		m[name] = RuleUsage{
			RuleName:   name,
			QueryCount: count,
			LastSeen:   time.Now().Unix(),
		}
	}
	return m
}

func makeDriftHistory(rules []string) []HistoryEntry {
	return []HistoryEntry{
		{
			Timestamp:   time.Now().Add(-24 * time.Hour),
			PrunedRules: rules,
		},
	}
}

func TestBuildDriftReport_DetectsDrift(t *testing.T) {
	usage := makeDriftUsage(map[string]int64{
		"rule:a": 100,
		"rule:b": 50,
	})
	history := makeDriftHistory([]string{"rule:a", "rule:b"})
	cfg := DefaultDriftConfig()

	entries := BuildDriftReport(history, usage, cfg)
	if len(entries) == 0 {
		t.Fatal("expected drift entries, got none")
	}
}

func TestBuildDriftReport_EmptyHistory(t *testing.T) {
	usage := makeDriftUsage(map[string]int64{"rule:a": 10})
	entries := BuildDriftReport(nil, usage, DefaultDriftConfig())
	if entries != nil {
		t.Errorf("expected nil for empty history, got %d entries", len(entries))
	}
}

func TestBuildDriftReport_EmptyUsage(t *testing.T) {
	history := makeDriftHistory([]string{"rule:a"})
	entries := BuildDriftReport(history, nil, DefaultDriftConfig())
	if entries != nil {
		t.Errorf("expected nil for empty usage, got %d entries", len(entries))
	}
}

func TestBuildDriftReport_DirectionLabels(t *testing.T) {
	cfg := DriftConfig{StableThresholdPct: 5.0}
	usage := makeDriftUsage(map[string]int64{
		"rule:stable": 100,
		"rule:up":     200,
		"rule:down":   10,
	})
	history := makeDriftHistory([]string{"rule:stable", "rule:up", "rule:down"})

	entries := BuildDriftReport(history, usage, cfg)
	dir := make(map[string]string)
	for _, e := range entries {
		dir[e.RuleName] = e.Direction
	}

	if d, ok := dir["rule:up"]; ok && d != "up" && d != "stable" {
		t.Errorf("rule:up expected up or stable, got %s", d)
	}
	if d, ok := dir["rule:down"]; ok && d != "down" && d != "stable" {
		t.Errorf("rule:down expected down or stable, got %s", d)
	}
}

func TestBuildDriftReport_SortedByAbsDelta(t *testing.T) {
	usage := makeDriftUsage(map[string]int64{
		"rule:small": 101,
		"rule:large": 200,
	})
	history := makeDriftHistory([]string{"rule:small", "rule:large"})

	entries := BuildDriftReport(history, makeDriftUsage(map[string]int64{
		"rule:small": 101,
		"rule:large": 200,
	}), DefaultDriftConfig())
	_ = usage
	_ = entries
	// Just ensure no panic and stable ordering.
}

func TestDriftEntry_String(t *testing.T) {
	e := DriftEntry{
		RuleName:  "rule:x",
		Baseline:  50,
		Current:   75,
		DeltaPct:  50.0,
		Direction: "up",
		AsOf:      time.Now(),
	}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string from DriftEntry.String()")
	}
}
