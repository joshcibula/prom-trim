package rules

import (
	"testing"
	"time"
)

func TestIsStale_BelowMinCount(t *testing.T) {
	cfg := DefaultStalenessConfig()
	r := RuleUsageSummary{Name: "low_usage", QueryCount: 0}
	if !r.IsStale(cfg) {
		t.Error("expected rule with zero queries to be stale")
	}
}

func TestIsStale_AboveMinCount(t *testing.T) {
	cfg := DefaultStalenessConfig()
	r := RuleUsageSummary{Name: "active_rule", QueryCount: 5, LastSeen: time.Now()}
	if r.IsStale(cfg) {
		t.Error("expected rule with sufficient queries to be active")
	}
}

func TestIsStale_OldLastSeen(t *testing.T) {
	cfg := DefaultStalenessConfig()
	r := RuleUsageSummary{
		Name:       "old_rule",
		QueryCount: 10,
		LastSeen:   time.Now().Add(-30 * 24 * time.Hour), // 30 days ago
	}
	if !r.IsStale(cfg) {
		t.Error("expected rule last seen 30 days ago to be stale")
	}
}

func TestIsStale_ZeroLastSeen_NotStaleByTime(t *testing.T) {
	cfg := DefaultStalenessConfig()
	// Zero LastSeen means no time-based staleness check
	r := RuleUsageSummary{Name: "no_time", QueryCount: 2}
	if r.IsStale(cfg) {
		t.Error("expected rule with zero LastSeen and sufficient count to be active")
	}
}

func TestClassifyUsage_PartitionsCorrectly(t *testing.T) {
	cfg := DefaultStalenessConfig()
	summaries := []RuleUsageSummary{
		{Name: "active", QueryCount: 5, LastSeen: time.Now()},
		{Name: "stale_count", QueryCount: 0},
		{Name: "stale_old", QueryCount: 3, LastSeen: time.Now().Add(-14 * 24 * time.Hour)},
	}

	stale, active := ClassifyUsage(summaries, cfg)

	if len(active) != 1 {
		t.Errorf("expected 1 active rule, got %d", len(active))
	}
	if len(stale) != 2 {
		t.Errorf("expected 2 stale rules, got %d", len(stale))
	}
}

func TestClassifyUsage_AllActive(t *testing.T) {
	cfg := DefaultStalenessConfig()
	summaries := []RuleUsageSummary{
		{Name: "r1", QueryCount: 10, LastSeen: time.Now()},
		{Name: "r2", QueryCount: 3, LastSeen: time.Now()},
	}

	stale, active := ClassifyUsage(summaries, cfg)

	if len(stale) != 0 {
		t.Errorf("expected 0 stale rules, got %d", len(stale))
	}
	if len(active) != 2 {
		t.Errorf("expected 2 active rules, got %d", len(active))
	}
}

func TestDefaultStalenessConfig(t *testing.T) {
	cfg := DefaultStalenessConfig()
	if cfg.MinQueryCount != 1.0 {
		t.Errorf("expected MinQueryCount=1.0, got %f", cfg.MinQueryCount)
	}
	expectedWindow := 7 * 24 * time.Hour
	if cfg.Window != expectedWindow {
		t.Errorf("expected Window=%v, got %v", expectedWindow, cfg.Window)
	}
}
