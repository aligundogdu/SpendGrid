# SpendGrid Komut Referansı

## Hızlı Başvuru Kartı

| Komut | Açıklama | Kullanım |
|-------|----------|----------|
| `init` | Yeni veritabanı oluştur | `spendgrid init` |
| `add` | İşlem ekle | `spendgrid add` veya `spendgrid add --direct "GÜN\|AÇIKLAMA\|TUTAR\|ETİKETLER"` |
| `list` | İşlemleri listele | `spendgrid list` veya `spendgrid list 01` |
| `edit` | İşlem düzenle | `spendgrid edit 5` |
| `remove` | İşlem sil | `spendgrid remove 3` veya `spendgrid rm 3` |
| `rules` | Kural yönetimi | `spendgrid rules list`, `spendgrid rules add` |
| `sync` | Kuralları senkronize et | `spendgrid sync` |
| `complete` | Kural tamamla | `spendgrid complete` veya `spendgrid complete ID` |
| `uncomplete` | Tamamlama iptal | `spendgrid uncomplete` veya `spendgrid uncomplete ID` |
| `complete-month` | Ayın tümünü tamamla | `spendgrid complete-month` veya `spendgrid complete-month 2026-02` |
| `report` | Rapor al | `spendgrid report monthly` |
| `status` | Durum gör | `spendgrid status` |
| `plan` | Plan raporu | `spendgrid plan` |
| `exchange` | Kur işlemleri | `spendgrid exchange show` |
| `investments` | Yatırımlar | `spendgrid investments` |
| `pool` | Bekleyen işlemler | `spendgrid pool list` |
| `config` | Ayarlar | `spendgrid config list` |
| `validate` | Doğrulama | `spendgrid validate` |
| `last` | Son dizinler | `spendgrid last` |

---

## Detaylı Komut Açıklamaları

### 1. init - Veritabanı Başlatma

Yeni bir SpendGrid veritabanı oluşturur.

```bash
spendgrid init
```

**Ne yapar?**
- `.spendgrid` dizini oluşturur
- `_config/` alt yapılandırma dosyaları oluşturur
- `_pool/` backlog dizini oluşturur
- `_share/` paylaşım dizini oluşturur
- Mevcut yıl dizinini (örn: `2026/`) oluşturur
- Ay dosyalarını (`01.md` - `12.md`) oluşturur

**Örnek:**
```bash
mkdir ~/finans
cd ~/finans
spendgrid init
# ✓ Veritabanı başlatıldı
```

---

### 2. add - İşlem Ekleme

Yeni finansal işlem ekler. İki modu vardır.

#### İnteraktif Mod (Önerilen)

```bash
spendgrid add
```

**Adım adım:**
1. Gün sorar (varsayılan: bugün)
2. Açıklama sorar (boşluklu yazabilirsiniz!)
3. Tutar ve para birimi sorar
4. Etiketler sorar (otomatik tamamlama var)
5. Projeler sorar (otomatik tamamlama var)
6. Not sorar (opsiyonel)

**Örnek Diyalog:**
```
Gün [7]: 15
Açıklama: Market Alışverişi  <- BOŞLUKLU YAZABİLİRSİNİZ!
Tutar ve Para Birimi: 450.50 TRY
Etiketler: #market #gida
Proje: @ev
Not: Haftalık alışveriş

✓ İşlem eklendi
```

#### Direkt Mod

```bash
spendgrid add --direct "GÜN|AÇIKLAMA|TUTAR PARA|ETİKETLER"
```

**Format:**
- `GÜN` - Ayın günü (1-31)
- `AÇIKLAMA` - İşlem açıklaması
- `TUTAR PARA` - Tutar ve para birimi (örn: -450.50 TRY, 1000 USD)
- `ETİKETLER` - # ile etiketler

**Örnekler:**
```bash
# Basit gider
spendgrid add --direct "15|Market Alışverişi|-450.50 TRY|#market #gida"

# Gelir
spendgrid add --direct "05|Maaş|+25000 TRY|#maas #sirketA"

# Dövizli işlem
spendgrid add --direct "20|Freelance Ödeme|+1000 USD|#freelance #gelir"

# Manuel kur ile
spendgrid add --direct "10|AWS Fatura|-120 USD @35.50|#fatura #aws"
```

---

### 3. list - İşlemleri Listeleme

Mevcut ay veya belirli bir ayın işlemlerini listeler.

```bash
# Mevcut ayı listele
spendgrid list

# Belirli ayı listele (01-12)
spendgrid list 02
spendgrid list 5
```

**Çıktı Örneği:**
```
┌────┬────────────────────┬───────────┬──────────┬─────────────────┐
│ Gün│ Açıklama           │ Tutar     │ Para     │ Etiketler       │
├────┼────────────────────┼───────────┼──────────┼─────────────────┤
│ 01 │ Maaş               │ +25000.00 │ TRY      │ #maas @@sirket  │
│ 05 │ Market             │  -450.50  │ TRY      │ #market #gida   │
│ 10 │ Kira               │ -5000.00  │ TRY      │ #kira #ev       │
│ 15 │ Elektrik           │  -350.00  │ TRY      │ #fatura         │
└────┴────────────────────┴───────────┴──────────┴─────────────────┘
```

---

### 4. edit - İşlem Düzenleme

Belirli bir satırdaki işlemi düzenler.

```bash
spendgrid edit SATIR_NO
```

**Örnek:**
```bash
# Önce listeyi gör
spendgrid list
# 3. satırı düzenle
spendgrid edit 3
```

**Not:** `list` komutundaki satır numarasını kullanın.

---

### 5. remove / rm - İşlem Silme

İşlem siler. Kısayol: `rm`

```bash
spendgrid remove SATIR_NO
# veya
spendgrid rm SATIR_NO
```

**Örnek:**
```bash
spendgrid list
# 5. satırı sil
spendgrid rm 5
```

**Dikkat:** Bu işlem geri alınamaz!

---

### 6. rules - Kural Yönetimi

Tekrarlayan işlemler için kural sistemi.

#### rules list - Kuralları Listele

```bash
spendgrid rules list
```

**Çıktı:**
```
✓ [INC] Maaş | maa_1770358056 | +25000.00 TRY | Monthly day 5
✗ [EXP] Kira | kira_1770236512 | -5000.00 TRY | Monthly day 5
```

#### rules add - Kural Ekle

**İnteraktif:**
```bash
spendgrid rules add
```

**Direkt:**
```bash
spendgrid rules add "KURAL_ADI" TUTAR PARA_BİRİMİ TİP [opsiyonlar]
```

**Parametreler:**
- `KURAL_ADI` - Kural adı (tırnak içinde, boşluklu olabilir)
- `TUTAR` - Tutar (pozitif sayı)
- `PARA_BİRİMİ` - TRY, USD, EUR
- `TİP` - `income` (gelir) veya `expense` (gider)

**Opsiyonel Flag'ler:**
- `--day N` - Ayın günü (1-31, varsayılan: 1)
- `--tags "etiket1,etiket2"` - Etiketler
- `--project "proje"` - Proje adı
- `--start-date YYYY-MM` - Başlangıç tarihi
- `--end-date YYYY-MM` - Bitiş tarihi
- `--metadata "açıklama"` - Açıklama

**Örnekler:**
```bash
# Basit kira kuralı
spendgrid rules add "Ev Kira" 5000 TRY expense --day 5 --tags "kira,ev"

# Maaş kuralı
spendgrid rules add "Aylık Maaş" 25000 TRY income --day 5 --tags "maas"

# Proje ile
spendgrid rules add "Proje X Ödeme" 10000 TRY income \
  --day 10 --tags "proje,odeme" --project "musteriX"

# Dövizli
spendgrid rules add "Freelance USD" 1000 USD income --day 15

# Tarih aralıklı
spendgrid rules add "Staj Maaşı" 5000 TRY income \
  --day 5 --start-date 2026-06 --end-date 2026-08
```

---

### 7. sync - Manuel Senkronizasyon

Kuralları ay dosyalarına senkronize eder. (Normalde otomatik yapılır)

```bash
spendgrid sync
```

**Ne zaman kullanılır?**
- Yeni kural eklediniz ama ay dosyasında göremiyorsunuz
- Manuel müdahale sonrası kontrol

---

### 8. complete - Kural Tamamlama

Kuralı "gerçekleşmiş" olarak işaretler. `[ ]` → `[x]`

```bash
# Interaktif mod (önerilen)
spendgrid complete

# Direkt ID ile
spendgrid complete KURAL_ID
```

**İnteraktif Mod Akışı:**
```
Uncompleted Rules (last 10):
----------------------------------------------------------------------
 1. ☐ 05 | Ev Kira                     | kira_001 | -5000.00 TRY
 2. ☐ 15 | Elektrik Faturası           | elek_001 | -350.00 TRY
 3. ☐ 20 | Maaş                        | maa_001  | +25000.00 TRY
----------------------------------------------------------------------

Enter rule number (1-N) or ID (or press Enter to cancel): 3
✓ Rule 'maa_001' marked as completed
```

**Örnekler:**
```bash
# Listeyi gör ve seç
spendgrid complete

# Direkt ID ile
spendgrid complete maa_1770358056

# Tamamlanan tekrar gösterilmez!
```

---

### 9. uncomplete - Tamamlama İptali

Kuralı "gerçekleşmemiş" yapar. `[x]` → `[ ]`

```bash
# Interaktif mod
spendgrid uncomplete

# Direkt ID ile
spendgrid uncomplete KURAL_ID
```

**Kullanım:** Yanlışlıkla tamamlanan kuralı geri almak için.

---

### 10. complete-month - Toplu Tamamlama

Bir ayın tüm kurallarını topluca tamamlar. **Dikkatli kullanın!**

```bash
# Mevcut ay
spendgrid complete-month

# Belirli ay
spendgrid complete-month 2026-02
```

**Uyarı:** Bu komut kontrolsüz olarak tüm rule'ları `[x]` yapar. Banka hesabınızı kontrol ettikten sonra kullanın!

---

### 11. report - Raporlama

Finansal raporlar alır.

#### report monthly - Aylık Rapor

```bash
# Mevcut ay
spendgrid report monthly

# Belirli ay
spendgrid report monthly 2
spendgrid report monthly 02
```

**Çıktı:**
```
Aylık Rapor February 2026
======================================================================

📊 GERÇEKLEŞEN
----------------------------------------------------------------------
Currency                      Income         Expense
----------------------------------------------------------------------
TRY                        25000.00         5350.00
----------------------------------------------------------------------
TOTAL (TRY)                25000.00         5350.00
NET                        19650.00

📅 PLANLANAN (Tamamlanmamış Rule'lar)
----------------------------------------------------------------------
  ☐ 25 | Su Faturası              | -150.00 TRY

🔮 PROJEKSİYON (Gerçekleşen + Planlanan)
----------------------------------------------------------------------
PROJ. TOTAL                25000.00         5500.00
PROJ. NET                  19500.00
```

#### report yearly - Yıllık Rapor

```bash
spendgrid report yearly
```

#### report web - HTML Rapor

```bash
# Mevcut ay HTML
spendgrid report web

# Yıllık HTML
spendgrid report web --year
```

HTML dosyası `_share/` dizinine kaydedilir.

---

### 12. status - Durum Görüntüleme

Veritabanı durumunu özetler.

```bash
spendgrid status
```

**Çıktı:**
```
Durum Özeti
========================================

📅 Current Period: February 2026

📊 Completed Transactions:
   Total: 9 (Income: 8, Expense: 1)
   Total Income:  350600.00
   Total Expense: 18000.00
   Net:           332600.00

📅 Planned (Uncompleted Rules):
   Total: 3
   Expected Income:  350000.00
   Expected Expense: 18000.00
   Expected Net:     332000.00

🏷️ Categories:
   Active Tags: 5
   Active Projects: 3

⚙️ Rules:
   Active Rules: 8
```

---

### 13. plan - Planlama Raporu

Planlanan vs gerçekleşen karşılaştırması.

```bash
# Mevcut ay
spendgrid plan

# Belirli ay
spendgrid plan 02
```

**Çıktı:**
```
Plan Raporu - February 2026
========================================

GELİRLER:
Planlanan:      +35000.00 TRY
Gerçekleşen:    +25000.00 TRY
Fark:           -10000.00 TRY (Eksik)

GİDERLER:
Planlanan:      -8000.00 TRY
Gerçekleşen:    -5350.00 TRY
Fark:           +2650.00 TRY (Düşük)

NET:
Planlanan:      +27000.00 TRY
Gerçekleşen:    +19650.00 TRY
Fark:           -7350.00 TRY

⚠️ Tamamlanmamış Rule'lar:
  ☐ 20 | Maaş: +10000.00 TRY
  ☐ 25 | Kredi: -2000.00 TRY
```

---

### 14. exchange - Kur İşlemleri

Döviz kuru yönetimi.

#### exchange show - Kurları Gör

```bash
spendgrid exchange show
```

**Çıktı:**
```
Güncel Kurlar (2026-02-07)
========================================
USD/TRY: 35.50
EUR/TRY: 38.20
GBP/TRY: 45.10
```

#### exchange refresh - Kurları Güncelle

```bash
spendgrid exchange refresh
```

TCMB veya Frankfurt ECB API'den güncel kurları çeker.

#### exchange set - Manuel Kur Belirle

```bash
spendgrid exchange set 2026-02-07 USD 35.50
```

Belirli bir tarih için kur tanımlar.

---

### 15. investments - Yatırım Portföyü

Yatırım takibi (cost basis hesaplama).

```bash
spendgrid investments
```

**Önce yatırım ekleme:**
```bash
spendgrid add --direct "01|Altın Alımı|+5000TRY ALTIN(10gr * 500TRY)|#investment# #altin"
spendgrid add --direct "15|Hisse Alımı|+10000TRY TUPRS(100 * 100TRY)|#investment# #borsa"
```

**Çıktı:**
```
Yatırım Portföyü
========================================

Altın (ALTIN):
  Toplam: 10.00 gr
  Maliyet: 5000.00 TRY
  Ort. Birim: 500.00 TRY/gr

Tüpraş (TUPRS):
  Toplam: 100.00 adet
  Maliyet: 10000.00 TRY
  Ort. Birim: 100.00 TRY/adet

Toplam Portföy Değeri: 15000.00 TRY
```

---

### 16. pool - Bekleyen İşlemler

Tarihsiz bekleyen işlemler (backlog).

#### pool list - Listele

```bash
spendgrid pool list
```

#### pool add - Ekle

```bash
spendgrid pool add "Açıklama" TUTAR PARA [ETİKETLER]
```

#### pool move - Aya Taşı

```bash
spendgrid pool move POOL_ID AY
```

**Örnek:**
```bash
# Bekleyen işlem ekle
spendgrid pool add "Yıllık Abonelik" 500 TRY "#yakinda"

# Şubat ayına taşı
spendgrid pool move 1 02
```

---

### 17. config - Ayarlar

Yapılandırma yönetimi.

#### config list - Ayarları Listele

```bash
spendgrid config list
```

#### config get - Değer Oku

```bash
spendgrid config get language
```

#### config set - Değer Ata

```bash
# Dil değiştir
spendgrid config set language tr

# Temel para birimi
spendgrid config set base_currency TRY
```

---

### 18. validate - Doğrulama

Veritabanı doğrulaması yapar.

```bash
spendgrid validate
```

**Kontrol eder:**
- Dosya yapısı
- İşlem formatı
- Kural sözdizimi
- Bozuk satırlar

---

### 19. last - Son Dizinler

Son kullanılan SpendGrid dizinlerini gösterir.

```bash
spendgrid last
```

**Çıktı:**
```
Son Kullanılan Dizinler:
 1. ~/finans/2026
 2. ~/is/fatura
 3. ~/kisisel/butce
```

---

## Komut Zincirleri ve İş Akışları

### Günlük Akış

```bash
# Durumu kontrol et
spendgrid status

# Tamamlananları işaretle
spendgrid complete

# Rapor al
spendgrid report monthly
```

### Haftalık Akış

```bash
# Tüm işlemleri listele
spendgrid list

# Eksik var mı kontrol et
spendgrid validate

# Plan raporu al
spendgrid plan
```

### Ay Sonu Akış

```bash
# Tümünü kontrol et
spendgrid report monthly
spendgrid plan

# Tamamlanmayan var mı?
spendgrid complete

# Bir sonraki ay hazırlığı
spendgrid rules list
```

---

## Hızlı Referans Tabloları

### İşlem Formatı

```
- GÜN | AÇIKLAMA | TUTAR PARA [@KUR] | ETİKETLER | [META]
```

| Bileşen | Örnek | Açıklama |
|---------|-------|----------|
| Gün | `15` | 1-31 arası |
| Açıklama | `Market` | Boşluklu olabilir |
| Tutar | `-450.50` | Negatif: gider, Pozitif: gelir |
| Para | `TRY` | TRY, USD, EUR, GBP |
| Kur | `@35.50` | Opsiyonel, manuel kur |
| Etiket | `#market` | # ile başlar |
| Proje | `@ev` | @ ile başlar |
| Meta | `[NOTE:aciklama]` | Köşeli parantez içinde |

### Para Birimleri

| Kod | Açıklama |
|-----|----------|
| TRY | Türk Lirası |
| TL | Türk Lirası (TRY alias) |
| USD | Amerikan Doları |
| $ | USD alias |
| EUR | Euro |
| € | EUR alias |
| GBP | İngiliz Sterlini |

### Etiket Türleri

| Etiket | Kullanım |
|--------|----------|
| #investment# | Yatırım takibi (sistem) |
| #loan# | Kredi takibi (sistem) |
| #maas | Maaş gelirleri |
| #kira | Kira giderleri |
| #fatura | Fatura ödemeleri |
| #market | Market alışverişleri |
| #yatirim | Yatırım harcamaları |

---

**Son Güncelleme:** 2026-02-07  
**Versiyon:** 1.0.0
