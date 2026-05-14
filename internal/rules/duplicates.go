package rules

import (
	"fmt"
	"sort"
)

// DuplicateEntry represents a recording rule name that appears more than once
// across the parsed rule groups in a file.
type DuplicateEntry struct {
	Name   string
	Count  int
	Groups []string
}

func (d DuplicateEntry) String() string {
	return fmt.Sprintf("%s (count=%d, groups=%v)", d.Name, d.Count, d.Groups)
}

// BuildDuplicatesReport parses the given rules file and returns entries for
// any recording rule names that appear in more than one group, or more than
// once within the same group.
func BuildDuplicatesReport(path string) ([]DuplicateEntry, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("duplicates: parse %s: %w", path, err)
	}

	type occurrence struct {
		count  int
		groups []string
		seen   map[string]bool
	}

	index := make(map[string]*occurrence)

	for _, g := range groups {
		for _, node := range g.Rules {
			name := ruleRecordName(node)
			if name == "" {
				continue
			}
			occ, ok := index[name]
			if !ok {
				occ = &occurrence{seen: make(map[string]bool)}
				index[name] = occ
			}
			occ.count++
			if !occ.seen[g.Name] {
				occ.seen[g.Name] = true
				occ.groups = append(occ.groups, g.Name)
			}
		}
	}

	var result []DuplicateEntry
	for name, occ := range index {
		if occ.count > 1 {
			sort.Strings(occ.groups)
			result = append(result, DuplicateEntry{
				Name:   name,
				Count:  occ.count,
				Groups: occ.groups,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}
