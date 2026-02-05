package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/transaction"
)

// ListCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list [month]",
	Short: "List transactions",
	Long:  `List transactions for the current month or specify a month (01-12).`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		month := ""
		if len(args) > 0 {
			month = args[0]
		}

		if err := transaction.ListTransactions(month); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}
