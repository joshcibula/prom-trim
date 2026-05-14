package rules

import (
	"fmt"
	"sort"
)

// OrphanEntry represents a recording rule whose metric is never referenced
// by any other rule expression in the same file.
type OrphanEntry struct {
	Group  string
	Record string
	Expr   string
}

func (e OrphanEntry) String() string {
	return fmt.Sprintf("%s/%s", e.Group, e.Record)
}

// BuildOrphansReport parses the given rules file and returns all recording
// rules that are not referenced (by metric name) in any other rule's expression.
func BuildOrphansReport(path string) ([]OrphanEntry, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return nil, err
	}

	// Collect all record names and their definitions.
	type ruleDef struct {
		group  string
		record string
		expr   string
	}
	var defs []ruleDef
	for _, g := range groups {
		for _, n := range g.Rules {
			if n.Record == "" {
				continue
			}
			defs = append(defs, ruleDef{group: g.Name, record: n.Record, expr: n.Expr.Value})
		}
	}

	// Build a set of all expressions to search for references.
	allExprs := make([]string, 0, len(defs))
	for _, d := range defs {
		allExprs = append(allExprs, d.expr)
	}

	var orphans []OrphanEntry
	for _, d := range defs {
		if !isReferencedIn(d.record, allExprs) {
			orphan = append(orphans, OrphanEntry{
				Group:  d.group,
				Record: d.record,
				Expr:   d.expr,
			})
		}
	}

	sort.Slice(orphans, func(i, j int) bool {
		if orphans[i].Group != orphans[j].Group {
			return orphans[i].Group < orphans[j].Group
		}
		return orphans[i].Record < orphans[j].Record
	})
	return orphans, nil
}

// isReferencedIn returns true if the metric name appears as a substring in
// any of the provided PromQL expression strings.
func isReferencedIn(metric string, exprs []string) bool {
	for _, expr := range exprs {
		if containsToken(expr, metric) {
			return true
		}
	}
	return false
}

// containsToken checks whether target appears as a word token in src.
func containsToken(src, target string) bool {
	if len(target) == 0 {
		return false
	}
	idx := 0
	for {
		pos := indexOf(src[idx:], target)
		if pos < 0 {
			return false
		}
		abs := idx + pos
		before := abs == 0 || !isIdentChar(rune(src[abs-1]))
		after := abs+len(target) >= len(src) || !isIdentChar(rune(src[abs+len(target)]))
		if before && after {
			return true
		}
		idx = abs + 1
	}
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func isIdentChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_' || r == ':'
}
