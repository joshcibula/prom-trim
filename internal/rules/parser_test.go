package rules

import (
	"os"
	"path/filepath"
	"testing"
)

const validRulesYAML = `
groups:
  - name: example
    rules:
      - record: job:requests:rate5m
        expr: rate(http_requests_total[5m])
      - record: job:errors:rate5m
        expr: rate(http_errors_total[5m])
  - name: infra
    rules:
      - record: node:cpu:avg
        expr: avg(rate(node_cpu_seconds_total[1m]))
`

func writeTempRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writing temp rules file: %v", err)
	}
	return p
}

func TestParseFile_Valid(t *testing.T) {
	p := writeTempRules(t, validRulesYAML)
	rf, err := ParseFile(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rf.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(rf.Groups))
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/rules.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestAllRecordNames(t *testing.T) {
	p := writeTempRules(t, validRulesYAML)
	rf, _ := ParseFile(p)
	names := rf.AllRecordNames()
	if len(names) != 3 {
		t.Errorf("expected 3 record names, got %d", len(names))
	}
}

func TestFilterStale(t *testing.T) {
	p := writeTempRules(t, validRulesYAML)
	rf, _ := ParseFile(p)

	stale := map[string]struct{}{
		"job:errors:rate5m": {},
	}
	pruned := rf.FilterStale(stale)
	if pruned != 1 {
		t.Errorf("expected 1 pruned rule, got %d", pruned)
	}

	names := rf.AllRecordNames()
	for _, n := range names {
		if n == "job:errors:rate5m" {
			t.Error("stale rule still present after filtering")
		}
	}
}

func TestFilterStale_NoneStale(t *testing.T) {
	p := writeTempRules(t, validRulesYAML)
	rf, _ := ParseFile(p)
	pruned := rf.FilterStale(map[string]struct{}{})
	if pruned != 0 {
		t.Errorf("expected 0 pruned, got %d", pruned)
	}
}
