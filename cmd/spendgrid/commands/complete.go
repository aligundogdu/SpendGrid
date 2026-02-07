package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/parser"
)

// CompleteCmd represents the complete command
var CompleteCmd = &cobra.Command{
	Use:   "complete [rule_id]",
	Short: "Mark a rule as completed",
	Long:  `Mark a recurring rule as completed by its ID. If no ID provided, shows recent rules to select from.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var ruleID string

		if len(args) == 0 {
			// Show recent uncompleted rules and let user select
			recentRules, err := getRecentUncompletedRules(10)
			if err != nil {
				color.Red("Error: %v", err)
				return
			}

			if len(recentRules) == 0 {
				color.Green("✓ All rules in current month are already completed!")
				return
			}

			fmt.Println()
			color.Cyan("Uncompleted Rules (last 10):")
			fmt.Println(strings.Repeat("-", 70))
			for i, rule := range recentRules {
				fmt.Printf("%2d. ☐ %02d | %-30s | %s | %s\n",
					i+1, rule.Day, rule.Description, rule.ID, rule.Amount)
			}
			fmt.Println(strings.Repeat("-", 70))
			fmt.Println()

			// Ask for rule selection
			fmt.Print("Enter rule number (1-N) or ID (or press Enter to cancel): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input == "" {
				color.Yellow("Cancelled")
				return
			}

			// Check if input is a number (index) or ID
			if idx, err := strconv.Atoi(input); err == nil && idx > 0 && idx <= len(recentRules) {
				// Input is a valid index
				ruleID = recentRules[idx-1].ID
			} else {
				// Input is treated as ID
				ruleID = input
			}
		} else {
			ruleID = args[0]
		}

		if err := toggleRuleCompletion(ruleID, true); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("✓ Rule '%s' marked as completed", ruleID)
	},
}

// UncompleteCmd represents the uncomplete command
var UncompleteCmd = &cobra.Command{
	Use:   "uncomplete [rule_id]",
	Short: "Mark a rule as not completed",
	Long:  `Mark a recurring rule as not completed by its ID. If no ID provided, shows recent rules to select from.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var ruleID string

		if len(args) == 0 {
			// Show recent completed rules and let user select
			recentRules, err := getRecentCompletedRules(10)
			if err != nil {
				color.Red("Error: %v", err)
				return
			}

			if len(recentRules) == 0 {
				color.Yellow("No completed rules found in current month.")
				return
			}

			fmt.Println()
			color.Cyan("Completed Rules (last 10):")
			fmt.Println(strings.Repeat("-", 70))
			for i, rule := range recentRules {
				fmt.Printf("%2d. ☑ %02d | %-30s | %s | %s\n",
					i+1, rule.Day, rule.Description, rule.ID, rule.Amount)
			}
			fmt.Println(strings.Repeat("-", 70))
			fmt.Println()

			// Ask for rule selection
			fmt.Print("Enter rule number (1-N) or ID (or press Enter to cancel): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input == "" {
				color.Yellow("Cancelled")
				return
			}

			// Check if input is a number (index) or ID
			if idx, err := strconv.Atoi(input); err == nil && idx > 0 && idx <= len(recentRules) {
				// Input is a valid index
				ruleID = recentRules[idx-1].ID
			} else {
				// Input is treated as ID
				ruleID = input
			}
		} else {
			ruleID = args[0]
		}

		if err := toggleRuleCompletion(ruleID, false); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("✓ Rule '%s' marked as not completed", ruleID)
	},
}

// CompleteMonthCmd represents the complete-month command
var CompleteMonthCmd = &cobra.Command{
	Use:   "complete-month [YYYY-MM]",
	Short: "Complete all rules in a month",
	Long:  `Mark all rules in a specified month (or current month) as completed.`,
	Run: func(cmd *cobra.Command, args []string) {
		var yearMonth string
		if len(args) > 0 {
			yearMonth = args[0]
		} else {
			now := time.Now()
			yearMonth = now.Format("2006-01")
		}

		if err := completeAllRulesInMonth(yearMonth); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("✓ All rules in %s marked as completed", yearMonth)
	},
}

// RuleInfo holds information about a rule for display
type RuleInfo struct {
	Day         int
	Description string
	ID          string
	Amount      string
	Completed   bool
}

// getRecentUncompletedRules gets uncompleted rules from the current month
func getRecentUncompletedRules(limit int) ([]RuleInfo, error) {
	return getRecentRulesWithFilter(limit, false)
}

// getRecentCompletedRules gets completed rules from the current month
func getRecentCompletedRules(limit int) ([]RuleInfo, error) {
	return getRecentRulesWithFilter(limit, true)
}

// getRecentRulesWithFilter gets rules from current month with completion filter
func getRecentRulesWithFilter(limit int, completed bool) ([]RuleInfo, error) {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return nil, fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	now := time.Now()
	monthFile := parser.GetMonthFile(int(now.Month()))
	yearDir := strconv.Itoa(now.Year())
	filePath := filepath.Join(yearDir, monthFile)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read month file: %v", err)
	}

	var rules []RuleInfo
	lines := strings.Split(string(content), "\n")
	inRulesSection := false

	// Pattern to match rule lines and extract info
	rulePattern := regexp.MustCompile(`^-\s*\[([x\s])\]\s+(\d+)\s*\|\s*([^[]+)\[(\w+)\]\s*\|\s*([^|]+)`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "## RULES" {
			inRulesSection = true
			continue
		}

		if inRulesSection && strings.HasPrefix(trimmed, "##") {
			break
		}

		if inRulesSection && trimmed != "" {
			if matches := rulePattern.FindStringSubmatch(trimmed); matches != nil {
				isCompleted := matches[1] == "x"
				day, _ := strconv.Atoi(matches[2])
				desc := strings.TrimSpace(matches[3])
				id := matches[4]
				amount := strings.TrimSpace(matches[5])

				// Filter by completion status
				if isCompleted == completed {
					rules = append(rules, RuleInfo{
						Day:         day,
						Description: desc,
						ID:          id,
						Amount:      amount,
						Completed:   isCompleted,
					})
				}
			}
		}
	}

	// Return up to limit rules
	if len(rules) > limit {
		return rules[:limit], nil
	}
	return rules, nil
}

// toggleRuleCompletion finds and toggles the completion status of a rule in month files
func toggleRuleCompletion(ruleID string, completed bool) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	now := time.Now()
	year := now.Year()

	// Search in all month files
	for month := 1; month <= 12; month++ {
		monthFile := parser.GetMonthFile(month)
		yearDir := strconv.Itoa(year)
		filePath := filepath.Join(yearDir, monthFile)

		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip if file doesn't exist
		}

		updated, found := updateRuleInContent(string(content), ruleID, completed)
		if found {
			if err := os.WriteFile(filePath, []byte(updated), 0644); err != nil {
				return fmt.Errorf("failed to update file: %v", err)
			}
			return nil
		}
	}

	return fmt.Errorf("rule with ID '%s' not found in any month file", ruleID)
}

// updateRuleInContent updates the checkbox status of a rule in file content
func updateRuleInContent(content string, ruleID string, completed bool) (string, bool) {
	lines := strings.Split(content, "\n")
	found := false
	newCheckbox := "[ ]"
	if completed {
		newCheckbox = "[x]"
	}

	// Pattern to match rule lines: - [ ] DD | Description [ID] | ...
	// or: - [x] DD | Description [ID] | ...
	checkboxPattern := regexp.MustCompile(`^(-\s*\[)[x\s](\]\s+\d+\s*\|[^[]*\[` + regexp.QuoteMeta(ruleID) + `\])`)

	for i, line := range lines {
		if checkboxPattern.MatchString(line) {
			lines[i] = checkboxPattern.ReplaceAllString(line, "${1}"+string(newCheckbox[1])+"${2}")
			found = true
			break
		}
	}

	return strings.Join(lines, "\n"), found
}

// completeAllRulesInMonth marks all uncompleted rules in a month as completed
func completeAllRulesInMonth(yearMonth string) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	// Parse year and month from YYYY-MM format
	parts := strings.Split(yearMonth, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format. Use YYYY-MM")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid year: %v", err)
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return fmt.Errorf("invalid month: %v", err)
	}

	monthFile := parser.GetMonthFile(month)
	yearDir := strconv.Itoa(year)
	filePath := filepath.Join(yearDir, monthFile)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("month file not found: %v", err)
	}

	updated, count := completeAllRulesInContent(string(content))
	if count == 0 {
		return fmt.Errorf("no uncompleted rules found in %s", yearMonth)
	}

	if err := os.WriteFile(filePath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to update file: %v", err)
	}

	color.Green("✓ %d rule(s) marked as completed", count)
	return nil
}

// completeAllRulesInContent updates all uncompleted rules to completed
func completeAllRulesInContent(content string) (string, int) {
	lines := strings.Split(content, "\n")
	count := 0

	// Pattern to match uncompleted rule lines: - [ ] ...
	// But not: - [x] ...
	uncompletedPattern := regexp.MustCompile(`^(-\s*\[)\s(\].*)`)

	for i, line := range lines {
		// Check if this line is in RULES section
		if strings.TrimSpace(line) == "## RULES" {
			// Start processing rules section
			for j := i + 1; j < len(lines); j++ {
				innerLine := lines[j]
				// Stop if we hit another section or empty line after rules
				if strings.HasPrefix(innerLine, "##") {
					break
				}
				// Skip empty lines
				if strings.TrimSpace(innerLine) == "" {
					continue
				}
				// Check for uncompleted rule
				if uncompletedPattern.MatchString(innerLine) {
					lines[j] = uncompletedPattern.ReplaceAllString(innerLine, "${1}x${2}")
					count++
				}
			}
			break
		}
	}

	return strings.Join(lines, "\n"), count
}

func init() {
	// These commands will be added to root command in main.go
}
