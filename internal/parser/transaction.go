package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Transaction represents a single financial transaction
type Transaction struct {
	Day         int
	Description string
	Amount      float64
	Currency    string
	Rate        float64 // Manual rate if specified (@rate)
	Tags        []string
	Projects    []string
	Meta        map[string]string
	Raw         string
	IsUnparsed  bool
	LineNumber  int
}

// IsExpense returns true if the amount is negative
func (t *Transaction) IsExpense() bool {
	return t.Amount < 0
}

// IsIncome returns true if the amount is positive
func (t *Transaction) IsIncome() bool {
	return t.Amount > 0
}

// ParseTransaction parses a single transaction line
// Format: - DAY | DESCRIPTION | AMOUNT CURRENCY [@RATE] | TAGS | [META]
// Example: - 15 | Market Alışverişi | -3.200,50 TRY | #mutfak | [NOTE:Misafir geldi]
func ParseTransaction(line string, lineNum int) *Transaction {
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	// Must start with "-"
	if !strings.HasPrefix(line, "-") {
		return nil
	}

	tx := &Transaction{
		Tags:       []string{},
		Projects:   []string{},
		Meta:       make(map[string]string),
		Raw:        line,
		LineNumber: lineNum,
	}

	// Remove the leading "- "
	content := strings.TrimPrefix(line, "-")
	content = strings.TrimSpace(content)

	// Handle checkbox format for rules: "- [ ] " or "- [x] "
	if strings.HasPrefix(content, "[ ]") || strings.HasPrefix(content, "[x]") {
		// Find the checkbox and remove it
		checkboxEnd := strings.Index(content, "]")
		if checkboxEnd > 0 {
			content = strings.TrimSpace(content[checkboxEnd+1:])
		}
	}

	// Split by pipe
	parts := splitByPipe(content)
	if len(parts) < 4 {
		tx.IsUnparsed = true
		return tx
	}

	// Parse day
	day, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || day < 1 || day > 31 {
		tx.IsUnparsed = true
		return tx
	}
	tx.Day = day

	// Parse description
	tx.Description = strings.TrimSpace(parts[1])

	// Parse amount and currency
	if err := parseAmountAndCurrency(parts[2], tx); err != nil {
		tx.IsUnparsed = true
		return tx
	}

	// Parse tags and projects
	parseTagsAndProjects(parts[3], tx)

	// Parse meta if present
	if len(parts) >= 5 {
		parseMeta(parts[4], tx)
	}

	return tx
}

// splitByPipe splits content by | while preserving content in brackets
func splitByPipe(content string) []string {
	var parts []string
	var current strings.Builder
	inBrackets := 0

	for _, char := range content {
		switch char {
		case '|':
			if inBrackets == 0 {
				parts = append(parts, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		case '[':
			inBrackets++
			current.WriteRune(char)
		case ']':
			inBrackets--
			current.WriteRune(char)
		default:
			current.WriteRune(char)
		}
	}

	// Add the last part
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	return parts
}

// parseAmountAndCurrency parses the amount and currency part
// Formats: -25000 TRY, -120 USD @35.50, +1000 TL, 500$
func parseAmountAndCurrency(part string, tx *Transaction) error {
	part = strings.TrimSpace(part)

	// Check for manual rate: @rate at the end
	rateRegex := regexp.MustCompile(`@([0-9]+[.,]?[0-9]*)\s*$`)
	if matches := rateRegex.FindStringSubmatch(part); matches != nil {
		rateStr := strings.ReplaceAll(matches[1], ",", ".")
		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil {
			return fmt.Errorf("invalid rate: %v", err)
		}
		tx.Rate = rate
		// Remove rate from part
		part = rateRegex.ReplaceAllString(part, "")
		part = strings.TrimSpace(part)
	}

	// Extract currency (last word/sequence)
	// Look for currency patterns: TRY, TL, USD, $, EUR, €
	currencyPatterns := []struct {
		pattern string
		value   string
	}{
		{`(?i)\b(TRY|TL|₺)\b`, "TRY"},
		{`(?i)\b(USD|\$|dolar|dollar)\b`, "USD"},
		{`(?i)\b(EUR|€|euro)\b`, "EUR"},
	}

	var foundCurrency string
	for _, cp := range currencyPatterns {
		re := regexp.MustCompile(cp.pattern)
		if re.MatchString(part) {
			foundCurrency = cp.value
			// Remove currency from part to get amount
			part = re.ReplaceAllString(part, "")
			break
		}
	}

	if foundCurrency == "" {
		return fmt.Errorf("no currency found")
	}
	tx.Currency = foundCurrency

	// Parse amount from remaining part
	amountStr := strings.TrimSpace(part)
	// Normalize: remove spaces, standardize decimal separator
	amountStr = strings.ReplaceAll(amountStr, " ", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "") // Remove thousand separator

	// Handle multiple minus signs (e.g., user wrote "--5000" instead of "-5000")
	// Count minus signs and normalize to single minus if odd number, positive if even
	minusCount := strings.Count(amountStr, "-")
	amountStr = strings.ReplaceAll(amountStr, "-", "")
	amountStr = strings.ReplaceAll(amountStr, "+", "")

	// If odd number of minus signs, add one back
	if minusCount%2 == 1 {
		amountStr = "-" + amountStr
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}
	tx.Amount = amount

	return nil
}

// parseTagsAndProjects parses tags (#tag) and projects (@project)
func parseTagsAndProjects(part string, tx *Transaction) {
	words := strings.Fields(part)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			tag := strings.TrimPrefix(word, "#")
			tx.Tags = append(tx.Tags, tag)
		} else if strings.HasPrefix(word, "@") {
			project := strings.TrimPrefix(word, "@")
			tx.Projects = append(tx.Projects, project)
		}
	}
}

// parseMeta parses metadata in brackets: [NOTE:xyz] [ID:123]
func parseMeta(part string, tx *Transaction) {
	part = strings.TrimSpace(part)
	if !strings.HasPrefix(part, "[") || !strings.HasSuffix(part, "]") {
		return
	}

	// Remove brackets
	content := part[1 : len(part)-1]

	// Split by comma or space if no colon
	items := strings.Split(content, ",")
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// Check for KEY:VALUE format
		if idx := strings.Index(item, ":"); idx > 0 {
			key := strings.TrimSpace(item[:idx])
			value := strings.TrimSpace(item[idx+1:])
			tx.Meta[key] = value
		} else {
			// Just store as note
			tx.Meta["NOTE"] = item
		}
	}
}

// FormatTransaction formats a transaction back to string format
func FormatTransaction(tx *Transaction) string {
	var parts []string

	// Day
	parts = append(parts, fmt.Sprintf("%02d", tx.Day))

	// Description
	parts = append(parts, tx.Description)

	// Amount, Currency, and optional Rate
	amountPart := formatAmount(tx.Amount, tx.Currency)
	if tx.Rate > 0 {
		amountPart += fmt.Sprintf(" @%.2f", tx.Rate)
	}
	parts = append(parts, amountPart)

	// Tags and Projects
	var tagsParts []string
	for _, tag := range tx.Tags {
		tagsParts = append(tagsParts, "#"+tag)
	}
	for _, project := range tx.Projects {
		tagsParts = append(tagsParts, "@"+project)
	}
	parts = append(parts, strings.Join(tagsParts, " "))

	// Meta
	if len(tx.Meta) > 0 {
		var metaParts []string
		for key, value := range tx.Meta {
			metaParts = append(metaParts, fmt.Sprintf("%s:%s", key, value))
		}
		parts = append(parts, fmt.Sprintf("[%s]", strings.Join(metaParts, ",")))
	}

	return "- " + strings.Join(parts, " | ")
}

// formatAmount formats the amount with proper formatting
func formatAmount(amount float64, currency string) string {
	sign := ""
	if amount < 0 {
		sign = "-"
	}
	// Use absolute value to avoid double negative
	if amount < 0 {
		amount = -amount
	}
	return fmt.Sprintf("%s%.2f %s", sign, amount, currency)
}

// ParseMonthFile parses all transactions from a month file
func ParseMonthFile(content string) ([]*Transaction, []*Transaction) {
	var parsed []*Transaction
	var unparsed []*Transaction

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		tx := ParseTransaction(line, i+1)
		if tx == nil {
			continue
		}
		if tx.IsUnparsed {
			unparsed = append(unparsed, tx)
		} else {
			parsed = append(parsed, tx)
		}
	}

	return parsed, unparsed
}

// GetCurrentMonthFile returns the current month filename
func GetCurrentMonthFile() string {
	now := time.Now()
	return fmt.Sprintf("%02d.md", now.Month())
}

// GetMonthFile returns a specific month filename
func GetMonthFile(month int) string {
	return fmt.Sprintf("%02d.md", month)
}
