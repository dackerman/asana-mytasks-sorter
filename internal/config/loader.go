package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
)

// LoadConfiguration loads the configuration from a file or returns defaults
func LoadConfiguration(configFile string) asana.SectionConfig {
	config := asana.DefaultSectionConfig()
	
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
func loadSectionConfig(configPath string) (asana.SectionConfig, error) {
	// Handle relative paths
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return asana.SectionConfig{}, fmt.Errorf("failed to resolve absolute path: %v", err)
		}
		configPath = absPath
	}

	// Read config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return asana.SectionConfig{}, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse JSON
	var config asana.SectionConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return asana.SectionConfig{}, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}