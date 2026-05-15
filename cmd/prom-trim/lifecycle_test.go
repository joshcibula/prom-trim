package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prom-trim/internal/rules"
)

func writeLifecycleCmdRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func captureLifecycleOutput(t *testing.T, args []string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	_ = runLifecycle(args)
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

const lifecycleRulesYAML = `groups:
  - name: svc
    rules:
      - record: http:requests:rate5m
        expr: sum(rate(http_requests_total[5m]))
      - record: http:errors:rate5m
        expr: sum(rate(http_errors_total[5m]))
`

func TestRunLifecycle_TableOutput(t *testing.T) {
	p := writeLifecycleCmdRules(t, lifecycleRulesYAML)
	out := captureLifecycleOutput(t, []string{"--rules", p})

	for _, want := range []string{"RULE", "STAGE", "TRANSITION", "http:requests:rate5m"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q\ngot: %s", want, out)
		}
	}
}

func TestRunLifecycle_JSONOutput(t *testing.T) {
	p := writeLifecycleCmdRules(t, lifecycleRulesYAML)
	out := captureLifecycleOutput(t, []string{"--rules", p, "--format", "json"})

	var entries []rules.LifecycleEntry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, out)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestRunLifecycle_MissingFile(t *testing.T) {
	err := runLifecycle([]string{"--rules", "/no/such/file.yaml"})
	if err == nil {
		t.Fatal("expected error for missing rules file")
	}
}

func TestRunLifecycle_MissingFlag(t *testing.T) {
	err := runLifecycle([]string{})
	if err == nil {
		t.Fatal("expected error when --rules is missing")
	}
	if !strings.Contains(err.Error(), "--rules") {
		t.Errorf("error should mention --rules flag, got: %v", err)
	}
}

func TestRunLifecycle_WithSnapshot(t *testing.T) {
	p := writeLifecycleCmdRules(t, lifecycleRulesYAML)

	dir := t.TempDir()
	snapPath := filepath.Join(dir, "snap.json")
	if err := rules.SaveSnapshot(snapPath, map[string]rules.UsageStats{
		"http:requests:rate5m": {QueryCount: 50},
	}); err != nil {
		t.Fatal(err)
	}

	out := captureLifecycleOutput(t, []string{"--rules", p, "--usage", snapPath})
	if !strings.Contains(out, "http:requests:rate5m") {
		t.Errorf("expected rule in output, got: %s", out)
	}
}
