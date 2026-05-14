package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ScheduleEntry records when a rule is scheduled for pruning.
type ScheduleEntry struct {
	RuleName    string    `json:"rule_name"`
	ScheduledAt time.Time `json:"scheduled_at"`
	PruneAfter  time.Time `json:"prune_after"`
	Reason      string    `json:"reason"`
}

// Schedule holds a collection of entries pending pruning.
type Schedule struct {
	Entries []ScheduleEntry `json:"entries"`
}

// SaveSchedule persists the schedule to a JSON file.
func SaveSchedule(path string, s Schedule) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schedule: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write schedule: %w", err)
	}
	return nil
}

// LoadSchedule reads a schedule from a JSON file.
func LoadSchedule(path string) (Schedule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Schedule{}, nil
		}
		return Schedule{}, fmt.Errorf("read schedule: %w", err)
	}
	var s Schedule
	if err := json.Unmarshal(data, &s); err != nil {
		return Schedule{}, fmt.Errorf("unmarshal schedule: %w", err)
	}
	return s, nil
}

// AddEntry appends a new entry, replacing any existing entry for the same rule.
func (s *Schedule) AddEntry(entry ScheduleEntry) {
	for i, e := range s.Entries {
		if e.RuleName == entry.RuleName {
			s.Entries[i] = entry
			return
		}
	}
	s.Entries = append(s.Entries, entry)
}

// DueEntries returns entries whose PruneAfter time is before or equal to now.
func (s *Schedule) DueEntries(now time.Time) []ScheduleEntry {
	var due []ScheduleEntry
	for _, e := range s.Entries {
		if !e.PruneAfter.After(now) {
			due = append(due, e)
		}
	}
	return due
}
