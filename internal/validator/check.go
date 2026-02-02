package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"spendgrid/internal/i18n"
	"spendgrid/internal/parser"
)

// ValidationResult holds validation results
type ValidationResult struct {
	TotalFiles    int
	TotalLines    int
	ParsedCount   int
	UnparsedCount int
	UnparsedLines []UnparsedLine
	Errors        []string
}

// UnparsedLine represents an unparsed line with context
type UnparsedLine struct {
	File    string
	LineNum int
	Content string
}

// ValidateAll validates all files in the database
func ValidateAll() error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	result := &ValidationResult{
		UnparsedLines: []UnparsedLine{},
		Errors:        []string{},
	}

	// Validate current year
	now := time.Now()
	year := strconv.Itoa(now.Year())
	yearDir := year

	// Check if year directory exists
	if _, err := os.Stat(yearDir); os.IsNotExist(err) {
		return fmt.Errorf("year directory not found: %s", yearDir)
	}

	// Validate each month file
	for month := 1; month <= 12; month++ {
		monthFile := parser.GetMonthFile(month)
		filePath := filepath.Join(yearDir, monthFile)

		if err := validateFile(filePath, result); err != nil {
			// File might not exist, that's ok
			continue
		}
	}

	// Validate config files
	validateConfigFile("_config/settings.yml", result)
	validateConfigFile("_config/rules.yml", result)

	// Print results
	printValidationResults(result)

	return nil
}

func validateFile(filePath string, result *ValidationResult) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	result.TotalFiles++

	parsed, unparsed := parser.ParseMonthFile(string(content))
	result.ParsedCount += len(parsed)
	result.UnparsedCount += len(unparsed)
	result.TotalLines += len(parsed) + len(unparsed)

	for _, tx := range unparsed {
		result.UnparsedLines = append(result.UnparsedLines, UnparsedLine{
			File:    filepath.Base(filePath),
			LineNum: tx.LineNumber,
			Content: tx.Raw,
		})
	}

	return nil
}

func validateConfigFile(filePath string, result *ValidationResult) {
	if _, err := os.Stat(filePath); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Config file missing: %s", filePath))
	}
}

func printValidationResults(result *ValidationResult) {
	fmt.Println()
	fmt.Println(i18n.T("validation.header"))
	fmt.Println("========================================")

	fmt.Printf("Files checked: %d\n", result.TotalFiles)
	fmt.Printf("Total lines: %d\n", result.TotalLines)
	fmt.Printf("Parsed successfully: %d\n", result.ParsedCount)
	fmt.Printf("Unparsed lines: %d\n", result.UnparsedCount)

	if len(result.UnparsedLines) > 0 {
		fmt.Println()
		fmt.Println(i18n.T("validation.unparsed_header"))
		fmt.Println("----------------------------------------")
		for _, line := range result.UnparsedLines {
			fmt.Printf("%s:%d | %s\n", line.File, line.LineNum, line.Content)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Println(i18n.T("validation.errors_header"))
		fmt.Println("----------------------------------------")
		for _, err := range result.Errors {
			fmt.Printf("- %s\n", err)
		}
	}

	if result.UnparsedCount == 0 && len(result.Errors) == 0 {
		fmt.Println()
		fmt.Println(i18n.T("validation.all_ok"))
	}

	fmt.Println()
}
