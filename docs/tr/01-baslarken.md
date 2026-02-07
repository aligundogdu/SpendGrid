# Başlarken

SpendGrid'e hoş geldiniz! Bu kılavuz, finansal yönetim aracını hızlıca kurmanıza ve kullanmaya başlamanıza yardımcı olacaktır.

## İçindekiler
1. [SpendGrid Nedir?](#spendgrid-nedir)
2. [Kurulum](#kurulum)
3. [Hızlı Başlangıç](#hızlı-başlangıç)
4. [Temel Kavramlar](#temel-kavramlar)
5. [İlk Ayarlar](#ilk-ayarlar)
6. [Sonraki Adımlar](#sonraki-adımlar)

---

## SpendGrid Nedir?

SpendGrid, tekrarlayan finansal işlemlerinizi takip etmenizi sağlayan, yerel-öncelikli, dosya-tabanlı bir finans yönetim aracıdır.

### Felsefe

- **Veri Sahibi Sizsiniz:** Verileriniz düz metin dosyalarında saklanır
- **İnsan Tarafından Okunabilir:** Markdown formatı, herkesin anlayabileceği yapı
- **Projeksiyon Odaklı:** Geleceği planlayın, geçmişi takip edin

### Neden SpendGrid?

**Geleneksel Yöntem:**
```
Excel tablosu, karmaşık formüller, bulut bağımlılığı
```

**SpendGrid ile:**
```bash
# Basit komutlar
spendgrid add
spendgrid report monthly
# Verileriniz yerel dosyalarda, kontrol sizde
```

---

## Kurulum

### Seçenek 1: Homebrew (Önerilen)

```bash
brew tap yourusername/spendgrid
brew install spendgrid
```

### Seçenek 2: Manuel Kurulum

```bash
# macOS/Linux
wget https://github.com/yourusername/spendgrid/releases/latest/download/spendgrid-darwin-amd64
chmod +x spendgrid-darwin-amd64
sudo mv spendgrid-darwin-amd64 /usr/local/bin/spendgrid
```

### Seçenek 3: Kaynak Kodundan Derleme

```bash
git clone https://github.com/yourusername/spendgrid.git
cd spendgrid/cli-app
go build -o spendgrid ./cmd/spendgrid
sudo mv spendgrid /usr/local/bin/
```

---

## Hızlı Başlangıç

### Adım 1: Veritabanı Oluştur

```bash
# Finans dizini oluştur
mkdir ~/finans
cd ~/finans

# SpendGrid'i başlat
spendgrid init
```

**Oluşturulan Dosyalar:**
```
~/finans/
├── .spendgrid/           # Ana yapılandırma
├── _config/              # Ayarlar
│   ├── settings.yml
│   ├── rules.yml
│   ├── categories.yml
│   └── projects.yml
├── _pool/                # Bekleyen işlemler
│   └── backlog.md
├── _share/               # Paylaşım dosyaları
└── 2026/                 # Yıllık veri
    ├── 01.md
    ├── 02.md
    ├── ...
    └── 12.md
```

### Adım 2: İlk İşlemi Ekle

```bash
spendgrid add
```

**Örnek Diyalog:**
```
Gün [7]: 15
Açıklama: Market Alışverişi
Tutar ve Para Birimi: 450.50 TRY
Etiketler: #market #gida
Proje: @ev
Not: Haftalık alışveriş

✓ İşlem eklendi
```

### Adım 3: Listele

```bash
spendgrid list
```

**Çıktı:**
```
┌────┬────────────────────┬───────────┬──────────┬─────────────────┐
│ Gün│ Açıklama           │ Tutar     │ Para     │ Etiketler       │
├────┼────────────────────┼───────────┼──────────┼─────────────────┤
│ 15 │ Market Alışverişi  │  -450.50  │ TRY      │ #market #gida   │
└────┴────────────────────┴───────────┴──────────┴─────────────────┘
```

### Adım 4: Rapor Al

```bash
spendgrid report monthly
```

---

## Temel Kavramlar

### 1. İşlem (Transaction)

Tek bir finansal hareket:
```markdown
- 15 | Market Alışverişi | -450.50 TRY | #market #gida
```

Format: `- GÜN | AÇIKLAMA | TUTAR PARA | ETİKETLER`

### 2. Kural (Rule)

Tekrarlayan işlem:
```yaml
rules:
  - id: kira_001
    name: Ev Kira
    amount: 5000
    currency: TRY
    type: expense
    schedule:
      frequency: monthly
      day: 5
```

### 3. Ay Dosyası

Her ay için ayrı dosya: `01.md`, `02.md`, ..., `12.md`

```markdown
# 2026 Ocak

## ROWS
- 01 | Maaş | +25000 TRY | #maas
- 05 | Kira | -5000 TRY | #kira

## RULES
- [ ] 15 | Elektrik | -350 TRY | #fatura
```

### 4. Tamamlama Sistemi

- `[ ]` - Planlanmış, henüz gerçekleşmemiş
- `[x]` - Tamamlanmış, gerçekleşmiş

```bash
# Rule'ları listele
spendgrid complete

# Tamamla
spendgrid complete kural_id
```

---

## İlk Ayarlar

### Dil Ayarı

```bash
# Türkçe
spendgrid config set language tr

# İngilizce
spendgrid config set language en
```

### Temel Para Birimi

```bash
spendgrid config set base_currency TRY
```

### İlk Kuralınızı Oluşturun

```bash
# Aylık kira
spendgrid rules add "Ev Kira" 5000 TRY expense \
  --day 5 \
  --tags "kira,ev"

# Aylık maaş
spendgrid rules add "Maaş" 25000 TRY income \
  --day 5 \
  --tags "maas"
```

---

## Örnek: İlk Ayınızı Tamamlayın

### Gün 1 - Ayın Başı

```bash
# Durumu kontrol et
spendgrid status

# Planlanan rule'ları gör
spendgrid report monthly
```

### Gün 5 - Maaş ve Kira

```bash
# Maaş yattı
spendgrid complete
# Maaş'ı seç

# Kira ödendi
spendgrid complete
# Kira'yı seç

# Raporu kontrol et
spendgrid report monthly
```

### Gün 15 - Fatura

```bash
spendgrid add
# 15 | Elektrik Faturası | -350 TRY | #fatura

# veya rule olarak tanımladıysan
spendgrid complete
# Elektrik'i seç
```

### Ay Sonu - Değerlendirme

```bash
# Aylık rapor
spendgrid report monthly

# Yıllık görünüm
spendgrid report yearly

# HTML export
spendgrid report web
```

---

## Sık Kullanılan Komutlar

```bash
# Durum gör
spendgrid status

# İşlem ekle
spendgrid add

# Listele
spendgrid list

# Rapor al
spendgrid report monthly

# Rule ekle
spendgrid rules add "Kural Adı" TUTAR PARA TİP --day GÜN

# Rule tamamla
spendgrid complete

# Kurları güncelle
spendgrid exchange refresh
```

---

## Sonraki Adımlar

### Detaylı Kılavuzlar
- [Kural Sistemi](./04-kural-sistemi.md) - Detaylı rule kullanımı
- [Komut Referansı](./02-komutlar.md) - Tüm komutlar
- [İşlem Formatı](./03-islem-formati.md) - İşlem yazım kuralları

### İleri Seviye
- [Raporlama](./05-raporlama.md) - Rapor çeşitleri
- [Yatırımlar](./07-yatirimlar.md) - Yatırım takibi
- [Yapılandırma](./06-yapilandirma.md) - Ayarlar

### Yardım
```bash
# Genel yardım
spendgrid --help

# Komut yardımı
spendgrid add --help
spendgrid rules add --help
```

---

**Tebrikler!** SpendGrid kullanımına başladınız. 🎉

---

**Son Güncelleme:** 2026-02-07  
**Versiyon:** 1.0.0
