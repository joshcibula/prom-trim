package main

import (
	"os"
	"strings"
	"testing"
)

const annotationsCmdRulesYAML = `
groups:
  - name: infra
    rules:
      - record: node:cpu:rate5m
        expr: rate(node_cpu_seconds_total[5m])
        annotations:
          owner: infra-team
          tier: "1"
      - record: node:mem:usage
        expr: node_memory_MemAvailable_bytes
`

func writeAnnotationsCmdRules(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "annot-rules-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func captureAnnotationsOutput(t *testing.T, args []string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	_ = runAnnotations(args)
	w.Close()
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return sb.String()
}

func TestRunAnnotations_TableOutput(t *testing.T) {
	path := writeAnnotationsCmdRules(t, annotationsCmdRulesYAML)
	out := captureAnnotationsOutput(t, []string{path})
	if !strings.Contains(out, "owner") {
		t.Errorf("expected 'owner' column in output, got:\n%s", out)
	}
	if !strings.Contains(out, "node:cpu:rate5m") {
		t.Errorf("expected rule name in output, got:\n%s", out)
	}
}

func TestRunAnnotations_JSONOutput(t *testing.T) {
	path := writeAnnotationsCmdRules(t, annotationsCmdRulesYAML)
	out := captureAnnotationsOutput(t, []string{"--format=json", path})
	if !strings.Contains(out, "\"owner\"") {
		t.Errorf("expected JSON with owner key, got:\n%s", out)
	}
}

func TestRunAnnotations_MissingFile(t *testing.T) {
	err := runAnnotations([]string{"/no/such/file.yaml"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRunAnnotations_NoArgs(t *testing.T) {
	err := runAnnotations([]string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}
}
