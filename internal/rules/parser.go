package rules

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RecordingRule represents a single Prometheus recording rule.
type RecordingRule struct {
	Record string `yaml:"record"`
	Expr   string `yaml:"expr"`
	Labels map[string]string `yaml:"labels,omitempty"`
}

// RuleGroup represents a group of Prometheus rules.
type RuleGroup struct {
	Name     string          `yaml:"name"`
	Interval string          `yaml:"interval,omitempty"`
	Rules    []RecordingRule `yaml:"rules"`
}

// RuleFile represents the top-level structure of a Prometheus rules file.
type RuleFile struct {
	Groups []RuleGroup `yaml:"groups"`
}

// ParseFile reads and parses a Prometheus rules YAML file.
func ParseFile(path string) (*RuleFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading rules file %q: %w", path, err)
	}

	var rf RuleFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, fmt.Errorf("parsing rules file %q: %w", path, err)
	}

	return &rf, nil
}

// AllRecordNames returns a deduplicated slice of all recording rule names
// across all groups in the file.
func (rf *RuleFile) AllRecordNames() []string {
	seen := make(map[string]struct{})
	var names []string
	for _, g := range rf.Groups {
		for _, r := range g.Rules {
			if r.Record == "" {
				continue
			}
			if _, ok := seen[r.Record]; !ok {
				seen[r.Record] = struct{}{}
				names = append(names, r.Record)
			}
		}
	}
	return names
}

// FilterStale removes recording rules whose names appear in the stale set
// and returns the number of rules pruned.
func (rf *RuleFile) FilterStale(stale map[string]struct{}) int {
	pruned := 0
	for i, g := range rf.Groups {
		var kept []RecordingRule
		for _, r := range g.Rules {
			if _, isStale := stale[r.Record]; isStale {
				pruned++
				continue
			}
			kept = append(kept, r)
		}
		rf.Groups[i].Rules = kept
	}
	return pruned
}
