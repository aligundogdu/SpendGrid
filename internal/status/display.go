package status

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"spendgrid/internal/i18n"
	"spendgrid/internal/parser"
	"spendgrid/internal/rules"
)

// ShowStatus displays the current status of the spendgrid database
func ShowStatus() error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	now := time.Now()
	year := strconv.Itoa(now.Year())
	currentMonth := int(now.Month())

	// Get active rules count
	activeRules, err := rules.GetActiveRules()
	rulesCount := 0
	if err == nil {
		rulesCount = len(activeRules)
	}

	// Count transactions this month
	monthFile := parser.GetMonthFile(currentMonth)
	filePath := filepath.Join(year, monthFile)

	var txCount, incomeCount, expenseCount int
	var plannedCount int
	var totalIncome, totalExpense float64
	var plannedIncome, plannedExpense float64

	content, err := os.ReadFile(filePath)
	if err == nil {
		parsed, _ := parser.ParseMonthFile(string(content))

		for _, tx := range parsed {
			// Count uncompleted rules separately
			if tx.IsRule && !tx.Completed {
				plannedCount++
				if tx.IsIncome() {
					plannedIncome += tx.Amount
				} else {
					plannedExpense += -tx.Amount
				}
				continue
			}

			// Count completed transactions and non-rule transactions
			txCount++
			if tx.IsIncome() {
				incomeCount++
				totalIncome += tx.Amount
			} else {
				expenseCount++
				totalExpense += -tx.Amount
			}
		}
	}

	// Count categories and projects
	tagsCount := countUniqueTags(year, currentMonth)
	projectsCount := countUniqueProjects(year, currentMonth)

	// Print status
	fmt.Println()
	fmt.Println(i18n.T("status.header"))
	fmt.Println("========================================")
	fmt.Println()

	fmt.Printf("📅 Current Period: %s %d\n", time.Month(currentMonth), now.Year())
	fmt.Println()

	fmt.Println("📊 Completed Transactions:")
	fmt.Printf("   Total: %d (Income: %d, Expense: %d)\n", txCount, incomeCount, expenseCount)
	fmt.Printf("   Total Income:  %.2f\n", totalIncome)
	fmt.Printf("   Total Expense: %.2f\n", totalExpense)
	fmt.Printf("   Net:           %.2f\n", totalIncome-totalExpense)
	fmt.Println()

	if plannedCount > 0 {
		fmt.Println("📅 Planned (Uncompleted Rules):")
		fmt.Printf("   Total: %d\n", plannedCount)
		fmt.Printf("   Expected Income:  %.2f\n", plannedIncome)
		fmt.Printf("   Expected Expense: %.2f\n", plannedExpense)
		fmt.Printf("   Expected Net:     %.2f\n", plannedIncome-plannedExpense)
		fmt.Println()
	}

	fmt.Println("🏷️ Categories:")
	fmt.Printf("   Active Tags: %d\n", tagsCount)
	fmt.Printf("   Active Projects: %d\n", projectsCount)
	fmt.Println()

	fmt.Println("⚙️ Rules:")
	fmt.Printf("   Active Rules: %d\n", rulesCount)
	fmt.Println()

	// Check for unparsed lines
	unparsedCount := countUnparsedLines(year, currentMonth)
	if unparsedCount > 0 {
		fmt.Printf("⚠️  Warning: %d unparsed lines in current month\n", unparsedCount)
		fmt.Println("   Run 'spendgrid validate' for details")
		fmt.Println()
	}

	// Check exchange rates
	cachePath := filepath.Join(os.Getenv("HOME"), ".local", "share", "spendgrid", "data", "exchange_rates.json")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		fmt.Println("💱 Exchange rates: Not cached")
		fmt.Println("   Run 'spendgrid exchange refresh' to update")
	} else {
		fmt.Println("💱 Exchange rates: Cached")
	}
	fmt.Println()

	fmt.Println(i18n.T("status.footer"))
	fmt.Println()

	return nil
}

func countUniqueTags(year string, month int) int {
	monthFile := parser.GetMonthFile(month)
	filePath := filepath.Join(year, monthFile)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}

	parsed, _ := parser.ParseMonthFile(string(content))

	tagSet := make(map[string]bool)
	for _, tx := range parsed {
		for _, tag := range tx.Tags {
			tagSet[tag] = true
		}
	}

	return len(tagSet)
}

func countUniqueProjects(year string, month int) int {
	monthFile := parser.GetMonthFile(month)
	filePath := filepath.Join(year, monthFile)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}

	parsed, _ := parser.ParseMonthFile(string(content))

	projectSet := make(map[string]bool)
	for _, tx := range parsed {
		for _, proj := range tx.Projects {
			projectSet[proj] = true
		}
	}

	return len(projectSet)
}

func countUnparsedLines(year string, month int) int {
	monthFile := parser.GetMonthFile(month)
	filePath := filepath.Join(year, monthFile)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}

	_, unparsed := parser.ParseMonthFile(string(content))
	return len(unparsed)
}

// ShowStatusForPath displays the status for a specific directory path
func ShowStatusForPath(dirPath string) error {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	// Change to target directory
	if err := os.Chdir(dirPath); err != nil {
		return fmt.Errorf("failed to change directory: %v", err)
	}

	// Show status
	err = ShowStatus()

	// Change back to original directory
	os.Chdir(originalDir)

	return err
}
