package exchange

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/xdg"
)

const appName = "spendgrid"

// TCMBClient implements the TCMB (Central Bank of Turkey) API
type TCMBClient struct {
	BaseURL string
}

// NewTCMBClient creates a new TCMB client
func NewTCMBClient() *TCMBClient {
	return &TCMBClient{
		BaseURL: "https://www.tcmb.gov.tr/kurlar",
	}
}

// TCMBResponse represents the XML response from TCMB
type TCMBResponse struct {
	XMLName xml.Name   `xml:"Tarih_Date"`
	Date    string     `xml:"Tarih,attr"`
	Rates   []TCMBRate `xml:"Currency"`
}

// TCMBRate represents a single currency rate from TCMB
type TCMBRate struct {
	Code            string `xml:"CurrencyCode,attr"`
	Name            string `xml:"CurrencyName"`
	Unit            int    `xml:"Unit"`
	ForexBuying     string `xml:"ForexBuying"`
	ForexSelling    string `xml:"ForexSelling"`
	BanknoteBuying  string `xml:"BanknoteBuying"`
	BanknoteSelling string `xml:"BanknoteSelling"`
}

// FetchRates fetches exchange rates from TCMB for a specific date
func (c *TCMBClient) FetchRates(date time.Time) (map[string]float64, error) {
	// TCMB format: https://www.tcmb.gov.tr/kurlar/202602/02022026.xml
	// Format: YYYYMM/DDMMYYYY.xml
	dateStr := date.Format("02012006")
	monthStr := date.Format("200601")

	url := fmt.Sprintf("%s/%s/%s.xml", c.BaseURL, monthStr, dateStr)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch TCMB rates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Try today if specific date fails (weekend/holiday)
		if !isToday(date) {
			return c.FetchRates(time.Now())
		}
		return nil, fmt.Errorf("TCMB API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var tcmbResp TCMBResponse
	if err := xml.Unmarshal(body, &tcmbResp); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	rates := make(map[string]float64)
	// TCMB rates are per unit and in TRY
	for _, rate := range tcmbResp.Rates {
		if rate.ForexBuying != "" {
			val, err := strconv.ParseFloat(strings.ReplaceAll(rate.ForexBuying, ",", "."), 64)
			if err == nil {
				// Rate is for the unit specified (usually 1)
				rates[rate.Code] = val / float64(rate.Unit)
			}
		}
	}

	return rates, nil
}

// FrankfurtClient implements the European Central Bank (Frankfurt) API
type FrankfurtClient struct {
	BaseURL string
}

// NewFrankfurtClient creates a new Frankfurt/ECB client
func NewFrankfurtClient() *FrankfurtClient {
	return &FrankfurtClient{
		BaseURL: "https://api.frankfurter.app",
	}
}

// FrankfurtResponse represents the JSON response from Frankfurt API
type FrankfurtResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

// FetchRates fetches exchange rates from Frankfurt API for a specific date
func (c *FrankfurtClient) FetchRates(date time.Time) (map[string]float64, error) {
	dateStr := date.Format("2006-01-02")
	url := fmt.Sprintf("%s/%s?from=EUR", c.BaseURL, dateStr)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Frankfurt rates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Try today if specific date fails
		if !isToday(date) {
			return c.FetchRates(time.Now())
		}
		return nil, fmt.Errorf("Frankfurt API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var frankfurtResp FrankfurtResponse
	if err := json.Unmarshal(body, &frankfurtResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Frankfurt rates are EUR-based, convert to direct rates
	rates := make(map[string]float64)
	for currency, rate := range frankfurtResp.Rates {
		rates[currency] = rate
	}

	// Add EUR rate
	rates["EUR"] = 1.0

	return rates, nil
}

// ExchangeRateCache handles caching of exchange rates
type ExchangeRateCache struct {
	Rates map[string]map[string]float64 `json:"rates"` // date -> currency -> rate
}

// GetCachePath returns the path to the exchange rates cache file
func GetCachePath() string {
	return filepath.Join(xdg.DataHome, appName, "data", "exchange_rates.json")
}

// LoadCache loads the exchange rate cache from disk
func LoadCache() (*ExchangeRateCache, error) {
	cachePath := GetCachePath()

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return &ExchangeRateCache{
			Rates: make(map[string]map[string]float64),
		}, nil
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache: %v", err)
	}

	var cache ExchangeRateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %v", err)
	}

	if cache.Rates == nil {
		cache.Rates = make(map[string]map[string]float64)
	}

	return &cache, nil
}

// SaveCache saves the exchange rate cache to disk
func (c *ExchangeRateCache) SaveCache() error {
	cachePath := GetCachePath()

	// Ensure directory exists
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %v", err)
	}

	return os.WriteFile(cachePath, data, 0644)
}

// GetRate gets a specific exchange rate for a date
func (c *ExchangeRateCache) GetRate(date string, currency string) (float64, bool) {
	if dateRates, ok := c.Rates[date]; ok {
		if rate, ok := dateRates[currency]; ok {
			return rate, true
		}
	}
	return 0, false
}

// SetRate sets an exchange rate for a date
func (c *ExchangeRateCache) SetRate(date string, currency string, rate float64) {
	if c.Rates[date] == nil {
		c.Rates[date] = make(map[string]float64)
	}
	c.Rates[date][currency] = rate
}

// FetchAndCacheRates fetches rates from appropriate API and caches them
func FetchAndCacheRates(date time.Time, preferTCMB bool) error {
	cache, err := LoadCache()
	if err != nil {
		return err
	}

	dateStr := date.Format("2006-01-02")

	// Check if we already have rates for this date
	if _, ok := cache.GetRate(dateStr, "USD"); ok {
		return nil // Already cached
	}

	var rates map[string]float64

	if preferTCMB {
		// Try TCMB first
		tcmb := NewTCMBClient()
		rates, err = tcmb.FetchRates(date)
		if err != nil {
			// Fallback to Frankfurt
			frankfurt := NewFrankfurtClient()
			rates, err = frankfurt.FetchRates(date)
		}
	} else {
		// Try Frankfurt first
		frankfurt := NewFrankfurtClient()
		rates, err = frankfurt.FetchRates(date)
		if err != nil {
			// Fallback to TCMB
			tcmb := NewTCMBClient()
			rates, err = tcmb.FetchRates(date)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to fetch rates: %v", err)
	}

	// Cache the rates
	for currency, rate := range rates {
		cache.SetRate(dateStr, currency, rate)
	}

	return cache.SaveCache()
}

// GetExchangeRate gets the exchange rate for converting to base currency (TRY)
func GetExchangeRate(date time.Time, fromCurrency string) (float64, error) {
	fromCurrency = strings.ToUpper(fromCurrency)

	// If already TRY, return 1
	if fromCurrency == "TRY" || fromCurrency == "TL" {
		return 1.0, nil
	}

	cache, err := LoadCache()
	if err != nil {
		return 0, err
	}

	dateStr := date.Format("2006-01-02")

	// Try to get from cache
	if rate, ok := cache.GetRate(dateStr, fromCurrency); ok {
		return rate, nil
	}

	// Fetch from API (use TCMB by default for Turkish users)
	preferTCMB := true // This could be configurable
	if err := FetchAndCacheRates(date, preferTCMB); err != nil {
		return 0, err
	}

	// Try again from cache
	if rate, ok := cache.GetRate(dateStr, fromCurrency); ok {
		return rate, nil
	}

	return 0, fmt.Errorf("exchange rate not found for %s on %s", fromCurrency, dateStr)
}

// ConvertAmount converts an amount from one currency to another
func ConvertAmount(amount float64, fromCurrency, toCurrency string, date time.Time) (float64, error) {
	fromCurrency = strings.ToUpper(fromCurrency)
	toCurrency = strings.ToUpper(toCurrency)

	if fromCurrency == toCurrency {
		return amount, nil
	}

	// Convert to base currency (TRY) first
	fromRate, err := GetExchangeRate(date, fromCurrency)
	if err != nil {
		return 0, err
	}

	// amount in TRY
	amountInBase := amount * fromRate

	if toCurrency == "TRY" || toCurrency == "TL" {
		return amountInBase, nil
	}

	// Convert from base to target
	toRate, err := GetExchangeRate(date, toCurrency)
	if err != nil {
		return 0, err
	}

	return amountInBase / toRate, nil
}

// RefreshRates forces a refresh of exchange rates
func RefreshRates() error {
	today := time.Now()
	preferTCMB := true
	return FetchAndCacheRates(today, preferTCMB)
}

// SetManualRate sets a manual exchange rate for a specific date
func SetManualRate(dateStr, currency string, rate float64) error {
	cache, err := LoadCache()
	if err != nil {
		return err
	}

	cache.SetRate(dateStr, strings.ToUpper(currency), rate)
	return cache.SaveCache()
}

func isToday(date time.Time) bool {
	today := time.Now()
	return date.Year() == today.Year() &&
		date.Month() == today.Month() &&
		date.Day() == today.Day()
}
