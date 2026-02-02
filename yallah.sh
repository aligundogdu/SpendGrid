#!/bin/bash

# Renkler
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}SpendGrid Git Geçmişi Oluşturuluyor...${NC}"

# 1. Faz 1: Temel Altyapı
echo -e "${GREEN}Faz 1: Temel Altyapı Commitleniyor...${NC}"
git add go.mod go.sum
git commit -m "chore: initialize project with go modules"

git add internal/filesystem
git commit -m "feat: implement local filesystem and initialization logic"

git add internal/config
git commit -m "feat: add global configuration management"

git add internal/i18n locales
git commit -m "feat: add i18n support (TR/EN)"

# 2. Faz 2: Transaction Sistemi
echo -e "${GREEN}Faz 2: Transaction Sistemi Commitleniyor...${NC}"
git add internal/transaction
git commit -m "feat: implement transaction CRUD operations"

git add internal/parser
git commit -m "feat: implement transaction parser with regex support"

# 3. Faz 3: Rules ve Smart Sync
echo -e "${GREEN}Faz 3: Rules ve Smart Sync Commitleniyor...${NC}"
git add internal/rules
git commit -m "feat: implement recurring rules engine and storage"

git add internal/status
git commit -m "feat: add status command for visual feedback"

# 4. Faz 4: Kur ve Raporlama
echo -e "${GREEN}Faz 4: Kur ve Raporlama Commitleniyor...${NC}"
git add internal/exchange
git commit -m "feat: integrate TCMB and Frankfurt APIs for exchange rates"

git add internal/reports
git commit -m "feat: implement monthly and yearly reporting with ASCII tables"

# 5. Faz 5: Ek Özellikler
echo -e "${GREEN}Faz 5: Ek Özellikler Commitleniyor...${NC}"
git add internal/investment
git commit -m "feat: add investment portfolio tracking"

git add internal/pool
git commit -m "feat: add backlog (pool) management system"

git add internal/validator
git commit -m "feat: add validation command for data integrity check"

# 6. Main Application & Documentation
echo -e "${GREEN}Main App ve Dokümantasyon Commitleniyor...${NC}"
git add cmd/spendgrid
git commit -m "feat: implement main CLI entry point and command routing"

git add README.md
git commit -m "docs: add comprehensive README with usage instructions"

# 7. Bug Fixes & Refinements
echo -e "${GREEN}Düzeltmeler Commitleniyor...${NC}"
git add .
git commit -m "fix: improve command parsing and misc refinements"

echo -e "${BLUE}Geçmiş başarıyla oluşturuldu! 🚀${NC}"
git log --oneline --graph --decorate --all
