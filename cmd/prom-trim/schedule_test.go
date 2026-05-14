package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/prom-trim/internal/rules"
)

func writeScheduleFile(t *testing.T, dir string, s rules.Schedule) string {
	t.Helper()
	path := filepath.Join(dir, "schedule.json")
	data, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(path, data, 0o644)
	return path
}

func TestRunSchedule_PrintsTable(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	s := rules.Schedule{
		Entries: []rules.ScheduleEntry{
			{RuleName: "job:req_rate", ScheduledAt: now, PruneAfter: now.Add(48 * time.Hour), Reason: "low usage"},
		},
	}
	path := writeScheduleFile(t, dir, s)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runSchedule([]string{"--schedule-file", path})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf strings.Builder
	data := make([]byte, 4096)
	n, _ := r.Read(data)
	buf.Write(data[:n])
	out := buf.String()

	if !strings.Contains(out, "job:req_rate") {
		t.Errorf("expected rule name in output, got:\n%s", out)
	}
	if !strings.Contains(out, "RULE") {
		t.Errorf("expected header in output, got:\n%s", out)
	}
}

func TestRunSchedule_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-schedule.json")

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runSchedule([]string{"--schedule-file", path})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error for missing schedule: %v", err)
	}
}

func TestRunSchedule_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	s := rules.Schedule{
		Entries: []rules.ScheduleEntry{
			{RuleName: "svc:errors", ScheduledAt: now, PruneAfter: now.Add(24 * time.Hour), Reason: "zero queries"},
		},
	}
	path := writeScheduleFile(t, dir, s)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runSchedule([]string{"--schedule-file", path, "--format", "json"})
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out rules.Schedule
	if err := json.NewDecoder(r).Decode(&out); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(out.Entries) != 1 || out.Entries[0].RuleName != "svc:errors" {
		t.Errorf("unexpected output: %+v", out)
	}
}
