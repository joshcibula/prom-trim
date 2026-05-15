package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const agingCmdRulesYAML = `
groups:
  - name: sre
    rules:
      - record: job:up:avg
        expr: avg(up)
      - record: job:latency:p99
        expr: histogram_quantile(0.99, rate(http_duration_seconds_bucket[5m]))
`

func writeAgingCmdRules(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "aging-cmd-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(agingCmdRulesYAML); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func captureAgingOutput(t *testing.T, args []string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	_ = runAging(args)
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestRunAging_TableOutput(t *testing.T) {
	path := writeAgingCmdRules(t)
	out := captureAgingOutput(t, []string{"--file", path})
	if !strings.Contains(out, "RULE") {
		t.Error("expected table header RULE")
	}
	if !strings.Contains(out, "job:up:avg") {
		t.Error("expected rule name in output")
	}
}

func TestRunAging_JSONOutput(t *testing.T) {
	path := writeAgingCmdRules(t)
	out := captureAgingOutput(t, []string{"--file", path, "--format", "json"})
	if !strings.Contains(out, "RuleName") && !strings.Contains(out, "ruleName") && !strings.Contains(out, "job:up:avg") {
		t.Error("expected JSON output with rule name")
	}
}

func TestRunAging_MissingFile(t *testing.T) {
	err := runAging([]string{"--file", "/no/such/file.yaml"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRunAging_MissingFlag(t *testing.T) {
	err := runAging([]string{})
	if err == nil {
		t.Error("expected error when --file not provided")
	}
}
