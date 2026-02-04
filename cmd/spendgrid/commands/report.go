package commands

import (
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"spendgrid/internal/reports"
)

// ReportCmd represents the report command
var ReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reports",
	Long:  `Generate financial reports for current month, year, or as HTML.`,
}

// ReportMonthlyCmd generates monthly report
var ReportMonthlyCmd = &cobra.Command{
	Use:   "monthly [month]",
	Short: "Generate monthly report",
	Long:  `Generate report for specified month (01-12) or current month if not specified.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		month := 0
		if len(args) > 0 {
			// Parse month argument
			if m, err := strconv.Atoi(args[0]); err == nil && m >= 1 && m <= 12 {
				month = m
			}
		}

		if err := reports.GenerateMonthlyReport(month); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}

// ReportYearlyCmd generates yearly report
var ReportYearlyCmd = &cobra.Command{
	Use:   "yearly",
	Short: "Generate yearly report",
	Long:  `Generate report for the entire year.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := reports.GenerateYearlyReport(); err != nil {
			color.Red("Error: %v", err)
			return
		}
	},
}

// ReportWebCmd generates HTML report
var ReportWebCmd = &cobra.Command{
	Use:   "web",
	Short: "Generate HTML report",
	Long:  `Generate interactive HTML report for web browser.`,
	Run: func(cmd *cobra.Command, args []string) {
		yearly, _ := cmd.Flags().GetBool("year")

		if err := reports.GenerateHTMLReport(yearly); err != nil {
			color.Red("Error: %v", err)
			return
		}
		color.Green("✓ HTML report generated!")
	},
}

func init() {
	ReportCmd.AddCommand(ReportMonthlyCmd)
	ReportCmd.AddCommand(ReportYearlyCmd)
	ReportCmd.AddCommand(ReportWebCmd)

	ReportWebCmd.Flags().BoolP("year", "y", false, "Generate yearly HTML report")
}
