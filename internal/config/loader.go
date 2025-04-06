package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dackerman/asana-tasks-sorter/internal/core"
)

// LoadConfiguration loads the configuration from a file or returns defaults
func LoadConfiguration(configFile string) core.SectionConfig {
	// If the user explicitly asked for defaults
	if configFile == "default" {
		return core.DefaultSectionConfig()
	}
	
	// Try to load from the file
	loadedConfig, err := loadSectionConfig(configFile)
	if err == nil {
		return loadedConfig
	}
	
	fmt.Printf("Error loading section config: %v\nUsing default configuration\n", err)
	return core.DefaultSectionConfig()
}

// loadSectionConfig loads the section configuration from a JSON file
func loadSectionConfig(configPath string) (core.SectionConfig, error) {
	// Handle relative paths
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return core.SectionConfig{}, fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		configPath = absPath
	}

	// Read config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return core.SectionConfig{}, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config core.SectionConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return core.SectionConfig{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}