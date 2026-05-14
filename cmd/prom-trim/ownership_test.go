package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeOwnershipCmdRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const ownershipRulesYAML = `
groups:
  - name: platform
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
        annotations:
          owner: alice
          team: platform-eng
      - record: job:errors:rate5m
        expr: rate(errors_total[5m])
        annotations:
          team: sre
`

func captureOwnershipOutput(t *testing.T, args []string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runOwnership(args)
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("runOwnership error: %v", err)
	}
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestRunOwnership_TableOutput(t *testing.T) {
	path := writeOwnershipCmdRules(t, ownershipRulesYAML)
	out := captureOwnershipOutput(t, []string{path})
	if !strings.Contains(out, "RULE") {
		t.Error("expected RULE header in table output")
	}
	if !strings.Contains(out, "alice") {
		t.Error("expected owner alice in output")
	}
	if !strings.Contains(out, "platform-eng") {
		t.Error("expected team platform-eng in output")
	}
}

func TestRunOwnership_JSONOutput(t *testing.T) {
	path := writeOwnershipCmdRules(t, ownershipRulesYAML)
	out := captureOwnershipOutput(t, []string{"-format=json", path})
	if !strings.Contains(out, "\"rule\"") {
		t.Error("expected json key 'rule' in output")
	}
	if !strings.Contains(out, "alice") {
		t.Error("expected owner alice in json output")
	}
}

func TestRunOwnership_ByTeam(t *testing.T) {
	path := writeOwnershipCmdRules(t, ownershipRulesYAML)
	out := captureOwnershipOutput(t, []string{"-by-team", path})
	if !strings.Contains(out, "TEAM") {
		t.Error("expected TEAM header in by-team output")
	}
	if !strings.Contains(out, "sre") {
		t.Error("expected team sre in by-team output")
	}
}

func TestRunOwnership_MissingFile(t *testing.T) {
	err := runOwnership([]string{"/no/such/file.yaml"})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestRunOwnership_NoArgs(t *testing.T) {
	err := runOwnership([]string{})
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}
