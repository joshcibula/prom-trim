package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeRedundancyRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestBuildRedundancyReport_DetectsOverlap(t *testing.T) {
	path := writeRedundancyRules(t, `
groups:
  - name: test
    rules:
      - record: rule_a
        expr: sum(rate(http_requests_total[5m])) by (job)
      - record: rule_b
        expr: sum(rate(http_requests_total[5m])) by (job, status)
      - record: rule_c
        expr: avg(cpu_usage) by (instance)
`)

	entries, err := BuildRedundancyReport(path, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one redundancy entry")
	}
	found := false
	for _, e := range entries {
		for _, o := range e.OverlapsBy {
			if (e.Rule == "rule_a" && o == "rule_b") || (e.Rule == "rule_b" && o == "rule_a") {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected rule_a and rule_b to overlap")
	}
}

func TestBuildRedundancyReport_NoOverlap(t *testing.T) {
	path := writeRedundancyRules(t, `
groups:
  - name: test
    rules:
      - record: rule_x
        expr: sum(rate(http_requests_total[5m]))
      - record: rule_y
        expr: avg(memory_bytes) by (host)
`)

	entries, err := BuildRedundancyReport(path, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no entries, got %d", len(entries))
	}
}

func TestBuildRedundancyReport_FileNotFound(t *testing.T) {
	_, err := BuildRedundancyReport("/nonexistent/path.yaml", 0.5)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildRedundancyReport_Empty(t *testing.T) {
	path := writeRedundancyRules(t, `
groups:
  - name: empty
    rules: []
`)
	entries, err := BuildRedundancyReport(path, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestRedundancyEntry_String(t *testing.T) {
	e := RedundancyEntry{
		Rule:       "my_rule",
		OverlapsBy: []string{"other_rule", "another_rule"},
	}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
	if !contains(s, "my_rule") {
		t.Errorf("expected rule name in string, got: %s", s)
	}
}

func TestJaccardSimilarity_IdenticalSets(t *testing.T) {
	a := tokenSet("sum(rate(http_requests_total[5m]))")
	b := tokenSet("sum(rate(http_requests_total[5m]))")
	sim := jaccardSimilarity(a, b)
	if sim != 1.0 {
		t.Errorf("expected 1.0 for identical sets, got %f", sim)
	}
}

func TestJaccardSimilarity_DisjointSets(t *testing.T) {
	a := tokenSet("sum(foo)")
	b := tokenSet("avg(bar)")
	sim := jaccardSimilarity(a, b)
	if sim >= 0.5 {
		t.Errorf("expected low similarity for disjoint sets, got %f", sim)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
