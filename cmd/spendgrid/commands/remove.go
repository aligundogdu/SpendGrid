package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/transaction"
)

// RemoveCmd represents the remove command
var RemoveCmd = &cobra.Command{
	Use:     "remove <line_number>",
	Aliases: []string{"rm"},
	Short:   "Remove a transaction",
	Long:    `Remove a transaction by its line number in the current month.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := transaction.RemoveTransaction(args[0]); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("✓ Transaction removed successfully!")
	},
}
