package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeGroupStatsRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp rules: %v", err)
	}
	return p
}

const groupStatsYAML = `
groups:
  - name: alpha
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
      - record: job:errors:rate5m
        expr: rate(errors_total[5m])
  - name: beta
    rules:
      - record: job:latency:p99
        expr: histogram_quantile(0.99, rate(latency_bucket[5m]))
`

func TestBuildGroupStats_Basic(t *testing.T) {
	file := writeGroupStatsRules(t, groupStatsYAML)
	usage := map[string]float64{
		"job:requests:rate5m": 42,
	}
	stats, err := BuildGroupStats(file, usage)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(stats))
	}
	alpha := stats[0]
	if alpha.GroupName != "alpha" {
		t.Errorf("expected alpha, got %s", alpha.GroupName)
	}
	if alpha.RuleCount != 2 {
		t.Errorf("expected 2 rules, got %d", alpha.RuleCount)
	}
	if alpha.StaleCount != 1 {
		t.Errorf("expected 1 stale, got %d", alpha.StaleCount)
	}
	if alpha.Coverage != 50.0 {
		t.Errorf("expected 50%% coverage, got %.1f", alpha.Coverage)
	}
}

func TestBuildGroupStats_AllStale(t *testing.T) {
	file := writeGroupStatsRules(t, groupStatsYAML)
	stats, err := BuildGroupStats(file, map[string]float64{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, s := range stats {
		if s.StaleCount != s.RuleCount {
			t.Errorf("group %s: expected all stale", s.GroupName)
		}
		if s.Coverage != 0 {
			t.Errorf("group %s: expected 0%% coverage", s.GroupName)
		}
	}
}

func TestBuildGroupStats_FileNotFound(t *testing.T) {
	_, err := BuildGroupStats("/nonexistent/rules.yaml", nil)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestGroupStat_String(t *testing.T) {
	g := GroupStat{GroupName: "mygroup", RuleCount: 5, StaleCount: 2, Coverage: 60.0}
	s := g.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
