package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"spendgrid/internal/config"
	"spendgrid/internal/exchange"
	"spendgrid/internal/filesystem"
	"spendgrid/internal/i18n"
	"spendgrid/internal/investment"
	"spendgrid/internal/parser"
	"spendgrid/internal/pool"
	"spendgrid/internal/reports"
	"spendgrid/internal/rules"
	"spendgrid/internal/status"
	"spendgrid/internal/transaction"
	"spendgrid/internal/validator"
)

// version is set during build using -ldflags
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Load i18n (this also initializes global config)
	if err := i18n.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading translations: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]

	// Auto-sync rules on startup (except for init and certain commands)
	if command != "init" && command != "version" && command != "--help" && command != "-h" {
		if _, err := rules.SyncRules(); err != nil {
			// Silent fail - don't block user on sync errors
			// Just log to stderr
			fmt.Fprintf(os.Stderr, "Warning: auto-sync failed: %v\n", err)
		}
	}

	switch command {
	case "init":
		handleInit()
	case "add":
		handleAdd(os.Args[2:])
	case "list":
		handleList(os.Args[2:])
	case "edit":
		handleEdit(os.Args[2:])
	case "remove", "rm":
		handleRemove(os.Args[2:])
	case "rules":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid rules <list|add|edit|toggle|remove>\n")
			os.Exit(1)
		}
		handleRules(os.Args[2:])
	case "sync":
		handleSync()
	case "report":
		handleReport(os.Args[2:])
	case "exchange":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid exchange [show|refresh|set]\n")
			fmt.Fprintf(os.Stderr, "  show   - Display current exchange rates (default)\n")
			fmt.Fprintf(os.Stderr, "  refresh - Fetch latest rates from API\n")
			fmt.Fprintf(os.Stderr, "  set    - Set manual exchange rate\n")
			os.Exit(1)
		}
		handleExchange(os.Args[2:])
	case "investments":
		handleInvestments()
	case "pool":
		if len(os.Args) < 3 {
			handlePool([]string{"list"})
		} else {
			handlePool(os.Args[2:])
		}
	case "validate":
		handleValidate()
	case "status":
		handleStatus()
	case "set":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid set <config> [arguments]\n")
			os.Exit(1)
		}
		handleSet(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf(i18n.Tfmt("commands.version.format", version))
		fmt.Println()
	case "help", "--help", "-h":
		printUsage()
	default:
		// Check if we have more than one argument (spendgrid command arg1 arg2...)
		// This happens if the user forgot quotes: spendgrid -100 food
		if len(os.Args) > 2 {
			fmt.Fprintln(os.Stderr, "Error: Too many arguments.")
			fmt.Fprintln(os.Stderr, "Did you forget quotes around your transaction?")
			fmt.Fprintln(os.Stderr, "Correct usage: spendgrid \"-100TL description #tag\"")
			os.Exit(1)
		}

		// We have exactly one argument (spendgrid "something")
		input := os.Args[1]

		// Heuristic: Is this a transaction or a typo of a command?
		// A transaction usually:
		// 1. Contains spaces (description, tags)
		// 2. Starts with a number or currency symbol (+, -, digit, $, €, ₺)
		isTransaction := false

		if strings.Contains(input, " ") {
			// "-100TL food" -> has space -> likely transaction
			isTransaction = true
		} else {
			// "100TL" -> no space -> check start char
			firstChar := string([]rune(input)[0])
			if strings.ContainsAny(firstChar, "0123456789+-$€₺") {
				isTransaction = true
			}
		}

		if isTransaction {
			handleQuickInput(input)
		} else {
			// Likely a typo like "addd", "stats", "ls", etc.
			fmt.Fprintf(os.Stderr, i18n.Tfmt("errors.unknown_command", command))
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "Did you mean to use quick input? Use quotes and ensure it looks like a transaction:")
			fmt.Fprintln(os.Stderr, "  spendgrid \"-100TL description #tag\"")
			printUsage()
			os.Exit(1)
		}
	}
}

func handleInit() {
	if err := filesystem.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleAdd(args []string) {
	if len(args) > 0 && args[0] == "--direct" {
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid add --direct \"DAY|DESC|AMOUNT|TAGS\"\n")
			os.Exit(1)
		}
		directInput := strings.Join(args[1:], " ")
		if err := transaction.AddDirectTransaction(directInput); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := transaction.AddTransaction(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func handleList(args []string) {
	month := ""
	if len(args) > 0 {
		month = args[0]
	}
	if err := transaction.ListTransactions(month); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleEdit(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: spendgrid edit <line_number>\n")
		os.Exit(1)
	}
	if err := transaction.EditTransaction(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleRemove(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: spendgrid remove <line_number>\n")
		os.Exit(1)
	}
	if err := transaction.RemoveTransaction(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleRules(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: spendgrid rules <list|add|edit|toggle|remove>\n")
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "list":
		if err := rules.ListRules(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "add":
		if len(args) == 1 {
			// Interactive mode
			if err := rules.AddRuleInteractive(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Direct mode: spendgrid rules add <name> <amount> <currency> <type> [flags]
			if len(args) < 5 {
				fmt.Fprintf(os.Stderr, "Usage: spendgrid rules add <name> <amount> <currency> <type> [flags]\n")
				fmt.Fprintf(os.Stderr, "  type: income or expense\n")
				fmt.Fprintf(os.Stderr, "  flags: --day <day> --tags <tag1,tag2> --project <project>\n")
				fmt.Fprintf(os.Stderr, "Example: spendgrid rules add \"Maaş\" 50000 TRY income --day 1 --tags maaş,gelir\n")
				os.Exit(1)
			}
			if err := rules.AddRuleDirect(args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	case "edit":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid rules edit <rule_id>\n")
			os.Exit(1)
		}
		if err := rules.EditRuleInteractive(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "toggle":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid rules toggle <rule_id>\n")
			os.Exit(1)
		}
		if err := rules.ToggleRuleStatus(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "remove":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid rules remove <rule_id>\n")
			os.Exit(1)
		}
		if err := rules.RemoveRule(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown rules command: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleSync() {
	if err := rules.SyncNow(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleReport(args []string) {
	if len(args) > 0 && args[0] == "--year" {
		if err := reports.GenerateYearlyReport(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else if len(args) > 0 && args[0] == "--web" {
		year := false
		if len(args) > 1 && args[1] == "--year" {
			year = true
		}
		if err := reports.GenerateHTMLReport(year); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		month := 0
		if len(args) > 0 {
			if m, err := strconv.Atoi(args[0]); err == nil && m >= 1 && m <= 12 {
				month = m
			}
		}
		if err := reports.GenerateMonthlyReport(month); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func handleExchange(args []string) {
	if len(args) == 0 {
		// Default to show if no subcommand
		if err := exchange.ShowRates(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	subcommand := args[0]

	switch subcommand {
	case "show", "list":
		if err := exchange.ShowRates(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "refresh":
		if err := exchange.RefreshRates(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Exchange rates refreshed successfully!")
	case "set":
		if len(args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid exchange set <date> <currency> <rate>\n")
			fmt.Fprintf(os.Stderr, "Example: spendgrid exchange set 2026-02-02 USD 35.50\n")
			os.Exit(1)
		}
		date := args[1]
		currency := args[2]
		rate, err := strconv.ParseFloat(args[3], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid rate: %v\n", err)
			os.Exit(1)
		}
		if err := exchange.SetManualRate(date, currency, rate); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exchange rate set: %s = %.4f on %s\n", currency, rate, date)
	default:
		fmt.Fprintf(os.Stderr, "Unknown exchange command: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleInvestments() {
	if err := investment.GenerateInvestmentReport(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handlePool(args []string) {
	subcommand := "list"
	if len(args) > 0 {
		subcommand = args[0]
	}

	switch subcommand {
	case "list":
		if err := pool.ShowPool(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "add":
		if err := pool.AddPoolItem(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "move":
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid pool move <line> <month>\n")
			os.Exit(1)
		}
		if err := pool.MovePoolItem(args[1], args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "remove":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid pool remove <line>\n")
			os.Exit(1)
		}
		if err := pool.RemovePoolItem(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown pool command: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleValidate() {
	if err := validator.ValidateAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleStatus() {
	if err := status.ShowStatus(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleSet(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: spendgrid set config [key] [value]\n")
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "config":
		if len(args) == 1 {
			// Show current config
			showConfig()
		} else if len(args) >= 3 {
			// Set config value
			key := args[1]
			value := strings.Join(args[2:], " ")
			setConfigValue(key, value)
		} else {
			fmt.Fprintf(os.Stderr, "Usage: spendgrid set config <key> <value>\n")
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown set command: %s\n", subcommand)
		os.Exit(1)
	}
}

func showConfig() {
	cfg := config.GetGlobalConfig()
	fmt.Println("Global Configuration:")
	fmt.Printf("  language: %s\n", cfg.Language)
}

func setConfigValue(key, value string) {
	switch key {
	case "language", "lang":
		if err := config.SetLanguage(value); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting language: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Language set to: %s\n", value)
	default:
		fmt.Fprintf(os.Stderr, "Unknown config key: %s\n", key)
		os.Exit(1)
	}
}

func handleQuickInput(input string) {
	// Parse the quick input
	tx, err := parser.QuickInputParser(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing input: %v\n", err)
		fmt.Fprintf(os.Stderr, "Input: %s\n", input)
		fmt.Fprintln(os.Stderr, "\nUsage: spendgrid \"<amount> <description> #tag @project\"")
		fmt.Fprintln(os.Stderr, "Example: spendgrid \"-100TL market alışverişi #mutfak @ev\"")
		os.Exit(1)
	}

	// Add to current month file
	monthFile := parser.GetCurrentMonthFile()
	currentYear := strconv.Itoa(time.Now().Year())
	filePath := filepath.Join(currentYear, monthFile)

	if err := addQuickTransactionToFile(filePath, tx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Auto-save tags and projects
	if err := autoSaveTagsAndProjects(tx.Tags, tx.Projects); err != nil {
		// Non-fatal, just warn
		fmt.Fprintf(os.Stderr, "Warning: could not auto-save tags: %v\n", err)
	}

	// Build display string
	var parts []string
	parts = append(parts, fmt.Sprintf("Added: %s", tx.Description))
	parts = append(parts, fmt.Sprintf("%.2f %s", tx.Amount, tx.Currency))
	if len(tx.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("tags: %s", strings.Join(tx.Tags, ", ")))
	}
	if len(tx.Projects) > 0 {
		parts = append(parts, fmt.Sprintf("project: %s", strings.Join(tx.Projects, ", ")))
	}
	fmt.Println(strings.Join(parts, " | "))
}

func addQuickTransactionToFile(filePath string, tx *parser.Transaction) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the ROWS section
	inRows := false
	insertIndex := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ROWS") {
			inRows = true
			insertIndex = i + 1
			continue
		}
		if strings.HasPrefix(trimmed, "## RULES") && inRows {
			// Insert before RULES
			insertIndex = i
			break
		}
	}

	if insertIndex < 0 {
		return fmt.Errorf("could not find ROWS section")
	}

	formatted := parser.FormatTransaction(tx)

	// Insert the new transaction
	lines = append(lines[:insertIndex], append([]string{formatted}, lines[insertIndex:]...)...)

	if err := os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

func autoSaveTagsAndProjects(tags, projects []string) error {
	// Save tags to categories.yml
	if len(tags) > 0 {
		categoriesPath := filepath.Join("_config", "categories.yml")
		if err := appendToYamlList(categoriesPath, "categories", tags); err != nil {
			return fmt.Errorf("failed to save tags: %v", err)
		}
	}

	// Save projects to projects.yml
	if len(projects) > 0 {
		projectsPath := filepath.Join("_config", "projects.yml")
		if err := appendToYamlList(projectsPath, "projects", projects); err != nil {
			return fmt.Errorf("failed to save projects: %v", err)
		}
	}

	return nil
}

func appendToYamlList(filePath, key string, items []string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Find existing items
	existing := make(map[string]bool)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			item := strings.TrimPrefix(trimmed, "- ")
			existing[item] = true
		}
	}

	// Add new items
	var added bool
	for _, item := range items {
		if !existing[item] {
			lines = append(lines, "- "+item)
			added = true
		}
	}

	if added {
		return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
	}

	return nil
}

func printUsage() {
	fmt.Println(i18n.T("app.description"))
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  spendgrid <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init              Initialize SpendGrid in current directory")
	fmt.Println("  add               Add a new transaction (interactive)")
	fmt.Println("  add --direct      Add a transaction directly: DAY|DESC|AMOUNT|TAGS")
	fmt.Println("  list [month]      List transactions (current month or specify 01-12)")
	fmt.Println("  edit <line>       Edit a transaction by line number")
	fmt.Println("  remove <line>     Remove a transaction by line number")
	fmt.Println("  rules             Manage recurring rules (list, add, edit, toggle, remove)")
	fmt.Println("  sync              Sync rules to month files")
	fmt.Println("  report            Generate monthly report")
	fmt.Println("  report --year     Generate yearly report")
	fmt.Println("  report --web      Generate HTML report")
	fmt.Println("  exchange refresh  Refresh exchange rates from API")
	fmt.Println("  exchange set      Set manual exchange rate")
	fmt.Println("  investments       Show investment portfolio")
	fmt.Println("  pool              Show backlog items")
	fmt.Println("  pool add          Add item to backlog")
	fmt.Println("  pool move         Move item from backlog to month")
	fmt.Println("  pool remove       Remove item from backlog")
	fmt.Println("  validate          Validate all files")
	fmt.Println("  status            Show database status")
	fmt.Println("  set config        View or set global configuration")
	fmt.Println("  version           Show version information")
	fmt.Println("  help              Show this help message")
}
