package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeOrphansRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestBuildOrphansReport_DetectsOrphan(t *testing.T) {
	path := writeOrphansRules(t, `
groups:
  - name: g1
    rules:
      - record: job:requests:rate5m
        expr: rate(http_requests_total[5m])
      - record: job:errors:rate5m
        expr: rate(http_errors_total[5m])
      - record: job:error_ratio
        expr: job:errors:rate5m / job:requests:rate5m
`)
	orphan, err := BuildOrphansReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// job:error_ratio is not referenced by anyone
	if len(orphans) != 1 {
		t.Fatalf("expected 1 orphan, got %d", len(orphans))
	}
	if orphans[0].Record != "job:error_ratio" {
		t.Errorf("expected job:error_ratio, got %s", orphans[0].Record)
	}
}

func TestBuildOrphansReport_NoOrphans(t *testing.T) {
	path := writeOrphansRules(t, `
groups:
  - name: g1
    rules:
      - record: job:requests:rate5m
        expr: rate(http_requests_total[5m])
      - record: job:error_ratio
        expr: rate(http_errors_total[5m]) / job:requests:rate5m
`)
	orphan, err := BuildOrphansReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// job:requests:rate5m is referenced by job:error_ratio
	// job:error_ratio is a leaf — still an orphan
	if len(orphans) != 1 {
		t.Errorf("expected 1 orphan (leaf), got %d", len(orphans))
	}
}

func TestBuildOrphansReport_FileNotFound(t *testing.T) {
	_, err := BuildOrphansReport("/no/such/file.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildOrphansReport_Empty(t *testing.T) {
	path := writeOrphansRules(t, `groups: []`)
	orphan, err := BuildOrphansReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orphans) != 0 {
		t.Errorf("expected 0 orphans, got %d", len(orphans))
	}
}

func TestContainsToken(t *testing.T) {
	cases := []struct {
		src, target string
		want        bool
	}{
		{"job:requests:rate5m / 2", "job:requests:rate5m", true},
		{"rate(http_total[5m])", "http_total", true},
		{"foo_bar", "foo", false},
		{"foo", "foo_bar", false},
		{"", "foo", false},
	}
	for _, c := range cases {
		got := containsToken(c.src, c.target)
		if got != c.want {
			t.Errorf("containsToken(%q, %q) = %v, want %v", c.src, c.target, got, c.want)
		}
	}
}
