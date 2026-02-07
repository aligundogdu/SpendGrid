# SpendGrid 💰

**Financial Projection and Cash Flow Management** | **Finansal Projeksiyon ve Nakit Akışı Yönetimi**

[![Version](https://img.shields.io/badge/version-v0.2.5-blue.svg)](https://github.com/aligundogdu/SpendGrid/releases)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 🇬🇧 English

### Philosophy

SpendGrid is a **local-first, file-based** financial management tool designed for people who want complete ownership of their financial data. Unlike cloud-based solutions, all your data stays on your local machine in plain-text Markdown files that you can read and edit with any text editor.

**Core Principles:**
- 🏠 **Data Ownership**: Your financial data belongs to you, not a corporation
- 📄 **Human Readable**: Open files in any text editor - no special software required
- 🔮 **Projection Focused**: Plan your next 12 months, not just report the past
- 🔄 **Hybrid Structure**: Personal and business finances in one timeline
- 🧠 **Smart Sync**: Automatic synchronization between rules and transaction files

### Features

✨ **Transaction Management**
- Natural language quick input: `./spendgrid "-100TL groceries #food @home"`
- Interactive and direct transaction entry modes
- Multi-currency support (TRY, USD, EUR)
- Tag and project categorization
- Investment tracking with cost basis calculation

🔄 **Smart Rules Engine**
- Automatic recurring transactions
- Syncs rules to month files automatically
- Respects manual edits (checked items are preserved)
- Only affects current and future months

📊 **Powerful Reporting**
- Monthly and yearly financial reports
- ASCII table format in terminal
- HTML export for web viewing
- Currency conversion with TCMB/Frankfurt API

🌍 **Localization**
- Full Turkish and English support
- Language detection based on system settings
- All CLI messages localized

### Installation

#### Option 1: Homebrew (Recommended for macOS/Linux)

```bash
brew tap aligundogdu/spendgrid
brew install spendgrid
```

To upgrade:
```bash
brew upgrade spendgrid
```

#### Option 2: Install Script

One-line installation (works on macOS and Linux):

```bash
curl -fsSL https://raw.githubusercontent.com/aligundogdu/SpendGrid/main/cli-app/install.sh | bash
```

This will automatically detect your OS and architecture, download the appropriate binary, and install it to `/usr/local/bin`.

#### Option 3: Manual Download

Download the pre-built binary for your system from the [Releases](https://github.com/aligundogdu/SpendGrid/releases) page:

**macOS (Intel):**
```bash
curl -L -o spendgrid https://github.com/aligundogdu/SpendGrid/releases/latest/download/spendgrid-darwin-amd64
chmod +x spendgrid
sudo mv spendgrid /usr/local/bin/
```

**macOS (Apple Silicon):**
```bash
curl -L -o spendgrid https://github.com/aligundogdu/SpendGrid/releases/latest/download/spendgrid-darwin-arm64
chmod +x spendgrid
sudo mv spendgrid /usr/local/bin/
```

**Linux:**
```bash
curl -L -o spendgrid https://github.com/aligundogdu/SpendGrid/releases/latest/download/spendgrid-linux-amd64
chmod +x spendgrid
sudo mv spendgrid /usr/local/bin/
```

#### Option 4: Build from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/aligundogdu/SpendGrid.git
cd SpendGrid/cli-app

# Build the binary
go build -o spendgrid cmd/spendgrid/main.go

# Optional: Move to PATH
mv spendgrid /usr/local/bin/
```

### Quick Start

```bash
# Initialize a new SpendGrid database
spendgrid init

# Add transactions with natural language
spendgrid "-2500TRY rent payment #housing"
spendgrid "5000 USD consulting fee #business @client1"

# View current month
spendgrid list

# Generate reports
spendgrid report          # Monthly report
spendgrid report --year   # Yearly report
spendgrid report --web    # HTML export

# Check status
spendgrid status
```

### Transaction Format

Transactions use a simple pipe-delimited format:

```markdown
- DAY | DESCRIPTION | AMOUNT CURRENCY | TAGS | [METADATA]

Examples:
- 15 | Grocery Shopping | -350.50 TRY | #food #weekly |
- 01 | Rent | -2500 TRY | #housing |
- 05 | Freelance Payment | 1000 USD | #business @project1 | [NOTE:Client X]
```

### Directory Structure

```
your-finances/
├── .spendgrid              # Version marker
├── _config/
│   ├── settings.yml        # Local settings
│   ├── rules.yml           # Recurring rules
│   ├── categories.yml      # #tag definitions
│   └── projects.yml        # @project definitions
├── _pool/
│   └── backlog.md          # Pending transactions
├── _share/
│   └── report_*.html       # HTML exports
└── 2026/
    ├── 01.md ... 12.md     # Monthly files
    └── 2026_Projection.md  # Year summary
```

### Command Reference

| Command | Description |
|---------|-------------|
| `init` | Initialize database |
| `add` | Add transaction (interactive) |
| `list [month]` | List transactions |
| `edit <line>` | Edit transaction |
| `remove <line>` | Remove transaction |
| `rules` | Manage recurring rules |
| `sync` | Sync rules to months |
| `report` | Generate reports |
| `exchange refresh` | Update exchange rates |
| `investments` | View investment portfolio |
| `pool` | Manage backlog |
| `validate` | Validate all files |
| `status` | Show database status |
| `set config` | Configure settings |

---

## 🇹🇷 Türkçe

### Felsefe

SpendGrid, finansal verilerinin tamamına sahip olmak isteyenler için tasarlanmış **local-first, dosya-tabanlı** bir finans yönetim aracıdır. Bulut tabanlı çözümlerin aksine, tüm verileriniz yerel makinenizde düz metin Markdown dosyalarında saklanır ve herhangi bir metin editörüyle okunup düzenlenebilir.

**Temel Prensipler:**
- 🏠 **Veri Sahipliği**: Finansal verileriniz size ait, bir şirkete değil
- 📄 **İnsan Tarafından Okunabilir**: Dosyaları herhangi bir editörde açın - özel yazılım gerekmez
- 🔮 **Projeksiyon Odaklı**: Geçmişi raporlamak yerine önümüzdeki 12 ayı planlayın
- 🔄 **Hibrit Yapı**: Kişisel ve iş finansmanı tek zaman çizelgesinde
- 🧠 **Akıllı Senkronizasyon**: Kurallar ve işlem dosyaları arasında otomatik senkronizasyon

### Özellikler

✨ **İşlem Yönetimi**
- Doğal dil hızlı giriş: `./spendgrid "-100TL market alışverişi #mutfak @ev"`
- İnteraktif ve doğrudan işlem giriş modları
- Çoklu para birimi desteği (TRY, USD, EUR)
- Etiket ve proje kategorizasyonu
- Maliyet bazlı yatırım takibi

🔄 **Akıllı Kurallar Motoru**
- Otomatik tekrarlayan işlemler
- Kuralları ay dosyalarına otomatik senkronize eder
- Manuel düzenlemelere saygı (işaretli öğeler korunur)
- Sadece mevcut ve gelecek ayları etkiler

📊 **Güçlü Raporlama**
- Aylık ve yıllık finansal raporlar
- Terminalde ASCII tablo formatı
- Web görüntüleme için HTML dışa aktarım
- TCMB/Frankfurt API ile kur dönüşümü

🌍 **Yerelleştirme**
- Tam Türkçe ve İngilizce desteği
- Sistem ayarlarına göre dil algılama
- Tüm CLI mesajları yerelleştirildi

### Kurulum

```bash
# Depoyu klonlayın
git clone https://github.com/yourusername/spendgrid.git
cd spendgrid/cli-app

# İkili dosyayı derleyin
go build -o spendgrid cmd/spendgrid/main.go

# İsteğe bağlı: PATH'e taşıyın
mv spendgrid /usr/local/bin/
```

### Hızlı Başlangıç

```bash
# Yeni bir SpendGrid veritabanı başlatın
spendgrid init

# Doğal dil ile işlem ekleyin
spendgrid "-2500TRY kira ödemesi #konut"
spendgrid "5000 USD danışmanlık ücreti #iş @musteri1"

# Mevcut ayı görüntüleyin
spendgrid list

# Raporlar oluşturun
spendgrid report          # Aylık rapor
spendgrid report --year   # Yıllık rapor
spendgrid report --web    # HTML dışa aktarım

# Durumu kontrol edin
spendgrid status
```

### İşlem Formatı

İşlemler basit bir pipe-ayrılmış format kullanır:

```markdown
- GÜN | AÇIKLAMA | TUTAR PARA_BİRİMİ | ETİKETLER | [META_VERİ]

Örnekler:
- 15 | Market Alışverişi | -350.50 TRY | #gıda #haftalık |
- 01 | Kira | -2500 TRY | #konut |
- 05 | Serbest Çalışma | 1000 USD | #iş @proje1 | [NOTE:Müşteri X]
```

### Dizin Yapısı

```
finansman/
├── .spendgrid              # Versiyon belirteci
├── _config/
│   ├── settings.yml        # Yerel ayarlar
│   ├── rules.yml           # Tekrarlayan kurallar
│   ├── categories.yml      # #etiket tanımları
│   └── projects.yml        # @proje tanımları
├── _pool/
│   └── backlog.md          # Bekleyen işlemler
├── _share/
│   └── report_*.html       # HTML dışa aktarımlar
└── 2026/
    ├── 01.md ... 12.md     # Aylık dosyalar
    └── 2026_Projection.md  # Yıl özeti
```

### Komut Referansı

| Komut | Açıklama |
|-------|----------|
| `init` | Veritabanı başlat |
| `add` | İşlem ekle (interaktif) |
| `list [ay]` | İşlemleri listele |
| `edit <satır>` | İşlem düzenle |
| `remove <satır>` | İşlem sil |
| `rules` | Tekrarlayan kuralları yönet |
| `sync` | Kuralları aylara senkronize et |
| `report` | Raporlar oluştur |
| `exchange refresh` | Kur güncelle |
| `investments` | Yatırım portföyü görüntüle |
| `pool` | Backlog yönetimi |
| `validate` | Tüm dosyaları doğrula |
| `status` | Veritabanı durumunu göster |
| `set config` | Ayarları yapılandır |

---

## 💡 Philosophy in Action

**The "why" behind SpendGrid:**

> "We believe your financial data is yours. Not a database vendor's, not a cloud service's, not a startup's. Yours. When you open your SpendGrid folder, you see plain text files. You can read them with Notepad. You can version control them with Git. You can sync them with Dropbox if you want. But most importantly, you understand exactly where your money is going without any vendor lock-in."

**SpendGrid'un Felsefesi:**

> "Finansal verilerinizin size ait olduğuna inanıyoruz. Bir veritabanı satıcısına değil, bir bulut hizmetine değil, bir startup'a değil. Size ait. SpendGrid klasörünü açtığınızda, düz metin dosyaları görürsünüz. Onları Notepad ile okuyabilirsiniz. Git ile versiyon kontrolü yapabilirsiniz. İsterseniz Dropbox ile senkronize edebilirsiniz. Ama en önemlisi, herhangi bir satıcı bağımlılığı olmadan paranızın nereye gittiğini tam olarak anlarsınız."

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Katkılarınızı bekliyoruz! Lütfen [CONTRIBUTING.md](CONTRIBUTING.md) dosyasına bakın.

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details.

---

## 🌟 Acknowledgments

- Inspired by [Plain Text Accounting](https://plaintextaccounting.org/)
- Exchange rates powered by TCMB (Turkey) and Frankfurt ECB API
- Built with ❤️ in Go

---

**Made with 💚 for people who care about their data.**  
**Verileri için endişelenen insanlar için 💚 ile yapıldı.**
