package rules

import (
	"os"
	"testing"
)

func writeLabelRules(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "labels-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

const labelRulesYAML = `
groups:
  - name: test
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
        labels:
          env: prod
          team: platform
      - record: job:errors:rate5m
        expr: rate(errors_total[5m])
        labels:
          env: prod
`

func TestExtractLabels_Basic(t *testing.T) {
	path := writeLabelRules(t, labelRulesYAML)
	summaries, err := ExtractLabels(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].Labels["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", summaries[0].Labels["env"])
	}
}

func TestExtractLabels_NotFound(t *testing.T) {
	_, err := ExtractLabels("/no/such/file.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildLabelIndex_Basic(t *testing.T) {
	path := writeLabelRules(t, labelRulesYAML)
	summaries, _ := ExtractLabels(path)
	idx := BuildLabelIndex(summaries)

	rules, ok := idx["env"]
	if !ok {
		t.Fatal("expected 'env' key in index")
	}
	if len(rules) != 2 {
		t.Errorf("expected 2 rules for 'env', got %d", len(rules))
	}

	teamRules := idx["team"]
	if len(teamRules) != 1 {
		t.Errorf("expected 1 rule for 'team', got %d", len(teamRules))
	}
}

func TestLabelKeys_Sorted(t *testing.T) {
	path := writeLabelRules(t, labelRulesYAML)
	summaries, _ := ExtractLabels(path)
	keys := LabelKeys(summaries)
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "env" || keys[1] != "team" {
		t.Errorf("unexpected key order: %v", keys)
	}
}

func TestFormatLabelPairs_Output(t *testing.T) {
	labels := map[string]string{"env": "prod", "team": "platform"}
	out := formatLabelPairs(labels)
	expected := "{env=prod, team=platform}"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestExtractLabels_NoLabels(t *testing.T) {
	content := `
groups:
  - name: bare
    rules:
      - record: bare:metric
        expr: up
`
	path := writeLabelRules(t, content)
	summaries, err := ExtractLabels(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if len(summaries[0].Labels) != 0 {
		t.Errorf("expected empty labels, got %v", summaries[0].Labels)
	}
}
