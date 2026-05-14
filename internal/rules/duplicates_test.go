package rules

import (
	"os"
	"testing"
)

func writeDuplicatesRules(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "dup-rules-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestBuildDuplicatesReport_NoDuplicates(t *testing.T) {
	path := writeDuplicatesRules(t, `
groups:
  - name: g1
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
      - record: job:errors:rate5m
        expr: rate(errors_total[5m])
`)
	entries, err := BuildDuplicatesReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no duplicates, got %d", len(entries))
	}
}

func TestBuildDuplicatesReport_DetectsDuplicate(t *testing.T) {
	path := writeDuplicatesRules(t, `
groups:
  - name: g1
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
  - name: g2
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
`)
	entries, err := BuildDuplicatesReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 duplicate, got %d", len(entries))
	}
	if entries[0].Name != "job:requests:rate5m" {
		t.Errorf("unexpected name: %s", entries[0].Name)
	}
	if entries[0].Count != 2 {
		t.Errorf("expected count 2, got %d", entries[0].Count)
	}
	if len(entries[0].Groups) != 2 {
		t.Errorf("expected 2 groups, got %v", entries[0].Groups)
	}
}

func TestBuildDuplicatesReport_SortedByCountDesc(t *testing.T) {
	path := writeDuplicatesRules(t, `
groups:
  - name: g1
    rules:
      - record: rule:alpha:total
        expr: sum(alpha)
      - record: rule:beta:total
        expr: sum(beta)
      - record: rule:beta:total
        expr: sum(beta)
      - record: rule:alpha:total
        expr: sum(alpha)
      - record: rule:alpha:total
        expr: sum(alpha)
`)
	entries, err := BuildDuplicatesReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(entries))
	}
	if entries[0].Count < entries[1].Count {
		t.Errorf("expected descending count order, got %d then %d", entries[0].Count, entries[1].Count)
	}
}

func TestBuildDuplicatesReport_FileNotFound(t *testing.T) {
	_, err := BuildDuplicatesReport("/no/such/file.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestDuplicateEntry_String(t *testing.T) {
	e := DuplicateEntry{Name: "job:req:rate5m", Count: 3, Groups: []string{"g1", "g2"}}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
