package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeOwnershipRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExtractOwnership_Basic(t *testing.T) {
	path := writeOwnershipRules(t, `
groups:
  - name: platform
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
        annotations:
          owner: alice
          team: platform-eng
      - record: job:errors:rate5m
        expr: rate(errors_total[5m])
        annotations:
          team: platform-eng
`)
	entries, err := ExtractOwnership(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Owner != "alice" {
		t.Errorf("expected owner alice, got %q", entries[0].Owner)
	}
	if entries[0].Team != "platform-eng" {
		t.Errorf("expected team platform-eng, got %q", entries[0].Team)
	}
	if entries[1].Owner != "" {
		t.Errorf("expected empty owner for second entry, got %q", entries[1].Owner)
	}
}

func TestExtractOwnership_SkipsAlertRules(t *testing.T) {
	path := writeOwnershipRules(t, `
groups:
  - name: alerts
    rules:
      - alert: HighErrorRate
        expr: rate(errors_total[5m]) > 0.1
        annotations:
          team: sre
`)
	entries, err := ExtractOwnership(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries (alert rules skipped), got %d", len(entries))
	}
}

func TestExtractOwnership_NotFound(t *testing.T) {
	_, err := ExtractOwnership("/no/such/file.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildOwnerIndex_Basic(t *testing.T) {
	entries := []OwnershipEntry{
		{Rule: "a", Team: "alpha"},
		{Rule: "b", Team: "alpha"},
		{Rule: "c", Team: "beta"},
		{Rule: "d", Team: ""},
	}
	index := BuildOwnerIndex(entries)
	if len(index["alpha"]) != 2 {
		t.Errorf("expected 2 alpha entries, got %d", len(index["alpha"]))
	}
	if len(index["beta"]) != 1 {
		t.Errorf("expected 1 beta entry, got %d", len(index["beta"]))
	}
	if len(index["(unowned)"]) != 1 {
		t.Errorf("expected 1 unowned entry, got %d", len(index["(unowned)"]))
	}
}

func TestOwnershipEntry_String(t *testing.T) {
	e := OwnershipEntry{Rule: "job:req:rate5m", Group: "platform", Owner: "alice", Team: "eng"}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
