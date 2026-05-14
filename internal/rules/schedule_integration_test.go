package rules

import (
	"path/filepath"
	"testing"
	"time"
)

func TestScheduleRoundTrip_WithDueEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schedule.json")

	now := time.Now().UTC().Truncate(time.Second)

	s := Schedule{}
	s.AddEntry(ScheduleEntry{
		RuleName:    "job:req_rate",
		ScheduledAt: now.Add(-72 * time.Hour),
		PruneAfter:  now.Add(-24 * time.Hour),
		Reason:      "low usage",
	})
	s.AddEntry(ScheduleEntry{
		RuleName:    "svc:error_ratio",
		ScheduledAt: now,
		PruneAfter:  now.Add(48 * time.Hour),
		Reason:      "zero queries",
	})

	if err := SaveSchedule(path, s); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadSchedule(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	due := loaded.DueEntries(now)
	if len(due) != 1 {
		t.Fatalf("expected 1 due entry, got %d", len(due))
	}
	if due[0].RuleName != "job:req_rate" {
		t.Errorf("unexpected due rule: %s", due[0].RuleName)
	}

	// Re-schedule the due entry with a new grace period
	loaded.AddEntry(ScheduleEntry{
		RuleName:    "job:req_rate",
		ScheduledAt: now,
		PruneAfter:  now.Add(7 * 24 * time.Hour),
		Reason:      "rescheduled",
	})

	if err := SaveSchedule(path, loaded); err != nil {
		t.Fatalf("resave: %v", err)
	}

	final, err := LoadSchedule(path)
	if err != nil {
		t.Fatalf("final load: %v", err)
	}
	if len(final.DueEntries(now)) != 0 {
		t.Errorf("expected no due entries after reschedule")
	}
}
