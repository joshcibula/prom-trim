package rules

import (
	"testing"

	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
)

func makeRuleNode(record string) rulefmt.RuleNode {
	var node rulefmt.RuleNode
	node.Record.SetString(record)
	var exprNode yaml.Node
	exprNode.SetString("up")
	node.Expr = exprNode
	return node
}

func TestExtractNamespace_WithSlash(t *testing.T) {
	if got := ExtractNamespace("platform/infra"); got != "platform" {
		t.Errorf("expected platform, got %s", got)
	}
}

func TestExtractNamespace_WithColon(t *testing.T) {
	if got := ExtractNamespace("team:backend"); got != "team" {
		t.Errorf("expected team, got %s", got)
	}
}

func TestExtractNamespace_NoSeparator(t *testing.T) {
	if got := ExtractNamespace("standalone"); got != "standalone" {
		t.Errorf("expected standalone, got %s", got)
	}
}

func TestBuildNamespaceStats_Basic(t *testing.T) {
	groups := []rulefmt.RuleGroup{
		{
			Name: "platform/infra",
			Rules: []rulefmt.RuleNode{
				makeRuleNode("cpu:usage"),
				makeRuleNode("mem:usage"),
			},
		},
		{
			Name: "platform/network",
			Rules: []rulefmt.RuleNode{
				makeRuleNode("net:rx"),
			},
		},
		{
			Name: "team:backend",
			Rules: []rulefmt.RuleNode{
				makeRuleNode("rpc:latency"),
			},
		},
	}
	stale := map[string]bool{"mem:usage": true, "net:rx": true}

	stats := BuildNamespaceStats(groups, stale)

	if len(stats) != 2 {
		t.Fatalf("expected 2 namespaces, got %d", len(stats))
	}

	// sorted: platform, team
	platform := stats[0]
	if platform.Namespace != "platform" {
		t.Errorf("expected platform, got %s", platform.Namespace)
	}
	if platform.GroupCount != 2 {
		t.Errorf("expected 2 groups, got %d", platform.GroupCount)
	}
	if platform.RuleCount != 3 {
		t.Errorf("expected 3 rules, got %d", platform.RuleCount)
	}
	if platform.StaleCount != 2 {
		t.Errorf("expected 2 stale, got %d", platform.StaleCount)
	}

	team := stats[1]
	if team.StaleCount != 0 {
		t.Errorf("expected 0 stale for team, got %d", team.StaleCount)
	}
}

func TestBuildNamespaceStats_Empty(t *testing.T) {
	stats := BuildNamespaceStats([]rulefmt.RuleGroup{}, map[string]bool{})
	if len(stats) != 0 {
		t.Errorf("expected empty stats, got %d", len(stats))
	}
}

func TestNamespaceStat_String(t *testing.T) {
	n := NamespaceStat{Namespace: "platform", GroupCount: 2, RuleCount: 5, StaleCount: 1}
	s := n.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
