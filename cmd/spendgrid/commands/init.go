package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/filesystem"
)

// InitCmd represents the init command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SpendGrid in current directory",
	Long:  `Initialize SpendGrid in the current directory by creating the necessary structure.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := filesystem.Init(); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("✓ SpendGrid initialized successfully!")
	},
}
