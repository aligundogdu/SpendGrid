package transaction

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"spendgrid/internal/currency"
	"spendgrid/internal/i18n"
	"spendgrid/internal/parser"
)

// AddTransaction adds a new transaction interactively
func AddTransaction() error {
	// Check if we're in a spendgrid directory
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	reader := bufio.NewReader(os.Stdin)

	// Get current month file
	monthFile := parser.GetCurrentMonthFile()
	currentYear := strconv.Itoa(time.Now().Year())
	filePath := filepath.Join(currentYear, monthFile)

	// Ask for day
	fmt.Print(i18n.T("transaction.day_prompt") + " ")
	dayStr, _ := reader.ReadString('\n')
	dayStr = strings.TrimSpace(dayStr)
	day, err := strconv.Atoi(dayStr)
	if err != nil || day < 1 || day > 31 {
		return fmt.Errorf("invalid day: %s", dayStr)
	}

	// Ask for description
	fmt.Print(i18n.T("transaction.description_prompt") + " ")
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)

	// Ask for amount and currency
	fmt.Print(i18n.T("transaction.amount_prompt") + " ")
	amountInput, _ := reader.ReadString('\n')
	amountInput = strings.TrimSpace(amountInput)

	amount, curr, err := parseAmountInput(amountInput)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// Normalize currency
	curr = currency.Normalize(curr)

	// Ask for tags
	fmt.Print(i18n.T("transaction.tags_prompt") + " ")
	tagsInput, _ := reader.ReadString('\n')
	tagsInput = strings.TrimSpace(tagsInput)
	tags := parseTags(tagsInput)

	// Ask for projects
	fmt.Print(i18n.T("transaction.projects_prompt") + " ")
	projInput, _ := reader.ReadString('\n')
	projInput = strings.TrimSpace(projInput)
	projects := parseProjects(projInput)

	// Ask for note (optional)
	fmt.Print(i18n.T("transaction.note_prompt") + " ")
	note, _ := reader.ReadString('\n')
	note = strings.TrimSpace(note)

	// Create transaction
	tx := &parser.Transaction{
		Day:         day,
		Description: desc,
		Amount:      amount,
		Currency:    curr,
		Tags:        tags,
		Projects:    projects,
		Meta:        make(map[string]string),
	}

	if note != "" {
		tx.Meta["NOTE"] = note
	}

	// Add to file
	if err := addTransactionToFile(filePath, tx); err != nil {
		return err
	}

	// Auto-save tags and projects
	if err := autoSaveTagsAndProjects(tags, projects); err != nil {
		// Non-fatal, just warn
		fmt.Fprintf(os.Stderr, "Warning: could not auto-save tags: %v\n", err)
	}

	fmt.Println(i18n.T("transaction.add_success"))
	return nil
}

// AddDirectTransaction adds a transaction from a direct input string
func AddDirectTransaction(input string) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	// Parse the direct input
	// Format expected: DAY | DESC | AMOUNT CURR | TAGS
	parts := strings.Split(input, "|")
	if len(parts) < 4 {
		return fmt.Errorf("invalid format. Expected: DAY|DESC|AMOUNT CURR|TAGS")
	}

	day, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return fmt.Errorf("invalid day: %v", err)
	}

	desc := strings.TrimSpace(parts[1])

	amount, curr, err := parseAmountInput(strings.TrimSpace(parts[2]))
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}
	curr = currency.Normalize(curr)

	tags := parseTags(strings.TrimSpace(parts[3]))

	tx := &parser.Transaction{
		Day:         day,
		Description: desc,
		Amount:      amount,
		Currency:    curr,
		Tags:        tags,
		Projects:    []string{},
		Meta:        make(map[string]string),
	}

	monthFile := parser.GetCurrentMonthFile()
	currentYear := strconv.Itoa(time.Now().Year())
	filePath := filepath.Join(currentYear, monthFile)

	if err := addTransactionToFile(filePath, tx); err != nil {
		return err
	}

	if err := autoSaveTagsAndProjects(tags, []string{}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not auto-save tags: %v\n", err)
	}

	fmt.Println(i18n.T("transaction.add_success"))
	return nil
}

// ListTransactions lists all transactions for current or specified month
func ListTransactions(month string) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	var monthFile string
	if month == "" {
		monthFile = parser.GetCurrentMonthFile()
	} else {
		monthInt, err := strconv.Atoi(month)
		if err != nil || monthInt < 1 || monthInt > 12 {
			return fmt.Errorf("invalid month: %s", month)
		}
		monthFile = parser.GetMonthFile(monthInt)
	}

	currentYear := strconv.Itoa(time.Now().Year())
	filePath := filepath.Join(currentYear, monthFile)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read month file: %v", err)
	}

	parsed, unparsed := parser.ParseMonthFile(string(content))

	// Print header
	fmt.Printf("\n%s %s\n", currentYear, monthFile)
	fmt.Println(strings.Repeat("=", 60))

	// Print parsed transactions
	if len(parsed) > 0 {
		fmt.Println(i18n.T("transaction.parsed_header"))
		fmt.Println(strings.Repeat("-", 60))
		for i, tx := range parsed {
			fmt.Printf("%3d | %02d | %-20s | %10.2f %s | %s\n",
				i+1, tx.Day, truncate(tx.Description, 20), tx.Amount, tx.Currency,
				formatTagsAndProjects(tx.Tags, tx.Projects))
		}
	}

	// Print unparsed lines
	if len(unparsed) > 0 {
		fmt.Println()
		fmt.Println(i18n.T("transaction.unparsed_header"))
		fmt.Println(strings.Repeat("-", 60))
		for _, tx := range unparsed {
			fmt.Printf("Line %d: %s\n", tx.LineNumber, tx.Raw)
		}
	}

	if len(parsed) == 0 && len(unparsed) == 0 {
		fmt.Println(i18n.T("transaction.no_transactions"))
	}

	fmt.Println()
	return nil
}

// EditTransaction edits a transaction by line number
func EditTransaction(lineNum string) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	line, err := strconv.Atoi(lineNum)
	if err != nil || line < 1 {
		return fmt.Errorf("invalid line number: %s", lineNum)
	}

	monthFile := parser.GetCurrentMonthFile()
	currentYear := strconv.Itoa(time.Now().Year())
	filePath := filepath.Join(currentYear, monthFile)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read month file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the line in ROWS section
	inRows := false
	txLine := 0
	actualLine := 0

	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "## ROWS") {
			inRows = true
			continue
		}
		if strings.HasPrefix(trimmed, "## RULES") {
			inRows = false
		}
		if inRows && strings.HasPrefix(trimmed, "-") {
			txLine++
			if txLine == line {
				actualLine = i
				break
			}
		}
	}

	if actualLine == 0 {
		return fmt.Errorf("transaction not found at line %d", line)
	}

	reader := bufio.NewReader(os.Stdin)

	// Parse existing transaction
	existing := parser.ParseTransaction(lines[actualLine], actualLine+1)
	if existing == nil || existing.IsUnparsed {
		return fmt.Errorf("cannot edit unparsed transaction")
	}

	// Show current and ask for new values
	fmt.Printf("Current: %s\n", lines[actualLine])
	fmt.Println("Press Enter to keep current value, or enter new value:")

	// Day
	fmt.Printf("Day [%d]: ", existing.Day)
	dayStr, _ := reader.ReadString('\n')
	dayStr = strings.TrimSpace(dayStr)
	if dayStr != "" {
		if d, err := strconv.Atoi(dayStr); err == nil && d >= 1 && d <= 31 {
			existing.Day = d
		}
	}

	// Description
	fmt.Printf("Description [%s]: ", existing.Description)
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)
	if desc != "" {
		existing.Description = desc
	}

	// Amount
	fmt.Printf("Amount [%.2f %s]: ", existing.Amount, existing.Currency)
	amtStr, _ := reader.ReadString('\n')
	amtStr = strings.TrimSpace(amtStr)
	if amtStr != "" {
		if amt, curr, err := parseAmountInput(amtStr); err == nil {
			existing.Amount = amt
			existing.Currency = currency.Normalize(curr)
		}
	}

	// Tags
	fmt.Printf("Tags [%s]: ", strings.Join(existing.Tags, " "))
	tagsStr, _ := reader.ReadString('\n')
	tagsStr = strings.TrimSpace(tagsStr)
	if tagsStr != "" {
		existing.Tags = parseTags(tagsStr)
	}

	// Update the line
	lines[actualLine] = parser.FormatTransaction(existing)

	// Write back
	if err := os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to save: %v", err)
	}

	fmt.Println(i18n.T("transaction.edit_success"))
	return nil
}

// RemoveTransaction removes a transaction by line number
func RemoveTransaction(lineNum string) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	line, err := strconv.Atoi(lineNum)
	if err != nil || line < 1 {
		return fmt.Errorf("invalid line number: %s", lineNum)
	}

	monthFile := parser.GetCurrentMonthFile()
	currentYear := strconv.Itoa(time.Now().Year())
	filePath := filepath.Join(currentYear, monthFile)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read month file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find and remove the line
	inRows := false
	txLine := 0

	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "## ROWS") {
			inRows = true
			continue
		}
		if strings.HasPrefix(trimmed, "## RULES") {
			inRows = false
		}
		if inRows && strings.HasPrefix(trimmed, "-") {
			txLine++
			if txLine == line {
				// Remove this line
				lines = append(lines[:i], lines[i+1:]...)
				break
			}
		}
	}

	if txLine != line {
		return fmt.Errorf("transaction not found at line %d", line)
	}

	if err := os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to save: %v", err)
	}

	fmt.Println(i18n.T("transaction.remove_success"))
	return nil
}

// Helper functions

func parseAmountInput(input string) (float64, string, error) {
	// Remove spaces
	input = strings.ReplaceAll(input, " ", "")

	// Regex to match amount and currency
	// Examples: -25000TRY, 500.50USD, +1000TL, -3.200,50TRY
	re := regexp.MustCompile(`^([+-]?[0-9.,]+)([A-Za-z$€₺]+)$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) != 3 {
		return 0, "", fmt.Errorf("invalid format")
	}

	amountStr := matches[1]
	currency := matches[2]

	// Normalize thousand separators and handle multiple signs
	amountStr = strings.ReplaceAll(amountStr, ",", "")

	// Count minus signs and normalize
	minusCount := strings.Count(amountStr, "-")
	amountStr = strings.ReplaceAll(amountStr, "-", "")
	amountStr = strings.ReplaceAll(amountStr, "+", "")

	// If odd number of minus signs, add one back
	if minusCount%2 == 1 {
		amountStr = "-" + amountStr
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, "", err
	}

	return amount, currency, nil
}

func parseTags(input string) []string {
	var tags []string
	words := strings.Fields(input)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			tags = append(tags, strings.TrimPrefix(word, "#"))
		}
	}
	return tags
}

func parseProjects(input string) []string {
	var projects []string
	words := strings.Fields(input)
	for _, word := range words {
		if strings.HasPrefix(word, "@") {
			projects = append(projects, strings.TrimPrefix(word, "@"))
		}
	}
	return projects
}

func formatTagsAndProjects(tags, projects []string) string {
	var parts []string
	for _, t := range tags {
		parts = append(parts, "#"+t)
	}
	for _, p := range projects {
		parts = append(parts, "@"+p)
	}
	return strings.Join(parts, " ")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func addTransactionToFile(filePath string, tx *parser.Transaction) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the ROWS section and add transaction
	inRows := false
	insertIndex := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ROWS") {
			inRows = true
			insertIndex = i + 1
			continue
		}
		if strings.HasPrefix(trimmed, "## RULES") && inRows {
			// Insert before RULES
			insertIndex = i
			break
		}
	}

	if insertIndex < 0 {
		return fmt.Errorf("could not find ROWS section")
	}

	formatted := parser.FormatTransaction(tx)

	// Insert the new transaction
	lines = append(lines[:insertIndex], append([]string{formatted}, lines[insertIndex:]...)...)

	if err := os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

func autoSaveTagsAndProjects(tags, projects []string) error {
	// Save tags to categories.yml
	if len(tags) > 0 {
		categoriesPath := filepath.Join("_config", "categories.yml")
		if err := appendToYamlList(categoriesPath, "categories", tags); err != nil {
			return fmt.Errorf("failed to save tags: %v", err)
		}
	}

	// Save projects to projects.yml
	if len(projects) > 0 {
		projectsPath := filepath.Join("_config", "projects.yml")
		if err := appendToYamlList(projectsPath, "projects", projects); err != nil {
			return fmt.Errorf("failed to save projects: %v", err)
		}
	}

	return nil
}

func appendToYamlList(filePath, key string, items []string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Find existing items
	existing := make(map[string]bool)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			item := strings.TrimPrefix(trimmed, "- ")
			existing[item] = true
		}
	}

	// Add new items
	var added bool
	for _, item := range items {
		if !existing[item] {
			lines = append(lines, "- "+item)
			added = true
		}
	}

	if added {
		return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
	}

	return nil
}
