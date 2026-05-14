package rules

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// OwnershipEntry holds ownership metadata for a single recording rule.
type OwnershipEntry struct {
	Rule   string `json:"rule" yaml:"rule"`
	Group  string `json:"group" yaml:"group"`
	Owner  string `json:"owner" yaml:"owner"`
	Team   string `json:"team" yaml:"team"`
}

// String returns a human-readable representation of an ownership entry.
func (e OwnershipEntry) String() string {
	return fmt.Sprintf("%s (group: %s, owner: %s, team: %s)", e.Rule, e.Group, e.Owner, e.Team)
}

// ExtractOwnership parses a rules file and extracts ownership metadata from
// annotations. It looks for "owner" and "team" annotation keys.
func ExtractOwnership(path string) ([]OwnershipEntry, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	var entries []OwnershipEntry
	for _, g := range groups {
		for _, r := range g.Rules {
			if r.Record == "" {
				continue
			}
			entry := OwnershipEntry{
				Rule:  r.Record,
				Group: g.Name,
			}
			for k, v := range r.Annotations {
				switch strings.ToLower(k) {
				case "owner":
					entry.Owner = v
				case "team":
					entry.Team = v
				}
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// BuildOwnerIndex groups ownership entries by team name.
func BuildOwnerIndex(entries []OwnershipEntry) map[string][]OwnershipEntry {
	index := make(map[string][]OwnershipEntry)
	for _, e := range entries {
		key := e.Team
		if key == "" {
			key = "(unowned)"
		}
		index[key] = append(index[key], e)
	}
	return index
}

// OwnershipKeys returns a sorted list of annotation keys used for ownership
// detection — exported for documentation/help purposes.
func OwnershipKeys() []string {
	return []string{"owner", "team"}
}

// ownershipNode is a minimal struct for yaml unmarshalling (reuses yaml.Node logic).
type ownershipNode = yaml.Node
