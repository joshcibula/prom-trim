package rules

import (
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// LabelSummary holds aggregated label metadata for a set of recording rules.
type LabelSummary struct {
	RuleName string            `json:"rule_name"`
	Labels   map[string]string `json:"labels"`
}

// LabelIndex maps label keys to the set of rule names that carry them.
type LabelIndex map[string][]string

// ExtractLabels returns a LabelSummary for every recording rule found in the
// parsed groups of a rules file.
func ExtractLabels(path string) ([]LabelSummary, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return nil, err
	}

	var summaries []LabelSummary
	for _, g := range groups {
		for _, r := range g.Rules {
			name := ruleRecordName(r)
			if name == "" {
				continue
			}
			labels := make(map[string]string)
			for k, v := range r.Labels {
				labels[k] = v
			}
			summaries = append(summaries, LabelSummary{
				RuleName: name,
				Labels:   labels,
			})
		}
	}
	return summaries, nil
}

// BuildLabelIndex inverts a slice of LabelSummary into a LabelIndex so callers
// can look up which rules carry a given label key.
func BuildLabelIndex(summaries []LabelSummary) LabelIndex {
	idx := make(LabelIndex)
	for _, s := range summaries {
		for k := range s.Labels {
			idx[k] = append(idx[k], s.RuleName)
		}
	}
	for k := range idx {
		sort.Strings(idx[k])
	}
	return idx
}

// LabelKeys returns a sorted slice of all distinct label keys present across
// the provided summaries.
func LabelKeys(summaries []LabelSummary) []string {
	seen := make(map[string]struct{})
	for _, s := range summaries {
		for k := range s.Labels {
			seen[k] = struct{}{}
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// formatLabelPairs renders a label map as a compact sorted string for display.
func formatLabelPairs(labels map[string]string) string {
	pairs := make([]string, 0, len(labels))
	for k, v := range labels {
		pairs = append(pairs, k+"="+v)
	}
	sort.Strings(pairs)
	return "{" + strings.Join(pairs, ", ") + "}"
}

// silence unused import from yaml (used transitively via ParseFile)
var _ = yaml.Node{}
