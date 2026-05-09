package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadHistory_NotFound(t *testing.T) {
	h, err := LoadHistory("/nonexistent/history.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(h.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(h.Entries))
	}
}

func TestAppendHistory_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	entry := HistoryEntry{
		Timestamp:  time.Now().UTC().Truncate(time.Second),
		RulesFile:  "rules.yaml",
		TotalRules: 10,
		StaleRules: 3,
		DryRun:     false,
		Pruned:     []string{"rule_a", "rule_b", "rule_c"},
	}

	if err := AppendHistory(path, entry); err != nil {
		t.Fatalf("AppendHistory: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(h.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(h.Entries))
	}
	if h.Entries[0].StaleRules != 3 {
		t.Errorf("stale_rules mismatch: got %d", h.Entries[0].StaleRules)
	}
}

func TestAppendHistory_Accumulates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	for i := 0; i < 3; i++ {
		entry := HistoryEntry{TotalRules: i + 1, RulesFile: "rules.yaml"}
		if err := AppendHistory(path, entry); err != nil {
			t.Fatalf("AppendHistory iteration %d: %v", i, err)
		}
	}

	h, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(h.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(h.Entries))
	}
}

func TestLoadHistory_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadHistory(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
