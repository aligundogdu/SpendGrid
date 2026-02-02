package currency

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

const appName = "spendgrid"

var (
	currencyMaps map[string][]string
	mapsLoaded   bool
)

// Normalize converts various currency representations to standard codes
// Examples: TL -> TRY, $ -> USD, € -> EUR, dolar -> USD
func Normalize(input string) string {
	if !mapsLoaded {
		if err := loadCurrencyMaps(); err != nil {
			// Fallback to hardcoded mappings
			return hardcodedNormalize(input)
		}
	}

	input = strings.ToUpper(strings.TrimSpace(input))

	// Check if it's already a standard code
	if _, exists := currencyMaps[input]; exists {
		return input
	}

	// Look through mappings
	for standardCode, variants := range currencyMaps {
		for _, variant := range variants {
			if strings.EqualFold(input, variant) {
				return standardCode
			}
		}
	}

	// Return original if not found
	return input
}

// IsValid checks if a currency code is valid
func IsValid(code string) bool {
	code = Normalize(code)
	validCodes := []string{"TRY", "USD", "EUR"}
	for _, valid := range validCodes {
		if code == valid {
			return true
		}
	}
	return false
}

// GetAllCurrencies returns all supported currency codes
func GetAllCurrencies() []string {
	return []string{"TRY", "USD", "EUR"}
}

func loadCurrencyMaps() error {
	mapsPath := filepath.Join(xdg.DataHome, appName, "data", "currency_maps.json")

	data, err := os.ReadFile(mapsPath)
	if err != nil {
		return fmt.Errorf("failed to read currency maps: %v", err)
	}

	if err := json.Unmarshal(data, &currencyMaps); err != nil {
		return fmt.Errorf("failed to parse currency maps: %v", err)
	}

	mapsLoaded = true
	return nil
}

// hardcodedNormalize provides fallback normalization
func hardcodedNormalize(input string) string {
	input = strings.ToUpper(strings.TrimSpace(input))

	switch input {
	case "TRY", "TL", "₺":
		return "TRY"
	case "USD", "$", "DOLAR", "DOLLAR":
		return "USD"
	case "EUR", "€", "EURO":
		return "EUR"
	default:
		return input
	}
}
