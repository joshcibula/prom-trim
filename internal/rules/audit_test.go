package rules

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAudit_NotFound(t *testing.T) {
	_, err := LoadAudit("/nonexistent/audit.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestAppendAudit_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	events := []AuditEvent{
		{
			Timestamp: time.Now().UTC(),
			Action:    "prune",
			RuleName:  "job:requests:rate5m",
			RuleFile:  "rules.yaml",
			Reason:    "stale",
			DryRun:    false,
		},
	}
	if err := AppendAudit(path, events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestAppendAudit_Accumulates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	e1 := AuditEvent{Timestamp: time.Now().UTC(), Action: "prune", RuleName: "rule_a", DryRun: true}
	e2 := AuditEvent{Timestamp: time.Now().UTC(), Action: "schedule", RuleName: "rule_b", DryRun: false}

	if err := AppendAudit(path, []AuditEvent{e1}); err != nil {
		t.Fatal(err)
	}
	if err := AppendAudit(path, []AuditEvent{e2}); err != nil {
		t.Fatal(err)
	}

	log, err := LoadAudit(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(log.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(log.Events))
	}
}

func TestLoadAudit_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")
	_ = os.WriteFile(path, []byte("not-json"), 0o644)

	_, err := LoadAudit(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAuditEvent_Fields(t *testing.T) {
	ts := time.Now().UTC()
	e := AuditEvent{
		Timestamp: ts,
		Action:    "prune",
		RuleName:  "my:rule",
		RuleFile:  "rules/main.yaml",
		Reason:    "low usage",
		DryRun:    true,
	}
	if e.Action != "prune" {
		t.Errorf("unexpected action: %s", e.Action)
	}
	if !e.DryRun {
		t.Error("expected dry_run to be true")
	}
}
