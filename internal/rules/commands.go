package rules

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"spendgrid/internal/cache"
	"spendgrid/internal/i18n"
	"spendgrid/internal/parser"
)

// ListRules displays all rules
func ListRules() error {
	ruleSet, err := LoadRules()
	if err != nil {
		return err
	}

	if len(ruleSet.Rules) == 0 {
		fmt.Println(i18n.T("rules.no_rules"))
		return nil
	}

	fmt.Println(i18n.T("rules.header"))
	fmt.Println(strings.Repeat("=", 80))

	for _, r := range ruleSet.Rules {
		status := "✓"
		if !r.Active {
			status = "✗"
		}

		typeStr := "INC"
		if r.Type == "expense" {
			typeStr = "EXP"
		}

		fmt.Printf("%s [%s] %s | %s | %.2f %s | Monthly day %d\n",
			status,
			r.ID,
			typeStr,
			r.Name,
			r.Amount,
			r.Currency,
			r.Schedule.Day)
	}

	return nil
}

// AddRuleInteractive adds a new rule interactively with autocomplete
func AddRuleInteractive() error {
	// Load cache for autocomplete
	cacheStore, err := cache.LoadCache()
	if err != nil {
		cacheStore = &cache.Cache{Tags: []string{}, Projects: []string{}}
	}

	// Scan existing files to populate cache
	if err := refreshCacheFromFiles(cacheStore); err != nil {
		// Non-fatal, continue with empty cache
		fmt.Fprintf(os.Stderr, "Warning: could not scan existing files: %v\n", err)
	}

	// Get rule name
	fmt.Println(i18n.T("rules.name_prompt"))
	name, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading name: %v", err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Generate ID
	id := GenerateRuleID(name)

	// Get type
	fmt.Println("Tür [income/expense] [expense]:")
	ruleTypeInput, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading type: %v", err)
	}
	ruleType := strings.TrimSpace(strings.ToLower(ruleTypeInput))
	if ruleType != "income" && ruleType != "expense" {
		ruleType = "expense" // default
	}

	// Get amount and currency
	fmt.Println("Tutar ve Para Birimi (örn: 25000TRY, 500 USD, -150.50 EUR):")
	amountInput, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading amount: %v", err)
	}
	amountInput = strings.TrimSpace(amountInput)
	amount, currency, err := parseAmountInput(amountInput)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// Get schedule day with default
	today := time.Now().Day()
	fmt.Printf("Ayın günü [%d]:\n", today)
	dayStr, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading day: %v", err)
	}
	dayStr = strings.TrimSpace(dayStr)
	day := today
	if dayStr != "" {
		d, err := strconv.Atoi(dayStr)
		if err == nil && d >= 1 && d <= 31 {
			day = d
		}
	}

	// Ask about duration
	fmt.Println("Tüm yıl boyunca mu? (e/h) [e]:")
	fullYear, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading duration: %v", err)
	}
	fullYear = strings.TrimSpace(strings.ToLower(fullYear))

	startDate := ""
	endDate := ""

	if fullYear == "h" || fullYear == "hayır" || fullYear == "hayir" {
		// Ask for start and end dates
		fmt.Println("Başlangıç tarihi (YYYY-MM):")
		startDate, err = readSimpleLine()
		if err != nil {
			return fmt.Errorf("error reading start date: %v", err)
		}
		startDate = strings.TrimSpace(startDate)

		fmt.Println("Bitiş tarihi (YYYY-MM):")
		endDate, err = readSimpleLine()
		if err != nil {
			return fmt.Errorf("error reading end date: %v", err)
		}
		endDate = strings.TrimSpace(endDate)
	}

	// Ask about installment/credit
	fmt.Println("Taksitli/kredili ödeme mi? (e/h) [h]:")
	isInstallment, err := readSimpleLine()
	if err != nil {
		return fmt.Errorf("error reading installment: %v", err)
	}
	isInstallment = strings.TrimSpace(strings.ToLower(isInstallment))

	totalAmount := 0.0
	metadata := ""

	if isInstallment == "e" || isInstallment == "evet" {
		fmt.Println("Toplam tutar (örn: 25000):")
		totalStr, err := readSimpleLine()
		if err != nil {
			return fmt.Errorf("error reading total amount: %v", err)
		}
		totalStr = strings.TrimSpace(totalStr)
		if ta, err := strconv.ParseFloat(totalStr, 64); err == nil {
			totalAmount = ta
		}

		fmt.Println("Açıklama (örn: 3 taksit - iPhone 15):")
		metadata, err = readSimpleLine()
		if err != nil {
			return fmt.Errorf("error reading metadata: %v", err)
		}
		metadata = strings.TrimSpace(metadata)
	}

	// Get tags with autocomplete
	fmt.Println("Etiketler:")
	fmt.Println("  (Type to filter, Tab to autocomplete, 1-9 to select, Enter to confirm)")
	tagsInput, err := readWithAutocomplete("  > ", cacheStore.GetTags(), "#")
	if err != nil {
		return fmt.Errorf("error reading tags: %v", err)
	}
	tags := parseTags(tagsInput)

	// Get project with autocomplete
	fmt.Println("Proje:")
	fmt.Println("  (Type to filter, Tab to autocomplete, 1-9 to select, Enter to confirm)")
	project, err := readWithAutocomplete("  > ", cacheStore.GetProjects(), "@")
	if err != nil {
		return fmt.Errorf("error reading project: %v", err)
	}
	project = strings.TrimSpace(project)

	rule := Rule{
		ID:       id,
		Name:     name,
		Amount:   amount,
		Currency: currency,
		Type:     ruleType,
		Tags:     tags,
		Project:  project,
		Schedule: Schedule{
			Frequency: "monthly",
			Day:       day,
		},
		Active:      true,
		StartDate:   startDate,
		EndDate:     endDate,
		TotalAmount: totalAmount,
		Metadata:    metadata,
	}

	if err := AddRule(rule); err != nil {
		return err
	}

	// Auto-save tags and project to cache
	for _, tag := range tags {
		cacheStore.AddTag(tag)
	}
	if project != "" {
		cacheStore.AddProject(strings.TrimPrefix(project, "@"))
	}
	cacheStore.SaveCache()

	// Show summary
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✓ Kural başarıyla eklendi!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Ad: %s\n", name)
	fmt.Printf("Tutar: %.2f %s\n", amount, currency)
	fmt.Printf("Tür: %s\n", ruleType)
	fmt.Printf("Gün: %d\n", day)
	if startDate != "" && endDate != "" {
		fmt.Printf("Tarih aralığı: %s - %s\n", startDate, endDate)
	} else {
		fmt.Println("Süre: Tüm yıl")
	}
	if totalAmount > 0 {
		fmt.Printf("Toplam tutar: %.2f %s\n", totalAmount, currency)
	}
	if metadata != "" {
		fmt.Printf("Açıklama: %s\n", metadata)
	}
	if len(tags) > 0 {
		fmt.Printf("Etiketler: %s\n", strings.Join(tags, ", "))
	}
	if project != "" {
		fmt.Printf("Proje: %s\n", project)
	}
	fmt.Printf("ID: %s\n", id)
	fmt.Println(strings.Repeat("=", 50))

	return nil
}

// readSimpleLine reads a line of input using keyboard mode
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
		case keyboard.KeySpace:
			input = append(input, ' ')
			fmt.Print(" ")
		default:
			if char != 0 {
				input = append(input, char)
				fmt.Printf("%c", char)
			}
		}
	}
}

// readWithAutocomplete reads input with real-time autocomplete suggestions
func readWithAutocomplete(prompt string, items []string, prefix string) (string, error) {
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

		case keyboard.KeySpace:
			input = append(input, ' ')
			selectedIndex = -1
			searchTerm := strings.TrimPrefix(string(input), prefix)
			matches = filterItems(items, searchTerm)
			suggestionsShown = updateDisplay(prompt, input, matches, selectedIndex, prefix)

		default:
			if char != 0 {
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

				searchTerm := strings.TrimPrefix(string(input), prefix)
				matches = filterItems(items, searchTerm)

				suggestionsShown = updateDisplay(prompt, input, matches, selectedIndex, prefix)
			}
		}
	}
}

// clearSuggestions clears the suggestions line from the terminal
func clearSuggestions() {
	fmt.Print("\n\033[K\033[F")
}

// updateDisplay updates the terminal display with current input and suggestions
func updateDisplay(prompt string, input []rune, matches []string, selectedIndex int, prefix string) bool {
	fmt.Printf("\r\033[K")
	fmt.Printf("%s%s", prompt, string(input))

	if len(matches) > 0 {
		fmt.Println()
		fmt.Printf("\033[2K")
		fmt.Printf("\r")

		maxShow := 5
		if len(matches) < maxShow {
			maxShow = len(matches)
		}

		for i := 0; i < maxShow; i++ {
			if i == selectedIndex {
				fmt.Printf("\033[7m %d. %s%s \033[0m", i+1, prefix, matches[i])
			} else {
				fmt.Printf(" %d. %s%s", i+1, prefix, matches[i])
			}
		}

		fmt.Printf("\033[F")
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
	currentYear := strconv.Itoa(time.Now().Year())

	return filepath.Walk(currentYear, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Parse transactions from file
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

// EditRuleInteractive edits a rule interactively
func EditRuleInteractive(id string) error {
	rule, err := GetRule(id)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Editing rule: %s\n", rule.Name)
	fmt.Println("Press Enter to keep current value, or enter new value:")

	// Name
	fmt.Printf("Name [%s]: ", rule.Name)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name != "" {
		rule.Name = name
	}

	// Type
	fmt.Printf("Type [%s]: ", rule.Type)
	ruleType, _ := reader.ReadString('\n')
	ruleType = strings.TrimSpace(strings.ToLower(ruleType))
	if ruleType == "income" || ruleType == "expense" {
		rule.Type = ruleType
	}

	// Amount
	fmt.Printf("Amount [%.2f]: ", rule.Amount)
	amountStr, _ := reader.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	if amountStr != "" {
		if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
			rule.Amount = amount
		}
	}

	// Currency
	fmt.Printf("Currency [%s]: ", rule.Currency)
	currency, _ := reader.ReadString('\n')
	currency = strings.TrimSpace(strings.ToUpper(currency))
	if currency != "" {
		rule.Currency = currency
	}

	// Schedule day
	fmt.Printf("Day of month [%d]: ", rule.Schedule.Day)
	dayStr, _ := reader.ReadString('\n')
	dayStr = strings.TrimSpace(dayStr)
	if dayStr != "" {
		if day, err := strconv.Atoi(dayStr); err == nil && day >= 1 && day <= 31 {
			rule.Schedule.Day = day
		}
	}

	// Tags
	fmt.Printf("Tags [%s]: ", strings.Join(rule.Tags, " "))
	tagsInput, _ := reader.ReadString('\n')
	tagsInput = strings.TrimSpace(tagsInput)
	if tagsInput != "" {
		rule.Tags = parseTags(tagsInput)
	}

	if err := UpdateRule(id, *rule); err != nil {
		return err
	}

	fmt.Println(i18n.T("rules.edit_success"))
	return nil
}

// ToggleRuleStatus toggles a rule's active status
func ToggleRuleStatus(id string) error {
	rule, err := GetRule(id)
	if err != nil {
		return err
	}

	if err := ToggleRule(id); err != nil {
		return err
	}

	status := "activated"
	if rule.Active {
		status = "deactivated"
	}

	fmt.Printf("Rule '%s' %s\n", rule.Name, status)
	return nil
}

// RemoveRule removes a rule
func RemoveRule(id string) error {
	rule, err := GetRule(id)
	if err != nil {
		return err
	}

	fmt.Printf("Remove rule '%s'? [y/n]: ", rule.Name)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Cancelled")
		return nil
	}

	if err := DeleteRule(id); err != nil {
		return err
	}

	fmt.Println(i18n.T("rules.remove_success"))
	return nil
}

// SyncNow performs a manual sync
func SyncNow() error {
	fmt.Println(i18n.T("rules.sync_start"))

	result, err := SyncRules()
	if err != nil {
		return fmt.Errorf("sync failed: %v", err)
	}

	fmt.Printf(i18n.T("rules.sync_complete"), result.Added, result.Updated, result.Skipped)
	fmt.Println()

	if len(result.Errors) > 0 {
		fmt.Println(i18n.T("rules.sync_errors"))
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	return nil
}

func parseTags(input string) []string {
	var tags []string
	words := strings.Fields(input)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			tags = append(tags, strings.TrimPrefix(word, "#"))
		} else {
			tags = append(tags, word)
		}
	}
	return tags
}

// AddRuleDirect adds a rule from command line arguments
func AddRuleDirect(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("insufficient arguments: name, amount, currency, type required")
	}

	name := args[0]
	amountStr := args[1]
	currency := strings.ToUpper(args[2])
	ruleType := strings.ToLower(args[3])

	// Validate type
	if ruleType != "income" && ruleType != "expense" {
		return fmt.Errorf("type must be 'income' or 'expense'")
	}

	// Parse amount and currency (combined format supported: 25000TRY)
	amount, parsedCurrency, err := parseAmountInput(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// If currency was parsed from amount string, use it
	if parsedCurrency != "" {
		currency = strings.ToUpper(parsedCurrency)
	}

	// Parse optional flags
	day := 1
	var tags []string
	project := ""
	startDate := ""
	endDate := ""
	totalAmount := 0.0
	metadata := ""

	for i := 4; i < len(args); i++ {
		switch args[i] {
		case "--day":
			if i+1 < len(args) {
				d, err := strconv.Atoi(args[i+1])
				if err == nil && d >= 1 && d <= 31 {
					day = d
				}
				i++
			}
		case "--tags":
			if i+1 < len(args) {
				tagList := strings.Split(args[i+1], ",")
				for _, tag := range tagList {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						tags = append(tags, tag)
					}
				}
				i++
			}
		case "--project":
			if i+1 < len(args) {
				project = args[i+1]
				i++
			}
		case "--start-date":
			if i+1 < len(args) {
				startDate = args[i+1]
				i++
			}
		case "--end-date":
			if i+1 < len(args) {
				endDate = args[i+1]
				i++
			}
		case "--total-amount":
			if i+1 < len(args) {
				if ta, err := strconv.ParseFloat(args[i+1], 64); err == nil {
					totalAmount = ta
				}
				i++
			}
		case "--metadata":
			if i+1 < len(args) {
				metadata = args[i+1]
				i++
			}
		}
	}

	id := GenerateRuleID(name)

	rule := Rule{
		ID:       id,
		Name:     name,
		Amount:   amount,
		Currency: currency,
		Type:     ruleType,
		Tags:     tags,
		Project:  project,
		Schedule: Schedule{
			Frequency: "monthly",
			Day:       day,
		},
		Active:      true,
		StartDate:   startDate,
		EndDate:     endDate,
		TotalAmount: totalAmount,
		Metadata:    metadata,
	}

	if err := AddRule(rule); err != nil {
		return err
	}

	fmt.Printf("Rule added successfully: %s (ID: %s)\n", name, id)
	fmt.Printf("  Amount: %.2f %s\n", amount, currency)
	fmt.Printf("  Type: %s\n", ruleType)
	fmt.Printf("  Schedule: Monthly on day %d\n", day)
	if len(tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
	}
	if project != "" {
		fmt.Printf("  Project: %s\n", project)
	}
	if startDate != "" {
		fmt.Printf("  Start Date: %s\n", startDate)
	}
	if endDate != "" {
		fmt.Printf("  End Date: %s\n", endDate)
	}
	if totalAmount > 0 {
		fmt.Printf("  Total Amount: %.2f %s\n", totalAmount, currency)
	}
	if metadata != "" {
		fmt.Printf("  Metadata: %s\n", metadata)
	}

	return nil
}

// parseAmountInput parses amount and currency from input string
// Supports formats like: 25000TRY, 500 USD, -150.50 EUR, 25.000,50 TRY
func parseAmountInput(input string) (float64, string, error) {
	input = strings.ReplaceAll(input, " ", "")

	currencyPatterns := []string{"TL", "TRY", "USD", "EUR", "GBP", "$", "€", "₺"}

	var currency string
	amountStr := input

	for _, curr := range currencyPatterns {
		if strings.HasSuffix(strings.ToUpper(input), strings.ToUpper(curr)) {
			currency = curr
			amountStr = input[:len(input)-len(curr)]
			break
		}
	}

	if currency == "" {
		return 0, "", fmt.Errorf("invalid format: cannot parse amount and currency from '%s'", input)
	}

	// Normalize thousand separators
	if strings.Contains(amountStr, ",") && strings.Contains(amountStr, ".") {
		lastComma := strings.LastIndex(amountStr, ",")
		lastDot := strings.LastIndex(amountStr, ".")

		if lastComma > lastDot {
			amountStr = strings.ReplaceAll(amountStr, ".", "")
			amountStr = strings.Replace(amountStr, ",", ".", 1)
		} else {
			amountStr = strings.ReplaceAll(amountStr, ",", "")
		}
	} else if strings.Contains(amountStr, ",") {
		parts := strings.Split(amountStr, ",")
		if len(parts) == 2 && len(parts[1]) <= 2 {
			amountStr = strings.Replace(amountStr, ",", ".", 1)
		} else {
			amountStr = strings.ReplaceAll(amountStr, ",", "")
		}
	}

	// Handle multiple signs
	minusCount := strings.Count(amountStr, "-")
	amountStr = strings.ReplaceAll(amountStr, "-", "")
	amountStr = strings.ReplaceAll(amountStr, "+", "")

	if minusCount%2 == 1 {
		amountStr = "-" + amountStr
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, "", fmt.Errorf("cannot parse amount '%s': %v", amountStr, err)
	}

	return amount, currency, nil
}
