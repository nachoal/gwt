package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version  int      `yaml:"version"`
	Copy     []string `yaml:"copy"`
	Setup    []string `yaml:"setup"`
	Settings Settings `yaml:"settings"`
}

type Settings struct {
	Root            string `yaml:"root"`
	AutoCleanMerged bool   `yaml:"auto_clean_merged"`
	ConfirmDelete   bool   `yaml:"confirm_delete"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Version: 1,
		Copy: []string{
			".env",
			".env.local",
		},
		Setup: []string{
			"npm install",
		},
		Settings: Settings{
			Root:            filepath.Join(homeDir, "git-worktrees"),
			AutoCleanMerged: true,
			ConfirmDelete:   true,
		},
	}
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(".worktree.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults for missing values
	if config.Settings.Root == "" {
		homeDir, _ := os.UserHomeDir()
		config.Settings.Root = filepath.Join(homeDir, "git-worktrees")
	} else if strings.HasPrefix(config.Settings.Root, "~/") {
		// Expand tilde to home directory
		homeDir, _ := os.UserHomeDir()
		config.Settings.Root = filepath.Join(homeDir, config.Settings.Root[2:])
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(".worktree.yaml", data, 0644)
}
