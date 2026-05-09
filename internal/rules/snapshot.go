package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot represents a point-in-time record of rule usage data.
type Snapshot struct {
	CreatedAt time.Time         `json:"created_at"`
	RulesFile string            `json:"rules_file"`
	Usage     map[string]int64  `json:"usage"`
}

// SaveSnapshot writes a usage snapshot to the given path as JSON.
func SaveSnapshot(path, rulesFile string, usage map[string]int64) error {
	snap := Snapshot{
		CreatedAt: time.Now().UTC(),
		RulesFile: rulesFile,
		Usage:     usage,
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot: marshal failed: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("snapshot: write failed: %w", err)
	}

	return nil
}

// LoadSnapshot reads a previously saved snapshot from the given path.
func LoadSnapshot(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: read failed: %w", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("snapshot: unmarshal failed: %w", err)
	}

	return &snap, nil
}
