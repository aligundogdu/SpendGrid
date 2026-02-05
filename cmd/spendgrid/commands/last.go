package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/last"
)

// LastCmd represents the last command
var LastCmd = &cobra.Command{
	Use:   "last",
	Short: "Show last 10 SpendGrid directories",
	Long:  `Display the last 10 SpendGrid directories you have worked in.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := last.ShowRecentDirs(); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}
