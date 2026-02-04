package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/transaction"
)

// AddCmd represents the add command
var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new transaction",
	Long:  `Add a new transaction to the current month. Interactive mode will prompt for details.`,
	Run: func(cmd *cobra.Command, args []string) {
		direct, _ := cmd.Flags().GetBool("direct")

		if direct {
			// Direct mode
			if len(args) == 0 {
				color.Red("Error: Direct mode requires an argument")
				return
			}
			directInput := args[0]
			if err := transaction.AddDirectTransaction(directInput); err != nil {
				color.Red("Error: %v", err)
				return
			}
			color.Green("✓ Transaction added successfully!")
		} else {
			// Interactive mode
			if err := transaction.AddTransaction(); err != nil {
				color.Red("Error: %v", err)
				return
			}
		}
	},
}

func init() {
	AddCmd.Flags().BoolP("direct", "d", false, "Add transaction directly with format: DAY|DESC|AMOUNT|TAGS")
}
