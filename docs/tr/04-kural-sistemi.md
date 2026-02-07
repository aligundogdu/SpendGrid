# SpendGrid Kural Sistemi - Detaylı Kullanım Kılavuzu

## İçindekiler
1. [Kural Sistemi Nedir?](#kural-sistemi-nedir)
2. [Temel Kavramlar](#temel-kavramlar)
3. [Kural Yapısı](#kural-yapısı)
4. [Senkronizasyon Mekanizması](#senkronizasyon-mekanizması)
5. [Tamamlama Sistemi](#tamamlama-sistemi)
6. [Senaryolar ve Örnekler](#senaryolar-ve-örnekler)
   - [Maaş Planlaması Senaryoları](#maaş-planlaması-senaryoları)
   - [Kredi Ödeme Senaryoları](#kredi-ödeme-senaryoları)
   - [Aylık Gider ve Harcama Kuralları](#aylık-gider-ve-harcama-kuralları)
   - [Planlama ve Tamamlama Senaryoları](#planlama-ve-tamamlama-senaryoları)
7. [İleri Seviye Kullanım](#ileri-seviye-kullanım)
8. [Sık Karşılaşılan Sorunlar](#sık-karşılaşılan-sorunlar)

---

## Kural Sistemi Nedir?

SpendGrid'in kural sistemi, tekrarlayan finansal işlemlerinizi otomatik olarak takip etmenizi sağlar. Aylık kira, maaş, fatura ödemeleri gibi düzenli gelir ve giderlerinizi bir kez tanımlayın, sistem her ay otomatik olarak bunları ay dosyalarınıza ekler.

### Neden Kural Sistemi?

**Geleneksel Yöntem:**
```bash
# Her ay manuel olarak ekleme yapmak
spendgrid add
# 05 | Kira | -5000 TRY | #kira @ev
# 15 | Elektrik | -300 TRY | #fatura
# 20 | Maaş | +25000 TRY | #maas @sirket
```

**Kural Sistemi ile:**
```bash
# Bir kez tanımla
spendgrid rules add "Aylık Kira" 5000 TRY expense --day 5 --tags "kira" --project "ev"

# Her ay otomatik senkronize olur
# Tamamlandığında işaretle
spendgrid complete kira_xxx
```

---

## Temel Kavramlar

### 1. Kural (Rule)
Bir kural, belirli bir tarihte tekrarlanan bir finansal işlemi temsil eder. Kurallar `_config/rules.yml` dosyasında saklanır.

### 2. Senkronizasyon (Sync)
Kuralların ay dosyalarına (`01.md`, `02.md`, vb.) otomatik olarak kopyalanması işlemidir. Her SpendGrid komutu çalıştığında otomatik olarak gerçekleşir.

### 3. Checkbox Durumu
Kurallar ay dosyalarına iki şekilde eklenir:
- `[ ]` - Planlanmış, henüz gerçekleşmemiş
- `[x]` - Tamamlanmış, gerçekleşmiş

### 4. Complete/Uncomplete
Rule'ların checkbox durumunu değiştirme işlemidir. Sadece tamamlanmış rule'lar raporlara dahil edilir.

---

## Kural Yapısı

### YAML Formatı

```yaml
rules:
  - id: maa_1770358056
    name: Aylık Net Maaş
    amount: 25000
    currency: TRY
    type: income
    category: gelir
    tags:
      - maas
      - net
    project: '@sirketA'
    schedule:
      frequency: monthly
      day: 5
    active: true
    start_date: "2026-01"
    end_date: "2026-12"
    metadata: "Her ayın 5'inde yatıyor"
```

### Alan Açıklamaları

| Alan | Zorunlu | Açıklama |
|------|---------|----------|
| `id` | Otomatik | Benzersiz tanımlayıcı (otomatik oluşturulur) |
| `name` | Evet | Kural adı (açıklayıcı) |
| `amount` | Evet | Tutar (pozitif sayı) |
| `currency` | Evet | Para birimi (TRY, USD, EUR) |
| `type` | Evet | `income` veya `expense` |
| `tags` | Hayır | Etiketler listesi |
| `project` | Hayır | Proje adı (@ ile başlar) |
| `schedule.frequency` | Evet | `monthly`, `weekly`, `yearly` |
| `schedule.day` | Evet | Ayın günü (1-31) |
| `active` | Hayır | Aktif/pasif durumu (default: true) |
| `start_date` | Hayır | Başlangıç tarihi (YYYY-MM) |
| `end_date` | Hayır | Bitiş tarihi (YYYY-MM) |
| `metadata` | Hayır | Açıklama/not |

---

## Senkronizasyon Mekanizması

### Nasıl Çalışır?

1. **Her Komut Sonrası:** `spendgrid` her çalıştığında otomatik senkronizasyon yapılır
2. **Ay Dosyası Kontrolü:** Mevcut ay dosyasına (`02.md`, vb.) bakılır
3. **Eksik Kurallar:** `_config/rules.yml` içindeki aktif kurallar ay dosyasına eklenir
4. **Checkbox Formatı:** Kurallar `- [ ] GÜN | AÇIKLAMA [ID] | TUTAR PARA | #etiketler` formatında eklenir

### Örnek Senkronizasyon

**rules.yml:**
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
    tags: [kira, ev]
```

**Senkronizasyon Sonrası 02.md:**
```markdown
## ROWS
- 01 | Market | -250 TRY | #market

## RULES
- [ ] 05 | Ev Kira [kira_001] | -5000.00 TRY | #kira #ev
```

### Önemli Notlar

- ✅ Mevcut kurallar **üzerine yazılmaz**
- ✅ Manuel düzenlemeler korunur
- ✅ Sadece `[ ]` (işaretlenmemiş) kurallar senkronize edilir
- ✅ `[x]` (işaretlenmiş) kurallar dokunulmaz

---

## Tamamlama Sistemi

### Felsefe: Planlama vs Gerçekleşme

SpendGrid'de kural sistemi iki aşamalıdır:

1. **Planlama Aşaması:** Kural senkronize edilir, `[ ]` olarak işaretlenir
2. **Gerçekleşme Aşaması:** Para geldiğinde/gittiğinde `[x]` yapılır

### Neden Bu Sistem?

**Sorun:**
```
Her ay 5'inde 25.000 TL maaş gelmesi bekleniyor.
Ama ayın 3'ünde henüz para gelmedi.
Rapora bakıyorsun: "Gelir: 25.000 TL" - YANLIŞ!
```

**Çözüm:**
```
Planlanan: +25.000 TL (henüz gelmedi)
Gerçekleşen: 0 TL
Projeksiyon: +25.000 TL (beklenen)

Para geldikten sonra:
Gerçekleşen: +25.000 TL
```

### Komutlar

#### complete - Rule Tamamlama
```bash
# Interaktif mod (önerilen)
spendgrid complete
# Liste gösterilir, numara veya ID girilir

# Direkt ID ile
spendgrid complete maa_1770358056

# Ayın tümünü topluca tamamla
spendgrid complete-month 2026-02
```

#### uncomplete - Tamamlama İptali
```bash
# Interaktif mod
spendgrid uncomplete

# Direkt ID ile
spendgrid uncomplete maa_1770358056
```

### Üç Bölümlü Rapor

Raporlarda artık üç bölüm görürsünüz:

```
📊 GERÇEKLEŞEN (Gerçekleşmiş İşlemler)
   Gelir: 15.000 TL
   Gider: 8.000 TL
   Net: +7.000 TL

📅 PLANLANAN (Tamamlanmamış Rule'lar)
   Gelir: +25.000 TL (Maaş)
   Gider: -5.000 TL (Kira)

🔮 PROJEKSİYON (Gerçekleşen + Planlanan)
   Beklenen Net: +27.000 TL
```

---

## Senaryolar ve Örnekler

### Maaş Planlaması Senaryoları

#### Senaryo 1: Standart Aylık Maaş

**Durum:** Her ayın 5'inde 25.000 TL net maaş alınıyor.

```bash
# Kural oluştur
spendgrid rules add "Aylık Net Maaş" 25000 TRY income \
  --day 5 \
  --tags "maas,net" \
  --project "@sirketA"

# Senkronizasyon sonrası ay dosyasına eklenir
# - [ ] 05 | Aylık Net Maaş [maa_xxx] | +25000.00 TRY | #maas #net @@sirketA

# Maaş yattığında tamamla
spendgrid complete maa_xxx
# veya
spendgrid complete
# 1 yaz (listede 1. sırada ise)
```

#### Senaryo 2: İki Ayrı Maaş Ödemesi

**Durum:** 5'inde ana maaş (25.000 TL), 20'sinde ek ödeme (5.000 TL)

```bash
# Ana maaş
spendgrid rules add "Ana Maaş" 25000 TRY income --day 5 --tags "maas,ana"

# Ek ödeme
spendgrid rules add "Ek Ödeme" 5000 TRY income --day 20 --tags "maas,ek"

# Rapor görünümü:
# Planlanan Gelir: +30.000 TL
#   - 05 | Ana Maaş: +25.000 TL
#   - 20 | Ek Ödeme: +5.000 TL
```

#### Senaryo 3: Döviz Üzerinden Maaş (USD)

**Durum:** Freelance çalışma, her ay 15'inde 1.000 USD ödeme

```bash
# USD cinsinden kural
spendgrid rules add "Freelance Ödeme" 1000 USD income \
  --day 15 \
  --tags "freelance,usd" \
  --project "@musteriX"

# Ay dosyasına eklenir:
# - [ ] 15 | Freelance Ödeme [fre_xxx] | +1000.00 USD | #freelance #usd @@musteriX

# Ödeme geldiğinde
spendgrid complete fre_xxx

# Rapor otomatik kur çevrimi yapar (örn: 1 USD = 35 TL)
# Gelir: +35.000 TRY (1000 USD @ 35.00)
```

---

### Kredi Ödeme Senaryoları

#### Senaryo 1: Konut Kredisi (Eşit Taksit)

**Durum:** Her ayın 10'unda 4.500 TL konut kredisi ödemesi

```bash
# Kredi kuralı
spendgrid rules add "Konut Kredisi" 4500 TRY expense \
  --day 10 \
  --tags "kredi,konut,banka" \
  --project "@ziraat"

# Ay dosyası:
# - [ ] 10 | Konut Kredisi [kre_xxx] | -4500.00 TRY | #kredi #konut #banka @@ziraat

# Ödeme çekildiğinde
spendgrid complete kre_xxx
```

#### Senaryo 2: İhtiyaç Kredisi (Aylık Takip)

**Durum:** Her ayın 1'inde 2.800 TL ihtiyaç kredisi, ekstra bilgi ile

```yaml
# rules.yml içinde:
rules:
  - id: kre_ihtiyac_001
    name: İhtiyaç Kredisi Taksiti
    amount: 2800
    currency: TRY
    type: expense
    tags: [kredi, ihtiyac, akbank]
    project: '@akbank'
    schedule:
      frequency: monthly
      day: 1
    metadata: "24 ay taksit, Kalan: 18 ay"
```

```bash
# Senkronizasyon sonrası
# - [ ] 01 | İhtiyaç Kredisi Taksiti [kre_ihtiyac_001] | -2800.00 TRY | #kredi #ihtiyac #akbank @@akbank

# Her ay metadata güncellenebilir
spendgrid rules edit kre_ihtiyac_001
# Metadata: "24 ay taksit, Kalan: 17 ay"
```

#### Senaryo 3: Kredi Kartı Ödeme (Asgari + Ek)

**Durum:** Her ayın 5'inde asgari ödeme 1.500 TL, ama tam ödeme planı

```bash
# Asgari ödeme kuralı (sabit)
spendgrid rules add "KK Asgari Ödeme" 1500 TRY expense --day 5 --tags "kredi,kart,asgari"

# Ay içinde ek ödeme (manuel ekleme)
spendgrid add
# 15 | KK Ek Ödeme | -3000 TRY | #kredi #kart #ek
```

---

### Aylık Gider ve Harcama Kuralları

#### 1. Kira Ödemesi
```bash
spendgrid rules add "Ev Kira" 5000 TRY expense --day 5 --tags "kira,ev,konut"
```

#### 2. Elektrik Faturası
```bash
spendgrid rules add "Elektrik Faturası" 350 TRY expense --day 15 --tags "fatura,elektrik"
```

#### 3. Doğalgaz Faturası
```bash
spendgrid rules add "Doğalgaz Faturası" 450 TRY expense --day 15 --tags "fatura,dogalgaz"
```

#### 4. Su Faturası
```bash
spendgrid rules add "Su Faturası" 150 TRY expense --day 20 --tags "fatura,su"
```

#### 5. İnternet Ücreti
```bash
spendgrid rules add "İnternet Ücreti" 120 TRY expense --day 1 --tags "fatura,internet"
```

#### 6. Telefon Faturası
```bash
spendgrid rules add "Telefon Faturası" 250 TRY expense --day 5 --tags "fatura,telefon"
```

#### 7. Spor Salonu Üyeliği
```bash
spendgrid rules add "Spor Salonu" 300 TRY expense --day 1 --tags "spor,uyelik"
```

#### 8. Netflix Üyeliği
```bash
spendgrid rules add "Netflix" 50 TRY expense --day 15 --tags "abonelik,dijital"
```

#### 9. Spotify Üyeliği
```bash
spendgrid rules add "Spotify" 35 TRY expense --day 20 --tags "abonelik,muzik"
```

#### 10. Aylık Yatırım (Otomatik)
```bash
spendgrid rules add "BES Ödemesi" 1000 TRY expense --day 10 --tags "yatirim,bes,emeklilik"
```

**Tüm Giderlerin Rapor Görünümü:**
```
📅 PLANLANAN Giderler:
   - 01 | İnternet: -120 TRY
   - 01 | Spor Salonu: -300 TRY
   - 05 | Ev Kira: -5000 TRY
   - 05 | Telefon: -250 TRY
   - 10 | BES: -1000 TRY
   - 15 | Elektrik: -350 TRY
   - 15 | Doğalgaz: -450 TRY
   - 15 | Netflix: -50 TRY
   - 20 | Su: -150 TRY
   - 20 | Spotify: -35 TRY
   
   Toplam Planlanan Gider: -7.705 TRY
```

---

### Planlama ve Tamamlama Senaryoları

#### Senaryo 1: Günlük Takip (Tavsiye Edilen)

**Gün 1 - Ayın Başı:**
```bash
spendgrid status
# Planlanan 3 rule var gösterilir

spendgrid report monthly
# 📅 PLANLANAN: +25.000 TL (Maaş)
# 📅 PLANLANAN: -5.000 TL (Kira)
```

**Gün 5 - Maaş Günü:**
```bash
# Maaş yattı, kontrol et
spendgrid complete
# Listeden 1 seç (Maaş)

spendgrid report monthly
# 📊 GERÇEKLEŞEN: +25.000 TL
# 📅 PLANLANAN: -5.000 TL (Kira bekleniyor)
# 🔮 PROJEKSİYON: +20.000 TL
```

**Gün 5 - Kira Ödeme:**
```bash
# Kira ödendi
spendgrid complete
# Listeden Kira'yı seç

spendgrid report monthly
# 📊 GERÇEKLEŞEN: +25.000 TL / -5.000 TL
# Net: +20.000 TL
```

#### Senaryo 2: Haftalık Toplu Tamamlama

```bash
# Her Cumartesi haftalık kontrol
spendgrid complete
# Bu hafta tamamlananları işaretle

# veya topluca
spendgrid complete-month
# Ayın tüm rule'larını tamamla (dikkatli kullan!)
```

#### Senaryo 3: Manuel Doğrulama

```bash
# Banka hesabını kontrol et
# Gelen para: 25.000 TL (Maaş)

# SpendGrid'de kontrol
spendgrid complete maa_001

# Raporu tekrar kontrol et
spendgrid report monthly
# GERÇEKLEŞEN kısmında maaş görünmeli
```

#### Senaryo 4: Yanlış Tamamlama Düzeltme

```bash
# Yanlışlıkla kira tamamlandı ama henüz ödenmedi
spendgrid uncomplete kira_001

# Rapor tekrar hesaplanır
# Kira PLANLANAN'a geri döner
```

#### Senaryo 5: Kısmi Tamamlama (Maaş Gecikmesi)

```bash
# Ayın 5'inde maaş yatması gerekti ama yatmadı
spendgrid report monthly
# PLANLANAN'da hâlâ maaş bekliyor

# Ayın 7'sinde yattı
spendgrid complete maa_001

# GERÇEKLEŞEN'e geçti
```

#### Senaryo 6: Fatura Tutarı Değişikliği

```bash
# Elektrik faturası her ay farklı
# Kural: 350 TRY (ortalama)

# Ayın 15'inde gerçek fatura: 420 TRY
# 1. Yöntem: Kuralı tamamla + manuel ekle
spendgrid complete ele_001
spendgrid add
# 15 | Elektrik Fatura Farkı | -70 TRY | #fatura #elektrik

# 2. Yöntem: Direkt manuel ekle (kuralı silme)
spendgrid add
# 15 | Elektrik Faturası (Gerçek) | -420 TRY | #fatura #elektrik
```

#### Senaryo 7: Aylık Özet ve Toplu İşlem

```bash
# Ay sonu kontrol
spendgrid report monthly

# Eksik tamamlamaları kontrol et
spendgrid complete

# Tümü tamamlandı mı?
spendgrid status
# "All rules completed" mesajı

# Sonraki ay hazırlığı
spendgrid rules list
# Pasif olanları kontrol et
```

---

## İleri Seviye Kullanım

### 1. Tarih Aralıklı Kurallar

```yaml
rules:
  - id: staj_maas
    name: Stajyer Maaşı
    amount: 5000
    currency: TRY
    type: income
    schedule:
      frequency: monthly
      day: 5
    start_date: "2026-06"  # Haziran'da başla
    end_date: "2026-08"    # Ağustos'ta bitir
```

### 2. Proje Bazlı Takip

```bash
# Aynı projeye ait gelir/gider
spendgrid rules add "Proje X Ödeme" 10000 TRY income --day 10 --project "@projeX"
spendgrid rules add "Proje X Maliyet" 2000 TRY expense --day 15 --project "@projeX"

# Raporlama proje bazlı yapılır
```

### 3. Çoklu Para Birimi

```bash
# USD gelir
spendgrid rules add "Freelance USD" 1000 USD income --day 15

# EUR gider  
spendgrid rules add "Hosting EUR" 50 EUR expense --day 1

# Rapor otomatik kur çevrimi yapar
```

---

## Sık Karşılaşılan Sorunlar

### Sorun 1: "Rule not found"

**Neden:** ID yanlış veya rule senkronize olmamış

**Çözüm:**
```bash
# Önce listeyi gör
spendgrid complete
# Numara ile seç

# veya ID'yi doğrula
spendgrid rules list
```

### Sorun 2: Rule senkronize olmuyor

**Neden:** 
- Rule pasif (active: false)
- Tarih aralığı dışında
- Zaten ay dosyasında var

**Çözüm:**
```bash
# Rule durumunu kontrol et
spendgrid rules list

# Aktif değilse düzenle
spendgrid rules edit rule_id
```

### Sorun 3: Yanlışlıkla tamamlanan rule

**Çözüm:**
```bash
spendgrid uncomplete rule_id
```

### Sorun 4: Ay dosyasında rule yok

**Neden:** Henüz senkronizasyon yapılmamış

**Çözüm:**
```bash
# Manuel senkronizasyon
spendgrid sync

# veya herhangi bir komut çalıştır (auto-sync)
spendgrid status
```

---

## Özet ve En İyi Pratikler

### ✅ Yapılması Gerekenler

1. **Her rule'a anlamlı isim verin** - "Maaş" yerine "Aylık Net Maaş"
2. **Etiketleri düzenli kullanın** - `#maas`, `#kira`, `#fatura`
3. **Projeleri takip edin** - `@sirketA`, `@ev`
4. **Düzenli tamamlayın** - Her gün veya hafta kontrol
5. **Raporları inceleyin** - Planlanan vs gerçekleşen farkı

### ❌ Yapılmaması Gerekenler

1. Aynı gün/amaç için çoklu rule oluşturma
2. Tutarları negatif olarak girmeye çalışma (type kullan)
3. Tüm ayı `complete-month` ile otomatik tamamlama (kontrolsüz)
4. Rule ID'lerini manuel değiştirme

### 🎯 İdeal Akış

```bash
# 1. Ay başı - kuralları kontrol et
spendgrid rules list

# 2. Düzenli olarak (günlük/haftalık)
spendgrid complete  # Tamamlananları işaretle
spendgrid report monthly  # Durumu gör

# 3. Ay sonu - değerlendirme
spendgrid report monthly
# Tüm rule'lar tamamlandı mı kontrol et
```

---

**Son Güncelleme:** 2026-02-07  
**Versiyon:** 1.0.0
