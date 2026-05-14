package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeDepsRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeDepsRules: %v", err)
	}
	return p
}

const depsRulesYAML = `
groups:
  - name: test
    rules:
      - record: job:requests:rate5m
        expr: rate(http_requests_total[5m])
      - record: job:requests:rate1h
        expr: sum(job:requests:rate5m) by (job)
      - record: job:requests:ratio
        expr: job:requests:rate1h / job:requests:rate5m
`

func TestBuildDependencyGraph_Basic(t *testing.T) {
	p := writeDepsRules(t, depsRulesYAML)
	graph, edges, err := BuildDependencyGraph(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(graph) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph))
	}

	deps := graph["job:requests:rate1h"]
	if len(deps) != 1 || deps[0] != "job:requests:rate5m" {
		t.Errorf("expected rate1h to depend on rate5m, got %v", deps)
	}

	if len(edges) == 0 {
		t.Error("expected at least one edge")
	}
}

func TestBuildDependencyGraph_NoDeps(t *testing.T) {
	const yaml = `
groups:
  - name: g
    rules:
      - record: metric:a
        expr: sum(raw_a)
      - record: metric:b
        expr: sum(raw_b)
`
	p := writeDepsRules(t, yaml)
	graph, edges, err := BuildDependencyGraph(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(edges) != 0 {
		t.Errorf("expected no edges, got %v", edges)
	}
	for _, deps := range graph {
		if len(deps) != 0 {
			t.Errorf("expected empty deps, got %v", deps)
		}
	}
}

func TestBuildDependencyGraph_FileNotFound(t *testing.T) {
	_, _, err := BuildDependencyGraph("/nonexistent/rules.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestDependents_Basic(t *testing.T) {
	p := writeDepsRules(t, depsRulesYAML)
	graph, _, err := BuildDependencyGraph(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	deps := Dependents(graph, "job:requests:rate5m")
	if len(deps) < 1 {
		t.Errorf("expected at least one dependent of rate5m, got %v", deps)
	}
}

func TestDependents_NoMatch(t *testing.T) {
	graph := DependencyGraph{
		"a": {"b"},
	}
	result := Dependents(graph, "c")
	if len(result) != 0 {
		t.Errorf("expected no dependents, got %v", result)
	}
}

func TestDependencyEdge_String(t *testing.T) {
	e := DependencyEdge{From: "a:metric", To: "b:metric"}
	got := e.String()
	if got != "a:metric -> b:metric" {
		t.Errorf("unexpected string: %q", got)
	}
}
