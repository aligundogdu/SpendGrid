package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/investment"
)

// InvestmentsCmd represents the investments command
var InvestmentsCmd = &cobra.Command{
	Use:   "investments",
	Short: "Show investment portfolio",
	Long:  `Display your investment portfolio summary.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := investment.GenerateInvestmentReport(); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}
