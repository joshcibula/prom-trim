package rules

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// PruneResult holds the outcome of a prune operation.
type PruneResult struct {
	Removed []string
	Kept    []string
	DryRun  bool
}

// Prune removes stale recording rules from the given file.
// If dryRun is true, no changes are written to disk.
func Prune(filePath string, staleNames map[string]bool, dryRun bool) (*PruneResult, error) {
	groups, err := ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("parsing rules file: %w", err)
	}

	result := &PruneResult{DryRun: dryRun}

	var filteredGroups []RuleGroup
	for _, g := range groups {
		var kept []Rule
		for _, r := range g.Rules {
			if staleNames[r.Record] {
				result.Removed = append(result.Removed, r.Record)
			} else {
				kept = append(kept, r)
				result.Kept = append(result.Kept, r.Record)
			}
		}
		if len(kept) > 0 {
			g.Rules = kept
			filteredGroups = append(filteredGroups, g)
		}
	}

	if dryRun {
		return result, nil
	}

	if err := writeGroups(filePath, filteredGroups); err != nil {
		return nil, fmt.Errorf("writing pruned rules: %w", err)
	}

	return result, nil
}

// writeGroups serialises the rule groups back to YAML and writes them to path.
func writeGroups(path string, groups []RuleGroup) error {
	type rulesFile struct {
		Groups []RuleGroup `yaml:"groups"`
	}

	out, err := yaml.Marshal(rulesFile{Groups: groups})
	if err != nil {
		return err
	}

	// Prepend a small comment so the file is clearly machine-managed.
	header := "# Managed by prom-trim — do not edit stale entries manually.\n"
	content := header + strings.TrimSpace(string(out)) + "\n"

	return os.WriteFile(path, []byte(content), 0o644)
}
