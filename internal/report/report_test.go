package report_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/prom-trim/internal/report"
)

var sampleResults = []report.RuleResult{
	{Name: "job:requests:rate5m", Group: "api", QueryCount: 42, Stale: false},
	{Name: "job:errors:rate5m", Group: "api", QueryCount: 0, Stale: true},
	{Name: "instance:cpu:avg", Group: "infra", QueryCount: 7, Stale: false, LastSeen: time.Now()},
}

func TestWrite_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	summary := report.BuildSummary(sampleResults, false)
	if err := report.Write(&buf, sampleResults, summary); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, hdr := range []string{"RULE", "GROUP", "QUERY COUNT", "STATUS"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}

func TestWrite_MarksStaleRules(t *testing.T) {
	var buf bytes.Buffer
	summary := report.BuildSummary(sampleResults, false)
	_ = report.Write(&buf, sampleResults, summary)
	out := buf.String()
	if !strings.Contains(out, "STALE") {
		t.Error("expected at least one STALE entry in output")
	}
}

func TestWrite_DryRunNotice(t *testing.T) {
	var buf bytes.Buffer
	summary := report.BuildSummary(sampleResults, true)
	_ = report.Write(&buf, sampleResults, summary)
	if !strings.Contains(buf.String(), "dry-run") {
		t.Error("expected dry-run notice in output")
	}
}

func TestBuildSummary_Counts(t *testing.T) {
	s := report.BuildSummary(sampleResults, false)
	if s.Total != 3 {
		t.Errorf("Total: want 3, got %d", s.Total)
	}
	if s.Stale != 1 {
		t.Errorf("Stale: want 1, got %d", s.Stale)
	}
	if s.Kept != 2 {
		t.Errorf("Kept: want 2, got %d", s.Kept)
	}
}

func TestBuildSummary_Empty(t *testing.T) {
	s := report.BuildSummary(nil, false)
	if s.Total != 0 || s.Stale != 0 || s.Kept != 0 {
		t.Errorf("expected all zeros for empty input, got %+v", s)
	}
}
