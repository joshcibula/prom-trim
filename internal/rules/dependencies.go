package rules

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// metricRefPattern matches metric names used inside a PromQL expression.
var metricRefPattern = regexp.MustCompile(`\b([a-zA-Z_:][a-zA-Z0-9_:]*)\b`)

// DependencyEdge represents a directed dependency between two recording rules.
type DependencyEdge struct {
	From string `json:"from"` // rule whose expression references To
	To   string `json:"to"`   // rule whose record name is referenced
}

// String returns a human-readable representation of the edge.
func (e DependencyEdge) String() string {
	return fmt.Sprintf("%s -> %s", e.From, e.To)
}

// DependencyGraph maps each recording rule to the set of rules it depends on.
type DependencyGraph map[string][]string

// BuildDependencyGraph parses a rules file and builds an inter-rule dependency
// graph. A rule A depends on rule B when B's record name appears inside A's
// expression.
func BuildDependencyGraph(path string) (DependencyGraph, []DependencyEdge, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("dependency graph: %w", err)
	}

	// Collect all record names defined in the file.
	recordNames := map[string]struct{}{}
	for _, g := range groups {
		for _, r := range g.Rules {
			if r.Record != "" {
				recordNames[r.Record] = struct{}{}
			}
		}
	}

	graph := DependencyGraph{}
	var edges []DependencyEdge

	for _, g := range groups {
		for _, r := range g.Rules {
			if r.Record == "" {
				continue
			}
			matches := metricRefPattern.FindAllString(r.Expr.Value, -1)
			seen := map[string]bool{}
			for _, m := range matches {
				if _, ok := recordNames[m]; ok && m != r.Record && !seen[m] {
					seen[m] = true
					graph[r.Record] = append(graph[r.Record], m)
					edges = append(edges, DependencyEdge{From: r.Record, To: m})
				}
			}
			if _, exists := graph[r.Record]; !exists {
				graph[r.Record] = []string{}
			}
		}
	}

	// Sort for determinism.
	for k := range graph {
		sort.Strings(graph[k])
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	_ = yaml.Marshal // ensure import used via ParseFile transitive dep
	_ = strings.ToLower

	return graph, edges, nil
}

// Dependents returns all rules that directly depend on the given rule name.
func Dependents(graph DependencyGraph, target string) []string {
	var result []string
	for rule, deps := range graph {
		for _, d := range deps {
			if d == target {
				result = append(result, rule)
				break
			}
		}
	}
	sort.Strings(result)
	return result
}
