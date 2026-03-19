package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	VaultPath  string            `yaml:"vault_path"`
	Editor     string            `yaml:"editor"`
	Ignore     []string          `yaml:"ignore"`
	Validation ValidationConfig  `yaml:"validation"`
}

type ValidationConfig struct {
	Enabled        bool     `yaml:"enabled"`
	IgnorePatterns []string `yaml:"ignore_patterns"`
	FuzzyThreshold float64  `yaml:"fuzzy_threshold"`
	StrictMode     bool     `yaml:"strict_mode"`
}

func LoadFrom(path string) (*Config, error) {
	cfg := &Config{}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("config: failed to read %s: %w", path, err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("config: failed to parse %s: %w", path, err)
		}
	}

	if cfg.VaultPath == "" {
		if envPath := os.Getenv("BEACON_VAULT_PATH"); envPath != "" {
			cfg.VaultPath = envPath
		}
	}

	if cfg.VaultPath == "" {
		return nil, fmt.Errorf("config: vault_path is required (set in config file or BEACON_VAULT_PATH env)")
	}

	applyDefaults(cfg)
	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Editor == "" {
		cfg.Editor = "vim"
	}
	if len(cfg.Ignore) == 0 {
		cfg.Ignore = []string{".obsidian"}
	}
	if cfg.Validation.FuzzyThreshold == 0 {
		cfg.Validation.FuzzyThreshold = 0.8
	}
}
