package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/prom-trim/internal/rules"
)

func writeHistoryFile(t *testing.T, entries []rules.HistoryEntry) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	h := rules.History{Entries: entries}
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func TestRunHistory_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	// file does not exist — should print "No history entries found."
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runHistory(path)
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var buf bytes.Buffer
	buf.ReadFrom(r)
	if !strings.Contains(buf.String(), "No history") {
		t.Errorf("expected 'No history' message, got: %s", buf.String())
	}
}

func TestRunHistory_PrintsRows(t *testing.T) {
	entries := []rules.HistoryEntry{
		{
			Timestamp:  time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
			RulesFile:  "rules.yaml",
			TotalRules: 20,
			StaleRules: 5,
			DryRun:     true,
			Pruned:     []string{"a", "b"},
		},
	}
	path := writeHistoryFile(t, entries)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runHistory(path)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()
	for _, want := range []string{"TIMESTAMP", "rules.yaml", "20", "5", "yes"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}
