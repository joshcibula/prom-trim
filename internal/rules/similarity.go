package rules

import (
	"math"
	"sort"

	"github.com/prometheus/prometheus/model/rulefmt"
)

// SimilarityEntry represents a pair of rules with a computed similarity score.
type SimilarityEntry struct {
	RuleA      string  `json:"rule_a"`
	RuleB      string  `json:"rule_b"`
	Score      float64 `json:"score"`
	SharedTerms int    `json:"shared_terms"`
}

// SimilarityConfig controls thresholds for similarity detection.
type SimilarityConfig struct {
	MinScore float64
}

// DefaultSimilarityConfig returns sensible defaults.
func DefaultSimilarityConfig() SimilarityConfig {
	return SimilarityConfig{MinScore: 0.6}
}

// BuildSimilarityReport parses a rules file and returns pairs of recording
// rules whose expression token sets exceed the minimum Jaccard similarity.
func BuildSimilarityReport(path string, cfg SimilarityConfig) ([]SimilarityEntry, error) {
	groups, err := ParseFile(path)
	if err != nil {
		return nil, err
	}

	type ruleExpr struct {
		name string
		tokens map[string]struct{}
	}

	var records []ruleExpr
	for _, g := range groups.Groups {
		for _, r := range g.Rules {
			if ruleRecordName(rulefmt.RuleNode(r)) == "" {
				continue
			}
			records = append(records, ruleExpr{
				name:   string(r.Record),
				tokens: tokenSet(string(r.Expr)),
			})
		}
	}

	var entries []SimilarityEntry
	for i := 0; i < len(records); i++ {
		for j := i + 1; j < len(records); j++ {
			a, b := records[i], records[j]
			score, shared := jaccardWithCount(a.tokens, b.tokens)
			if score >= cfg.MinScore {
				entries = append(entries, SimilarityEntry{
					RuleA:       a.name,
					RuleB:       b.name,
					Score:       math.Round(score*1000) / 1000,
					SharedTerms: shared,
				})
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})
	return entries, nil
}

func jaccardWithCount(a, b map[string]struct{}) (float64, int) {
	if len(a) == 0 && len(b) == 0 {
		return 0, 0
	}
	shared := 0
	for k := range a {
		if _, ok := b[k]; ok {
			shared++
		}
	}
	union := len(a) + len(b) - shared
	if union == 0 {
		return 0, 0
	}
	return float64(shared) / float64(union), shared
}
