package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	VaultPath    string            `yaml:"vault_path"`
	TemplatesDir string            `yaml:"templates_dir"`
	Editor       string            `yaml:"editor"`
	Ignore       []string          `yaml:"ignore"`
	Validation   ValidationConfig  `yaml:"validation"`
	TypePaths    map[string]string `yaml:"type_paths"`
	Daily        DailyConfig       `yaml:"daily"`
}

type DailyConfig struct {
	DateFormat string `yaml:"date_format"` // Go reference time format, default: "2006-01-02"
	Folder     string `yaml:"folder"`      // Overrides type_paths["daily"] if set
	Template   string `yaml:"template"`    // Template name, default: "daily"
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
	if cfg.TemplatesDir == "" {
		cfg.TemplatesDir = "700 - Recursos/Templates"
	}
	if len(cfg.Ignore) == 0 {
		cfg.Ignore = []string{".obsidian"}
	}
	if cfg.Validation.FuzzyThreshold == 0 {
		cfg.Validation.FuzzyThreshold = 0.8
	}
	if cfg.TypePaths == nil {
		cfg.TypePaths = map[string]string{
			"daily":     "100 - Diário",
			"projects":  "200 - Projetos",
			"resources": "700 - Recursos",
			"work":      "300 - Trabalho",
			"personal":  "400 - Pessoal",
		}
	}
	if cfg.Daily.DateFormat == "" {
		cfg.Daily.DateFormat = "2006-01-02"
	}
	if cfg.Daily.Template == "" {
		cfg.Daily.Template = "daily"
	}
	if cfg.Daily.Folder == "" {
		cfg.Daily.Folder = cfg.TypePaths["daily"]
	}
	// Fallback to default if type_paths was set without a "daily" key
	if cfg.Daily.Folder == "" {
		cfg.Daily.Folder = "100 - Diário"
	}
}
