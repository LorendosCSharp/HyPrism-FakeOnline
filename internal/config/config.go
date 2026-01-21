package config

import (
	"os"
	"path/filepath"

	"HyPrism/internal/env"

	"github.com/pelletier/go-toml/v2"
)

func configPath() string {
	return filepath.Join(env.GetDefaultAppDir(), "config.toml")
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configPath(), data, 0644)
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if err := Save(cfg); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
