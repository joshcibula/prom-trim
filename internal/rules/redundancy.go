package rules

import (
	"fmt"
	"sort"
	"strings"
)

// RedundancyEntry describes a recording rule whose expr is a subset or
// superset of another rule in the same file.
type RedundancyEntry struct {
	Rule       string
	Expr       string
	OverlapsBy []string // names of rules whose expressions overlap
	Group      string
}

func (r RedundancyEntry) String() string {
	return fmt.Sprintf("%s overlaps with: %s", r.Rule, strings.Join(r.OverlapsBy, ", "))
}

// BuildRedundancyReport identifies recording rules whose PromQL expressions
// share significant token overlap (>=50% Jaccard similarity).
func BuildRedundancyReport(path string, threshold float64) ([]RedundancyEntry, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("redundancy: parse %s: %w", path, err)
	}

	type ruleInfo struct {
		name  string
		expr  string
		group string
		tokens map[string]struct{}
	}

	var rules []ruleInfo
	for _, g := range groups {
		for _, n := range g.Rules {
			if n.Record == "" {
				continue
			}
			rules = append(rules, ruleInfo{
				name:   n.Record,
				expr:   n.Expr,
				group:  g.Name,
				tokens: tokenSet(n.Expr),
			})
		}
	}

	resultMap := map[string]*RedundancyEntry{}
	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			sim := jaccardSimilarity(rules[i].tokens, rules[j].tokens)
			if sim >= threshold {
				if _, ok := resultMap[rules[i].name]; !ok {
					resultMap[rules[i].name] = &RedundancyEntry{
						Rule:  rules[i].name,
						Expr:  rules[i].expr,
						Group: rules[i].group,
					}
				}
				resultMap[rules[i].name].OverlapsBy = append(resultMap[rules[i].name].OverlapsBy, rules[j].name)

				if _, ok := resultMap[rules[j].name]; !ok {
					resultMap[rules[j].name] = &RedundancyEntry{
						Rule:  rules[j].name,
						Expr:  rules[j].expr,
						Group: rules[j].group,
					}
				}
				resultMap[rules[j].name].OverlapsBy = append(resultMap[rules[j].name].OverlapsBy, rules[i].name)
			}
		}
	}

	var out []RedundancyEntry
	for _, e := range resultMap {
		out = append(out, *e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Rule < out[j].Rule })
	return out, nil
}

func tokenSet(expr string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, t := range strings.FieldsFunc(expr, func(r rune) bool {
		return strings.ContainsRune(" \t\n(){}[],=!<>+-*/^|&", r)
	}) {
		if t != "" {
			set[t] = struct{}{}
		}
	}
	return set
}

func jaccardSimilarity(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	intersection := 0
	for k := range a {
		if _, ok := b[k]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}
