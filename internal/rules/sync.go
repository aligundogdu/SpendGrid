package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SyncResult holds the results of a sync operation
type SyncResult struct {
	Added   int
	Updated int
	Skipped int
	Errors  []string
}

// SyncRules syncs rules to month files for the current and future months
func SyncRules() (*SyncResult, error) {
	result := &SyncResult{
		Errors: []string{},
	}

	// Load all active rules
	rules, err := GetActiveRules()
	if err != nil {
		return nil, fmt.Errorf("failed to load rules: %v", err)
	}

	if len(rules) == 0 {
		return result, nil
	}

	// Get current date
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// Sync for current year (current month and future months)
	for month := currentMonth; month <= 12; month++ {
		r, err := syncMonth(currentYear, month, rules)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%04d-%02d: %v", currentYear, month, err))
		} else {
			result.Added += r.Added
			result.Updated += r.Updated
			result.Skipped += r.Skipped
		}
	}

	return result, nil
}

// syncMonth syncs rules to a specific month file
func syncMonth(year, month int, rules []Rule) (*SyncResult, error) {
	result := &SyncResult{}

	// Build file path
	monthFile := fmt.Sprintf("%02d.md", month)
	yearDir := strconv.Itoa(year)
	filePath := filepath.Join(yearDir, monthFile)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create year directory if needed
		if err := os.MkdirAll(yearDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create year directory: %v", err)
		}
		// Create month file with default structure
		content := fmt.Sprintf("# %d %s\n\n## ROWS\n\n## RULES\n",
			year, getMonthName(month))
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to create month file: %v", err)
		}
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read month file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find RULES section
	rulesStartIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## RULES" {
			rulesStartIdx = i
			break
		}
	}

	if rulesStartIdx == -1 {
		// Add RULES section at end
		lines = append(lines, "", "## RULES")
		rulesStartIdx = len(lines) - 1
	}

	// Track which rules are already in the file
	existingRules := make(map[string]bool)
	existingLines := make(map[string]int) // rule ID -> line index

	// Parse existing rule lines
	re := regexp.MustCompile(`- \[([ x])\] \d+ \| .+`) // Match rule lines
	for i := rulesStartIdx + 1; i < len(lines); i++ {
		line := lines[i]
		if re.MatchString(line) {
			// Check if this line is checked [x]
			if strings.Contains(line, "- [x]") {
				// This is a user-modified line, extract rule info if possible
				// For now, just mark as existing
				existingRules[line] = true
				existingLines[line] = i
			}
		}
	}

	// Process each rule
	for _, rule := range rules {
		if !rule.ShouldApplyInMonth(year, month) {
			continue
		}

		// Generate the rule line
		scheduledDay := rule.GetScheduledDay(year, month)
		ruleLine := formatRuleLine(&rule, scheduledDay)

		// Check if this rule already exists in file
		// We need to match by rule ID embedded in the line
		found := false
		for existingLine := range existingRules {
			if isSameRule(existingLine, ruleLine) {
				found = true
				// Check if it's checked [x]
				if strings.Contains(existingLine, "- [x]") {
					// User has modified this, don't touch it
					result.Skipped++
				} else {
					// Update if needed
					if existingLine != ruleLine {
						// Update the line
						idx := existingLines[existingLine]
						lines[idx] = ruleLine
						result.Updated++
					}
				}
				break
			}
		}

		if !found {
			// Add new rule line
			lines = append(lines[:rulesStartIdx+1], append([]string{ruleLine}, lines[rulesStartIdx+1:]...)...)
			result.Added++
			// Update index since we added a line
			rulesStartIdx++
		}
	}

	// Write back to file
	if err := os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return nil, fmt.Errorf("failed to write month file: %v", err)
	}

	return result, nil
}

// formatRuleLine formats a rule as a transaction line
func formatRuleLine(rule *Rule, day int) string {
	sign := ""
	if rule.Type == "expense" || rule.Amount < 0 {
		sign = "-"
	}

	// Build tags
	tags := ""
	for _, t := range rule.Tags {
		tags += " #" + t
	}
	if rule.Project != "" {
		tags += " @" + rule.Project
	}

	return fmt.Sprintf("- [ ] %02d | %s | %s%.2f %s |%s",
		day,
		rule.Name,
		sign,
		abs(rule.Amount),
		rule.Currency,
		tags)
}

// isSameRule checks if two rule lines represent the same rule
// This is a simple check based on name and day
func isSameRule(existing, new string) bool {
	// Extract name from both lines
	// Format: - [ ] DAY | NAME | AMOUNT CURRENCY | TAGS

	existingParts := strings.Split(existing, "|")
	newParts := strings.Split(new, "|")

	if len(existingParts) < 2 || len(newParts) < 2 {
		return false
	}

	// Compare names (trim spaces)
	existingName := strings.TrimSpace(existingParts[1])
	newName := strings.TrimSpace(newParts[1])

	return existingName == newName
}

func getMonthName(month int) string {
	months := []string{
		"", "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
		"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık",
	}
	if month >= 1 && month <= 12 {
		return months[month]
	}
	return ""
}
