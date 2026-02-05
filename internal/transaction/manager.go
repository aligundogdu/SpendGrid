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

	"github.com/eiannone/keyboard"

	"spendgrid/internal/cache"
	"spendgrid/internal/currency"
	"spendgrid/internal/i18n"
	"spendgrid/internal/parser"
)

// AddTransaction adds a new transaction interactively with real-time autocomplete
func AddTransaction() error {
	// Check if we're in a spendgrid directory
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	// Get current month file
	monthFile := parser.GetCurrentMonthFile()
	currentYear := strconv.Itoa(time.Now().Year())
	filePath := filepath.Join(currentYear, monthFile)

	// Load and refresh cache from existing transactions
	cacheStore, err := cache.LoadCache()
	if err != nil {
		cacheStore = &cache.Cache{Tags: []string{}, Projects: []string{}}
	}

	// Scan existing files to populate cache
	if err := refreshCacheFromFiles(cacheStore); err != nil {
		// Non-fatal, continue with empty cache
		fmt.Fprintf(os.Stderr, "Warning: could not scan existing files: %v\n", err)
	}

	// Ask for day (default to today)
	today := time.Now().Day()
	fmt.Printf("%s [%d]: ", i18n.T("transaction.day_prompt"), today)
	dayStr, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading day: %v", err)
	}
	dayStr = strings.TrimSpace(dayStr)
	day := today
	if dayStr != "" {
		parsedDay, err := strconv.Atoi(dayStr)
		if err != nil || parsedDay < 1 || parsedDay > 31 {
			return fmt.Errorf("invalid day: %s", dayStr)
		}
		day = parsedDay
	}

	// Ask for description
	fmt.Println(i18n.T("transaction.description_prompt"))
	desc, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading description: %v", err)
	}
	desc = strings.TrimSpace(desc)

	// Ask for amount and currency
	fmt.Println(i18n.T("transaction.amount_prompt"))
	amountInput, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading amount: %v", err)
	}
	amountInput = strings.TrimSpace(amountInput)

	amount, curr, err := parseAmountInput(amountInput)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// Normalize currency
	curr = currency.Normalize(curr)

	// Ask for tags with real-time autocomplete
	fmt.Println(i18n.T("transaction.tags_prompt") + " ")
	fmt.Println("  (Type to filter, Tab to autocomplete, 1-9 to select, Enter to confirm)")
	tagsInput, err := readWithAutocomplete("  > ", cacheStore.GetTags(), "#")
	if err != nil {
		return fmt.Errorf("error reading tags: %v", err)
	}
	tags := parseTags(tagsInput)

	// Ask for projects with real-time autocomplete
	fmt.Println(i18n.T("transaction.projects_prompt") + " ")
	fmt.Println("  (Type to filter, Tab to autocomplete, 1-9 to select, Enter to confirm)")
	projInput, err := readWithAutocomplete("  > ", cacheStore.GetProjects(), "@")
	if err != nil {
		return fmt.Errorf("error reading projects: %v", err)
	}
	projects := parseProjects(projInput)

	// Ask for note (optional)
	fmt.Println(i18n.T("transaction.note_prompt"))
	note, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading note: %v", err)
	}
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

// readSimpleLine reads a line of input using keyboard mode (for consistency)
func readSimpleLine() (string, error) {
	if err := keyboard.Open(); err != nil {
		// Fallback to regular input
		reader := bufio.NewReader(os.Stdin)
		return reader.ReadString('\n')
	}
	defer keyboard.Close()

	var input []rune
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return "", err
		}

		switch key {
		case keyboard.KeyEnter:
			fmt.Println()
			return string(input), nil
		case keyboard.KeyCtrlC, keyboard.KeyEsc:
			return "", fmt.Errorf("cancelled")
		case keyboard.KeyBackspace, keyboard.KeyBackspace2:
			if len(input) > 0 {
				input = input[:len(input)-1]
				fmt.Print("\b \b")
			}
		default:
			if char != 0 {
				input = append(input, char)
				fmt.Printf("%c", char)
			}
		}
	}
}

// readWithAutocomplete reads input with real-time autocomplete suggestions
// prefix: "#" for tags, "@" for projects
func readWithAutocomplete(prompt string, items []string, prefix string) (string, error) {
	// Save terminal state and open keyboard
	if err := keyboard.Open(); err != nil {
		// Fallback to regular input if keyboard fails
		fmt.Print(prompt)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		return strings.TrimSpace(input), nil
	}
	defer keyboard.Close()

	var input []rune
	selectedIndex := -1
	matches := []string{}
	suggestionsShown := false

	fmt.Print(prompt)

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			return "", err
		}

		switch key {
		case keyboard.KeyEnter:
			// Clear suggestions if shown
			if suggestionsShown {
				clearSuggestions()
			}
			fmt.Println()
			result := string(input)
			if selectedIndex >= 0 && selectedIndex < len(matches) {
				result = prefix + matches[selectedIndex]
			}
			return result, nil

		case keyboard.KeyCtrlC, keyboard.KeyEsc:
			if suggestionsShown {
				clearSuggestions()
			}
			fmt.Println("\nCancelled")
			return "", fmt.Errorf("cancelled")

		case keyboard.KeyBackspace, keyboard.KeyBackspace2:
			if len(input) > 0 {
				input = input[:len(input)-1]
				selectedIndex = -1
				suggestionsShown = updateDisplay(prompt, input, matches, selectedIndex, prefix)
			}

		case keyboard.KeyTab:
			// Autocomplete with first match
			if len(matches) > 0 {
				input = []rune(prefix + matches[0])
				selectedIndex = 0
				suggestionsShown = updateDisplay(prompt, input, matches, selectedIndex, prefix)
			}

		case keyboard.KeyArrowUp:
			if selectedIndex > 0 {
				selectedIndex--
				suggestionsShown = updateDisplay(prompt, input, matches, selectedIndex, prefix)
			}

		case keyboard.KeyArrowDown:
			if selectedIndex < len(matches)-1 {
				selectedIndex++
				suggestionsShown = updateDisplay(prompt, input, matches, selectedIndex, prefix)
			}

		default:
			// Regular character input
			if char != 0 {
				// Handle number keys 1-9 for quick selection
				if char >= '1' && char <= '9' {
					idx := int(char - '1')
					if idx < len(matches) {
						input = []rune(prefix + matches[idx])
						selectedIndex = idx
						suggestionsShown = updateDisplay(prompt, input, matches, selectedIndex, prefix)
						continue
					}
				}

				input = append(input, char)
				selectedIndex = -1

				// Filter matches based on input (without prefix)
				searchTerm := strings.TrimPrefix(string(input), prefix)
				matches = filterItems(items, searchTerm)

				suggestionsShown = updateDisplay(prompt, input, matches, selectedIndex, prefix)
			}
		}
	}
}

// clearSuggestions clears the suggestions line from the terminal
func clearSuggestions() {
	// Move cursor down one line, clear it, move back up
	fmt.Print("\n\033[K\033[F")
}

// updateDisplay updates the terminal display with current input and suggestions
// Returns true if suggestions were shown
func updateDisplay(prompt string, input []rune, matches []string, selectedIndex int, prefix string) bool {
	// Clear current line and move cursor to beginning
	fmt.Printf("\r\033[K")

	// Print prompt and current input
	fmt.Printf("%s%s", prompt, string(input))

	// Show suggestions below
	if len(matches) > 0 {
		// Move to next line
		fmt.Println()
		// Clear the entire line
		fmt.Printf("\033[2K")
		// Move to beginning of line
		fmt.Printf("\r")

		// Show up to 5 matches
		maxShow := 5
		if len(matches) < maxShow {
			maxShow = len(matches)
		}

		for i := 0; i < maxShow; i++ {
			if i == selectedIndex {
				// Inverted colors for selected item
				fmt.Printf("\033[7m %d. %s%s \033[0m", i+1, prefix, matches[i])
			} else {
				fmt.Printf(" %d. %s%s", i+1, prefix, matches[i])
			}
		}

		// Move cursor back up to input line
		fmt.Printf("\033[F")
		// Move cursor to end of input (after the prompt)
		fmt.Printf("\033[%dC", len(prompt)+len(input))
		return true
	}
	return false
}

// filterItems returns items that contain the search term (case-insensitive)
func filterItems(items []string, searchTerm string) []string {
	if searchTerm == "" {
		return items
	}

	searchTerm = strings.ToLower(searchTerm)
	var matches []string

	for _, item := range items {
		if strings.Contains(strings.ToLower(item), searchTerm) {
			matches = append(matches, item)
		}
	}

	return matches
}

// refreshCacheFromFiles scans all transaction files and populates cache
func refreshCacheFromFiles(cacheStore *cache.Cache) error {
	// Get current year
	currentYear := strconv.Itoa(time.Now().Year())

	// Walk through year directory
	return filepath.Walk(currentYear, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		// Parse transactions
		parsed, _ := parser.ParseMonthFile(string(content))

		// Extract tags and projects
		for _, tx := range parsed {
			for _, tag := range tx.Tags {
				cacheStore.AddTag(tag)
			}
			for _, proj := range tx.Projects {
				cacheStore.AddProject(proj)
			}
		}

		return nil
	})
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

	// Also save to cache for autocomplete
	cacheStore, err := cache.LoadCache()
	if err != nil {
		return err
	}

	for _, tag := range tags {
		cacheStore.AddTag(tag)
	}
	for _, proj := range projects {
		cacheStore.AddProject(proj)
	}

	return cacheStore.SaveCache()
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
