package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSimilarityRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestBuildSimilarityReport_DetectsSimilar(t *testing.T) {
	path := writeSimilarityRules(t, `groups:
  - name: test
    rules:
      - record: rule:a
        expr: sum(rate(http_requests_total[5m])) by (job)
      - record: rule:b
        expr: sum(rate(http_requests_total[5m])) by (env)
`)
	cfg := DefaultSimilarityConfig()
	cfg.MinScore = 0.5
	entries, err := BuildSimilarityReport(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one similar pair")
	}
	if entries[0].RuleA == "" || entries[0].RuleB == "" {
		t.Error("expected rule names to be populated")
	}
	if entries[0].Score <= 0 {
		t.Error("expected positive score")
	}
}

func TestBuildSimilarityReport_NoSimilar(t *testing.T) {
	path := writeSimilarityRules(t, `groups:
  - name: test
    rules:
      - record: rule:x
        expr: up
      - record: rule:y
        expr: kube_pod_info{namespace="default"}
`)
	cfg := DefaultSimilarityConfig()
	entries, err := BuildSimilarityReport(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no similar pairs, got %d", len(entries))
	}
}

func TestBuildSimilarityReport_FileNotFound(t *testing.T) {
	_, err := BuildSimilarityReport("/nonexistent/path.yaml", DefaultSimilarityConfig())
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildSimilarityReport_Empty(t *testing.T) {
	path := writeSimilarityRules(t, `groups:
  - name: empty
    rules: []
`)
	entries, err := BuildSimilarityReport(path, DefaultSimilarityConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty report, got %d entries", len(entries))
	}
}

func TestBuildSimilarityReport_SortedByScore(t *testing.T) {
	path := writeSimilarityRules(t, `groups:
  - name: test
    rules:
      - record: rule:a
        expr: sum(rate(http_requests_total[5m])) by (job)
      - record: rule:b
        expr: sum(rate(http_requests_total[5m])) by (job, env)
      - record: rule:c
        expr: sum(rate(http_requests_total[5m])) by (job, env, region)
`)
	cfg := SimilarityConfig{MinScore: 0.3}
	entries, err := BuildSimilarityReport(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(entries); i++ {
		if entries[i].Score > entries[i-1].Score {
			t.Errorf("entries not sorted by score desc at index %d", i)
		}
	}
}
