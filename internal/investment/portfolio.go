package investment

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"spendgrid/internal/i18n"
	"spendgrid/internal/parser"
)

// Investment represents a single investment position
type Investment struct {
	Symbol       string
	Name         string
	Type         string // stock, gold, crypto, etc.
	TotalShares  float64
	TotalCost    float64
	Currency     string
	Transactions []InvestmentTransaction
}

// InvestmentTransaction represents a single buy transaction
type InvestmentTransaction struct {
	Date     time.Time
	Shares   float64
	Price    float64
	Currency string
}

// Portfolio holds all investments
type Portfolio map[string]*Investment

// ParseInvestmentFormat parses the special investment format
// Format: +1000TRY TUPRS(10 * 100TRY) or +1000TRY Altın(10gr * 100TRY)
func ParseInvestmentFormat(desc string) (*InvestmentTransaction, string, bool) {
	// Regex to match: SYMBOL(QUANTITY * PRICE)
	// Examples: TUPRS(10 * 100TRY), Altın(10gr * 100TL), BTC(0.5 * 50000USD)
	re := regexp.MustCompile(`([A-Za-z0-9ğüşöçıİĞÜŞÖÇ]+)\(([\d.,]+)([A-Za-z]*)\s*\*\s*([\d.,]+)\s*([A-Za-z₺$€]+)\)`)
	matches := re.FindStringSubmatch(desc)

	if len(matches) != 6 {
		return nil, "", false
	}

	symbol := matches[1]
	quantityStr := matches[2]
	unitType := matches[3] // gr, kg, etc. for commodities
	priceStr := matches[4]
	currency := matches[5]

	// Normalize currency
	currency = normalizeCurrency(currency)

	// Parse quantity
	quantity, err := strconv.ParseFloat(strings.ReplaceAll(quantityStr, ",", ""), 64)
	if err != nil {
		return nil, "", false
	}

	// Parse price
	price, err := strconv.ParseFloat(strings.ReplaceAll(priceStr, ",", ""), 64)
	if err != nil {
		return nil, "", false
	}

	// Determine investment type
	invType := determineInvestmentType(symbol, unitType)

	return &InvestmentTransaction{
		Shares:   quantity,
		Price:    price,
		Currency: currency,
	}, invType, true
}

// CalculatePortfolio scans all transactions and calculates portfolio
type CalculatePortfolio func() (*Portfolio, error)

func CalculatePortfolioFromTransactions() (*Portfolio, error) {
	portfolio := make(Portfolio)

	// Get current year
	now := time.Now()
	year := strconv.Itoa(now.Year())

	// Parse all month files
	for month := 1; month <= 12; month++ {
		monthFile := parser.GetMonthFile(month)
		filePath := filepath.Join(year, monthFile)

		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip if file doesn't exist
		}

		parsed, _ := parser.ParseMonthFile(string(content))

		for _, tx := range parsed {
			// Check if this is an investment transaction
			// Look for #invesment# tag (system tag)
			isInvestment := false
			for _, tag := range tx.Tags {
				if tag == "invesment" || tag == "investment" {
					isInvestment = true
					break
				}
			}

			if !isInvestment {
				continue
			}

			// Try to parse investment format from description
			invTx, invType, ok := ParseInvestmentFormat(tx.Description)
			if !ok {
				// Try alternative format: description contains symbol in parentheses
				continue
			}

			invTx.Date = time.Date(now.Year(), time.Month(month), tx.Day, 0, 0, 0, 0, time.UTC)

			// Extract symbol from description
			symbol := extractSymbol(tx.Description)
			if symbol == "" {
				continue
			}

			// Add to portfolio
			if portfolio[symbol] == nil {
				portfolio[symbol] = &Investment{
					Symbol:       symbol,
					Type:         invType,
					Currency:     invTx.Currency,
					Transactions: []InvestmentTransaction{},
				}
			}

			portfolio[symbol].TotalShares += invTx.Shares
			portfolio[symbol].TotalCost += invTx.Shares * invTx.Price
			portfolio[symbol].Transactions = append(portfolio[symbol].Transactions, *invTx)
		}
	}

	return &portfolio, nil
}

// GenerateInvestmentReport generates the investment report
func GenerateInvestmentReport() error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	portfolio, err := CalculatePortfolioFromTransactions()
	if err != nil {
		return err
	}

	if len(*portfolio) == 0 {
		fmt.Println(i18n.T("investment.no_investments"))
		return nil
	}

	// Print header
	fmt.Println()
	fmt.Println(i18n.T("investment.header"))
	fmt.Println(strings.Repeat("=", 80))

	// Print each investment
	fmt.Printf("%-15s %10s %15s %15s %15s\n", "Symbol", "Adet", "Ort. Maliyet", "Toplam Maliyet", "Para Birimi")
	fmt.Println(strings.Repeat("-", 80))

	for symbol, inv := range *portfolio {
		avgCost := 0.0
		if inv.TotalShares > 0 {
			avgCost = inv.TotalCost / inv.TotalShares
		}

		fmt.Printf("%-15s %10.2f %15.2f %15.2f %15s\n",
			symbol,
			inv.TotalShares,
			avgCost,
			inv.TotalCost,
			inv.Currency)

		// Show details
		if len(inv.Transactions) > 1 {
			fmt.Printf("  └─ %d işlem\n", len(inv.Transactions))
		}
	}

	fmt.Println()
	return nil
}

// Helper functions

func normalizeCurrency(curr string) string {
	curr = strings.ToUpper(strings.TrimSpace(curr))
	switch curr {
	case "TL", "₺":
		return "TRY"
	case "$":
		return "USD"
	case "€":
		return "EUR"
	default:
		return curr
	}
}

func determineInvestmentType(symbol, unitType string) string {
	symbolUpper := strings.ToUpper(symbol)

	// Gold/Silver
	if strings.Contains(symbolUpper, "ALTIN") || strings.Contains(symbolUpper, "GOLD") ||
		strings.Contains(symbolUpper, "GRAM") || strings.Contains(symbolUpper, "GR") {
		return "gold"
	}
	if strings.Contains(symbolUpper, "GÜMÜŞ") || strings.Contains(symbolUpper, "SILVER") {
		return "silver"
	}

	// Crypto
	if symbolUpper == "BTC" || symbolUpper == "ETH" || symbolUpper == "XRP" ||
		symbolUpper == "SOL" || symbolUpper == "ADA" || len(symbolUpper) <= 5 {
		// Check if it looks like a crypto symbol (short and uppercase)
		if len(symbolUpper) <= 5 && strings.ToUpper(symbol) == symbol {
			return "crypto"
		}
	}

	// Default to stock
	return "stock"
}

func extractSymbol(desc string) string {
	// Try to extract symbol from parentheses
	re := regexp.MustCompile(`([A-Za-z0-9ğüşöçıİĞÜŞÖÇ]+)\(`)
	matches := re.FindStringSubmatch(desc)
	if len(matches) > 1 {
		return strings.ToUpper(matches[1])
	}
	return ""
}

func (p *Portfolio) GetSymbols() []string {
	symbols := make([]string, 0, len(*p))
	for symbol := range *p {
		symbols = append(symbols, symbol)
	}
	return symbols
}
