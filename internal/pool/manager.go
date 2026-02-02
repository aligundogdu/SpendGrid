package pool

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"spendgrid/internal/i18n"
	"spendgrid/internal/parser"
)

const backlogFile = "_pool/backlog.md"

// ShowPool displays all items in the backlog
func ShowPool() error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	content, err := os.ReadFile(backlogFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(i18n.T("pool.empty"))
			return nil
		}
		return fmt.Errorf("failed to read backlog: %v", err)
	}

	parsed, unparsed := parser.ParseMonthFile(string(content))

	if len(parsed) == 0 && len(unparsed) == 0 {
		fmt.Println(i18n.T("pool.empty"))
		return nil
	}

	fmt.Println()
	fmt.Println(i18n.T("pool.header"))
	fmt.Println(strings.Repeat("=", 80))

	// Show parsed items
	if len(parsed) > 0 {
		fmt.Println(i18n.T("pool.items"))
		fmt.Println(strings.Repeat("-", 80))
		for i, tx := range parsed {
			fmt.Printf("%3d | %s | %10.2f %s | %s\n",
				i+1,
				truncate(tx.Description, 25),
				tx.Amount,
				tx.Currency,
				strings.Join(tx.Tags, " "))
		}
	}

	// Show unparsed items
	if len(unparsed) > 0 {
		fmt.Println()
		fmt.Println(i18n.T("pool.unparsed"))
		fmt.Println(strings.Repeat("-", 80))
		for _, tx := range unparsed {
			fmt.Printf("Line %d: %s\n", tx.LineNumber, tx.Raw)
		}
	}

	fmt.Println()
	return nil
}

// AddPoolItem adds a new item to the backlog
func AddPoolItem() error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	reader := bufio.NewReader(os.Stdin)

	// Description
	fmt.Print(i18n.T("pool.desc_prompt") + " ")
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return fmt.Errorf("description cannot be empty")
	}

	// Amount
	fmt.Print(i18n.T("pool.amount_prompt") + " ")
	amountStr, _ := reader.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)

	// Expected month (optional)
	fmt.Print(i18n.T("pool.month_prompt") + " ")
	monthStr, _ := reader.ReadString('\n')
	monthStr = strings.TrimSpace(monthStr)

	// Tags
	fmt.Print(i18n.T("pool.tags_prompt") + " ")
	tagsStr, _ := reader.ReadString('\n')
	tagsStr = strings.TrimSpace(tagsStr)

	// Format: - DESC | AMOUNT | MONTH | TAGS
	line := fmt.Sprintf("- %s | %s | %s | %s", desc, amountStr, monthStr, tagsStr)

	// Append to file
	f, err := os.OpenFile(backlogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open backlog: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write to backlog: %v", err)
	}

	fmt.Println(i18n.T("pool.add_success"))
	return nil
}

// MovePoolItem moves an item from backlog to a specific month
func MovePoolItem(lineNumStr, monthStr string) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	lineNum, err := strconv.Atoi(lineNumStr)
	if err != nil || lineNum < 1 {
		return fmt.Errorf("invalid line number: %s", lineNumStr)
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return fmt.Errorf("invalid month: %s", monthStr)
	}

	// Read backlog
	content, err := os.ReadFile(backlogFile)
	if err != nil {
		return fmt.Errorf("failed to read backlog: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the line
	txLine := 0
	actualLine := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-") {
			txLine++
			if txLine == lineNum {
				actualLine = i
				break
			}
		}
	}

	if actualLine == -1 {
		return fmt.Errorf("item not found at line %d", lineNum)
	}

	// Parse the transaction
	tx := parser.ParseTransaction(lines[actualLine], actualLine+1)
	if tx == nil || tx.IsUnparsed {
		return fmt.Errorf("cannot move unparsed item")
	}

	// Update day if not set
	if tx.Day == 0 {
		tx.Day = 1
	}

	// Add to month file
	year := strconv.Itoa(time.Now().Year())
	monthFile := parser.GetMonthFile(month)
	filePath := filepath.Join(year, monthFile)

	if err := addTransactionToFile(filePath, tx); err != nil {
		return err
	}

	// Remove from backlog
	lines = append(lines[:actualLine], lines[actualLine+1:]...)
	if err := os.WriteFile(backlogFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to update backlog: %v", err)
	}

	fmt.Printf(i18n.T("pool.move_success"), lineNum, month)
	fmt.Println()
	return nil
}

// RemovePoolItem removes an item from the backlog
func RemovePoolItem(lineNumStr string) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	lineNum, err := strconv.Atoi(lineNumStr)
	if err != nil || lineNum < 1 {
		return fmt.Errorf("invalid line number: %s", lineNumStr)
	}

	// Read backlog
	content, err := os.ReadFile(backlogFile)
	if err != nil {
		return fmt.Errorf("failed to read backlog: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the line
	txLine := 0
	actualLine := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-") {
			txLine++
			if txLine == lineNum {
				actualLine = i
				break
			}
		}
	}

	if actualLine == -1 {
		return fmt.Errorf("item not found at line %d", lineNum)
	}

	// Confirm
	fmt.Printf("Remove '%s'? [y/n]: ", truncate(lines[actualLine], 40))
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println(i18n.T("common.cancel"))
		return nil
	}

	// Remove
	lines = append(lines[:actualLine], lines[actualLine+1:]...)
	if err := os.WriteFile(backlogFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to update backlog: %v", err)
	}

	fmt.Println(i18n.T("pool.remove_success"))
	return nil
}

func addTransactionToFile(filePath string, tx *parser.Transaction) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the ROWS section
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

	// Insert
	lines = append(lines[:insertIndex], append([]string{formatted}, lines[insertIndex:]...)...)

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
