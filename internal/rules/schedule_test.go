package rules

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadSchedule_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schedule.json")

	now := time.Now().UTC().Truncate(time.Second)
	s := Schedule{
		Entries: []ScheduleEntry{
			{RuleName: "rule:a", ScheduledAt: now, PruneAfter: now.Add(48 * time.Hour), Reason: "low usage"},
		},
	}

	if err := SaveSchedule(path, s); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadSchedule(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].RuleName != "rule:a" {
		t.Errorf("unexpected rule name: %s", loaded.Entries[0].RuleName)
	}
}

func TestLoadSchedule_NotFound(t *testing.T) {
	s, err := LoadSchedule("/nonexistent/schedule.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(s.Entries) != 0 {
		t.Errorf("expected empty schedule")
	}
}

func TestLoadSchedule_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte("not-json"), 0o644)

	_, err := LoadSchedule(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSchedule_AddEntry_Deduplicates(t *testing.T) {
	now := time.Now()
	s := Schedule{}
	s.AddEntry(ScheduleEntry{RuleName: "rule:a", PruneAfter: now.Add(24 * time.Hour)})
	s.AddEntry(ScheduleEntry{RuleName: "rule:a", PruneAfter: now.Add(48 * time.Hour)})

	if len(s.Entries) != 1 {
		t.Fatalf("expected 1 entry after dedup, got %d", len(s.Entries))
	}
	if s.Entries[0].PruneAfter != now.Add(48*time.Hour) {
		t.Errorf("expected updated prune_after")
	}
}

func TestSchedule_DueEntries(t *testing.T) {
	now := time.Now()
	s := Schedule{
		Entries: []ScheduleEntry{
			{RuleName: "past", PruneAfter: now.Add(-1 * time.Hour)},
			{RuleName: "future", PruneAfter: now.Add(1 * time.Hour)},
			{RuleName: "exact", PruneAfter: now},
		},
	}
	due := s.DueEntries(now)
	if len(due) != 2 {
		t.Fatalf("expected 2 due entries, got %d", len(due))
	}
}
