package rules

import (
	"testing"
	"time"
)

func makeStaleEntry(name string, lastSeen time.Time) StaleEntry {
	var usage *RuleUsage
	if !lastSeen.IsZero() {
		usage = &RuleUsage{Name: name, LastSeen: lastSeen}
	}
	return StaleEntry{Name: name, Usage: usage}
}

func TestBuildRetentionReport_HighPriority(t *testing.T) {
	policy := DefaultRetentionPolicy()
	old := time.Now().AddDate(0, 0, -60)
	entries := []StaleEntry{makeStaleEntry("old:rule", old)}

	report := BuildRetentionReport(entries, policy)

	if len(report) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report))
	}
	if report[0].Priority != "high" {
		t.Errorf("expected high priority, got %s", report[0].Priority)
	}
	if !report[0].Eligible {
		t.Error("expected eligible=true for old rule")
	}
}

func TestBuildRetentionReport_NeverSeen(t *testing.T) {
	policy := DefaultRetentionPolicy()
	entries := []StaleEntry{{Name: "ghost:rule", Usage: nil}}

	report := BuildRetentionReport(entries, policy)

	if len(report) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report))
	}
	if report[0].Priority != "high" {
		t.Errorf("expected high priority for never-seen rule, got %s", report[0].Priority)
	}
	if !report[0].Eligible {
		t.Error("expected eligible=true for never-seen rule")
	}
}

func TestBuildRetentionReport_MediumPriority(t *testing.T) {
	policy := DefaultRetentionPolicy()
	recent := time.Now().AddDate(0, 0, -10)
	entries := []StaleEntry{makeStaleEntry("mid:rule", recent)}

	report := BuildRetentionReport(entries, policy)

	if report[0].Priority != "medium" {
		t.Errorf("expected medium priority, got %s", report[0].Priority)
	}
}

func TestBuildRetentionReport_LowPriority(t *testing.T) {
	policy := DefaultRetentionPolicy()
	veryRecent := time.Now().AddDate(0, 0, -2)
	entries := []StaleEntry{makeStaleEntry("new:rule", veryRecent)}

	report := BuildRetentionReport(entries, policy)

	if report[0].Priority != "low" {
		t.Errorf("expected low priority, got %s", report[0].Priority)
	}
	if report[0].Eligible {
		t.Error("expected eligible=false for recently-seen rule")
	}
}

func TestBuildRetentionReport_SortedByAgeDays(t *testing.T) {
	policy := DefaultRetentionPolicy()
	entries := []StaleEntry{
		makeStaleEntry("young:rule", time.Now().AddDate(0, 0, -5)),
		makeStaleEntry("old:rule", time.Now().AddDate(0, 0, -45)),
		makeStaleEntry("mid:rule", time.Now().AddDate(0, 0, -15)),
	}

	report := BuildRetentionReport(entries, policy)

	if report[0].Name != "old:rule" {
		t.Errorf("expected old:rule first, got %s", report[0].Name)
	}
	if report[2].Name != "young:rule" {
		t.Errorf("expected young:rule last, got %s", report[2].Name)
	}
}

func TestRetentionEntry_String(t *testing.T) {
	e := RetentionEntry{Name: "foo:bar", AgeDays: 10, Priority: "medium", Eligible: true}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string representation")
	}
}
