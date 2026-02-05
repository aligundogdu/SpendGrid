package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/exchange"
)

// ExchangeCmd represents the exchange command
var ExchangeCmd = &cobra.Command{
	Use:   "exchange",
	Short: "Manage exchange rates",
	Long:  `View, refresh, or set exchange rates for currency conversion.`,
}

// ExchangeShowCmd shows current rates
var ExchangeShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current exchange rates",
	Long:  `Display the current cached exchange rates.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := exchange.ShowRates(); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}

// ExchangeRefreshCmd refreshes rates from API
var ExchangeRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh exchange rates from API",
	Long:  `Fetch latest exchange rates from TCMB (Central Bank of Turkey).`,
	Run: func(cmd *cobra.Command, args []string) {
		color.Yellow("Fetching exchange rates...")

		if err := exchange.RefreshRates(); err != nil {
			color.Red("Error: %v", err)
			return
		}

		color.Green("✓ Exchange rates refreshed successfully!")
	},
}

// ExchangeSetCmd sets manual rate
var ExchangeSetCmd = &cobra.Command{
	Use:   "set <date> <currency> <rate>",
	Short: "Set manual exchange rate",
	Long:  `Set a manual exchange rate for a specific date. Format: YYYY-MM-DD`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		dateStr := args[0]
		currency := args[1]
		rateStr := args[2]

		// Parse rate
		var rate float64
		if _, err := fmt.Sscanf(rateStr, "%f", &rate); err != nil {
			color.Red("Error: Invalid rate format")
			return
		}

		if err := exchange.SetManualRate(dateStr, currency, rate); err != nil {
			color.Red("Error: %v", err)
			return
		}

		color.Green("✓ Exchange rate set: %s = %.4f on %s", currency, rate, dateStr)
	},
}

func init() {
	ExchangeCmd.AddCommand(ExchangeShowCmd)
	ExchangeCmd.AddCommand(ExchangeRefreshCmd)
	ExchangeCmd.AddCommand(ExchangeSetCmd)
}
