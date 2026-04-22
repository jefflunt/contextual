package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Atlassian        AtlassianConfig `yaml:"atlassian"`
	Planner          string          `yaml:"planner"`
	MaxContextLength int             `yaml:"max_context_length"`
}

type AtlassianConfig struct {
	Host           string `yaml:"host"`
	APIUser        string `yaml:"api_user"`
	APIToken       string `yaml:"api_token"`
	MaxSpiderJumps int    `yaml:"max_spider_jumps"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, ".contextual", "config.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.MaxContextLength == 0 {
		return nil, errors.New("max_context_length is required in config.yml")
	}
	return &cfg, nil
}
