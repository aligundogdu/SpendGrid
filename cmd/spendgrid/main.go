package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/cmd/spendgrid/commands"
	"spendgrid/internal/config"
	"spendgrid/internal/i18n"
	"spendgrid/internal/rules"
)

// version is set during build using -ldflags
var version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spendgrid",
	Short: "SpendGrid - Financial management tool",
	Long: color.CyanString("SpendGrid") + ` - Financial management tool

A local-first, file-based financial management tool.
Track your income, expenses, and budgets with simple markdown files.
`,
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load i18n and config
		if err := i18n.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading translations: %v\n", err)
			os.Exit(1)
		}

		// Auto-sync rules (except for init, version, and help commands)
		if cmd.Name() != "init" && cmd.Name() != "version" && cmd.Name() != "help" {
			if _, err := rules.SyncRules(); err != nil {
				// Silent fail - don't block user on sync errors
				fmt.Fprintf(os.Stderr, "Warning: auto-sync failed: %v\n", err)
			}
		}

		// Save current directory to recent list (if it's a SpendGrid directory)
		if cmd.Name() != "last" && cmd.Name() != "version" && cmd.Name() != "help" {
			if err := config.SaveCurrentDirectory(); err != nil {
				// Silent fail
				_ = err
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add all commands to root
	rootCmd.AddCommand(commands.InitCmd)
	rootCmd.AddCommand(commands.StatusCmd)
	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.EditCmd)
	rootCmd.AddCommand(commands.RemoveCmd)
	rootCmd.AddCommand(commands.SyncCmd)
	rootCmd.AddCommand(commands.LastCmd)
	rootCmd.AddCommand(commands.RulesCmd)
	rootCmd.AddCommand(commands.ValidateCmd)
	rootCmd.AddCommand(commands.InvestmentsCmd)
	rootCmd.AddCommand(commands.ReportCmd)
	rootCmd.AddCommand(commands.ExchangeCmd)
	rootCmd.AddCommand(commands.ConfigCmd)
	rootCmd.AddCommand(commands.PoolCmd)
}

func main() {
	Execute()
}
