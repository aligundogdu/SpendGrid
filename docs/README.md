# SpendGrid Documentation

Welcome to SpendGrid documentation!

## Available Languages / Mevcut Diller

- [🇬🇧 English](./en/)
- [🇹🇷 Türkçe](./tr/)

## Quick Navigation / Hızlı Erişim

### Getting Started
- [Installation & Quick Start (EN)](./en/01-getting-started.md)
- [Kurulum & Hızlı Başlangıç (TR)](./tr/01-baslarken.md)

### Core Features
- [Rule System Guide (EN)](./en/04-rules-system.md) - Complete guide with 23+ examples
- [Kural Sistemi Kılavuzu (TR)](./tr/04-kural-sistemi.md) - 23+ örnek ile detaylı kılavuz

### Command Reference
- [Command Reference (EN)](./en/02-commands.md)
- [Komut Referansı (TR)](./tr/02-komutlar.md)

### What's New
- [Changelog](./development/changelog.md) - Latest updates and features

---

## Documentation Structure

```
docs/
├── README.md                 # This file
├── en/                       # English documentation
│   ├── 01-getting-started.md
│   ├── 02-commands.md
│   ├── 03-transaction-format.md
│   ├── 04-rules-system.md    ⭐ NEW: Complete rule system guide
│   ├── 05-reporting.md
│   ├── 06-configuration.md
│   ├── 07-investments.md
│   ├── 08-exchange-rates.md
│   ├── 09-troubleshooting.md
│   └── 10-examples.md
├── tr/                       # Turkish documentation
│   ├── 01-baslarken.md
│   ├── 02-komutlar.md        ⭐ NEW: Complete command reference
│   ├── 03-islem-formati.md
│   ├── 04-kural-sistemi.md   ⭐ NEW: Detaylı kural sistemi
│   ├── 05-raporlama.md
│   ├── 06-yapilandirma.md
│   ├── 07-yatirimlar.md
│   ├── 08-doviz-kurlari.md
│   ├── 09-sorun-cozme.md
│   └── 10-ornekler.md
└── development/
    ├── changelog.md          ⭐ NEW: All changes documented
    ├── architecture.md
    └── contributing.md
```

---

## Recent Updates (2026-02-07)

### 🎯 Rule Completion System
- ✅ Checkbox tracking: `[ ]` → `[x]`
- ✅ Three-section reports: Actual, Planned, Projection
- ✅ `complete` and `uncomplete` commands
- ✅ Interactive selection with rule numbers

### ⌨️ Space Support
- ✅ Interactive inputs now support spaces
- ✅ Description fields accept full sentences

### 📚 New Documentation
- ✅ Comprehensive Turkish rule system guide (23+ examples)
- ✅ Complete command reference with all scenarios
- ✅ English rule system documentation
- ✅ Detailed changelog

---

## Key Features

### Local-First, File-Based
- All data in plain text Markdown files
- Human-readable format
- Version control friendly
- No vendor lock-in

### Smart Rule Engine
- Recurring transactions
- Auto-sync to month files
- Completion tracking
- Projection reports

### Multi-Currency
- TRY, USD, EUR, GBP support
- Automatic exchange rates
- Manual rate override (@rate)

### Complete Tracking
- Plan vs Actual comparison
- Checkbox-based completion
- Three-section reports
- Real-time status

---

## Support

- **Issues:** [GitHub Issues](https://github.com/yourusername/spendgrid/issues)
- **Discussions:** [GitHub Discussions](https://github.com/yourusername/spendgrid/discussions)
- **Documentation:** You're reading it! 📖

---

**Version:** 1.0.0  
**Last Updated:** 2026-02-07
