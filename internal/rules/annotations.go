package rules

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
	"os"
)

// AnnotationSummary holds annotation key/value pairs for a single recording rule.
type AnnotationSummary struct {
	Group       string            `json:"group"`
	Record      string            `json:"record"`
	Annotations map[string]string `json:"annotations"`
}

// ExtractAnnotations reads a rules file and returns annotation summaries
// for every recording rule that carries at least one annotation.
func ExtractAnnotations(path string) ([]AnnotationSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var root struct {
		Groups []struct {
			Name  string `yaml:"name"`
			Rules []struct {
				Record      string            `yaml:"record"`
				Annotations map[string]string `yaml:"annotations"`
			} `yaml:"rules"`
		} `yaml:"groups"`
	}

	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	var summaries []AnnotationSummary
	for _, g := range root.Groups {
		for _, r := range g.Rules {
			if r.Record == "" || len(r.Annotations) == 0 {
				continue
			}
			summaries = append(summaries, AnnotationSummary{
				Group:       g.Name,
				Record:      r.Record,
				Annotations: r.Annotations,
			})
		}
	}
	return summaries, nil
}

// AnnotationKeys returns a sorted, deduplicated list of all annotation keys
// present across the provided summaries.
func AnnotationKeys(summaries []AnnotationSummary) []string {
	seen := make(map[string]struct{})
	for _, s := range summaries {
		for k := range s.Annotations {
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
