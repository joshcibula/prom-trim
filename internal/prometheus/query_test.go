package prometheus

import (
	"testing"
	"time"
)

func TestDurationToPromQL(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{"hours", 24 * time.Hour, "24h"},
		{"minutes", 30 * time.Minute, "30m"},
		{"seconds", 45 * time.Second, "45s"},
		{"mixed falls back to seconds", 90 * time.Second, "90s"},
		{"one hour", time.Hour, "1h"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := durationToPromQL(tc.input)
			if got != tc.expected {
				t.Errorf("durationToPromQL(%v) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestFetchRuleUsage_EmptyInput(t *testing.T) {
	c := &Client{} // zero-value client is fine for empty input

	result, err := c.FetchRuleUsage(nil, nil, time.Hour) //nolint:staticcheck
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty rule list, got %v", result)
	}
}

func TestRuleUsage_Fields(t *testing.T) {
	now := time.Now()
	ru := RuleUsage{
		RuleName:   "job:requests:rate5m",
		QueryCount: 42,
		LastSeen:   now,
	}

	if ru.RuleName != "job:requests:rate5m" {
		t.Errorf("unexpected RuleName: %s", ru.RuleName)
	}
	if ru.QueryCount != 42 {
		t.Errorf("unexpected QueryCount: %f", ru.QueryCount)
	}
	if !ru.LastSeen.Equal(now) {
		t.Errorf("unexpected LastSeen: %v", ru.LastSeen)
	}
}

func TestFetchRuleUsage_NilRules(t *testing.T) {
	c := &Client{}

	// Both nil and empty slice should return nil without error.
	for _, rules := range [][]string{nil, {}} {
		result, err := c.FetchRuleUsage(nil, rules, time.Hour)
		if err != nil {
			t.Fatalf("unexpected error for rules=%v: %v", rules, err)
		}
		if result != nil {
			t.Errorf("expected nil result for rules=%v, got %v", rules, result)
		}
	}
}
