package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration structure
type Config struct {
	Monitors []*Monitor `yaml:"monitors"`
}

// loadConfig loads and parses the YAML configuration file
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate configuration
	if len(config.Monitors) == 0 {
		return nil, fmt.Errorf("no monitors defined in configuration")
	}

	for _, monitor := range config.Monitors {
		if monitor.Name == "" {
			return nil, fmt.Errorf("monitor name is required")
		}
		if monitor.Endpoint == "" {
			return nil, fmt.Errorf("endpoint is required for monitor %s", monitor.Name)
		}
		if monitor.CheckInterval <= 0 {
			return nil, fmt.Errorf("invalid check interval for monitor %s: must be positive", monitor.Name)
		}
		if monitor.FailThreshold <= 0 {
			return nil, fmt.Errorf("invalid fail threshold for monitor %s: must be positive", monitor.Name)
		}

		// Initialize runtime state
		monitor.Status = "unknown"
		monitor.UnchangedCount = 0
	}

	return &config, nil
}
