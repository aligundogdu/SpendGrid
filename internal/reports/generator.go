package reports

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"spendgrid/internal/exchange"
	"spendgrid/internal/i18n"
	"spendgrid/internal/parser"
)

// Renk tanımlamaları
var (
	// Soft renkler (soluk) - parlak versiyonlar
	softGreen = color.New(color.FgHiGreen) // Parlak yeşil
	softRed   = color.New(color.FgHiRed)   // Parlak kırmızı (koyu değil!)

	// Güçlü renkler (kalın ve parlak)
	strongGreen = color.New(color.FgHiGreen, color.Bold)
	strongRed   = color.New(color.FgHiRed, color.Bold)

	// Nötr renkler
	whiteBold = color.New(color.FgWhite, color.Bold)

	// Background renkler (TOTAL satırı için)
	bgWhite  = color.New(color.BgWhite, color.FgBlack)
	bgYellow = color.New(color.BgYellow, color.FgBlack, color.Bold)
)

// MonthlyReport represents a monthly financial report
type MonthlyReport struct {
	Year         int
	Month        int
	Income       map[string]float64            // by currency
	Expenses     map[string]float64            // by currency
	ByCategory   map[string]map[string]float64 // category -> currency -> amount
	ByProject    map[string]map[string]float64 // project -> currency -> amount
	Transactions []*parser.Transaction
}

// YearlyReport represents a yearly financial report
type YearlyReport struct {
	Year          int
	Months        []*MonthlyReport
	TotalIncome   map[string]float64
	TotalExpenses map[string]float64
	NetByMonth    map[int]float64 // month -> net amount in base currency
}

// GenerateMonthlyReport generates a report for the current or specified month
func GenerateMonthlyReport(month int) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	now := time.Now()
	year := now.Year()

	if month == 0 {
		month = int(now.Month())
	}

	// Parse month file
	monthFile := parser.GetMonthFile(month)
	yearDir := strconv.Itoa(year)
	filePath := filepath.Join(yearDir, monthFile)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read month file: %v", err)
	}

	parsed, unparsed := parser.ParseMonthFile(string(content))

	// Generate report
	report := &MonthlyReport{
		Year:         year,
		Month:        month,
		Income:       make(map[string]float64),
		Expenses:     make(map[string]float64),
		ByCategory:   make(map[string]map[string]float64),
		ByProject:    make(map[string]map[string]float64),
		Transactions: parsed,
	}

	// Aggregate data
	for _, tx := range parsed {
		if tx.IsIncome() {
			report.Income[tx.Currency] += tx.Amount
		} else {
			report.Expenses[tx.Currency] += -tx.Amount // Store as positive
		}

		// By category
		for _, tag := range tx.Tags {
			if report.ByCategory[tag] == nil {
				report.ByCategory[tag] = make(map[string]float64)
			}
			if tx.IsIncome() {
				report.ByCategory[tag][tx.Currency] += tx.Amount
			} else {
				report.ByCategory[tag][tx.Currency] += tx.Amount
			}
		}

		// By project
		for _, proj := range tx.Projects {
			if report.ByProject[proj] == nil {
				report.ByProject[proj] = make(map[string]float64)
			}
			if tx.IsIncome() {
				report.ByProject[proj][tx.Currency] += tx.Amount
			} else {
				report.ByProject[proj][tx.Currency] += tx.Amount
			}
		}
	}

	// Print report
	printMonthlyReport(report, unparsed)

	return nil
}

// GenerateYearlyReport generates a report for the entire year
func GenerateYearlyReport() error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	now := time.Now()
	year := now.Year()
	yearDir := strconv.Itoa(year)

	report := &YearlyReport{
		Year:          year,
		Months:        make([]*MonthlyReport, 0),
		TotalIncome:   make(map[string]float64),
		TotalExpenses: make(map[string]float64),
		NetByMonth:    make(map[int]float64),
	}

	// Parse all months
	for month := 1; month <= 12; month++ {
		monthFile := parser.GetMonthFile(month)
		filePath := filepath.Join(yearDir, monthFile)

		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip if file doesn't exist
		}

		parsed, _ := parser.ParseMonthFile(string(content))

		monthly := &MonthlyReport{
			Year:         year,
			Month:        month,
			Income:       make(map[string]float64),
			Expenses:     make(map[string]float64),
			ByCategory:   make(map[string]map[string]float64),
			ByProject:    make(map[string]map[string]float64),
			Transactions: parsed,
		}

		// Aggregate data
		for _, tx := range parsed {
			if tx.IsIncome() {
				monthly.Income[tx.Currency] += tx.Amount
				report.TotalIncome[tx.Currency] += tx.Amount
			} else {
				monthly.Expenses[tx.Currency] += -tx.Amount
				report.TotalExpenses[tx.Currency] += -tx.Amount
			}
		}

		report.Months = append(report.Months, monthly)
	}

	// Print report
	printYearlyReport(report)

	return nil
}

// GenerateHTMLReport generates an HTML report
func GenerateHTMLReport(year bool) error {
	if _, err := os.Stat(".spendgrid"); err != nil {
		return fmt.Errorf("not a spendgrid directory. Run 'spendgrid init' first")
	}

	now := time.Now()
	dateStr := now.Format("2006_01_02")
	filename := fmt.Sprintf("report_%s.html", dateStr)
	filePath := filepath.Join("_share", filename)

	// Generate HTML content
	var html strings.Builder
	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html>\n<head>\n")
	html.WriteString("<meta charset='UTF-8'>\n")
	html.WriteString("<title>SpendGrid Report</title>\n")
	html.WriteString("<style>\n")
	html.WriteString("body { font-family: Arial, sans-serif; margin: 40px; }\n")
	html.WriteString("table { border-collapse: collapse; width: 100%; margin: 20px 0; }\n")
	html.WriteString("th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }\n")
	html.WriteString("th { background-color: #4CAF50; color: white; }\n")
	html.WriteString("tr:nth-child(even) { background-color: #f2f2f2; }\n")
	html.WriteString(".income { color: green; }\n")
	html.WriteString(".expense { color: red; }\n")
	html.WriteString(".summary { font-weight: bold; font-size: 1.2em; margin: 20px 0; }\n")
	html.WriteString("</style>\n")
	html.WriteString("</head>\n<body>\n")
	html.WriteString("<h1>SpendGrid Financial Report</h1>\n")
	html.WriteString(fmt.Sprintf("<p>Generated: %s</p>\n", now.Format("2006-01-02 15:04:05")))

	if year {
		// Yearly report HTML
		yearly := &YearlyReport{
			Year:          now.Year(),
			Months:        make([]*MonthlyReport, 0),
			TotalIncome:   make(map[string]float64),
			TotalExpenses: make(map[string]float64),
		}

		yearDir := strconv.Itoa(now.Year())
		for month := 1; month <= 12; month++ {
			monthFile := parser.GetMonthFile(month)
			filePath := filepath.Join(yearDir, monthFile)

			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			parsed, _ := parser.ParseMonthFile(string(content))
			monthly := &MonthlyReport{
				Year:         now.Year(),
				Month:        month,
				Income:       make(map[string]float64),
				Expenses:     make(map[string]float64),
				Transactions: parsed,
			}

			for _, tx := range parsed {
				if tx.IsIncome() {
					monthly.Income[tx.Currency] += tx.Amount
					yearly.TotalIncome[tx.Currency] += tx.Amount
				} else {
					monthly.Expenses[tx.Currency] += -tx.Amount
					yearly.TotalExpenses[tx.Currency] += -tx.Amount
				}
			}
			yearly.Months = append(yearly.Months, monthly)
		}

		// Monthly summary table
		html.WriteString("<h2>Monthly Summary</h2>\n")
		html.WriteString("<table>\n")
		html.WriteString("<tr><th>Month</th><th>Income</th><th>Expenses</th><th>Net</th></tr>\n")

		for _, m := range yearly.Months {
			totalIncome := 0.0
			totalExpense := 0.0

			for _, amt := range m.Income {
				totalIncome += amt
			}
			for _, amt := range m.Expenses {
				totalExpense += amt
			}

			net := totalIncome - totalExpense
			monthName := time.Month(m.Month).String()

			html.WriteString(fmt.Sprintf("<tr><td>%s</td><td class='income'>%.2f</td><td class='expense'>%.2f</td><td>%.2f</td></tr>\n",
				monthName, totalIncome, totalExpense, net))
		}
		html.WriteString("</table>\n")

		// Yearly totals
		html.WriteString("<div class='summary'>\n")
		html.WriteString("<h2>Yearly Totals</h2>\n")

		totalInc := 0.0
		totalExp := 0.0
		for _, amt := range yearly.TotalIncome {
			totalInc += amt
		}
		for _, amt := range yearly.TotalExpenses {
			totalExp += amt
		}

		html.WriteString(fmt.Sprintf("<p class='income'>Total Income: %.2f</p>\n", totalInc))
		html.WriteString(fmt.Sprintf("<p class='expense'>Total Expenses: %.2f</p>\n", totalExp))
		html.WriteString(fmt.Sprintf("<p>Net: %.2f</p>\n", totalInc-totalExp))
		html.WriteString("</div>\n")
	} else {
		// Monthly report HTML
		month := int(now.Month())
		monthFile := parser.GetMonthFile(month)
		yearDir := strconv.Itoa(now.Year())
		filePath := filepath.Join(yearDir, monthFile)

		content, err := os.ReadFile(filePath)
		if err == nil {
			parsed, _ := parser.ParseMonthFile(string(content))

			html.WriteString(fmt.Sprintf("<h2>Transactions for %s %d</h2>\n", time.Month(month), now.Year()))
			html.WriteString("<table>\n")
			html.WriteString("<tr><th>Day</th><th>Description</th><th>Amount</th><th>Currency</th><th>Tags</th></tr>\n")

			for _, tx := range parsed {
				tags := strings.Join(tx.Tags, ", ")
				class := "expense"
				if tx.IsIncome() {
					class = "income"
				}
				html.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%s</td><td class='%s'>%.2f</td><td>%s</td><td>%s</td></tr>\n",
					tx.Day, tx.Description, class, tx.Amount, tx.Currency, tags))
			}
			html.WriteString("</table>\n")
		}
	}

	html.WriteString("</body>\n</html>")

	// Write to file
	if err := os.WriteFile(filePath, []byte(html.String()), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report: %v", err)
	}

	fmt.Printf("HTML report generated: %s\n", filePath)
	return nil
}

func printMonthlyReport(report *MonthlyReport, unparsed []*parser.Transaction) {
	// Header
	fmt.Printf("\n%s %s %d\n", i18n.T("reports.monthly_title"), time.Month(report.Month), report.Year)
	fmt.Println(strings.Repeat("=", 70))

	// Summary section
	fmt.Printf("\n%s\n", i18n.T("reports.summary"))
	fmt.Println(strings.Repeat("-", 70))

	// Calculate totals in base currency (TRY)
	date := time.Date(report.Year, time.Month(report.Month), 1, 0, 0, 0, 0, time.UTC)
	totalIncome := 0.0
	totalExpense := 0.0

	// Print by currency
	fmt.Printf("%-20s %15s %15s\n", "Currency", "Income", "Expense")
	fmt.Println(strings.Repeat("-", 70))

	allCurrencies := getAllCurrencies(report.Income, report.Expenses)
	for _, curr := range allCurrencies {
		inc := report.Income[curr]
		exp := report.Expenses[curr]

		// Convert to base currency
		incInBase, _ := exchange.ConvertAmount(inc, curr, "TRY", date)
		expInBase, _ := exchange.ConvertAmount(exp, curr, "TRY", date)

		totalIncome += incInBase
		totalExpense += expInBase

		// Renkli yazdırma - Soft renkler
		incomeStr := softGreen.Sprintf("%15.2f", inc)
		expenseStr := softRed.Sprintf("%15.2f", exp)
		fmt.Printf("%-20s %s %s\n", curr, incomeStr, expenseStr)
	}

	fmt.Println(strings.Repeat("-", 70))

	// TOTAL satırı - Soft renklerle
	totalIncomeStr := softGreen.Sprintf("%15.2f", totalIncome)
	totalExpenseStr := softRed.Sprintf("%15.2f", totalExpense)
	whiteBold.Printf("%-20s ", "TOTAL (TRY)")
	fmt.Printf("%s %s\n", totalIncomeStr, totalExpenseStr)

	// NET satırı - Güçlü renklerle (bakiye)
	net := totalIncome - totalExpense
	whiteBold.Printf("%-20s ", "NET")
	if net >= 0 {
		// Artı bakiye -> Güçlü yeşil
		fmt.Printf("%s\n", strongGreen.Sprintf("%15.2f", net))
	} else {
		// Eksi bakiye -> Güçlü kırmızı
		fmt.Printf("%s\n", strongRed.Sprintf("%15.2f", net))
	}

	// By category
	if len(report.ByCategory) > 0 {
		fmt.Printf("\n%s\n", i18n.T("reports.by_category"))
		fmt.Println(strings.Repeat("-", 70))

		categories := getSortedKeys(report.ByCategory)
		for _, cat := range categories {
			fmt.Printf("#%-19s", cat)
			for curr, amt := range report.ByCategory[cat] {
				fmt.Printf(" %10.2f %s", amt, curr)
			}
			fmt.Println()
		}
	}

	// Unparsed transactions
	if len(unparsed) > 0 {
		fmt.Printf("\n%s\n", i18n.T("transaction.unparsed_header"))
		fmt.Println(strings.Repeat("-", 70))
		for _, tx := range unparsed {
			fmt.Printf("Line %d: %s\n", tx.LineNumber, tx.Raw)
		}
	}

	fmt.Println()
}

func printYearlyReport(report *YearlyReport) {
	// Header
	fmt.Printf("\n%s %d\n", i18n.T("reports.yearly_title"), report.Year)
	fmt.Println(strings.Repeat("=", 70))

	// Monthly breakdown
	fmt.Printf("\n%s\n", i18n.T("reports.monthly_breakdown"))
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("%-10s %15s %15s %15s\n", "Month", "Income", "Expense", "Net")
	fmt.Println(strings.Repeat("-", 70))

	for _, m := range report.Months {
		totalIncome := 0.0
		totalExpense := 0.0

		for _, amt := range m.Income {
			totalIncome += amt
		}
		for _, amt := range m.Expenses {
			totalExpense += amt
		}

		net := totalIncome - totalExpense

		// Soft renklerle income ve expense
		incomeStr := softGreen.Sprintf("%15.2f", totalIncome)
		expenseStr := softRed.Sprintf("%15.2f", totalExpense)

		// Net için güçlü renkler
		var netStr string
		if net >= 0 {
			netStr = strongGreen.Sprintf("%15.2f", net)
		} else {
			netStr = strongRed.Sprintf("%15.2f", net)
		}

		fmt.Printf("%-10s %s %s %s\n",
			time.Month(m.Month).String(), incomeStr, expenseStr, netStr)
	}

	// Yearly totals - Background ile vurgulanmış
	fmt.Println(strings.Repeat("-", 70))

	totalInc := 0.0
	totalExp := 0.0
	for _, amt := range report.TotalIncome {
		totalInc += amt
	}
	for _, amt := range report.TotalExpenses {
		totalExp += amt
	}

	// TOTAL satırı - Sarı background ile vurgulu + renkli rakamlar
	netTotal := totalInc - totalExp

	// Background başlat
	bgYellow.Printf("%-10s", "TOTAL")

	// Renkli rakamlar (background'ın üzerinde)
	fmt.Printf(" %s", softGreen.Sprintf("%15.2f", totalInc))
	fmt.Printf(" %s", softRed.Sprintf("%15.2f", totalExp))

	// Net total (güçlü renklerle)
	if netTotal >= 0 {
		fmt.Printf(" %s", strongGreen.Sprintf("%15.2f", netTotal))
	} else {
		fmt.Printf(" %s", strongRed.Sprintf("%15.2f", netTotal))
	}
	fmt.Println() // Satır sonu

	// By currency summary
	fmt.Printf("\n%s\n", i18n.T("reports.by_currency"))
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("%-15s %15s %15s\n", "Currency", "Income", "Expense")
	fmt.Println(strings.Repeat("-", 70))

	allCurrencies := getAllCurrencies(report.TotalIncome, report.TotalExpenses)
	for _, curr := range allCurrencies {
		inc := report.TotalIncome[curr]
		exp := report.TotalExpenses[curr]

		incomeStr := softGreen.Sprintf("%15.2f", inc)
		expenseStr := softRed.Sprintf("%15.2f", exp)

		fmt.Printf("%-15s %s %s\n", curr, incomeStr, expenseStr)
	}

	fmt.Println()
}

func getAllCurrencies(income, expense map[string]float64) []string {
	currencySet := make(map[string]bool)
	for curr := range income {
		currencySet[curr] = true
	}
	for curr := range expense {
		currencySet[curr] = true
	}

	currencies := make([]string, 0, len(currencySet))
	for curr := range currencySet {
		currencies = append(currencies, curr)
	}
	sort.Strings(currencies)
	return currencies
}

func getSortedKeys(m map[string]map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
