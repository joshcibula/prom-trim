package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// LifecycleStage represents the current stage of a recording rule.
type LifecycleStage string

const (
	StageActive     LifecycleStage = "active"
	StageWatched    LifecycleStage = "watched"
	StageCandidate  LifecycleStage = "candidate"
	StageDeprecated LifecycleStage = "deprecated"
)

// LifecycleEntry records the stage transition of a rule.
type LifecycleEntry struct {
	Rule       string         `json:"rule"`
	Stage      LifecycleStage `json:"stage"`
	Transition time.Time      `json:"transition"`
	Reason     string         `json:"reason,omitempty"`
}

func (e LifecycleEntry) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s",
		e.Rule, e.Stage, e.Transition.Format(time.DateOnly), e.Reason)
}

// BuildLifecycleReport classifies each recording rule into a lifecycle stage
// based on usage data and staleness configuration.
func BuildLifecycleReport(rulesFile string, usage map[string]UsageStats, cfg StalenessConfig) ([]LifecycleEntry, error) {
	groups, err := ParseFile(rulesFile)
	if err != nil {
		return nil, fmt.Errorf("lifecycle: parse rules: %w", err)
	}

	now := time.Now()
	var entries []LifecycleEntry

	for _, g := range groups {
		for _, r := range g.Rules {
			name := ruleRecordName(r)
			if name == "" {
				continue
			}
			u, ok := usage[name]
			stage, reason := classifyLifecycleStage(u, ok, now, cfg)
			entries = append(entries, LifecycleEntry{
				Rule:       name,
				Stage:      stage,
				Transition: now,
				Reason:     reason,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Rule < entries[j].Rule
	})
	return entries, nil
}

func classifyLifecycleStage(u UsageStats, found bool, now time.Time, cfg StalenessConfig) (LifecycleStage, string) {
	if !found || u.LastSeen.IsZero() {
		return StageCandidate, "no usage data found"
	}
	age := now.Sub(u.LastSeen)
	switch {
	case u.QueryCount >= int64(cfg.MinQueryCount) && age <= cfg.MaxAge/2:
		return StageActive, "frequently queried and recently seen"
	case u.QueryCount >= int64(cfg.MinQueryCount):
		return StageWatched, "sufficient query count but aging"
	case age > cfg.MaxAge:
		return StageDeprecated, "exceeded max age threshold"
	default:
		return StageCandidate, "low query count"
	}
}

// SaveLifecycle persists lifecycle entries to a JSON file.
func SaveLifecycle(path string, entries []LifecycleEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("lifecycle: create file: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(entries)
}

// LoadLifecycle reads lifecycle entries from a JSON file.
func LoadLifecycle(path string) ([]LifecycleEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("lifecycle: open file: %w", err)
	}
	defer f.Close()
	var entries []LifecycleEntry
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return nil, fmt.Errorf("lifecycle: decode: %w", err)
	}
	return entries, nil
}
