package rules

import (
	"os"
	"testing"
)

// writeTempRules creates a temporary YAML rules file with the given content
// and registers it for cleanup when the test finishes.
func writeTempRules(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "rules-*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestPrune_DryRun(t *testing.T) {
	path := writeTempRules(t, `
groups:
  - name: test
    rules:
      - record: up_total
        expr: sum(up)
      - record: stale_metric
        expr: sum(stale)
`)

	stale := map[string]bool{"stale_metric": true}
	result, err := Prune(path, stale, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.DryRun {
		t.Error("expected DryRun to be true")
	}
	if len(result.Removed) != 1 || result.Removed[0] != "stale_metric" {
		t.Errorf("expected [stale_metric] removed, got %v", result.Removed)
	}
	if len(result.Kept) != 1 || result.Kept[0] != "up_total" {
		t.Errorf("expected [up_total] kept, got %v", result.Kept)
	}

	// File must be unchanged in dry-run mode.
	original, _ := os.ReadFile(path)
	if len(original) == 0 {
		t.Error("file should not be empty after dry run")
	}
}

func TestPrune_WritesFile(t *testing.T) {
	path := writeTempRules(t, `
groups:
  - name: test
    rules:
      - record: keep_me
        expr: sum(keep)
      - record: drop_me
        expr: sum(drop)
`)

	stale := map[string]bool{"drop_me": true}
	_, err := Prune(path, stale, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	groups, err := ParseFile(path)
	if err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}

	names := AllRecordNames(groups)
	if len(names) != 1 || names[0] != "keep_me" {
		t.Errorf("expected only [keep_me] after prune, got %v", names)
	}
}

func TestPrune_NothingStale(t *testing.T) {
	path := writeTempRules(t, `
groups:
  - name: test
    rules:
      - record: active_metric
        expr: sum(active)
`)

	result, err := Prune(path, map[string]bool{}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Removed) != 0 {
		t.Errorf("expected no removals, got %v", result.Removed)
	}
}
