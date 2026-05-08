package rules

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PruneOptions controls the behaviour of Prune.
type PruneOptions struct {
	DryRun bool
	// Backup creates a timestamped backup of the original file before writing.
	Backup bool
}

// Prune removes stale recording rules from the file at path.
// staleNames is the set of rule names to remove.
// It returns the number of rules removed.
func Prune(path string, staleNames map[string]struct{}, opts PruneOptions) (int, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}

	removed := 0
	for gi := range groups {
		var kept []yaml.Node
		for ri := 0; ri < len(groups[gi].Rules); ri++ {
			name := ruleRecordName(&groups[gi].Rules[ri])
			if _, stale := staleNames[name]; stale {
				removed++
				continue
			}
			kept = append(kept, groups[gi].Rules[ri])
		}
		groups[gi].Rules = kept
	}

	if opts.DryRun {
		return removed, nil
	}

	if opts.Backup {
		if _, err := BackupFile(path, BackupOptions{}); err != nil {
			return 0, fmt.Errorf("backup: %w", err)
		}
	}

	if err := writeGroups(path, groups); err != nil {
		return 0, fmt.Errorf("write %s: %w", path, err)
	}
	return removed, nil
}

func writeGroups(path string, groups []RuleGroup) error {
	type document struct {
		Groups []RuleGroup `yaml:"groups"`
	}
	out, err := yaml.Marshal(document{Groups: groups})
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func ruleRecordName(n *yaml.Node) string {
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == "record" {
			return n.Content[i+1].Value
		}
	}
	return ""
}
