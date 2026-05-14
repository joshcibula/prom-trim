package rules

import (
	"testing"
)

func makeStaleEntries() []DiffEntry {
	return []DiffEntry{
		{RuleName: "rule_high", Group: "grp1", Status: "removed"},
		{RuleName: "rule_medium", Group: "grp1", Status: "removed"},
		{RuleName: "rule_low", Group: "grp2", Status: "removed"},
		{RuleName: "rule_zero", Group: "grp2", Status: "removed"},
	}
}

func TestAssessImpact_Levels(t *testing.T) {
	usage := map[string]int64{
		"rule_high":   150,
		"rule_medium": 42,
		"rule_low":    3,
		"rule_zero":   0,
	}

	report := AssessImpact(makeStaleEntries(), usage)

	if len(report) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(report))
	}

	if report[0].Level != ImpactHigh {
		t.Errorf("expected first entry to be high impact, got %s", report[0].Level)
	}
	if report[1].Level != ImpactMedium {
		t.Errorf("expected second entry to be medium impact, got %s", report[1].Level)
	}
	if report[2].Level != ImpactLow {
		t.Errorf("expected third entry to be low impact, got %s", report[2].Level)
	}
	if report[3].Level != ImpactLow {
		t.Errorf("expected fourth entry to be low impact, got %s", report[3].Level)
	}
}

func TestAssessImpact_SortedByCount(t *testing.T) {
	usage := map[string]int64{
		"rule_high":   500,
		"rule_medium": 20,
		"rule_low":    1,
		"rule_zero":   0,
	}

	report := AssessImpact(makeStaleEntries(), usage)

	for i := 1; i < len(report); i++ {
		if report[i].QueryCount > report[i-1].QueryCount {
			t.Errorf("report not sorted descending at index %d", i)
		}
	}
}

func TestAssessImpact_MissingUsage(t *testing.T) {
	// Rules not present in usage map should default to 0 / low impact.
	entries := []DiffEntry{
		{RuleName: "unknown_rule", Group: "grp", Status: "removed"},
	}
	report := AssessImpact(entries, map[string]int64{})

	if len(report) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report))
	}
	if report[0].Level != ImpactLow {
		t.Errorf("expected low impact for unknown rule, got %s", report[0].Level)
	}
	if report[0].QueryCount != 0 {
		t.Errorf("expected query count 0, got %d", report[0].QueryCount)
	}
}

func TestImpactEntry_String(t *testing.T) {
	e := ImpactEntry{RuleName: "my_rule", Group: "grp", QueryCount: 5, Level: ImpactLow}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string from ImpactEntry.String()")
	}
}

func TestAssessImpact_Empty(t *testing.T) {
	report := AssessImpact([]DiffEntry{}, map[string]int64{})
	if len(report) != 0 {
		t.Errorf("expected empty report, got %d entries", len(report))
	}
}
