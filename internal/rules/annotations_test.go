package rules

import (
	"os"
	"testing"
)

func writeAnnotationRules(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "rules-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

const annotationRulesYAML = `
groups:
  - name: group_a
    rules:
      - record: metric:rate5m
        expr: rate(metric_total[5m])
        annotations:
          owner: team-a
          severity: low
      - record: metric:rate1m
        expr: rate(metric_total[1m])
  - name: group_b
    rules:
      - record: other:sum
        expr: sum(other_total)
        annotations:
          owner: team-b
`

func TestExtractAnnotations_Basic(t *testing.T) {
	path := writeAnnotationRules(t, annotationRulesYAML)
	summaries, err := ExtractAnnotations(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].Record != "metric:rate5m" {
		t.Errorf("expected metric:rate5m, got %s", summaries[0].Record)
	}
	if summaries[0].Annotations["owner"] != "team-a" {
		t.Errorf("expected team-a owner annotation")
	}
}

func TestExtractAnnotations_SkipsNoAnnotation(t *testing.T) {
	path := writeAnnotationRules(t, annotationRulesYAML)
	summaries, err := ExtractAnnotations(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range summaries {
		if s.Record == "metric:rate1m" {
			t.Error("rule without annotations should be skipped")
		}
	}
}

func TestExtractAnnotations_NotFound(t *testing.T) {
	_, err := ExtractAnnotations("/no/such/file.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestAnnotationKeys_Sorted(t *testing.T) {
	path := writeAnnotationRules(t, annotationRulesYAML)
	summaries, err := ExtractAnnotations(path)
	if err != nil {
		t.Fatal(err)
	}
	keys := AnnotationKeys(summaries)
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "owner" || keys[1] != "severity" {
		t.Errorf("unexpected keys order: %v", keys)
	}
}

func TestAnnotationKeys_Empty(t *testing.T) {
	keys := AnnotationKeys(nil)
	if len(keys) != 0 {
		t.Errorf("expected empty keys, got %v", keys)
	}
}
