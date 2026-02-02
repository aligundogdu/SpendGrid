package rules

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"spendgrid/internal/i18n"
)

// ListRules displays all rules
func ListRules() error {
	ruleSet, err := LoadRules()
	if err != nil {
		return err
	}

	if len(ruleSet.Rules) == 0 {
		fmt.Println(i18n.T("rules.no_rules"))
		return nil
	}

	fmt.Println(i18n.T("rules.header"))
	fmt.Println(strings.Repeat("=", 80))

	for _, r := range ruleSet.Rules {
		status := "✓"
		if !r.Active {
			status = "✗"
		}

		typeStr := "INC"
		if r.Type == "expense" {
			typeStr = "EXP"
		}

		fmt.Printf("%s [%s] %s | %s | %.2f %s | Monthly day %d\n",
			status,
			r.ID,
			typeStr,
			r.Name,
			r.Amount,
			r.Currency,
			r.Schedule.Day)
	}

	return nil
}

// AddRuleInteractive adds a new rule interactively
func AddRuleInteractive() error {
	reader := bufio.NewReader(os.Stdin)

	// Get rule name
	fmt.Print(i18n.T("rules.name_prompt") + " ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Generate ID
	id := GenerateRuleID(name)

	// Get type
	fmt.Print(i18n.T("rules.type_prompt") + " [income/expense]: ")
	ruleType, _ := reader.ReadString('\n')
	ruleType = strings.TrimSpace(strings.ToLower(ruleType))
	if ruleType != "income" && ruleType != "expense" {
		ruleType = "expense" // default
	}

	// Get amount
	fmt.Print(i18n.T("rules.amount_prompt") + " ")
	amountStr, _ := reader.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// Get currency
	fmt.Print(i18n.T("rules.currency_prompt") + " [TRY]: ")
	currency, _ := reader.ReadString('\n')
	currency = strings.TrimSpace(strings.ToUpper(currency))
	if currency == "" {
		currency = "TRY"
	}

	// Get schedule day
	fmt.Print(i18n.T("rules.day_prompt") + " [1]: ")
	dayStr, _ := reader.ReadString('\n')
	dayStr = strings.TrimSpace(dayStr)
	day := 1
	if dayStr != "" {
		d, err := strconv.Atoi(dayStr)
		if err == nil && d >= 1 && d <= 31 {
			day = d
		}
	}

	// Get tags
	fmt.Print(i18n.T("rules.tags_prompt") + " ")
	tagsInput, _ := reader.ReadString('\n')
	tagsInput = strings.TrimSpace(tagsInput)
	tags := parseTags(tagsInput)

	// Get project (optional)
	fmt.Print(i18n.T("rules.project_prompt") + " ")
	project, _ := reader.ReadString('\n')
	project = strings.TrimSpace(project)

	rule := Rule{
		ID:       id,
		Name:     name,
		Amount:   amount,
		Currency: currency,
		Type:     ruleType,
		Tags:     tags,
		Project:  project,
		Schedule: Schedule{
			Frequency: "monthly",
			Day:       day,
		},
		Active: true,
	}

	if err := AddRule(rule); err != nil {
		return err
	}

	fmt.Println(i18n.T("rules.add_success"))
	return nil
}

// EditRuleInteractive edits a rule interactively
func EditRuleInteractive(id string) error {
	rule, err := GetRule(id)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Editing rule: %s\n", rule.Name)
	fmt.Println("Press Enter to keep current value, or enter new value:")

	// Name
	fmt.Printf("Name [%s]: ", rule.Name)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name != "" {
		rule.Name = name
	}

	// Type
	fmt.Printf("Type [%s]: ", rule.Type)
	ruleType, _ := reader.ReadString('\n')
	ruleType = strings.TrimSpace(strings.ToLower(ruleType))
	if ruleType == "income" || ruleType == "expense" {
		rule.Type = ruleType
	}

	// Amount
	fmt.Printf("Amount [%.2f]: ", rule.Amount)
	amountStr, _ := reader.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	if amountStr != "" {
		if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
			rule.Amount = amount
		}
	}

	// Currency
	fmt.Printf("Currency [%s]: ", rule.Currency)
	currency, _ := reader.ReadString('\n')
	currency = strings.TrimSpace(strings.ToUpper(currency))
	if currency != "" {
		rule.Currency = currency
	}

	// Schedule day
	fmt.Printf("Day of month [%d]: ", rule.Schedule.Day)
	dayStr, _ := reader.ReadString('\n')
	dayStr = strings.TrimSpace(dayStr)
	if dayStr != "" {
		if day, err := strconv.Atoi(dayStr); err == nil && day >= 1 && day <= 31 {
			rule.Schedule.Day = day
		}
	}

	// Tags
	fmt.Printf("Tags [%s]: ", strings.Join(rule.Tags, " "))
	tagsInput, _ := reader.ReadString('\n')
	tagsInput = strings.TrimSpace(tagsInput)
	if tagsInput != "" {
		rule.Tags = parseTags(tagsInput)
	}

	if err := UpdateRule(id, *rule); err != nil {
		return err
	}

	fmt.Println(i18n.T("rules.edit_success"))
	return nil
}

// ToggleRuleStatus toggles a rule's active status
func ToggleRuleStatus(id string) error {
	rule, err := GetRule(id)
	if err != nil {
		return err
	}

	if err := ToggleRule(id); err != nil {
		return err
	}

	status := "activated"
	if rule.Active {
		status = "deactivated"
	}

	fmt.Printf("Rule '%s' %s\n", rule.Name, status)
	return nil
}

// RemoveRule removes a rule
func RemoveRule(id string) error {
	rule, err := GetRule(id)
	if err != nil {
		return err
	}

	fmt.Printf("Remove rule '%s'? [y/n]: ", rule.Name)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Cancelled")
		return nil
	}

	if err := DeleteRule(id); err != nil {
		return err
	}

	fmt.Println(i18n.T("rules.remove_success"))
	return nil
}

// SyncNow performs a manual sync
func SyncNow() error {
	fmt.Println(i18n.T("rules.sync_start"))

	result, err := SyncRules()
	if err != nil {
		return fmt.Errorf("sync failed: %v", err)
	}

	fmt.Printf(i18n.T("rules.sync_complete"), result.Added, result.Updated, result.Skipped)
	fmt.Println()

	if len(result.Errors) > 0 {
		fmt.Println(i18n.T("rules.sync_errors"))
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	return nil
}

func parseTags(input string) []string {
	var tags []string
	words := strings.Fields(input)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			tags = append(tags, strings.TrimPrefix(word, "#"))
		} else {
			tags = append(tags, word)
		}
	}
	return tags
}
