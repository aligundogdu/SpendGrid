# Getting Started

Welcome to SpendGrid! This guide will help you quickly set up and start using the financial management tool.

## Table of Contents
1. [What is SpendGrid?](#what-is-spendgrid)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Basic Concepts](#basic-concepts)
5. [Initial Setup](#initial-setup)
6. [Next Steps](#next-steps)

---

## What is SpendGrid?

SpendGrid is a local-first, file-based financial management tool that helps you track recurring financial transactions.

### Philosophy

- **You Own Your Data:** Your data is stored in plain text files
- **Human Readable:** Markdown format that anyone can understand
- **Projection Focused:** Plan the future, track the past

### Why SpendGrid?

**Traditional Method:**
```
Excel spreadsheet, complex formulas, cloud dependency
```

**With SpendGrid:**
```bash
# Simple commands
spendgrid add
spendgrid report monthly
# Your data in local files, you're in control
```

---

## Installation

### Option 1: Homebrew (Recommended)

```bash
brew tap yourusername/spendgrid
brew install spendgrid
```

### Option 2: Manual Installation

```bash
# macOS/Linux
wget https://github.com/yourusername/spendgrid/releases/latest/download/spendgrid-darwin-amd64
chmod +x spendgrid-darwin-amd64
sudo mv spendgrid-darwin-amd64 /usr/local/bin/spendgrid
```

### Option 3: Build from Source

```bash
git clone https://github.com/yourusername/spendgrid.git
cd spendgrid/cli-app
go build -o spendgrid ./cmd/spendgrid
sudo mv spendgrid /usr/local/bin/
```

---

## Quick Start

### Step 1: Create Database

```bash
# Create finance directory
mkdir ~/finance
cd ~/finance

# Initialize SpendGrid
spendgrid init
```

**Created Files:**
```
~/finance/
├── .spendgrid/           # Main configuration
├── _config/              # Settings
│   ├── settings.yml
│   ├── rules.yml
│   ├── categories.yml
│   └── projects.yml
├── _pool/                # Pending transactions
│   └── backlog.md
├── _share/               # Shared files
└── 2026/                 # Yearly data
    ├── 01.md
    ├── 02.md
    ├── ...
    └── 12.md
```

### Step 2: Add First Transaction

```bash
spendgrid add
```

**Sample Dialogue:**
```
Day [7]: 15
Description: Grocery Shopping
Amount and Currency: 450.50 TRY
Tags: #grocery #food
Project: @home
Note: Weekly shopping

✓ Transaction added
```

### Step 3: List

```bash
spendgrid list
```

**Output:**
```
┌────┬────────────────────┬───────────┬──────────┬─────────────────┐
│ Day│ Description        │ Amount    │ Currency │ Tags            │
├────┼────────────────────┼───────────┼──────────┼─────────────────┤
│ 15 │ Grocery Shopping   │  -450.50  │ TRY      │ #grocery #food  │
└────┴────────────────────┴───────────┴──────────┴─────────────────┘
```

### Step 4: Get Report

```bash
spendgrid report monthly
```

---

## Basic Concepts

### 1. Transaction

A single financial movement:
```markdown
- 15 | Grocery Shopping | -450.50 TRY | #grocery #food
```

Format: `- DAY | DESCRIPTION | AMOUNT CURRENCY | TAGS`

### 2. Rule

Recurring transaction:
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
```

### 3. Month File

Separate file for each month: `01.md`, `02.md`, ..., `12.md`

```markdown
# 2026 January

## ROWS
- 01 | Salary | +25000 TRY | #salary
- 05 | Rent | -5000 TRY | #rent

## RULES
- [ ] 15 | Electricity | -350 TRY | #bill
```

### 4. Completion System

- `[ ]` - Planned, not yet occurred
- `[x]` - Completed, occurred

```bash
# List rules
spendgrid complete

# Mark complete
spendgrid complete rule_id
```

---

## Initial Setup

### Language Setting

```bash
# English
spendgrid config set language en

# Turkish
spendgrid config set language tr
```

### Base Currency

```bash
spendgrid config set base_currency USD
```

### Create Your First Rule

```bash
# Monthly rent
spendgrid rules add "Home Rent" 5000 USD expense \
  --day 5 \
  --tags "rent,home"

# Monthly salary
spendgrid rules add "Salary" 5000 USD income \
  --day 5 \
  --tags "salary"
```

---

## Example: Complete Your First Month

### Day 1 - Start of Month

```bash
# Check status
spendgrid status

# See planned rules
spendgrid report monthly
```

### Day 5 - Salary and Rent

```bash
# Salary deposited
spendgrid complete
# Select Salary

# Rent paid
spendgrid complete
# Select Rent

# Check report
spendgrid report monthly
```

### Day 15 - Bill

```bash
spendgrid add
# 15 | Electricity Bill | -350 USD | #bill

# or if defined as rule
spendgrid complete
# Select Electricity
```

### End of Month - Evaluation

```bash
# Monthly report
spendgrid report monthly

# Yearly view
spendgrid report yearly

# HTML export
spendgrid report web
```

---

## Frequently Used Commands

```bash
# View status
spendgrid status

# Add transaction
spendgrid add

# List
spendgrid list

# Get report
spendgrid report monthly

# Add rule
spendgrid rules add "Rule Name" AMOUNT CURRENCY TYPE --day DAY

# Complete rule
spendgrid complete

# Update rates
spendgrid exchange refresh
```

---

## Next Steps

### Detailed Guides
- [Rule System](./04-rules-system.md) - Detailed rule usage
- [Command Reference](./02-commands.md) - All commands
- [Transaction Format](./03-transaction-format.md) - Transaction syntax

### Advanced
- [Reporting](./05-reporting.md) - Report types
- [Investments](./07-investments.md) - Investment tracking
- [Configuration](./06-configuration.md) - Settings

### Help
```bash
# General help
spendgrid --help

# Command help
spendgrid add --help
spendgrid rules add --help
```

---

**Congratulations!** You've started using SpendGrid. 🎉

---

**Last Updated:** 2026-02-07  
**Version:** 1.0.0
