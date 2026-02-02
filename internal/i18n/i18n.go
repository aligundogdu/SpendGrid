package i18n

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"spendgrid/internal/config"
)

var (
	translations map[string]interface{}
	currentLang  string
	basePath     string
)

// Load initializes the i18n system
func Load() error {
	// Initialize global config first
	if err := config.Init(); err != nil {
		return fmt.Errorf("failed to init config: %v", err)
	}

	// Use configured language or detect from system
	globalLang := config.GetLanguage()
	if globalLang == "" {
		globalLang = detectLanguage()
	}

	// Set base path for locales
	basePath = findLocalePath()

	return LoadLanguage(globalLang)
}

// LoadLanguage loads a specific language
func LoadLanguage(lang string) error {
	currentLang = lang

	localePath := filepath.Join(basePath, fmt.Sprintf("%s.yml", lang))
	data, err := os.ReadFile(localePath)
	if err != nil {
		// Fallback to English
		if lang != "en" {
			return LoadLanguage("en")
		}
		return fmt.Errorf("failed to load translations for %s: %v", lang, err)
	}

	if err := yaml.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse translations: %v", err)
	}

	return nil
}

// T returns a translation string
func T(key string) string {
	val := getNestedValue(translations, key)
	if val == nil {
		return key
	}

	if str, ok := val.(string); ok {
		return str
	}

	return key
}

// Tfmt returns a formatted translation string
func Tfmt(key string, args ...interface{}) string {
	format := T(key)
	return fmt.Sprintf(format, args...)
}

// GetLanguage returns current language
func GetLanguage() string {
	return currentLang
}

func findLocalePath() string {
	// Try different possible paths
	paths := []string{
		"locales",
		"../locales",
		"../../locales",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "locales"
}

func detectLanguage() string {
	// Check global config first
	// TODO: Implement after global config is ready

	// Check environment variable
	if lang := os.Getenv("SPENDGRID_LANG"); lang != "" {
		return lang
	}

	// Check system locale
	if lang := os.Getenv("LANG"); lang != "" {
		if len(lang) >= 2 {
			// Extract language code (e.g., "en_US.UTF-8" -> "en")
			langCode := lang[:2]
			if langCode == "tr" {
				return "tr"
			}
		}
	}

	return "en"
}

func getNestedValue(data map[string]interface{}, key string) interface{} {
	// Handle dot notation (e.g., "commands.init.confirm")
	parts := []rune(key)
	start := 0
	current := data

	for i := 0; i <= len(parts); i++ {
		if i == len(parts) || parts[i] == '.' {
			if i > start {
				part := string(parts[start:i])
				if val, ok := current[part]; ok {
					switch v := val.(type) {
					case map[string]interface{}:
						current = v
					case string:
						return v
					default:
						return val
					}
				} else {
					return nil
				}
			}
			start = i + 1
		}
	}

	// If we've traversed all parts and ended up with a map, something went wrong
	return nil
}
