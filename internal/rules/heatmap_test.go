package rules

import (
	"testing"
)

func makeUsageByPeriod() map[string]map[string]int64 {
	return map[string]map[string]int64{
		"2024-01-01": {"rule:a": 10, "rule:b": 0},
		"2024-01-02": {"rule:a": 20, "rule:b": 5},
		"2024-01-03": {"rule:a": 15, "rule:b": 3},
	}
}

func TestBuildHeatmapReport_Basic(t *testing.T) {
	rules := []string{"rule:a", "rule:b"}
	usage := makeUsageByPeriod()
	cfg := DefaultHeatmapConfig()

	entries := BuildHeatmapReport(rules, usage, cfg)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	a := entries[0]
	if a.RuleName != "rule:a" {
		t.Errorf("expected rule:a first, got %s", a.RuleName)
	}
	if a.Peak != 20 {
		t.Errorf("expected peak=20, got %d", a.Peak)
	}
	if a.Average < 14.9 || a.Average > 15.1 {
		t.Errorf("expected avg≈15.0, got %.2f", a.Average)
	}
}

func TestBuildHeatmapReport_BucketCount(t *testing.T) {
	rules := []string{"rule:a"}
	usage := map[string]map[string]int64{
		"2024-01-01": {"rule:a": 1},
		"2024-01-02": {"rule:a": 2},
		"2024-01-03": {"rule:a": 3},
		"2024-01-04": {"rule:a": 4},
		"2024-01-05": {"rule:a": 5},
	}
	cfg := HeatmapConfig{BucketCount: 3}

	entries := BuildHeatmapReport(rules, usage, cfg)
	if len(entries[0].Buckets) != 3 {
		t.Errorf("expected 3 buckets, got %d", len(entries[0].Buckets))
	}
	// Should keep the 3 most recent periods.
	if entries[0].Peak != 5 {
		t.Errorf("expected peak=5, got %d", entries[0].Peak)
	}
}

func TestBuildHeatmapReport_EmptyRules(t *testing.T) {
	entries := BuildHeatmapReport(nil, makeUsageByPeriod(), DefaultHeatmapConfig())
	if entries != nil {
		t.Errorf("expected nil for empty rules, got %v", entries)
	}
}

func TestBuildHeatmapReport_EmptyUsage(t *testing.T) {
	entries := BuildHeatmapReport([]string{"rule:a"}, nil, DefaultHeatmapConfig())
	if entries != nil {
		t.Errorf("expected nil for empty usage, got %v", entries)
	}
}

func TestHeatmapEntry_String(t *testing.T) {
	e := HeatmapEntry{RuleName: "rule:x", Peak: 42, Average: 7.5}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
	if s != "rule:x peak=42 avg=7.5" {
		t.Errorf("unexpected string: %s", s)
	}
}

func TestBuildHeatmapReport_MissingRuleInPeriod(t *testing.T) {
	// rule:c has no entries in any period — should have zero peak/avg.
	rules := []string{"rule:c"}
	usage := makeUsageByPeriod()
	entries := BuildHeatmapReport(rules, usage, DefaultHeatmapConfig())
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Peak != 0 {
		t.Errorf("expected peak=0 for unseen rule, got %d", entries[0].Peak)
	}
	if entries[0].Average != 0 {
		t.Errorf("expected avg=0 for unseen rule, got %.2f", entries[0].Average)
	}
}
