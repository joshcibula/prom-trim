package rules

import (
	"os"
	"path/filepath"
	"testing"
)

const diffOriginal = `groups:
  - name: group_a
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
      - record: job:errors:rate5m
        expr: rate(errors_total[5m])
  - name: group_b
    rules:
      - record: instance:cpu:avg
        expr: avg(cpu_usage)
`

const diffPruned = `groups:
  - name: group_a
    rules:
      - record: job:requests:rate5m
        expr: rate(requests_total[5m])
  - name: group_b
    rules:
      - record: instance:cpu:avg
        expr: avg(cpu_usage)
`

func writeDiffFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

func TestDiff_DetectsRemovedRule(t *testing.T) {
	dir := t.TempDir()
	orig := writeDiffFile(t, dir, "orig.yaml", diffOriginal)
	pruned := writeDiffFile(t, dir, "pruned.yaml", diffPruned)

	result, err := Diff(orig, pruned)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	if len(result.Removed) != 1 {
		t.Fatalf("expected 1 removed, got %d", len(result.Removed))
	}
	if result.Removed[0].Record != "job:errors:rate5m" {
		t.Errorf("unexpected removed record: %s", result.Removed[0].Record)
	}
	if result.Removed[0].Group != "group_a" {
		t.Errorf("unexpected removed group: %s", result.Removed[0].Group)
	}
}

func TestDiff_RetainedCount(t *testing.T) {
	dir := t.TempDir()
	orig := writeDiffFile(t, dir, "orig.yaml", diffOriginal)
	pruned := writeDiffFile(t, dir, "pruned.yaml", diffPruned)

	result, err := Diff(orig, pruned)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	if len(result.Retained) != 2 {
		t.Errorf("expected 2 retained, got %d", len(result.Retained))
	}
}

func TestDiff_NoDifference(t *testing.T) {
	dir := t.TempDir()
	orig := writeDiffFile(t, dir, "orig.yaml", diffPruned)
	pruned := writeDiffFile(t, dir, "pruned.yaml", diffPruned)

	result, err := Diff(orig, pruned)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	if len(result.Removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(result.Removed))
	}
}

func TestDiff_StringOutput(t *testing.T) {
	dir := t.TempDir()
	orig := writeDiffFile(t, dir, "orig.yaml", diffOriginal)
	pruned := writeDiffFile(t, dir, "pruned.yaml", diffPruned)

	result, err := Diff(orig, pruned)
	if err != nil {
		t.Fatalf("Diff error: %v", err)
	}

	out := result.String()
	if out == "" {
		t.Error("expected non-empty String() output")
	}
}

func TestDiff_OriginalNotFound(t *testing.T) {
	dir := t.TempDir()
	pruned := writeDiffFile(t, dir, "pruned.yaml", diffPruned)

	_, err := Diff(filepath.Join(dir, "missing.yaml"), pruned)
	if err == nil {
		t.Error("expected error for missing original file")
	}
}
