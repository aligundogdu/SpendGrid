package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Rule represents a recurring transaction rule
type Rule struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	Amount   float64  `yaml:"amount"`
	Currency string   `yaml:"currency"`
	Type     string   `yaml:"type"` // income or expense
	Category string   `yaml:"category"`
	Tags     []string `yaml:"tags"`
	Project  string   `yaml:"project,omitempty"`
	Schedule Schedule `yaml:"schedule"`
	Active   bool     `yaml:"active"`
}

// Schedule defines when the rule should be applied
type Schedule struct {
	Frequency string `yaml:"frequency"` // monthly, weekly, yearly
	Day       int    `yaml:"day"`       // Day of month (1-31) or day of week (1-7)
}

// RuleSet holds all rules
type RuleSet struct {
	Rules []Rule `yaml:"rules"`
}

// GetRulesFilePath returns the path to rules.yml
func GetRulesFilePath() string {
	return filepath.Join("_config", "rules.yml")
}

// LoadRules loads all rules from rules.yml
func LoadRules() (*RuleSet, error) {
	filePath := GetRulesFilePath()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty rule set
			return &RuleSet{Rules: []Rule{}}, nil
		}
		return nil, fmt.Errorf("failed to read rules file: %v", err)
	}

	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return nil, fmt.Errorf("failed to parse rules: %v", err)
	}

	return &ruleSet, nil
}

// SaveRules saves rules to rules.yml
func SaveRules(ruleSet *RuleSet) error {
	filePath := GetRulesFilePath()

	data, err := yaml.Marshal(ruleSet)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %v", err)
	}

	// Add header comment
	header := "# SpendGrid Rules\n# Otomatik oluşturulacak düzenli gelir/gider kuralları\n\n"
	data = append([]byte(header), data...)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write rules file: %v", err)
	}

	return nil
}

// AddRule adds a new rule
func AddRule(rule Rule) error {
	ruleSet, err := LoadRules()
	if err != nil {
		return err
	}

	// Check for duplicate ID
	for _, r := range ruleSet.Rules {
		if r.ID == rule.ID {
			return fmt.Errorf("rule with ID '%s' already exists", rule.ID)
		}
	}

	ruleSet.Rules = append(ruleSet.Rules, rule)
	return SaveRules(ruleSet)
}

// GetRule gets a rule by ID
func GetRule(id string) (*Rule, error) {
	ruleSet, err := LoadRules()
	if err != nil {
		return nil, err
	}

	for i := range ruleSet.Rules {
		if ruleSet.Rules[i].ID == id {
			return &ruleSet.Rules[i], nil
		}
	}

	return nil, fmt.Errorf("rule with ID '%s' not found", id)
}

// UpdateRule updates an existing rule
func UpdateRule(id string, updated Rule) error {
	ruleSet, err := LoadRules()
	if err != nil {
		return err
	}

	found := false
	for i := range ruleSet.Rules {
		if ruleSet.Rules[i].ID == id {
			ruleSet.Rules[i] = updated
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("rule with ID '%s' not found", id)
	}

	return SaveRules(ruleSet)
}

// DeleteRule removes a rule by ID
func DeleteRule(id string) error {
	ruleSet, err := LoadRules()
	if err != nil {
		return err
	}

	found := false
	newRules := make([]Rule, 0, len(ruleSet.Rules))
	for _, r := range ruleSet.Rules {
		if r.ID != id {
			newRules = append(newRules, r)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("rule with ID '%s' not found", id)
	}

	ruleSet.Rules = newRules
	return SaveRules(ruleSet)
}

// ToggleRule toggles a rule's active status
func ToggleRule(id string) error {
	ruleSet, err := LoadRules()
	if err != nil {
		return err
	}

	for i := range ruleSet.Rules {
		if ruleSet.Rules[i].ID == id {
			ruleSet.Rules[i].Active = !ruleSet.Rules[i].Active
			return SaveRules(ruleSet)
		}
	}

	return fmt.Errorf("rule with ID '%s' not found", id)
}

// GetActiveRules returns all active rules
func GetActiveRules() ([]Rule, error) {
	ruleSet, err := LoadRules()
	if err != nil {
		return nil, err
	}

	var active []Rule
	for _, r := range ruleSet.Rules {
		if r.Active {
			active = append(active, r)
		}
	}

	return active, nil
}

// FormatRuleAsTransaction formats a rule as a transaction line for a specific month/year
func FormatRuleAsTransaction(rule Rule, year, month int) string {
	// Determine sign based on type
	sign := ""
	if rule.Type == "expense" || rule.Amount < 0 {
		sign = "-"
	}

	// Build tags string
	tags := ""
	for _, t := range rule.Tags {
		tags += " #" + t
	}
	if rule.Project != "" {
		tags += " @" + rule.Project
	}

	// Format: - DAY | NAME | AMOUNT CURRENCY | TAGS
	return fmt.Sprintf("- [ ] %02d | %s | %s%.2f %s |%s",
		rule.Schedule.Day,
		rule.Name,
		sign,
		abs(rule.Amount),
		rule.Currency,
		tags)
}

// ShouldApplyInMonth checks if a rule should be applied in a given month/year
func (r *Rule) ShouldApplyInMonth(year, month int) bool {
	if !r.Active {
		return false
	}

	// For now, only support monthly frequency
	if r.Schedule.Frequency != "monthly" && r.Schedule.Frequency != "" {
		return false
	}

	// Check if the day is valid for this month
	day := r.Schedule.Day
	if day < 1 {
		day = 1
	}

	// Get last day of month
	lastDay := getLastDayOfMonth(year, month)
	if day > lastDay {
		day = lastDay
	}

	return true
}

// GetScheduledDay returns the actual day for this rule in a given month
func (r *Rule) GetScheduledDay(year, month int) int {
	day := r.Schedule.Day
	if day < 1 {
		return 1
	}

	lastDay := getLastDayOfMonth(year, month)
	if day > lastDay {
		return lastDay
	}

	return day
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func getLastDayOfMonth(year, month int) int {
	// Get first day of next month and subtract one day
	if month == 12 {
		year++
		month = 1
	} else {
		month++
	}

	t := time.Date(year, time.Month(month), 0, 0, 0, 0, 0, time.UTC)
	return t.Day()
}

// GenerateRuleID generates a unique rule ID from name
func GenerateRuleID(name string) string {
	// Simple ID generation: lowercase, replace spaces with underscore
	id := ""
	for _, c := range name {
		if c >= 'A' && c <= 'Z' {
			id += string(c + 32) // to lowercase
		} else if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			id += string(c)
		} else if c == ' ' {
			id += "_"
		}
	}

	// Add timestamp to ensure uniqueness
	id += "_" + strconv.FormatInt(time.Now().Unix(), 10)
	return id
}
