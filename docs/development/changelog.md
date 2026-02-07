# SpendGrid Changelog

All notable changes to SpendGrid project.

## [v0.2.3] - 2026-02-07

### Added

#### Rule Completion System
- **Checkbox-based completion tracking** - Rules now have `[ ]` (uncompleted) and `[x]` (completed) states
- **Three-section reports** - Reports now show: Actual (completed), Planned (uncompleted rules), and Projection (combined)
- **Complete command** - `spendgrid complete [rule_id]` to mark rules as completed
  - Interactive mode: Shows list of uncompleted rules, accepts number or ID
  - Direct mode: `spendgrid complete rule_id`
- **Uncomplete command** - `spendgrid uncomplete [rule_id]` to undo completion
  - Interactive mode: Shows list of completed rules
  - Direct mode: `spendgrid uncomplete rule_id`
- **Complete-month command** - `spendgrid complete-month [YYYY-MM]` to batch complete all rules in a month

#### Parser Updates
- **Transaction struct** - Added `IsRule` and `Completed` fields to track rule status
- **Checkbox parsing** - Parser now recognizes `[ ]` and `[x]` checkboxes and saves state

#### Space Support in Inputs
- **Interactive input** - Fixed space character handling in `readSimpleLine()` and `readWithAutocomplete()` functions
- **Description fields** - Users can now enter spaces in transaction descriptions and rule names during interactive mode

#### Documentation
- **Comprehensive Turkish documentation** (`docs/tr/`)
  - `04-kural-sistemi.md` - Detailed rule system guide with 23+ examples
  - `02-komutlar.md` - Complete command reference
- **English documentation** (`docs/en/`)
  - `04-rules-system.md` - English version of rule system guide

### Changed

#### Reports
- **MonthlyReport struct** - Added `PlannedIncome`, `PlannedExpenses`, and `PlannedTx` fields
- **Report generation** - Modified to separate completed transactions from uncompleted rules
- **Status display** - Now shows "Completed Transactions" and "Planned (Uncompleted Rules)" sections

#### Complete Command Behavior
- **Args validation** - Changed from `cobra.ExactArgs(1)` to `cobra.MaximumNArgs(1)`
- **Interactive fallback** - When run without arguments, shows list of recent rules
- **Smart input** - Accepts both rule number (1-N) and rule ID

### Technical Details

#### Files Modified
- `internal/parser/transaction.go` - Added IsRule and Completed fields
- `internal/reports/generator.go` - Three-section report implementation
- `internal/status/display.go` - Status display with rule separation
- `internal/transaction/manager.go` - Space character support
- `internal/rules/commands.go` - Space character support
- `cmd/spendgrid/commands/complete.go` - Complete/uncomplete commands
- `cmd/spendgrid/main.go` - Added new commands to root

#### New Files
- `docs/tr/04-kural-sistemi.md` - Turkish rule system documentation
- `docs/tr/02-komutlar.md` - Turkish commands reference
- `docs/en/04-rules-system.md` - English rule system documentation
- `docs/development/changelog.md` - This file

### Examples Added

#### Salary Planning (3 scenarios)
1. Standard monthly salary
2. Two separate salary payments
3. Foreign currency salary (USD)

#### Loan Payments (3 scenarios)
1. Housing loan (fixed installment)
2. Personal loan (monthly tracking)
3. Credit card payment (minimum + extra)

#### Monthly Expenses (10 examples)
- Rent, electricity, gas, water, internet
- Phone, gym, Netflix, Spotify, retirement plan

#### Planning & Completion (7 scenarios)
1. Daily tracking workflow
2. Weekly batch completion
3. Manual verification
4. Wrong completion correction
5. Partial completion (delay)
6. Bill amount change
7. Monthly summary workflow

---

## [Previous Versions]

### [1.x.x] - Pre-2026-02-07
- Initial SpendGrid implementation
- Basic transaction management
- Rule system without completion tracking
- Report generation
- Multi-currency support
- Exchange rate integration
- Investment tracking

---

## Migration Guide

### From Old Rule System

**Before:**
```markdown
## RULES
- 05 | Maaş [maa_001] | +25000.00 TRY | #maas
```

**After:**
```markdown
## RULES
- [ ] 05 | Maaş [maa_001] | +25000.00 TRY | #maas
```

**Action Required:**
1. Run `spendgrid sync` to add checkboxes
2. Use `spendgrid complete` to mark completed rules
3. Reports will now show correct actual vs planned

---

## Future Plans

- [ ] Rule editing from command line
- [ ] Rule deletion with confirmation
- [ ] Export/import rules
- [ ] Rule templates/presets
- [ ] Advanced filtering in reports
- [ ] Web dashboard for rule management

---

**Contributors:** opencode AI Assistant  
**Last Updated:** 2026-02-07
