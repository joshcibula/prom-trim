package rules

import (
	"os"
	"testing"
	"time"
)

func writeAgingRules(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "aging-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

const agingRulesYAML = `
groups:
  - name: team_a
    rules:
      - record: job:requests:rate5m
        expr: rate(http_requests_total[5m])
      - record: job:errors:rate5m
        expr: rate(http_errors_total[5m])
`

func TestBuildAgingReport_FileNotFound(t *testing.T) {
	_, err := BuildAgingReport("/no/such/file.yaml", nil, DefaultAgingConfig())
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildAgingReport_Empty(t *testing.T) {
	path := writeAgingRules(t, "groups: []")
	entries, err := BuildAgingReport(path, nil, DefaultAgingConfig())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestBuildAgingReport_NeverSeen(t *testing.T) {
	path := writeAgingRules(t, agingRulesYAML)
	cfg := DefaultAgingConfig()
	entries, err := BuildAgingReport(path, map[string]UsageStats{}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Risk != "high" {
			t.Errorf("expected high risk for never-seen rule %s, got %s", e.RuleName, e.Risk)
		}
	}
}

func TestBuildAgingReport_RiskClassification(t *testing.T) {
	path := writeAgingRules(t, agingRulesYAML)
	cfg := DefaultAgingConfig()
	now := time.Now()
	usage := map[string]UsageStats{
		"job:requests:rate5m": {Count: 100, LastSeen: now.Add(-2 * 24 * time.Hour)},
		"job:errors:rate5m":   {Count: 1, LastSeen: now.Add(-20 * 24 * time.Hour)},
	}
	entries, err := BuildAgingReport(path, usage, cfg)
	if err != nil {
		t.Fatal(err)
	}
	byName := make(map[string]AgingEntry)
	for _, e := range entries {
		byName[e.RuleName] = e
	}
	if byName["job:requests:rate5m"].Risk != "low" {
		t.Errorf("expected low risk, got %s", byName["job:requests:rate5m"].Risk)
	}
	if byName["job:errors:rate5m"].Risk != "medium" {
		t.Errorf("expected medium risk, got %s", byName["job:errors:rate5m"].Risk)
	}
}

func TestBuildAgingReport_SortedByAgeDays(t *testing.T) {
	path := writeAgingRules(t, agingRulesYAML)
	cfg := DefaultAgingConfig()
	now := time.Now()
	usage := map[string]UsageStats{
		"job:requests:rate5m": {Count: 10, LastSeen: now.Add(-5 * 24 * time.Hour)},
		"job:errors:rate5m":   {Count: 10, LastSeen: now.Add(-25 * 24 * time.Hour)},
	}
	entries, err := BuildAgingReport(path, usage, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 2 {
		t.Fatal("expected at least 2 entries")
	}
	if entries[0].AgeDays < entries[1].AgeDays {
		t.Error("entries should be sorted descending by AgeDays")
	}
}

func TestAgingEntry_String(t *testing.T) {
	e := AgingEntry{RuleName: "foo:bar", GroupName: "g", AgeDays: 10, Risk: "low"}
	s := e.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}
