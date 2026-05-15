package rules_test

import (
	"os"
	"testing"
	"time"

	"github.com/your-org/prom-trim/internal/rules"
)

const agingIntegrationYAML = `
groups:
  - name: infra
    rules:
      - record: node:cpu:rate1m
        expr: rate(node_cpu_seconds_total[1m])
      - record: node:mem:used_ratio
        expr: 1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes
      - record: node:disk:io_rate
        expr: rate(node_disk_io_time_seconds_total[5m])
`

func TestAgingRoundTrip_HighRiskNeverSeen(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "aging-int-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(agingIntegrationYAML)
	f.Close()

	cfg := rules.DefaultAgingConfig()
	entries, err := rules.BuildAgingReport(f.Name(), map[string]rules.UsageStats{}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Risk != "high" {
			t.Errorf("rule %s: expected high risk when never seen, got %s", e.RuleName, e.Risk)
		}
	}
}

func TestAgingRoundTrip_MixedRisk(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "aging-mix-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(agingIntegrationYAML)
	f.Close()

	now := time.Now()
	usage := map[string]rules.UsageStats{
		"node:cpu:rate1m":    {Count: 50, LastSeen: now.Add(-1 * 24 * time.Hour)},
		"node:mem:used_ratio": {Count: 2, LastSeen: now.Add(-16 * 24 * time.Hour)},
		// node:disk:io_rate intentionally absent → never seen
	}

	cfg := rules.DefaultAgingConfig()
	entries, err := rules.BuildAgingReport(f.Name(), usage, cfg)
	if err != nil {
		t.Fatal(err)
	}

	byName := make(map[string]rules.AgingEntry)
	for _, e := range entries {
		byName[e.RuleName] = e
	}

	if byName["node:cpu:rate1m"].Risk != "low" {
		t.Errorf("expected low, got %s", byName["node:cpu:rate1m"].Risk)
	}
	if byName["node:mem:used_ratio"].Risk != "medium" {
		t.Errorf("expected medium, got %s", byName["node:mem:used_ratio"].Risk)
	}
	if byName["node:disk:io_rate"].Risk != "high" {
		t.Errorf("expected high, got %s", byName["node:disk:io_rate"].Risk)
	}
}
