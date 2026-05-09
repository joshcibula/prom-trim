package rules

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// DiffResult holds the comparison between original and pruned rule groups.
type DiffResult struct {
	Removed []DiffEntry
	Retained []DiffEntry
}

// DiffEntry represents a single rule in a diff.
type DiffEntry struct {
	Group  string
	Record string
}

// String returns a human-readable unified-style diff summary.
func (d *DiffResult) String() string {
	var sb strings.Builder
	for _, e := range d.Removed {
		fmt.Fprintf(&sb, "- [%s] %s\n", e.Group, e.Record)
	}
	for _, e := range d.Retained {
		fmt.Fprintf(&sb, "  [%s] %s\n", e.Group, e.Record)
	}
	return sb.String()
}

// Diff compares two YAML rule files and returns a DiffResult describing
// which recording rules were removed and which were retained.
func Diff(originalPath, prunedPath string) (*DiffResult, error) {
	origGroups, err := ParseFile(originalPath)
	if err != nil {
		return nil, fmt.Errorf("reading original file: %w", err)
	}

	prunedGroups, err := ParseFile(prunedPath)
	if err != nil {
		return nil, fmt.Errorf("reading pruned file: %w", err)
	}

	_ = yaml.Marshal // ensure import used via ParseFile internals

	prunedSet := make(map[string]struct{})
	for _, g := range prunedGroups {
		for _, r := range g.Rules {
			if r.Record != "" {
				key := g.Name + "/" + r.Record
				prunedSet[key] = struct{}{}
			}
		}
	}

	result := &DiffResult{}
	for _, g := range origGroups {
		for _, r := range g.Rules {
			if r.Record == "" {
				continue
			}
			entry := DiffEntry{Group: g.Name, Record: r.Record}
			key := g.Name + "/" + r.Record
			if _, kept := prunedSet[key]; kept {
				result.Retained = append(result.Retained, entry)
			} else {
				result.Removed = append(result.Removed, entry)
			}
		}
	}

	return result, nil
}
