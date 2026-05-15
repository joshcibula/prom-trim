package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSimilarityCmdRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func captureSimilarityOutput(t *testing.T, args []string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	_ = runSimilarity(args)
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRunSimilarity_TableOutput(t *testing.T) {
	path := writeSimilarityCmdRules(t, `groups:
  - name: g
    rules:
      - record: rule:a
        expr: sum(rate(http_requests_total[5m])) by (job)
      - record: rule:b
        expr: sum(rate(http_requests_total[5m])) by (env)
`)
	out := captureSimilarityOutput(t, []string{"--rules", path, "--min-score", "0.4"})
	if !strings.Contains(out, "RULE_A") {
		t.Errorf("expected table header, got: %s", out)
	}
}

func TestRunSimilarity_JSONOutput(t *testing.T) {
	path := writeSimilarityCmdRules(t, `groups:
  - name: g
    rules:
      - record: rule:a
        expr: sum(rate(http_requests_total[5m])) by (job)
      - record: rule:b
        expr: sum(rate(http_requests_total[5m])) by (env)
`)
	out := captureSimilarityOutput(t, []string{"--rules", path, "--min-score", "0.4", "--format", "json"})
	if !strings.Contains(out, "rule_a") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestRunSimilarity_MissingFile(t *testing.T) {
	err := runSimilarity([]string{"--rules", "/no/such/file.yaml"})
	if err == nil {
		t.Fatal("expected error for missing rules file")
	}
}

func TestRunSimilarity_MissingFlag(t *testing.T) {
	err := runSimilarity([]string{})
	if err == nil {
		t.Fatal("expected error when --rules flag is absent")
	}
}
