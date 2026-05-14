package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/prom-trim/internal/rules"
)

func writeAuditFile(t *testing.T, events []rules.AuditEvent) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")
	data, _ := json.MarshalIndent(rules.AuditLog{Events: events}, "", "  ")
	_ = os.WriteFile(path, data, 0o644)
	return path
}

func TestRunAudit_Empty(t *testing.T) {
	path := writeAuditFile(t, []rules.AuditEvent{})
	if err := runAudit(path, "table"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAudit_NotFound(t *testing.T) {
	err := runAudit("/nonexistent/audit.json", "table")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
}

func TestRunAudit_PrintsTable(t *testing.T) {
	events := []rules.AuditEvent{
		{
			Timestamp: time.Now().UTC(),
			Action:    "prune",
			RuleName:  "job:requests:rate5m",
			RuleFile:  "rules.yaml",
			Reason:    "stale",
			DryRun:    false,
		},
	}
	path := writeAuditFile(t, events)
	if err := runAudit(path, "table"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAudit_JSONFormat(t *testing.T) {
	events := []rules.AuditEvent{
		{
			Timestamp: time.Now().UTC(),
			Action:    "schedule",
			RuleName:  "job:errors:rate1m",
			RuleFile:  "rules.yaml",
			Reason:    "low count",
			DryRun:    true,
		},
	}
	path := writeAuditFile(t, events)
	if err := runAudit(path, "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
