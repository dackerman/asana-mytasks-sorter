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
	config := core.DefaultSectionConfig()
	
	if configFile == "" {
		return config
	}
	
	loadedConfig, err := loadSectionConfig(configFile)
	if err == nil {
		return loadedConfig
	}
	
	fmt.Printf("Error loading section config: %v\nUsing default configuration\n", err)
	return config
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