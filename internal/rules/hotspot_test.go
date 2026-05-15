package rules

import (
	"os"
	"testing"
)

func writeHotspotRules(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "hotspot-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

const hotspotRulesYAML = `
groups:
  - name: platform
    rules:
      - record: job:requests:rate5m
        expr: rate(http_requests_total[5m])
      - record: job:errors:rate5m
        expr: rate(http_errors_total[5m])
      - record: job:latency:p99
        expr: histogram_quantile(0.99, rate(http_duration_bucket[5m]))
`

func makeHotspotUsage() map[string]UsageStats {
	return map[string]UsageStats{
		"job:requests:rate5m": {QueryCount: 1200, AvgRate: 4.0},
		"job:errors:rate5m":   {QueryCount: 50, AvgRate: 0.1},
		"job:latency:p99":     {QueryCount: 600, AvgRate: 2.0},
	}
}

func TestBuildHotspotReport_AboveThreshold(t *testing.T) {
	file := writeHotspotRules(t, hotspotRulesYAML)
	entries, err := BuildHotspotReport(file, makeHotspotUsage(), DefaultHotspotThreshold)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 hotspot entries, got %d", len(entries))
	}
	// sorted descending
	if entries[0].QueryCount <= entries[1].QueryCount {
		t.Errorf("expected descending order by QueryCount")
	}
}

func TestBuildHotspotReport_TierClassification(t *testing.T) {
	file := writeHotspotRules(t, hotspotRulesYAML)
	entries, err := BuildHotspotReport(file, makeHotspotUsage(), DefaultHotspotThreshold)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tiers := map[string]string{}
	for _, e := range entries {
		tiers[e.Name] = e.Tier
	}
	if tiers["job:requests:rate5m"] != "critical" {
		t.Errorf("expected critical for count=1200, got %s", tiers["job:requests:rate5m"])
	}
	if tiers["job:latency:p99"] != "high" {
		t.Errorf("expected high for count=600, got %s", tiers["job:latency:p99"])
	}
}

func TestBuildHotspotReport_BelowThreshold(t *testing.T) {
	file := writeHotspotRules(t, hotspotRulesYAML)
	entries, err := BuildHotspotReport(file, makeHotspotUsage(), 2000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries above threshold 2000, got %d", len(entries))
	}
}

func TestBuildHotspotReport_FileNotFound(t *testing.T) {
	_, err := BuildHotspotReport("/nonexistent/path.yaml", makeHotspotUsage(), DefaultHotspotThreshold)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestHotspotEntry_String(t *testing.T) {
	e := HotspotEntry{Name: "foo:bar", Group: "g1", QueryCount: 500, AvgRate: 1.5, Tier: "high"}
	s := e.String()
	for _, want := range []string{"foo:bar", "g1", "high", "500", "1.50"} {
		if !containsToken(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}
