package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level configuration for prom-trim.
type Config struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Rules      RulesConfig      `yaml:"rules"`
}

// PrometheusConfig holds connection settings for the Prometheus instance.
type PrometheusConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

// RulesConfig holds settings that control how stale rules are identified.
type RulesConfig struct {
	// RulesDir is the directory containing Prometheus rule files to evaluate.
	RulesDir string `yaml:"rules_dir"`
	// StalenessThreshold is the minimum period of zero usage before a rule is
	// considered stale.
	StalenessThreshold time.Duration `yaml:"staleness_threshold"`
	// DryRun prints proposed changes without modifying any files.
	DryRun bool `yaml:"dry_run"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Prometheus.URL == "" {
		return fmt.Errorf("prometheus.url must not be empty")
	}
	if c.Rules.RulesDir == "" {
		return fmt.Errorf("rules.rules_dir must not be empty")
	}
	if c.Rules.StalenessThreshold <= 0 {
		return fmt.Errorf("rules.staleness_threshold must be a positive duration")
	}
	if c.Prometheus.Timeout <= 0 {
		c.Prometheus.Timeout = 30 * time.Second
	}
	return nil
}
