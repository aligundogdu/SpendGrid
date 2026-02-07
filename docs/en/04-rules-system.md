# SpendGrid Rule System - Complete Guide

## Table of Contents
1. [What is the Rule System?](#what-is-the-rule-system)
2. [Basic Concepts](#basic-concepts)
3. [Rule Structure](#rule-structure)
4. [Synchronization Mechanism](#synchronization-mechanism)
5. [Completion System](#completion-system)
6. [Scenarios and Examples](#scenarios-and-examples)
7. [Advanced Usage](#advanced-usage)
8. [Troubleshooting](#troubleshooting)

---

## What is the Rule System?

SpendGrid's rule system allows you to automatically track recurring financial transactions. Define your regular income and expenses (like monthly rent, salary, bill payments) once, and the system automatically adds them to your month files.

### Why Rule System?

**Traditional Method:**
```bash
# Manually add every month
spendgrid add
# 05 | Rent | -5000 TRY | #rent @home
# 15 | Electricity | -300 TRY | #bill
# 20 | Salary | +25000 TRY | #salary @company
```

**With Rule System:**
```bash
# Define once
spendgrid rules add "Monthly Rent" 5000 TRY expense --day 5 --tags "rent" --project "home"

# Auto-syncs every month
# Mark as completed when done
spendgrid complete rent_xxx
```

---

## Basic Concepts

### 1. Rule
A rule represents a recurring financial transaction on a specific date. Rules are stored in `_config/rules.yml`.

### 2. Synchronization
The process of automatically copying rules to month files (`01.md`, `02.md`, etc.). Happens automatically on every SpendGrid command.

### 3. Checkbox Status
Rules are added to month files in two states:
- `[ ]` - Planned, not yet occurred
- `[x]` - Completed, occurred

### 4. Complete/Uncomplete
The process of changing a rule's checkbox status. Only completed rules are included in reports.

---

## Rule Structure

### YAML Format

```yaml
rules:
  - id: sal_1770358056
    name: Monthly Net Salary
    amount: 25000
    currency: TRY
    type: income
    category: income
    tags:
      - salary
      - net
    project: '@companyA'
    schedule:
      frequency: monthly
      day: 5
    active: true
    start_date: "2026-01"
    end_date: "2026-12"
    metadata: "Deposited on 5th of every month"
```

### Field Descriptions

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Auto | Unique identifier (auto-generated) |
| `name` | Yes | Rule name (descriptive) |
| `amount` | Yes | Amount (positive number) |
| `currency` | Yes | Currency (TRY, USD, EUR) |
| `type` | Yes | `income` or `expense` |
| `tags` | No | List of tags |
| `project` | No | Project name (starts with @) |
| `schedule.frequency` | Yes | `monthly`, `weekly`, `yearly` |
| `schedule.day` | Yes | Day of month (1-31) |
| `active` | No | Active/inactive status (default: true) |
| `start_date` | No | Start date (YYYY-MM) |
| `end_date` | No | End date (YYYY-MM) |
| `metadata` | No | Description/note |

---

## Synchronization Mechanism

### How It Works

1. **After Every Command:** Auto-sync happens when `spendgrid` runs
2. **Month File Check:** Checks current month file (`02.md`, etc.)
3. **Missing Rules:** Adds active rules from `_config/rules.yml`
4. **Checkbox Format:** Rules added as `- [ ] DAY | DESC [ID] | AMOUNT CURR | #tags`

### Example Synchronization

**rules.yml:**
```yaml
rules:
  - id: rent_001
    name: Home Rent
    amount: 5000
    currency: TRY
    type: expense
    schedule:
      frequency: monthly
      day: 5
    tags: [rent, home]
```

**After sync in 02.md:**
```markdown
## ROWS
- 01 | Grocery | -250 TRY | #grocery

## RULES
- [ ] 05 | Home Rent [rent_001] | -5000.00 TRY | #rent #home
```

### Important Notes

- ✅ Existing rules are **NOT overwritten**
- ✅ Manual edits are preserved
- ✅ Only `[ ]` (unchecked) rules are synced
- ✅ `[x]` (checked) rules are left untouched

---

## Completion System

### Philosophy: Planning vs Reality

SpendGrid's rule system has two stages:

1. **Planning Stage:** Rule is synced, marked as `[ ]`
2. **Reality Stage:** When money arrives/leaves, marked as `[x]`

### Why This System?

**Problem:**
```
Expected salary on 5th: 25,000 TL
But it's 3rd and money hasn't arrived
Report shows: "Income: 25,000 TL" - WRONG!
```

**Solution:**
```
Planned: +25,000 TL (hasn't arrived yet)
Actual: 0 TL
Projection: +25,000 TL (expected)

After money arrives:
Actual: +25,000 TL
```

### Commands

#### complete - Mark Rule as Complete
```bash
# Interactive mode (recommended)
spendgrid complete
# Shows list, enter number or ID

# Direct with ID
spendgrid complete sal_1770358056

# Complete entire month
spendgrid complete-month 2026-02
```

#### uncomplete - Undo Completion
```bash
# Interactive mode
spendgrid uncomplete

# Direct with ID
spendgrid uncomplete sal_1770358056
```

### Three-Section Report

Reports now show three sections:

```
📊 ACTUAL (Completed Transactions)
   Income: 15,000 TL
   Expense: 8,000 TL
   Net: +7,000 TL

📅 PLANNED (Uncompleted Rules)
   Income: +25,000 TL (Salary)
   Expense: -5,000 TL (Rent)

🔮 PROJECTION (Actual + Planned)
   Expected Net: +27,000 TL
```

---

## Scenarios and Examples

### Salary Planning Scenarios

#### Scenario 1: Standard Monthly Salary

**Situation:** Receive 25,000 TL net salary on 5th of every month.

```bash
# Create rule
spendgrid rules add "Monthly Net Salary" 25000 TRY income \
  --day 5 \
  --tags "salary,net" \
  --project "@companyA"

# After sync, added to month file
# - [ ] 05 | Monthly Net Salary [sal_xxx] | +25000.00 TRY | #salary #net @@companyA

# Mark complete when salary arrives
spendgrid complete sal_xxx
# or
spendgrid complete
# Enter 1 (if listed as #1)
```

#### Scenario 2: Two Separate Salary Payments

**Situation:** Main salary on 5th (25,000 TL), additional payment on 20th (5,000 TL)

```bash
# Main salary
spendgrid rules add "Main Salary" 25000 TRY income --day 5 --tags "salary,main"

# Additional payment
spendgrid rules add "Additional Payment" 5000 TRY income --day 20 --tags "salary,extra"

# Report view:
# Planned Income: +30,000 TL
#   - 05 | Main Salary: +25,000 TL
#   - 20 | Additional Payment: +5,000 TL
```

#### Scenario 3: Foreign Currency Salary (USD)

**Situation:** Freelance work, receive 1,000 USD on 15th of every month

```bash
# USD rule
spendgrid rules add "Freelance Payment" 1000 USD income \
  --day 15 \
  --tags "freelance,usd" \
  --project "@clientX"

# Added to month file:
# - [ ] 15 | Freelance Payment [fre_xxx] | +1000.00 USD | #freelance #usd @@clientX

# When payment arrives
spendgrid complete fre_xxx

# Report does automatic conversion (e.g., 1 USD = 35 TL)
# Income: +35,000 TRY (1000 USD @ 35.00)
```

---

### Loan Payment Scenarios

#### Scenario 1: Housing Loan (Fixed Installment)

**Situation:** 4,500 TL housing loan payment on 10th of every month

```bash
# Loan rule
spendgrid rules add "Housing Loan" 4500 TRY expense \
  --day 10 \
  --tags "loan,housing,bank" \
  --project "@ziraat"

# Month file:
# - [ ] 10 | Housing Loan [loa_xxx] | -4500.00 TRY | #loan #housing #bank @@ziraat

# When payment is deducted
spendgrid complete loa_xxx
```

#### Scenario 2: Personal Loan (Monthly Tracking)

**Situation:** 2,800 TL personal loan on 1st of every month with extra info

```yaml
# In rules.yml:
rules:
  - id: loa_personal_001
    name: Personal Loan Installment
    amount: 2800
    currency: TRY
    type: expense
    tags: [loan, personal, akbank]
    project: '@akbank'
    schedule:
      frequency: monthly
      day: 1
    metadata: "24 month installment, Remaining: 18 months"
```

```bash
# After sync
# - [ ] 01 | Personal Loan Installment [loa_personal_001] | -2800.00 TRY | #loan #personal #akbank @@akbank

# Update metadata each month
spendgrid rules edit loa_personal_001
# Metadata: "24 month installment, Remaining: 17 months"
```

#### Scenario 3: Credit Card Payment (Minimum + Extra)

**Situation:** Minimum payment 1,500 TL on 5th, but planning full payment

```bash
# Minimum payment rule (fixed)
spendgrid rules add "CC Minimum Payment" 1500 TRY expense --day 5 --tags "credit,card,minimum"

# Extra payment during month (manual add)
spendgrid add
# 15 | CC Extra Payment | -3000 TRY | #credit #card #extra
```

---

### Monthly Expense and Spending Rules

#### 1. Rent Payment
```bash
spendgrid rules add "Home Rent" 5000 TRY expense --day 5 --tags "rent,home,housing"
```

#### 2. Electricity Bill
```bash
spendgrid rules add "Electricity Bill" 350 TRY expense --day 15 --tags "bill,electricity"
```

#### 3. Natural Gas Bill
```bash
spendgrid rules add "Natural Gas Bill" 450 TRY expense --day 15 --tags "bill,gas"
```

#### 4. Water Bill
```bash
spendgrid rules add "Water Bill" 150 TRY expense --day 20 --tags "bill,water"
```

#### 5. Internet Fee
```bash
spendgrid rules add "Internet Fee" 120 TRY expense --day 1 --tags "bill,internet"
```

#### 6. Phone Bill
```bash
spendgrid rules add "Phone Bill" 250 TRY expense --day 5 --tags "bill,phone"
```

#### 7. Gym Membership
```bash
spendgrid rules add "Gym Membership" 300 TRY expense --day 1 --tags "gym,membership"
```

#### 8. Netflix Subscription
```bash
spendgrid rules add "Netflix" 50 TRY expense --day 15 --tags "subscription,digital"
```

#### 9. Spotify Subscription
```bash
spendgrid rules add "Spotify" 35 TRY expense --day 20 --tags "subscription,music"
```

#### 10. Monthly Investment (Auto)
```bash
spendgrid rules add "Retirement Plan" 1000 TRY expense --day 10 --tags "investment,retirement,pension"
```

**All Expenses Report View:**
```
📅 PLANNED Expenses:
   - 01 | Internet: -120 TRY
   - 01 | Gym: -300 TRY
   - 05 | Home Rent: -5000 TRY
   - 05 | Phone: -250 TRY
   - 10 | Retirement: -1000 TRY
   - 15 | Electricity: -350 TRY
   - 15 | Gas: -450 TRY
   - 15 | Netflix: -50 TRY
   - 20 | Water: -150 TRY
   - 20 | Spotify: -35 TRY
   
   Total Planned Expense: -7,705 TRY
```

---

### Planning and Completion Scenarios

#### Scenario 1: Daily Tracking (Recommended)

**Day 1 - Start of Month:**
```bash
spendgrid status
# Shows 3 planned rules

spendgrid report monthly
# 📅 PLANNED: +25,000 TL (Salary)
# 📅 PLANNED: -5,000 TL (Rent)
```

**Day 5 - Salary Day:**
```bash
# Salary deposited, check
spendgrid complete
# Select 1 from list (Salary)

spendgrid report monthly
# 📊 ACTUAL: +25,000 TL
# 📅 PLANNED: -5,000 TL (Rent pending)
# 🔮 PROJECTION: +20,000 TL
```

**Day 5 - Rent Payment:**
```bash
# Rent paid
spendgrid complete
# Select Rent from list

spendgrid report monthly
# 📊 ACTUAL: +25,000 TL / -5,000 TL
# Net: +20,000 TL
```

#### Scenario 2: Weekly Batch Completion

```bash
# Every Saturday - weekly check
spendgrid complete
# Mark completed items

# Or all at once (use carefully!)
spendgrid complete-month
# Completes all rules for the month
```

#### Scenario 3: Manual Verification

```bash
# Check bank account
# Received: 25,000 TL (Salary)

# Verify in SpendGrid
spendgrid complete sal_001

# Check report again
spendgrid report monthly
# Should show salary in ACTUAL section
```

#### Scenario 4: Wrong Completion Correction

```bash
# Accidentally marked rent as complete but not paid yet
spendgrid uncomplete rent_001

# Recalculates report
# Rent moves back to PLANNED
```

#### Scenario 5: Partial Completion (Salary Delay)

```bash
# Salary should arrive on 5th but didn't
spendgrid report monthly
# Still shows salary in PLANNED

# Arrived on 7th
spendgrid complete sal_001

# Moved to ACTUAL
```

#### Scenario 6: Bill Amount Change

```bash
# Electricity varies each month
# Rule: 350 TRY (average)

# On 15th actual bill: 420 TRY
# Method 1: Complete rule + manual add
spendgrid complete ele_001
spendgrid add
# 15 | Electricity Bill Difference | -70 TRY | #bill #electricity

# Method 2: Direct manual add (don't delete rule)
spendgrid add
# 15 | Electricity Bill (Actual) | -420 TRY | #bill #electricity
```

#### Scenario 7: Monthly Summary and Batch Processing

```bash
# End of month check
spendgrid report monthly

# Check missing completions
spendgrid complete

# All completed?
spendgrid status
# "All rules completed" message

# Prepare for next month
spendgrid rules list
# Check inactive ones
```

---

## Advanced Usage

### 1. Date Range Rules

```yaml
rules:
  - id: intern_salary
    name: Intern Salary
    amount: 5000
    currency: TRY
    type: income
    schedule:
      frequency: monthly
      day: 5
    start_date: "2026-06"  # Start in June
    end_date: "2026-08"    # End in August
```

### 2. Project-Based Tracking

```bash
# Same project income/expense
spendgrid rules add "Project X Payment" 10000 TRY income --day 10 --project "@projectX"
spendgrid rules add "Project X Cost" 2000 TRY expense --day 15 --project "@projectX"

# Report by project
```

### 3. Multiple Currencies

```bash
# USD income
spendgrid rules add "Freelance USD" 1000 USD income --day 15

# EUR expense
spendgrid rules add "Hosting EUR" 50 EUR expense --day 1

# Report does automatic conversion
```

---

## Troubleshooting

### Issue 1: "Rule not found"

**Cause:** Wrong ID or rule not synced

**Solution:**
```bash
# See the list first
spendgrid complete
# Select by number

# Or verify ID
spendgrid rules list
```

### Issue 2: Rule not syncing

**Causes:**
- Rule is inactive (active: false)
- Outside date range
- Already exists in month file

**Solution:**
```bash
# Check rule status
spendgrid rules list

# Edit if inactive
spendgrid rules edit rule_id
```

### Issue 3: Accidentally completed rule

**Solution:**
```bash
spendgrid uncomplete rule_id
```

### Issue 4: No rules in month file

**Cause:** Sync not done yet

**Solution:**
```bash
# Manual sync
spendgrid sync

# Or run any command (auto-sync)
spendgrid status
```

---

## Summary and Best Practices

### ✅ Do's

1. **Give meaningful names** - "Monthly Net Salary" instead of "Salary"
2. **Use tags consistently** - `#salary`, `#rent`, `#bill`
3. **Track projects** - `@companyA`, `@home`
4. **Complete regularly** - Daily or weekly check
5. **Review reports** - Planned vs actual difference

### ❌ Don'ts

1. Create multiple rules for same day/purpose
2. Enter negative amounts (use type instead)
3. Auto-complete entire month with `complete-month` (uncontrolled)
4. Manually change rule IDs

### 🎯 Ideal Flow

```bash
# 1. Start of month - check rules
spendgrid rules list

# 2. Regularly (daily/weekly)
spendgrid complete  # Mark completed
spendgrid report monthly  # See status

# 3. End of month - evaluation
spendgrid report monthly
# Check all rules completed
```

---

**Last Updated:** 2026-02-07  
**Version:** 1.0.0
