package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

const appName = "spendgrid"

// GlobalConfig represents the global configuration
type GlobalConfig struct {
	Language string `yaml:"language" json:"language"`
}

var (
	globalConfig     *GlobalConfig
	globalConfigPath string
	dataPath         string
)

// Init initializes the global config system
func Init() error {
	// Setup paths
	globalConfigPath = filepath.Join(xdg.ConfigHome, appName, "config")
	dataPath = filepath.Join(xdg.DataHome, appName, "data")

	// Create directories if they don't exist
	if err := os.MkdirAll(globalConfigPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Load or create global config
	if err := loadGlobalConfig(); err != nil {
		return err
	}

	// Initialize currency maps if they don't exist
	if err := initCurrencyMaps(); err != nil {
		return err
	}

	return nil
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *GlobalConfig {
	if globalConfig == nil {
		globalConfig = &GlobalConfig{
			Language: "en",
		}
	}
	return globalConfig
}

// SetLanguage sets the global language
func SetLanguage(lang string) error {
	if globalConfig == nil {
		globalConfig = &GlobalConfig{}
	}
	globalConfig.Language = lang
	return saveGlobalConfig()
}

// GetLanguage returns the configured language
func GetLanguage() string {
	if globalConfig == nil || globalConfig.Language == "" {
		return "en"
	}
	return globalConfig.Language
}

func loadGlobalConfig() error {
	configFile := filepath.Join(globalConfigPath, "settings.yml")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config
		globalConfig = &GlobalConfig{
			Language: "en",
		}
		return saveGlobalConfig()
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read global config: %v", err)
	}

	globalConfig = &GlobalConfig{}
	if err := yaml.Unmarshal(data, globalConfig); err != nil {
		return fmt.Errorf("failed to parse global config: %v", err)
	}

	return nil
}

func saveGlobalConfig() error {
	configFile := filepath.Join(globalConfigPath, "settings.yml")

	data, err := yaml.Marshal(globalConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal global config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write global config: %v", err)
	}

	return nil
}

func initCurrencyMaps() error {
	mapsFile := filepath.Join(dataPath, "currency_maps.json")

	if _, err := os.Stat(mapsFile); !os.IsNotExist(err) {
		return nil // Already exists
	}

	// Default currency mappings
	maps := map[string][]string{
		"TRY": {"TRY", "TL", "₺", "try", "tl"},
		"USD": {"USD", "$", "us", "dolar", "dollar"},
		"EUR": {"EUR", "€", "euro", "€"},
	}

	data, err := json.MarshalIndent(maps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal currency maps: %v", err)
	}

	if err := os.WriteFile(mapsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write currency maps: %v", err)
	}

	return nil
}

// GetDataPath returns the path to global data directory
func GetDataPath() string {
	return dataPath
}

// GetConfigPath returns the path to global config directory
func GetConfigPath() string {
	return globalConfigPath
}
