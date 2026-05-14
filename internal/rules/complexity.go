package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prometheus/prometheus/model/rulefmt"
)

// ComplexityLevel categorises a rule's PromQL expression complexity.
type ComplexityLevel string

const (
	ComplexityLow    ComplexityLevel = "low"
	ComplexityMedium ComplexityLevel = "medium"
	ComplexityHigh   ComplexityLevel = "high"
)

// ComplexityStat holds complexity metadata for a single recording rule.
type ComplexityStat struct {
	Group      string
	Record     string
	Expr       string
	TokenCount int
	Level      ComplexityLevel
}

func (c ComplexityStat) String() string {
	return fmt.Sprintf("%s/%s tokens=%d level=%s", c.Group, c.Record, c.TokenCount, c.Level)
}

// classifyComplexity assigns a level based on token count.
func classifyComplexity(tokens int) ComplexityLevel {
	switch {
	case tokens <= 5:
		return ComplexityLow
	case tokens <= 15:
		return ComplexityMedium
	default:
		return ComplexityHigh
	}
}

// tokenCount counts whitespace-delimited tokens in a PromQL expression.
func tokenCount(expr string) int {
	return len(strings.Fields(expr))
}

// BuildComplexityReport parses the given rules file and returns a
// ComplexityStat slice sorted from most complex to least complex.
func BuildComplexityReport(path string) ([]ComplexityStat, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return nil, err
	}

	var stats []ComplexityStat
	for _, g := range groups.Groups {
		for _, r := range g.Rules {
			node, ok := r.(rulefmt.RuleNode)
			if !ok {
				continue
			}
			if node.Record.Value == "" {
				continue
			}
			expr := node.Expr.Value
			tc := tokenCount(expr)
			stats = append(stats, ComplexityStat{
				Group:      g.Name,
				Record:     node.Record.Value,
				Expr:       expr,
				TokenCount: tc,
				Level:      classifyComplexity(tc),
			})
		}
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].TokenCount > stats[j].TokenCount
	})
	return stats, nil
}
