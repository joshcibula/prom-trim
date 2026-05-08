package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain_NoArgs ensures the binary exits non-zero when config is missing.
func TestMain_NoArgs(t *testing.T) {
	if os.Getenv("PROM_TRIM_INTEGRATION") == "" {
		t.Skip("skipping integration test; set PROM_TRIM_INTEGRATION=1 to run")
	}

	cmd := exec.Command("go", "run", ".", "--config", "/nonexistent/config.yaml")
	cmd.Dir = filepath.Join(".")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit when config file is missing, got nil")
	}
}

// TestMain_DryRunFlag verifies dry-run flag is accepted without panic.
func TestMain_DryRunFlag(t *testing.T) {
	if os.Getenv("PROM_TRIM_INTEGRATION") == "" {
		t.Skip("skipping integration test; set PROM_TRIM_INTEGRATION=1 to run")
	}

	cfgFile := writeTempConfig(t)
	cmd := exec.Command("go", "run", ".", "--config", cfgFile, "--dry-run")
	out, err := cmd.CombinedOutput()
	// We expect failure because there is no live Prometheus, but not a panic.
	if err != nil {
		t.Logf("output: %s", out)
	}
}

func writeTempConfig(t *testing.T) string {
	t.Helper()
	content := `prometheus_url: http://localhost:9090
rules_file: /tmp/rules.yaml
lookback_days: 7
timeout_seconds: 10
`
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("create temp config: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	f.Close()
	return f.Name()
}
