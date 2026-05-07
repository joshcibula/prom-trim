package rules

import (
	"testing"
	"time"

	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
)

func makeNode(record string) rulefmt.RuleNode {
	var recNode yaml.Node
	recNode.SetString(record)
	return rulefmt.RuleNode{
		Record: recNode,
		Expr:   yaml.Node{},
	}
}

func TestAllRecordNames_Basic(t *testing.T) {
	groups := []rulefmt.RuleGroup{
		{Name: "g1", Rules: []rulefmt.RuleNode{makeNode("rule_a"), makeNode("rule_b")}},
		{Name: "g2", Rules: []rulefmt.RuleNode{makeNode("rule_c")}},
	}
	names := AllRecordNames(groups)
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
}

func TestAllRecordNames_Empty(t *testing.T) {
	names := AllRecordNames(nil)
	if len(names) != 0 {
		t.Fatalf("expected 0 names, got %d", len(names))
	}
}

func TestFilterStale_RemovesStale(t *testing.T) {
	now := time.Now()
	cutoff := now.Add(-7 * 24 * time.Hour)

	groups := []rulefmt.RuleGroup{
		{
			Name: "g1",
			Rules: []rulefmt.RuleNode{
				makeNode("active_rule"),
				makeNode("stale_rule"),
			},
		},
	}

	usage := UsageMap{
		"active_rule": now.Add(-1 * 24 * time.Hour), // recent
		"stale_rule":  now.Add(-14 * 24 * time.Hour), // older than cutoff
	}

	active, stale := FilterStale(groups, usage, cutoff)

	if len(stale) != 1 || stale[0] != "stale_rule" {
		t.Errorf("expected [stale_rule] in stale list, got %v", stale)
	}
	if len(active) != 1 || len(active[0].Rules) != 1 {
		t.Errorf("expected 1 active group with 1 rule, got %v", active)
	}
}

func TestFilterStale_MissingUsageIsStale(t *testing.T) {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	groups := []rulefmt.RuleGroup{
		{Name: "g1", Rules: []rulefmt.RuleNode{makeNode("unknown_rule")}},
	}

	_, stale := FilterStale(groups, UsageMap{}, cutoff)
	if len(stale) != 1 {
		t.Errorf("expected unknown_rule to be stale, got %v", stale)
	}
}

func TestFilterStale_EmptyGroupDropped(t *testing.T) {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	groups := []rulefmt.RuleGroup{
		{Name: "g1", Rules: []rulefmt.RuleNode{makeNode("old_rule")}},
	}

	active, _ := FilterStale(groups, UsageMap{}, cutoff)
	if len(active) != 0 {
		t.Errorf("expected empty active groups when all rules are stale, got %v", active)
	}
}
