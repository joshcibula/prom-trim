package rules

import (
	"os"
	"testing"
	"time"
)

func writeFreshnessRules(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "freshness-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

const freshnessRulesYAML = `
groups:
  - name: infra
    rules:
      - record: job:requests:rate5m
        expr: rate(http_requests_total[5m])
      - record: job:errors:rate5m
        expr: rate(http_errors_total[5m])
      - record: job:latency:p99
        expr: histogram_quantile(0.99, http_latency_bucket)
`

func TestBuildFreshnessReport_FileNotFound(t *testing.T) {
	_, err := BuildFreshnessReport("/no/such/file.yaml", nil, DefaultFreshnessConfig())
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildFreshnessReport_NeverSeen(t *testing.T) {
	path := writeFreshnessRules(t, freshnessRulesYAML)
	entries, err := BuildFreshnessReport(path, map[string]UsageEntry{}, DefaultFreshnessConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Tier != "never-seen" {
			t.Errorf("%s: expected never-seen, got %s", e.Name, e.Tier)
		}
	}
}

func TestBuildFreshnessReport_TierClassification(t *testing.T) {
	path := writeFreshnessRules(t, freshnessRulesYAML)
	now := time.Now().UTC()
	cfg := DefaultFreshnessConfig()
	usage := map[string]UsageEntry{
		"job:requests:rate5m": {Count: 100, LastSeen: now.Add(-2 * 24 * time.Hour)},
		"job:errors:rate5m":   {Count: 10, LastSeen: now.Add(-15 * 24 * time.Hour)},
		"job:latency:p99":     {Count: 2, LastSeen: now.Add(-45 * 24 * time.Hour)},
	}
	entries, err := BuildFreshnessReport(path, usage, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{
		"job:requests:rate5m": "fresh",
		"job:errors:rate5m":   "aging",
		"job:latency:p99":     "stale",
	}
	for _, e := range entries {
		if got, ok := want[e.Name]; ok && got != e.Tier {
			t.Errorf("%s: expected tier %s, got %s", e.Name, got, e.Tier)
		}
	}
}

func TestBuildFreshnessReport_SortedByName(t *testing.T) {
	path := writeFreshnessRules(t, freshnessRulesYAML)
	entries, err := BuildFreshnessReport(path, nil, DefaultFreshnessConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(entries); i++ {
		if entries[i].Name < entries[i-1].Name {
			t.Errorf("entries not sorted: %s before %s", entries[i-1].Name, entries[i].Name)
		}
	}
}

func TestFreshnessEntry_String_NeverSeen(t *testing.T) {
	e := FreshnessEntry{Name: "foo:bar", Group: "g", Tier: "never-seen"}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
