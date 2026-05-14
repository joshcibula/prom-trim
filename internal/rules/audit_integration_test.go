package rules

import (
	"path/filepath"
	"testing"
	"time"
)

func TestAuditRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")

	now := time.Now().UTC().Truncate(time.Second)
	events := []AuditEvent{
		{
			Timestamp: now,
			Action:    "prune",
			RuleName:  "job:http_requests:rate5m",
			RuleFile:  "recording_rules.yaml",
			Reason:    "zero query count",
			DryRun:    false,
		},
		{
			Timestamp: now,
			Action:    "schedule",
			RuleName:  "job:errors:rate1m",
			RuleFile:  "recording_rules.yaml",
			Reason:    "last seen > 30d",
			DryRun:    true,
		},
	}

	if err := AppendAudit(path, events); err != nil {
		t.Fatalf("AppendAudit: %v", err)
	}

	loaded, err := LoadAudit(path)
	if err != nil {
		t.Fatalf("LoadAudit: %v", err)
	}

	if len(loaded.Events) != len(events) {
		t.Fatalf("expected %d events, got %d", len(events), len(loaded.Events))
	}
	for i, e := range loaded.Events {
		if e.RuleName != events[i].RuleName {
			t.Errorf("event[%d] RuleName: want %q, got %q", i, events[i].RuleName, e.RuleName)
		}
		if e.DryRun != events[i].DryRun {
			t.Errorf("event[%d] DryRun: want %v, got %v", i, events[i].DryRun, e.DryRun)
		}
	}
}
