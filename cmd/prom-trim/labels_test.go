package main

import (
	"os"
	"strings"
	"testing"
)

const labelsTestRules = `
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
          env: staging
`

func writeLabelsCmdRules(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "rules-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(labelsTestRules)
	f.Close()
	return f.Name()
}

func TestRunLabels_TableOutput(t *testing.T) {
	path := writeLabelsCmdRules(t)
	// Redirect stdout capture via pipe would be complex; test via runLabels not panicking.
	err := runLabels(path, "table", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLabels_JSONOutput(t *testing.T) {
	path := writeLabelsCmdRules(t)
	err := runLabels(path, "json", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLabels_IndexTable(t *testing.T) {
	path := writeLabelsCmdRules(t)
	err := runLabels(path, "table", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLabels_IndexJSON(t *testing.T) {
	path := writeLabelsCmdRules(t)
	err := runLabels(path, "json", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLabels_FileNotFound(t *testing.T) {
	err := runLabels("/no/such/file.yaml", "table", false)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "extracting labels") {
		t.Errorf("unexpected error message: %v", err)
	}
}
