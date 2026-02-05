package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/validator"
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate all files",
	Long:  `Validate all SpendGrid files for errors and inconsistencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validator.ValidateAll(); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("✓ Validation completed successfully!")
	},
}
