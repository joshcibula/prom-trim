package rules

import (
	"os"
	"testing"
)

func writeComplexityRules(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "complexity-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

const complexityRulesYAML = `groups:
  - name: test_group
    rules:
      - record: simple:rule
        expr: up
      - record: medium:rule
        expr: sum by (job) (rate(http_requests_total[5m]))
      - record: complex:rule
        expr: sum by (job, instance) (rate(http_errors_total[5m])) / sum by (job, instance) (rate(http_requests_total[5m]))
`

func TestBuildComplexityReport_SortedByTokens(t *testing.T) {
	path := writeComplexityRules(t, complexityRulesYAML)
	stats, err := BuildComplexityReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 3 {
		t.Fatalf("expected 3 stats, got %d", len(stats))
	}
	// Must be sorted descending by token count.
	for i := 1; i < len(stats); i++ {
		if stats[i].TokenCount > stats[i-1].TokenCount {
			t.Errorf("stats not sorted: index %d (%d) > index %d (%d)",
				i, stats[i].TokenCount, i-1, stats[i-1].TokenCount)
		}
	}
}

func TestBuildComplexityReport_Levels(t *testing.T) {
	path := writeComplexityRules(t, complexityRulesYAML)
	stats, err := BuildComplexityReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	levels := map[string]ComplexityLevel{}
	for _, s := range stats {
		levels[s.Record] = s.Level
	}
	if levels["simple:rule"] != ComplexityLow {
		t.Errorf("expected simple:rule to be low, got %s", levels["simple:rule"])
	}
	if levels["medium:rule"] != ComplexityMedium {
		t.Errorf("expected medium:rule to be medium, got %s", levels["medium:rule"])
	}
	if levels["complex:rule"] != ComplexityHigh {
		t.Errorf("expected complex:rule to be high, got %s", levels["complex:rule"])
	}
}

func TestBuildComplexityReport_FileNotFound(t *testing.T) {
	_, err := BuildComplexityReport("/no/such/file.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestBuildComplexityReport_Empty(t *testing.T) {
	path := writeComplexityRules(t, "groups: []\n")
	stats, err := BuildComplexityReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats, got %d", len(stats))
	}
}

func TestComplexityStat_String(t *testing.T) {
	s := ComplexityStat{Group: "g", Record: "r", TokenCount: 3, Level: ComplexityLow}
	out := s.String()
	if out == "" {
		t.Error("expected non-empty string representation")
	}
}
