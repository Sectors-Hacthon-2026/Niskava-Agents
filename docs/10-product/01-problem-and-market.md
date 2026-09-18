# 01 — Lanskap Masalah & Potensi Pasar

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

---

## 1. Lanskap Masalah di Pasar Modal Indonesia (IDX)

Pasar modal Indonesia memiliki karakteristik unik yang membedakannya dari bursa negara maju:
1. **Dominasi Investor Ritel yang Terfragmentasi**: Terdapat lebih dari 12 juta Single Investor Identification (SID) di Indonesia, namun mayoritas mengandalkan grup percakapan media sosial (Telegram, WhatsApp, forum Stockbit) yang dipenuhi spekulasi dan informasi tak terverifikasi.
2. **Keterpisahan Ekstrem Data Kuantitatif dan Kualitatif**:
   * Data pergerakan harga/volume tersedia di terminal broker atau aplikasi trading.
   * Data aksi korporasi dan laporan keuangan resmi berada di portal terpisah (IDXnet, KSEI).
   * Berita pasar modal tersebar di berbagai media finansial independen.
   * Tidak ada jembatan otomatis yang menghubungkan *"kapan volume melonjak"* dengan *"dokumen resmi apa yang mendasarinya"*.
3. **Waktu Respons Analisis yang Terlalu Lambat**:
   * Saat saham tertentu mengalami lonjakan tajam, seorang analis atau trader membutuhkan waktu 30 hingga 60 menit untuk:
     1. Memeriksa chart candlestick dan mengukur volume.
     2. Menarik data transaksi broker (*broker summary/bandarmology*).
     3. Mencari pengumuman keterbukaan informasi di IDX.
     4. Menelusuri portal berita terpercaya untuk memverifikasi katalis.
   * Di pasar yang bergerak cepat, keterlambatan 30 menit ini berakibat pada hilangnya peluang atau terjebak dalam aksi beli di pucuk (*buying at the top*).
4. **Bahaya Hukum Saran Finansial Ilegal**:
   * Banyak produk AI finansial memberikan saran beli/jual (*buy/sell*) tanpa izin penasihat investasi OJK, yang menimbulkan risiko regulasi besar bagi pengembang dan risiko kerugian modal bagi pengguna.

---

## 2. Keunggulan Data Sectors (The Sectors Moat)

Niskava Agent dirancang secara khusus untuk mengeksploitasi keunggulan data komparatif dari **Sectors Financial API v2**:

* **Cakupan Saham IDX Lengkap & Terstruktur**: Seluruh emiten di BEI memiliki data time series historis yang bersih dan siap diolah algoritma.
* **>500 Metrik Finansial & Standarisasi Laporan**: Rasio profitabilitas, likuiditas, leverage, serta laporan laba rugi terstandarisasi.
* **Fitur Khas Pasar Modal Indonesia di API v2**:
  * **Bandarmology & Broker Flow**: Data agregasi transaksi broker harian untuk membedakan transaksi ritel vs institusi/asing.
  * **Mining Extension**: Metrik operasional komoditas (produksi emas/nikel/batu bara, cadangan tambang) yang krusial bagi emiten sektor bahan baku (seperti `ANTM`, `PTBA`, `INCO`).
  * **Unified News API**: Agregasi berita terkurasi yang memangkas kebutuhan web scraping bebas yang boros bandwidth.
  * **Natural Language Screener**: Kemampuan menyaring emiten berdasarkan kueri semantik pasar.

---

## 3. Ukuran Pasar & Validasi Permintaan

* **Segmen Primer (Immediate Users)**:
  * Lebih dari 12 juta investor ritel terdaftar di KSEI yang membutuhkan alat bantu validasi berita instan.
  * Komunitas trading aktif yang memerlukan verifikasi rumor pasar agar tidak terjebak spekulasi gorengan saham.
* **Segmen Sekunder (Institutional & Research)**:
  * Analis junior di perusahaan sekuritas dan manajemen aset (asset management) yang menghabiskan 40% jam kerja mereka untuk mengumpulkan data latar belakang emiten (*preliminary background gathering*).
  * Jurnalis finansial yang memerlukan pengecekan silang fakta sebelum mempublikasikan berita pasar modal.
