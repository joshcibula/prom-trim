package rules

import (
	"testing"
	"time"
)

func makeVelocityHistory(names []string, count int) []HistoryEntry {
	var entries []HistoryEntry
	base := time.Now().Add(-30 * 24 * time.Hour)
	for i := 0; i < count; i++ {
		entries = append(entries, HistoryEntry{
			At:     JSONTime(base.Add(time.Duration(i) * 24 * time.Hour)),
			Pruned: names,
		})
	}
	return entries
}

func TestBuildVelocityReport_Empty(t *testing.T) {
	result := BuildVelocityReport(nil, DefaultVelocityConfig())
	if result != nil {
		t.Errorf("expected nil for empty history, got %v", result)
	}
}

func TestBuildVelocityReport_SingleEntry(t *testing.T) {
	h := makeVelocityHistory([]string{"rule:a"}, 1)
	result := BuildVelocityReport(h, DefaultVelocityConfig())
	// single entry per rule — not enough for velocity
	if len(result) != 0 {
		t.Errorf("expected 0 entries for single history row, got %d", len(result))
	}
}

func TestBuildVelocityReport_StableTrend(t *testing.T) {
	h := makeVelocityHistory([]string{"rule:stable"}, 10)
	result := BuildVelocityReport(h, DefaultVelocityConfig())
	if len(result) == 0 {
		t.Fatal("expected at least one entry")
	}
	for _, e := range result {
		if e.RuleName == "rule:stable" && e.Trend != "stable" {
			t.Errorf("expected stable trend, got %s", e.Trend)
		}
	}
}

func TestBuildVelocityReport_TrendClassification(t *testing.T) {
	cfg := VelocityConfig{
		RisingThreshold:  0.10,
		FallingThreshold: -0.10,
		WindowSplit:      0.5,
	}
	tests := []struct {
		delta    float64
		expected string
	}{
		{0.5, "rising"},
		{-0.5, "falling"},
		{0.0, "stable"},
	}
	for _, tt := range tests {
		got := classifyVelocityTrend(tt.delta, cfg)
		if got != tt.expected {
			t.Errorf("delta %.2f: expected %s, got %s", tt.delta, tt.expected, got)
		}
	}
}

func TestVelocityEntry_String(t *testing.T) {
	e := VelocityEntry{RuleName: "ns:rule", Delta: -0.75, Trend: "falling"}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
	for _, sub := range []string{"ns:rule", "-0.75", "falling"} {
		if !contains(s, sub) {
			t.Errorf("expected %q in output %q", sub, s)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
