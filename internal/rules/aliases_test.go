package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeAliasRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeAliasRules: %v", err)
	}
	return p
}

const aliasRulesYAML = `
groups:
  - name: test
    rules:
      - record: base:metric
        expr: sum(http_requests_total)
      - record: derived:metric
        expr: base:metric * 2
      - record: also:derived
        expr: base:metric / 100
      - record: standalone:metric
        expr: up
`

func TestBuildAliasReport_DetectsAliases(t *testing.T) {
	p := writeAliasRules(t, aliasRulesYAML)
	entries, err := BuildAliasReport(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one alias entry")
	}
	if entries[0].Name != "base:metric" {
		t.Errorf("expected base:metric first, got %s", entries[0].Name)
	}
	if entries[0].UsageCount != 2 {
		t.Errorf("expected usage count 2, got %d", entries[0].UsageCount)
	}
}

func TestBuildAliasReport_StandaloneNotIncluded(t *testing.T) {
	p := writeAliasRules(t, aliasRulesYAML)
	entries, err := BuildAliasReport(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range entries {
		if e.Name == "standalone:metric" {
			t.Errorf("standalone:metric should not appear in alias report")
		}
	}
}

func TestBuildAliasReport_FileNotFound(t *testing.T) {
	_, err := BuildAliasReport("/nonexistent/path/rules.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildAliasReport_Empty(t *testing.T) {
	p := writeAliasRules(t, "groups:\n  - name: empty\n    rules: []\n")
	entries, err := BuildAliasReport(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty report, got %d entries", len(entries))
	}
}

func TestAliasEntry_String(t *testing.T) {
	e := AliasEntry{Name: "foo:bar", UsedBy: []string{"a", "b"}, UsageCount: 2}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
