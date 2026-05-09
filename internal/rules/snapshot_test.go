package rules

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveSnapshot_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshot.json")

	usage := map[string]int64{
		"job:requests:rate5m": 42,
		"job:errors:rate5m":   0,
	}

	if err := SaveSnapshot(path, "rules.yaml", usage); err != nil {
		t.Fatalf("SaveSnapshot returned error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected snapshot file to exist")
	}
}

func TestSaveAndLoadSnapshot_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshot.json")

	usage := map[string]int64{
		"job:requests:rate5m": 7,
		"job:latency:p99":     3,
	}

	if err := SaveSnapshot(path, "myrules.yaml", usage); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	snap, err := LoadSnapshot(path)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}

	if snap.RulesFile != "myrules.yaml" {
		t.Errorf("expected RulesFile 'myrules.yaml', got %q", snap.RulesFile)
	}

	if snap.Usage["job:requests:rate5m"] != 7 {
		t.Errorf("expected usage 7, got %d", snap.Usage["job:requests:rate5m"])
	}

	if snap.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	if snap.CreatedAt.After(time.Now().Add(time.Second)) {
		t.Error("CreatedAt is in the future")
	}
}

func TestLoadSnapshot_NotFound(t *testing.T) {
	_, err := LoadSnapshot("/nonexistent/path/snapshot.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadSnapshot_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	if err := os.WriteFile(path, []byte("not-json{"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := LoadSnapshot(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
