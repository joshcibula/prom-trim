package rules

import (
	"testing"
	"time"
)

func makeHistory(entries [][]string) []HistoryEntry {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var history []HistoryEntry
	for i, rules := range entries {
		history = append(history, HistoryEntry{
			Timestamp:   base.Add(time.Duration(i) * 24 * time.Hour),
			PrunedRules: rules,
		})
	}
	return history
}

func TestBuildChurnReport_AboveThreshold(t *testing.T) {
	h := makeHistory([][]string{
		{"job:requests:rate5m", "job:errors:rate5m"},
		{"job:requests:rate5m"},
		{"job:requests:rate5m", "job:latency:p99"},
	})

	entries := BuildChurnReport(h, 2)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry above threshold, got %d", len(entries))
	}
	if entries[0].RuleName != "job:requests:rate5m" {
		t.Errorf("expected job:requests:rate5m, got %s", entries[0].RuleName)
	}
	if entries[0].ChangeCount != 3 {
		t.Errorf("expected ChangeCount 3, got %d", entries[0].ChangeCount)
	}
}

func TestBuildChurnReport_BelowThreshold(t *testing.T) {
	h := makeHistory([][]string{
		{"job:requests:rate5m"},
	})

	entries := BuildChurnReport(h, 2)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries below threshold, got %d", len(entries))
	}
}

func TestBuildChurnReport_SortedByCount(t *testing.T) {
	h := makeHistory([][]string{
		{"a:rule", "b:rule", "c:rule"},
		{"a:rule", "b:rule"},
		{"a:rule"},
	})

	entries := BuildChurnReport(h, 1)
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(entries))
	}
	if entries[0].ChangeCount < entries[1].ChangeCount {
		t.Errorf("entries not sorted descending by ChangeCount")
	}
}

func TestBuildChurnReport_Empty(t *testing.T) {
	entries := BuildChurnReport(nil, 1)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty history, got %d", len(entries))
	}
}

func TestChurnEntry_String(t *testing.T) {
	e := ChurnEntry{
		RuleName:    "job:requests:rate5m",
		ChangeCount: 3,
		Status:      "removed",
	}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string from ChurnEntry.String()")
	}
}
