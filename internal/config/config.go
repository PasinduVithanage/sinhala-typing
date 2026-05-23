package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all user-configurable settings.
type Config struct {
	MappingProfile   string `json:"mapping_profile"`
	AutoStart        bool   `json:"auto_start"`
	AutoFixClipboard bool   `json:"auto_fix_clipboard"`
	UseZWJClusters   bool   `json:"use_zwj_clusters"`
	ToggleHotkey     string `json:"toggle_hotkey"`
	RunAtStartup     bool   `json:"run_at_startup"`
}

func defaultConfig() *Config {
	return &Config{
		MappingProfile:   "phonetic",
		AutoStart:        false,
		AutoFixClipboard: false,
		UseZWJClusters:   true,
		ToggleHotkey:     "Ctrl+Alt+S",
	}
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sinhala-assistant", "config.json"), nil
}

// Load reads the config file, returning defaults if it doesn't exist.
func Load() *Config {
	path, err := configPath()
	if err != nil {
		return defaultConfig()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig()
	}
	cfg := defaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return defaultConfig()
	}
	return cfg
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
