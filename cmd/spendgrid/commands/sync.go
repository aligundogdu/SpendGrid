package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/rules"
)

// SyncCmd represents the sync command
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync rules to month files",
	Long:  `Synchronize all active rules to the current and future month files.`,
	Run: func(cmd *cobra.Command, args []string) {
		color.Yellow("Synchronizing rules...")

		result, err := rules.SyncRules()
		if err != nil {
			color.Red("Error: %v", err)
			return
		}

		color.Green("✓ Sync completed!")
		color.White("  Added: %d", result.Added)
		color.White("  Updated: %d", result.Updated)
		color.White("  Skipped: %d", result.Skipped)

		if len(result.Errors) > 0 {
			color.Yellow("\n⚠ Warnings:")
			for _, e := range result.Errors {
				color.Yellow("  - %s", e)
			}
		}
	},
}
