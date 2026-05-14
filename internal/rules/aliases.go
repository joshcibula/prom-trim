package rules

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
	"os"
)

// AliasEntry represents a recording rule that is referenced by another rule's expression.
type AliasEntry struct {
	Name       string   // the recording rule metric name
	UsedBy     []string // names of rules whose expressions reference this metric
	UsageCount int
}

func (a AliasEntry) String() string {
	return fmt.Sprintf("%s (used by %d rule(s))", a.Name, a.UsageCount)
}

// BuildAliasReport parses the given rules file and returns a list of recording
// rules that are referenced (aliased) within other rules' expressions, sorted
// by usage count descending.
func BuildAliasReport(path string) ([]AliasEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("aliases: read file: %w", err)
	}

	var root struct {
		Groups []struct {
			Rules []struct {
				Record string `yaml:"record"`
				Expr   string `yaml:"expr"`
			} `yaml:"rules"`
		} `yaml:"groups"`
	}

	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("aliases: parse yaml: %w", err)
	}

	// Collect all record names and all (name, expr) pairs.
	type ruleInfo struct {
		name string
		expr string
	}
	var allRules []ruleInfo
	recordNames := map[string]bool{}

	for _, g := range root.Groups {
		for _, r := range g.Rules {
			if r.Record == "" {
				continue
			}
			recordNames[r.Record] = true
			allRules = append(allRules, ruleInfo{name: r.Record, expr: r.Expr})
		}
	}

	// For each record name, find rules whose expressions reference it.
	usedBy := map[string][]string{}
	for _, candidate := range allRules {
		for _, r := range allRules {
			if r.name == candidate.name {
				continue
			}
			if strings.Contains(r.expr, candidate.name) {
				usedBy[candidate.name] = append(usedBy[candidate.name], r.name)
			}
		}
	}

	var entries []AliasEntry
	for name, users := range usedBy {
		entries = append(entries, AliasEntry{
			Name:       name,
			UsedBy:     users,
			UsageCount: len(users),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].UsageCount != entries[j].UsageCount {
			return entries[i].UsageCount > entries[j].UsageCount
		}
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}
