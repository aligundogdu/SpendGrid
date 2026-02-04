package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/rules"
)

// RulesCmd represents the rules command
var RulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage recurring rules",
	Long:  `Manage recurring transaction rules for automatic entries.`,
}

// RulesListCmd lists all rules
var RulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all rules",
	Run: func(cmd *cobra.Command, args []string) {
		if err := rules.ListRules(); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}

// RulesAddCmd adds a new rule
var RulesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new rule",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) >= 4 {
			// Direct mode: spendgrid rules add <name> <amount> <currency> <type> [flags]
			if err := rules.AddRuleDirect(args); err != nil {
				color.Red("Error: %v", err)
				return
			}
			color.Green("✓ Rule added successfully!")
		} else {
			// Interactive mode
			if err := rules.AddRuleInteractive(); err != nil {
				color.Red("Error: %v", err)
				return
			}
		}
	},
}

func init() {
	RulesCmd.AddCommand(RulesListCmd)
	RulesCmd.AddCommand(RulesAddCmd)
	// Add more subcommands here: edit, toggle, remove
}
