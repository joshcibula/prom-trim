package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/your-org/prom-trim/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "prom-trim-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	_ = f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTempConfig(t, `
prometheus:
  url: http://localhost:9090
  timeout: 10s
rules:
  rules_dir: /etc/prometheus/rules
  staleness_threshold: 168h
  dry_run: true
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Prometheus.URL != "http://localhost:9090" {
		t.Errorf("expected prometheus URL %q, got %q", "http://localhost:9090", cfg.Prometheus.URL)
	}
	if cfg.Prometheus.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", cfg.Prometheus.Timeout)
	}
	if cfg.Rules.StalenessThreshold != 168*time.Hour {
		t.Errorf("expected staleness_threshold 168h, got %v", cfg.Rules.StalenessThreshold)
	}
	if !cfg.Rules.DryRun {
		t.Error("expected dry_run to be true")
	}
}

func TestLoad_DefaultTimeout(t *testing.T) {
	path := writeTempConfig(t, `
prometheus:
  url: http://localhost:9090
rules:
  rules_dir: /tmp/rules
  staleness_threshold: 72h
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Prometheus.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Prometheus.Timeout)
	}
}

func TestLoad_MissingURL(t *testing.T) {
	path := writeTempConfig(t, `
rules:
  rules_dir: /tmp/rules
  staleness_threshold: 72h
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing prometheus.url")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
