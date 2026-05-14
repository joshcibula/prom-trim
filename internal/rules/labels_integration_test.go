package rules_test

import (
	"os"
	"testing"

	"github.com/yourorg/prom-trim/internal/rules"
)

const integrationLabelRules = `
groups:
  - name: infra
    rules:
      - record: node:cpu:rate5m
        expr: rate(node_cpu_seconds_total[5m])
        labels:
          tier: infra
          region: us-east-1
      - record: node:mem:used
        expr: node_memory_MemUsed_bytes
        labels:
          tier: infra
      - record: app:latency:p99
        expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
        labels:
          tier: app
          region: us-east-1
`

func TestLabels_ExtractAndIndex_Integration(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "integration-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(integrationLabelRules)
	f.Close()

	summaries, err := rules.ExtractLabels(f.Name())
	if err != nil {
		t.Fatalf("ExtractLabels: %v", err)
	}
	if len(summaries) != 3 {
		t.Fatalf("expected 3 summaries, got %d", len(summaries))
	}

	idx := rules.BuildLabelIndex(summaries)

	// 'tier' should appear in all 3 rules
	if len(idx["tier"]) != 3 {
		t.Errorf("expected 3 rules for tier, got %d", len(idx["tier"]))
	}

	// 'region' should appear in 2 rules
	if len(idx["region"]) != 2 {
		t.Errorf("expected 2 rules for region, got %d", len(idx["region"]))
	}

	keys := rules.LabelKeys(summaries)
	if len(keys) != 2 {
		t.Errorf("expected 2 distinct keys, got %d: %v", len(keys), keys)
	}
	if keys[0] != "region" || keys[1] != "tier" {
		t.Errorf("unexpected key order: %v", keys)
	}
}
