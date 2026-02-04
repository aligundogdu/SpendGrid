package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/status"
)

// StatusCmd represents the status command
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database status",
	Long:  `Display the current status of the SpendGrid database including transaction counts, rules, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := status.ShowStatus(); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}
