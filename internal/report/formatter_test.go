package report

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

var fixedTime = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func makeRows() []RuleRow {
	return []RuleRow{
		{Group: "g1", Name: "rule_active", LastSeen: "2h ago", Stale: false},
		{Group: "g1", Name: "rule_stale", LastSeen: "never", Stale: true},
	}
}

func TestFormatTable_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{Total: 2, Stale: 1, Active: 1, DryRun: false, RunAt: fixedTime}
	if err := FormatTable(&buf, makeRows(), s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, hdr := range []string{"GROUP", "NAME", "LAST SEEN", "STATUS"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}

func TestFormatTable_MarksStale(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{Total: 2, Stale: 1, Active: 1, RunAt: fixedTime}
	_ = FormatTable(&buf, makeRows(), s)
	out := buf.String()
	if !strings.Contains(out, "STALE") {
		t.Error("expected STALE marker in output")
	}
	if !strings.Contains(out, "active") {
		t.Error("expected active marker in output")
	}
}

func TestFormatTable_DryRunNotice(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{Total: 1, DryRun: true, RunAt: fixedTime}
	_ = FormatTable(&buf, makeRows(), s)
	if !strings.Contains(buf.String(), "DRY RUN") {
		t.Error("expected DRY RUN notice in output")
	}
}

func TestFormatJSON_Output(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{Total: 5, Active: 3, Stale: 2, DryRun: true, RunAt: fixedTime}
	if err := FormatJSON(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{`"total":5`, `"stale":2`, `"dry_run":true`} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in JSON output, got: %s", want, out)
		}
	}
}

func TestFormatJSON_NoDryRun(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{Total: 1, Active: 1, Stale: 0, DryRun: false, RunAt: fixedTime}
	_ = FormatJSON(&buf, s)
	if !strings.Contains(buf.String(), `"dry_run":false`) {
		t.Error("expected dry_run:false in output")
	}
}
