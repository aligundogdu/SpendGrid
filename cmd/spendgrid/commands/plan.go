package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/rules"
)

// PlanCmd represents the plan command
var PlanCmd = &cobra.Command{
	Use:   "plan [month]",
	Short: "Show planned vs actual spending",
	Long:  `Display a comparison between planned (rules) and actual (transactions) spending for the current or specified month.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		month := 0
		if len(args) > 0 {
			if m, err := strconv.Atoi(args[0]); err == nil && m >= 1 && m <= 12 {
				month = m
			}
		}

		if err := showPlanReport(month); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}

func showPlanReport(month int) error {
	// Get current date
	now := time.Now()
	year := now.Year()
	if month == 0 {
		month = int(now.Month())
	}

	// Update remaining amounts
	if err := rules.UpdateRemainingAmounts(year, month); err != nil {
		return err
	}

	// Load all active rules
	allRules, err := rules.GetActiveRules()
	if err != nil {
		return err
	}

	// Separate rules by type
	var incomeRules, expenseRules []rules.Rule
	for _, rule := range allRules {
		if rule.Type == "income" {
			incomeRules = append(incomeRules, rule)
		} else {
			expenseRules = append(expenseRules, rule)
		}
	}

	// Print header
	color.Cyan("\n📊 Planlanan vs Gerçekleşen - %s %d", time.Month(month), year)
	fmt.Println()
	color.White("========================================")

	// Income section
	if len(incomeRules) > 0 {
		fmt.Println()
		color.Green("💰 Gelirler:")
		color.White("----------------------------------------------------------------")
		for _, rule := range incomeRules {
			printRuleProgress(&rule)
		}
	}

	// Expense section
	if len(expenseRules) > 0 {
		fmt.Println()
		color.Red("💸 Giderler:")
		color.White("----------------------------------------------------------------")
		for _, rule := range expenseRules {
			printRuleProgress(&rule)
		}
	}

	// Summary
	fmt.Println()
	color.White("========================================")
	printSummary(incomeRules, expenseRules)

	return nil
}

func printRuleProgress(rule *rules.Rule) {
	planned := rule.Amount
	actual := planned - rule.RemainingAmount
	remaining := rule.RemainingAmount

	// Determine status and color
	var statusIcon, statusText string
	if remaining <= 0 {
		// Completed or over
		if actual > planned {
			// Over payment
			statusIcon = "+"
			statusText = fmt.Sprintf("[Fazla: %.2f]", actual-planned)
			color.Yellow("%-30s %10.2f / %-10.2f %s",
				rule.Name, actual, planned, statusText)
		} else {
			// Completed exactly
			statusIcon = "✓"
			statusText = "[Tamamlandı]"
			color.Green("%-30s %10.2f / %-10.2f %s",
				rule.Name, actual, planned, statusText)
		}
	} else {
		// Partial
		statusIcon = "⊘"
		statusText = fmt.Sprintf("[Kalan: %.2f]", remaining)
		color.Yellow("%-30s %10.2f / %-10.2f %s",
			rule.Name, actual, planned, statusText)
	}

	_ = statusIcon // Unused for now
}

func printSummary(incomeRules, expenseRules []rules.Rule) {
	var totalPlannedIncome, totalActualIncome float64
	var totalPlannedExpense, totalActualExpense float64

	for _, rule := range incomeRules {
		totalPlannedIncome += rule.Amount
		totalActualIncome += rule.Amount - rule.RemainingAmount
	}

	for _, rule := range expenseRules {
		totalPlannedExpense += rule.Amount
		totalActualExpense += rule.Amount - rule.RemainingAmount
	}

	plannedNet := totalPlannedIncome - totalPlannedExpense
	actualNet := totalActualIncome - totalActualExpense
	diff := actualNet - plannedNet

	fmt.Println()
	color.Cyan("📈 Toplam:")
	fmt.Printf("  Planlanan Net: %+10.2f\n", plannedNet)
	fmt.Printf("  Gerçekleşen Net: %+10.2f\n", actualNet)

	if diff >= 0 {
		color.Green("  Fark: %+10.2f (İyi)", diff)
	} else {
		color.Red("  Fark: %+10.2f (Dikkat)", diff)
	}
	fmt.Println()
}
