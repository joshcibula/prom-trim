package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// HistoryEntry records the result of a single prom-trim run.
type HistoryEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	RulesFile  string    `json:"rules_file"`
	TotalRules int       `json:"total_rules"`
	StaleRules int       `json:"stale_rules"`
	DryRun     bool      `json:"dry_run"`
	Pruned     []string  `json:"pruned"`
}

// History holds an ordered list of run entries.
type History struct {
	Entries []HistoryEntry `json:"entries"`
}

// AppendHistory loads an existing history file (if present), appends the new
// entry, and writes the result back to disk.
func AppendHistory(path string, entry HistoryEntry) error {
	h, err := LoadHistory(path)
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}
	h.Entries = append(h.Entries, entry)
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}

// LoadHistory reads a history file from disk. If the file does not exist an
// empty History is returned without error.
func LoadHistory(path string) (*History, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &History{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read history: %w", err)
	}
	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("unmarshal history: %w", err)
	}
	return &h, nil
}
