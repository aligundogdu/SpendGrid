package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"spendgrid/internal/currency"
)

// QuickInputParser parses natural language transaction input
// Examples:
//
//	"-100TL market alışverişi yaptım #mutfak @ev"
//	"5000 USD maaş geliri #iş #maaş"
//	"market alışverişi -100TL #mutfak"
//	"150 € restaurant #eğlence @tatil"
func QuickInputParser(input string) (*Transaction, error) {
	input = strings.TrimSpace(input)

	// Extract amount and currency
	amount, curr, remaining := extractAmount(input)

	// Extract tags
	tags, remaining := extractTags(remaining)

	// Extract projects
	projects, description := extractProjects(remaining)

	// Clean up description
	description = strings.TrimSpace(description)

	// If no description, use a default
	if description == "" {
		description = "İşlem"
	}

	// Normalize currency
	if curr == "" {
		curr = "TRY" // Default to TRY
	}
	curr = currency.Normalize(curr)

	// Get current day
	now := time.Now()

	tx := &Transaction{
		Day:         now.Day(),
		Description: description,
		Amount:      amount,
		Currency:    curr,
		Tags:        tags,
		Projects:    projects,
		Meta:        make(map[string]string),
	}

	return tx, nil
}

// extractAmount finds and extracts amount and currency from input
// Returns: amount, currency, remaining text
func extractAmount(input string) (float64, string, string) {
	// Regex pattern for amount: optional minus, digits with optional decimal/thousand separators, optional space, currency
	// Matches: 100TL, 100 TL, -100.50 USD, 1,500.00$, €50, etc.
	patterns := []string{
		`(-?[\d.,]+)\s*([A-Za-z$€₺]{1,4})`, // 100 TL, -100.50 USD
		`(-?[\d.,]+)\s*([$€₺])`,            // 100 $, -50 €
		`([$€₺])\s*(-?[\d.,]+)`,            // $100, €-50
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(input)

		if match != nil {
			// Get the matched text location
			loc := re.FindStringIndex(input)

			// Extract groups
			amountStr := match[1]
			currStr := match[2]

			// Check if we need to swap (for currency-first patterns like $100)
			if currStr == "$" || currStr == "€" || currStr == "₺" {
				// Check if second group is the amount
				if len(match) >= 3 {
					testAmount := match[2]
					testAmount = strings.ReplaceAll(testAmount, ",", "")
					if _, err := strconv.ParseFloat(testAmount, 64); err == nil {
						amountStr = testAmount
						currStr = match[1]
					}
				}
			}

			// Normalize amount string
			amountStr = strings.ReplaceAll(amountStr, ",", "")
			amountStr = strings.ReplaceAll(amountStr, " ", "")

			// Parse amount
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				continue
			}

			// Remove the matched text from input using strings.Replace
			before := input[:loc[0]]
			after := input[loc[1]:]
			remaining := strings.TrimSpace(before + " " + after)

			// Normalize currency
			curr := normalizeCurrencySymbol(currStr)

			return amount, curr, remaining
		}
	}

	// No amount found, return defaults
	return 0, "", input
}

// extractTags finds all #tags in input
// Returns: tags list, remaining text
func extractTags(input string) ([]string, string) {
	re := regexp.MustCompile(`#(\w+)`)
	matches := re.FindAllStringSubmatchIndex(input, -1)

	if matches == nil {
		return []string{}, input
	}

	var tags []string
	var toRemove []string

	for _, match := range matches {
		if len(match) >= 4 {
			// Get the tag text (without #)
			tagStart := match[2]
			tagEnd := match[3]
			tag := input[tagStart:tagEnd]
			tags = append(tags, tag)

			// Get the full match (with #) to remove
			fullMatch := input[match[0]:match[1]]
			toRemove = append(toRemove, fullMatch)
		}
	}

	// Remove tags from text
	remaining := input
	for _, removal := range toRemove {
		remaining = strings.Replace(remaining, removal, "", 1)
	}
	remaining = strings.TrimSpace(remaining)

	return tags, remaining
}

// extractProjects finds all @projects in input
// Returns: projects list, remaining text (description)
func extractProjects(input string) ([]string, string) {
	re := regexp.MustCompile(`@(\w+)`)
	matches := re.FindAllStringSubmatchIndex(input, -1)

	if matches == nil {
		return []string{}, input
	}

	var projects []string
	var toRemove []string

	for _, match := range matches {
		if len(match) >= 4 {
			// Get the project text (without @)
			projStart := match[2]
			projEnd := match[3]
			project := input[projStart:projEnd]
			projects = append(projects, project)

			// Get the full match (with @) to remove
			fullMatch := input[match[0]:match[1]]
			toRemove = append(toRemove, fullMatch)
		}
	}

	// Remove projects from text
	remaining := input
	for _, removal := range toRemove {
		remaining = strings.Replace(remaining, removal, "", 1)
	}
	remaining = strings.TrimSpace(remaining)

	return projects, remaining
}

func normalizeCurrencySymbol(curr string) string {
	curr = strings.TrimSpace(strings.ToUpper(curr))
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
