package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeValidatorRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp rules: %v", err)
	}
	return p
}

func TestValidateFile_Valid(t *testing.T) {
	p := writeValidatorRules(t, `
groups:
  - name: test_group
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
`)
	if err := ValidateFile(p); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateFile_NotFound(t *testing.T) {
	err := ValidateFile("/nonexistent/path/rules.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestValidateFile_EmptyGroups(t *testing.T) {
	p := writeValidatorRules(t, `groups: []`)
	err := ValidateFile(p)
	if err == nil || !strings.Contains(err.Error(), "no groups") {
		t.Errorf("expected 'no groups' error, got: %v", err)
	}
}

func TestValidateFile_EmptyGroupName(t *testing.T) {
	p := writeValidatorRules(t, `
groups:
  - name: ""
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
`)
	err := ValidateFile(p)
	if err == nil {
		t.Error("expected validation error for empty group name")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Issues) == 0 {
		t.Error("expected at least one issue")
	}
}

func TestValidationError_Message(t *testing.T) {
	ve := &ValidationError{Issues: []string{"issue one", "issue two"}}
	msg := ve.Error()
	if !strings.Contains(msg, "2 issue(s)") {
		t.Errorf("unexpected message: %s", msg)
	}
}
