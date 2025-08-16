package config

import (
	"fmt"
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
			Root:            "~/git-worktrees",
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

	needsMigration := false
	originalRoot := config.Settings.Root

	// Set defaults for missing values
	if config.Settings.Root == "" {
		homeDir, _ := os.UserHomeDir()
		config.Settings.Root = filepath.Join(homeDir, "git-worktrees")
	} else if strings.HasPrefix(config.Settings.Root, "~/") {
		// Expand tilde to home directory
		homeDir, _ := os.UserHomeDir()
		config.Settings.Root = filepath.Join(homeDir, config.Settings.Root[2:])
	} else if strings.HasPrefix(config.Settings.Root, "/Users/") || strings.HasPrefix(config.Settings.Root, "/home/") {
		// Detect absolute paths that should be portable
		homeDir, _ := os.UserHomeDir()
		
		// Check if this path points to a user's home directory
		if strings.HasPrefix(config.Settings.Root, "/Users/") {
			// macOS path format: /Users/username/...
			parts := strings.SplitN(config.Settings.Root, "/", 4)
			if len(parts) >= 4 {
				// Convert /Users/username/path to ~/path
				relativePath := parts[3]
				if strings.HasPrefix(relativePath, "git-worktrees") {
					config.Settings.Root = filepath.Join(homeDir, relativePath)
					needsMigration = true
				}
			}
		} else if strings.HasPrefix(config.Settings.Root, "/home/") {
			// Linux path format: /home/username/...
			parts := strings.SplitN(config.Settings.Root, "/", 4)
			if len(parts) >= 4 {
				// Convert /home/username/path to ~/path
				relativePath := parts[3]
				if strings.HasPrefix(relativePath, "git-worktrees") {
					config.Settings.Root = filepath.Join(homeDir, relativePath)
					needsMigration = true
				}
			}
		}
	}

	// Auto-migrate config if we detected non-portable paths
	if needsMigration {
		fmt.Fprintf(os.Stderr, "Migrating config: %s → ~/git-worktrees\n", originalRoot)
		if err := SaveConfig(&config); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not save migrated config: %v\n", err)
		}
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	// Create a copy to avoid modifying the original
	configCopy := *config
	
	// Contract home directory to ~ for portability
	homeDir, _ := os.UserHomeDir()
	if strings.HasPrefix(configCopy.Settings.Root, homeDir) {
		configCopy.Settings.Root = "~" + configCopy.Settings.Root[len(homeDir):]
	}
	
	data, err := yaml.Marshal(configCopy)
	if err != nil {
		return err
	}

	return os.WriteFile(".worktree.yaml", data, 0644)
}
