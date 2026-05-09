package rules_test

import (
	"path/filepath"
	"testing"
	"time"
)

// TestHistoryRoundTrip verifies that entries written via AppendHistory can be
// fully recovered by LoadHistory, preserving all fields.
func TestHistoryRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	now := time.Now().UTC().Truncate(time.Second)
	want := HistoryEntry{
		Timestamp:  now,
		RulesFile:  "/etc/prometheus/rules.yaml",
		TotalRules: 42,
		StaleRules: 7,
		DryRun:     true,
		Pruned:     []string{"job:requests:rate5m", "job:errors:rate5m"},
	}

	if err := AppendHistory(path, want); err != nil {
		t.Fatalf("AppendHistory: %v", err)
	}

	h, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(h.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(h.Entries))
	}
	got := h.Entries[0]
	if !got.Timestamp.Equal(want.Timestamp) {
		t.Errorf("timestamp: got %v want %v", got.Timestamp, want.Timestamp)
	}
	if got.RulesFile != want.RulesFile {
		t.Errorf("rules_file: got %q want %q", got.RulesFile, want.RulesFile)
	}
	if got.TotalRules != want.TotalRules {
		t.Errorf("total_rules: got %d want %d", got.TotalRules, want.TotalRules)
	}
	if got.StaleRules != want.StaleRules {
		t.Errorf("stale_rules: got %d want %d", got.StaleRules, want.StaleRules)
	}
	if got.DryRun != want.DryRun {
		t.Errorf("dry_run: got %v want %v", got.DryRun, want.DryRun)
	}
	if len(got.Pruned) != len(want.Pruned) {
		t.Errorf("pruned len: got %d want %d", len(got.Pruned), len(want.Pruned))
	}
}
