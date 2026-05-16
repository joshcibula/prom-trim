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

func writeVelocityHistory(t *testing.T, entries []rules.HistoryEntry) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	if err := rules.AppendHistory(path, entries[0]); err != nil {
		t.Fatalf("seed history: %v", err)
	}
	for _, e := range entries[1:] {
		if err := rules.AppendHistory(path, e); err != nil {
			t.Fatalf("append history: %v", err)
		}
	}
	return path
}

func captureVelocityOutput(t *testing.T, args []string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	_ = runVelocity(args)
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestRunVelocity_TableOutput(t *testing.T) {
	base := time.Now().Add(-20 * 24 * time.Hour)
	var history []rules.HistoryEntry
	for i := 0; i < 6; i++ {
		history = append(history, rules.HistoryEntry{
			At:     rules.JSONTime(base.Add(time.Duration(i) * 24 * time.Hour)),
			Pruned: []string{"ns:rule_a"},
		})
	}
	path := writeVelocityHistory(t, history)
	out := captureVelocityOutput(t, []string{"-history", path})
	if !strings.Contains(out, "RULE") {
		t.Errorf("expected header in output, got: %s", out)
	}
}

func TestRunVelocity_JSONOutput(t *testing.T) {
	base := time.Now().Add(-20 * 24 * time.Hour)
	var history []rules.HistoryEntry
	for i := 0; i < 4; i++ {
		history = append(history, rules.HistoryEntry{
			At:     rules.JSONTime(base.Add(time.Duration(i) * 24 * time.Hour)),
			Pruned: []string{"ns:rule_b"},
		})
	}
	path := writeVelocityHistory(t, history)
	out := captureVelocityOutput(t, []string{"-history", path, "-format", "json"})
	var parsed []rules.VelocityEntry
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, out)
	}
}

func TestRunVelocity_MissingFile(t *testing.T) {
	out := captureVelocityOutput(t, []string{"-history", "/nonexistent/history.json"})
	// missing file is tolerated — prints empty table
	if !strings.Contains(out, "no velocity data") && !strings.Contains(out, "RULE") {
		t.Errorf("unexpected output for missing file: %s", out)
	}
}
