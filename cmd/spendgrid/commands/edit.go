package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/transaction"
)

// EditCmd represents the edit command
var EditCmd = &cobra.Command{
	Use:   "edit <line_number>",
	Short: "Edit a transaction",
	Long:  `Edit a transaction by its line number in the current month.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := transaction.EditTransaction(args[0]); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}
