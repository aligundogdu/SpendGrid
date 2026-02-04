package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/config"
)

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and manage SpendGrid configuration settings.`,
}

// ConfigListCmd lists all config values
var ConfigListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `Display all current configuration settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetGlobalConfig()

		color.Cyan("Configuration:")
		fmt.Println()
		fmt.Printf("  Language: %s\n", cfg.Language)
		fmt.Println()

		color.Yellow("Recent directories stored in: ~/.config/spendgrid/")
	},
}

// ConfigGetCmd gets a specific config value
var ConfigGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long:  `Get the value of a specific configuration key.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])
		cfg := config.GetGlobalConfig()

		switch key {
		case "language", "lang":
			fmt.Println(cfg.Language)
		default:
			color.Yellow("Unknown config key: %s", key)
			color.Yellow("Supported keys: language (lang)")
		}
	},
}

// ConfigSetCmd sets a config value
var ConfigSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Supported keys:
  - language (or lang): Set the display language (en, tr)

Examples:
  spendgrid config set language tr    Set language to Turkish`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])
		value := args[1]

		switch key {
		case "language", "lang":
			if err := config.SetLanguage(value); err != nil {
				color.Red("Error: %v", err)
				return
			}
			color.Green("✓ Language set to: %s", value)
		default:
			color.Yellow("Unknown config key: %s", key)
			color.Yellow("Supported keys: language (lang)")
		}
	},
}

func init() {
	ConfigCmd.AddCommand(ConfigListCmd)
	ConfigCmd.AddCommand(ConfigGetCmd)
	ConfigCmd.AddCommand(ConfigSetCmd)
}
