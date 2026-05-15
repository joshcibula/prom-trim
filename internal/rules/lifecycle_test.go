package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeLifecycleRules(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestBuildLifecycleReport_FileNotFound(t *testing.T) {
	_, err := BuildLifecycleReport("/no/such/file.yaml", nil, DefaultStalenessConfig())
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildLifecycleReport_Empty(t *testing.T) {
	p := writeLifecycleRules(t, "groups: []\n")
	entries, err := BuildLifecycleReport(p, nil, DefaultStalenessConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestBuildLifecycleReport_StageClassification(t *testing.T) {
	rulesYAML := `groups:
  - name: test
    rules:
      - record: active:rule
        expr: sum(rate(http_requests_total[5m]))
      - record: candidate:rule
        expr: sum(rate(http_errors_total[5m]))
      - record: deprecated:rule
        expr: sum(rate(http_slow_total[5m]))
`
	p := writeLifecycleRules(t, rulesYAML)
	cfg := DefaultStalenessConfig()
	now := time.Now()

	usage := map[string]UsageStats{
		"active:rule":     {QueryCount: int64(cfg.MinQueryCount) + 5, LastSeen: now.Add(-1 * time.Hour)},
		"deprecated:rule": {QueryCount: 1, LastSeen: now.Add(-cfg.MaxAge * 2)},
	}

	entries, err := BuildLifecycleReport(p, usage, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	byName := make(map[string]LifecycleEntry)
	for _, e := range entries {
		byName[e.Rule] = e
	}

	if byName["active:rule"].Stage != StageActive {
		t.Errorf("expected active, got %s", byName["active:rule"].Stage)
	}
	if byName["candidate:rule"].Stage != StageCandidate {
		t.Errorf("expected candidate, got %s", byName["candidate:rule"].Stage)
	}
	if byName["deprecated:rule"].Stage != StageDeprecated {
		t.Errorf("expected deprecated, got %s", byName["deprecated:rule"].Stage)
	}
}

func TestSaveAndLoadLifecycle_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lifecycle.json")
	now := time.Now().Truncate(time.Second)

	orig := []LifecycleEntry{
		{Rule: "foo:bar", Stage: StageActive, Transition: now, Reason: "frequently queried"},
		{Rule: "baz:qux", Stage: StageDeprecated, Transition: now, Reason: "too old"},
	}

	if err := SaveLifecycle(path, orig); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := LoadLifecycle(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != len(orig) {
		t.Fatalf("expected %d entries, got %d", len(orig), len(loaded))
	}
	for i := range orig {
		if loaded[i].Rule != orig[i].Rule || loaded[i].Stage != orig[i].Stage {
			t.Errorf("entry %d mismatch", i)
		}
	}
}

func TestLoadLifecycle_NotFound(t *testing.T) {
	entries, err := LoadLifecycle("/no/such/lifecycle.json")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if entries != nil {
		t.Fatal("expected nil entries")
	}
}

func TestLoadLifecycle_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(p, []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadLifecycle(p)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLifecycleEntry_String(t *testing.T) {
	e := LifecycleEntry{
		Rule:       "my:rule",
		Stage:      StageWatched,
		Transition: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		Reason:     "aging",
	}
	s := e.String()
	for _, want := range []string{"my:rule", "watched", "2024-03-15", "aging"} {
		if !containsToken(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}

func TestSaveLifecycle_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	err := SaveLifecycle(path, []LifecycleEntry{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(path)
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}
