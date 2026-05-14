package rules

import (
	"testing"

	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
)

func makeRecordNode(name string) yaml.Node {
	n := yaml.Node{Kind: yaml.ScalarNode, Value: name}
	return n
}

func makeGroups(records ...string) []rulefmt.RuleGroup {
	var rules []rulefmt.RuleNode
	for _, rec := range records {
		rules = append(rules, rulefmt.RuleNode{
			Record: yaml.Node{Kind: yaml.ScalarNode, Value: rec},
			Expr:   yaml.Node{Kind: yaml.ScalarNode, Value: "vector(1)"},
		})
	}
	return []rulefmt.RuleGroup{{Name: "g", Rules: rules}}
}

func TestBuildCoverageReport_AllCovered(t *testing.T) {
	groups := makeGroups("rule:a", "rule:b")
	usage := map[string]float64{"rule:a": 5, "rule:b": 3}

	r := BuildCoverageReport(groups, usage)

	if r.Total != 2 {
		t.Fatalf("expected total 2, got %d", r.Total)
	}
	if r.Covered != 2 {
		t.Fatalf("expected covered 2, got %d", r.Covered)
	}
	if len(r.Uncovered) != 0 {
		t.Fatalf("expected no uncovered, got %v", r.Uncovered)
	}
	if r.Coverage() != 1.0 {
		t.Fatalf("expected 100%% coverage, got %.2f", r.Coverage())
	}
}

func TestBuildCoverageReport_PartialCoverage(t *testing.T) {
	groups := makeGroups("rule:a", "rule:b", "rule:c")
	usage := map[string]float64{"rule:a": 10}

	r := BuildCoverageReport(groups, usage)

	if r.Total != 3 {
		t.Fatalf("expected total 3, got %d", r.Total)
	}
	if r.Covered != 1 {
		t.Fatalf("expected covered 1, got %d", r.Covered)
	}
	if len(r.Uncovered) != 2 {
		t.Fatalf("expected 2 uncovered, got %v", r.Uncovered)
	}
}

func TestBuildCoverageReport_Empty(t *testing.T) {
	r := BuildCoverageReport([]rulefmt.RuleGroup{}, map[string]float64{})
	if r.Total != 0 || r.Coverage() != 0 {
		t.Fatalf("expected zero report, got %+v", r)
	}
}

func TestCoverageReport_String(t *testing.T) {
	r := CoverageReport{Total: 4, Covered: 2}
	s := r.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
	// Should contain the percentage
	if s != "2/4 rules covered (50.0%)" {
		t.Fatalf("unexpected string: %s", s)
	}
}

func TestBuildCoverageReport_ZeroUsage(t *testing.T) {
	groups := makeGroups("rule:a")
	usage := map[string]float64{"rule:a": 0}

	r := BuildCoverageReport(groups, usage)
	if r.Covered != 0 {
		t.Fatalf("zero usage should not count as covered, got %d", r.Covered)
	}
	if len(r.Uncovered) != 1 || r.Uncovered[0] != "rule:a" {
		t.Fatalf("expected rule:a in uncovered, got %v", r.Uncovered)
	}
}
