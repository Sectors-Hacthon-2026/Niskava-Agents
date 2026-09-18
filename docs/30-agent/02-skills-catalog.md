# 02 — Katalog Metodologi Analisa Agen (Skills Catalog)

**Status:** ACCEPTED  
**Versi Dokumen:** `1.1.0`  
**Terakhir Diperbarui:** 2026-09-16  
**Dokumen Terkait:** [`01-investigation-pipeline.md`](01-investigation-pipeline.md), [`03-evidence-and-causality.md`](03-evidence-and-causality.md)

Dalam Niskava Agent, **Skill** adalah Standar Operasional Prosedur (SOP) analisa terstruktur yang dijalankan oleh agen secara deterministik atau orkestrasi LLM sesuai kondisi investigasi.

---

## 1. Katalog Skills Resmi

| Nama Skill | Pemicu (Trigger) | Sumber Data Utama | Output Kunci |
|---|---|---|---|
| `market-anomaly` | Permintaan investigasi umum / lonjakan volume | Sectors `/v2/daily/` & `/v2/foreign-flow/` | Tanggal anomali, nilai $V_z$, $F_z$, klasifikasi divergensi |
| `event-correlation` | Adanya anomali kuantitatif yang belum terjelaskan | Sectors `/v2/news/`, `/v2/suspensions/`, `/v2/corporate-actions/`, Google News RSS | Kronologi kejadian (timeline), audit trail PDF bursa, verifikasi kausalitas |
| `insider-flow` | Deteksi anomali volume atau dugaan transaksi orang dalam | Sectors `/v2/filings/` & `/v2/broker-summary-top/` | Pola transaksi direksi/komisaris, top akumulasi/distribusi broker |
| `company-research` | Profiling fundamental / evaluasi kesehatan emiten | Sectors Company Report (`/v2/company/report/`) | Rasio profitabilitas, tren laba bersih, peer ranking |
| `financial-health` | Penurunan kinerja drastis atau rumor kebangkrutan | Sectors Balance Sheet & Cash Flow | Quick ratio, DER, deteksi *red flags* keuangan |
| `peer-comparison` | Analisis industri & perbandingan saham sejenis | Sectors Subsector Peers (`/v2/subsector/`) | Valuasi relatif (PE/PBV band), pangsa pasar subsektor |

---

## 2. Rincian Prosedur Standar (SOP) Setiap Skill

### A. Skill: `market-anomaly`
* **Tujuan**: Mengisolasi tanggal dan besaran anomali transaksi (volume, harga, dan arus modal asing) secara deterministik tanpa bias narasi.
* **Input**: `symbol` (str), `timeframe_days` (int, default: 30).
* **Alur Analisa**:
  1. Ambil deret waktu OHLCV harian dari `/v2/daily/{symbol}/`.
  2. Ambil deret Net Foreign Inflow dari `/v2/foreign-flow/{symbol}/`.
  3. Hitung rata-rata bergerak 20 hari ($\mu_{20}$) dan deviasi standar volume ($\sigma_{20}$) menggunakan NumPy.
  4. Hitung Z-Score volume ($V_z = \frac{V_t - \mu_{20}}{\sigma_{20}}$) dan abnormal return ($R_t$).
  5. Hitung Z-score arus modal asing ($F_z$) untuk mendeteksi lonjakan akumulasi/distribusi dana asing.
  6. Filter tanggal anomali: $(V_z \ge 2.5) \lor (|R_t| \ge 5.0\%) \lor (|F_z| \ge 2.5)$.
* **Keluaran Terstruktur**: Objek anomali JSON berisi metrik kuantitatif terverifikasi untuk diteruskan ke `event-correlation`.

---

### B. Skill: `event-correlation`
* **Tujuan**: Menemukan dan memvalidasi katalis informasi eksternal terhadap anomali transaksi.
* **Input**: `symbol`, `anomaly_date`, `divergence_pct`.
* **Alur Analisa**:
  1. Periksa catatan suspensi resmi bursa di `/v2/suspensions/?symbol={symbol}`. Jika ada pengumuman UMA atau suspensi bursa pada rentang tanggal, ekstrak nomor pengumuman dan URL PDF resmi BEI.
  2. Periksa jadwal aksi korporasi di `/v2/corporate-actions/{symbol}/` (apakah bertepatan dengan cum-date dividen atau rights issue).
  3. Generate 3 kueri pencarian bertarget pada jendela waktu $[T-2, T+1]$ dan ambil artikel dari Sectors Unified News (`/v2/news/`) serta Google News RSS Engine.
  4. Ekstrak entitas penting (nama fasilitas smelter, nilai kontrak, mitra bisnis, regulasi) menggunakan `trafilatura`.
  5. Cocokkan stempel waktu publikasi berita dengan waktu lonjakan transaksi untuk menilai kausalitas:
     * Berita mendahului lonjakan volume $\rightarrow$ `LIKELY_CATALYST`
     * Volume melonjak mendahului rilis berita $\rightarrow$ `PRECEDED_ANNOUNCEMENT`
     * Tanpa berita pemicu $\rightarrow$ `UNEXPLAINED_BY_NEWS`
* **Keluaran Terstruktur**: Daftar temuan (*findings*) dengan label `SUPPORTED`, `UNCERTAIN`, atau `CONTRADICTED`.

---

### C. Skill: `insider-flow` (Bandarmology & Insider Intelligence)
* **Tujuan**: Menyelidiki apakah anomali perdagangan dipicu atau disertai oleh transaksi kepemilikan orang dalam (*insiders*) atau broker institusi besar.
* **Input**: `symbol`, `start_date`, `end_date`.
* **Alur Analisa**:
  1. Panggil `/v2/filings/?symbol={symbol}` untuk mendeteksi transaksi beli/jual oleh direksi, komisaris, atau pemegang saham $\ge 5\%$.
  2. Panggil `/v2/broker-summary-top/{symbol}/` untuk mengetahui 3 broker pembeli bersih teratas (*top buyers*) dan penjual bersih (*top sellers*).
  3. Cocokkan dengan direktori `/v2/broker-registry/` untuk mengidentifikasi apakah akumulasi didominasi oleh broker asing, institusi domestik, atau ritel.
  4. Jika direksi melakukan akumulasi besar sebelum volume publik melonjak, hasilkan temuan berkategori `PRECEDED_ANNOUNCEMENT` dengan skor kepercayaan 0.95.
* **Keluaran Terstruktur**: Matriks kepemilikan orang dalam dan ringkasan konsentrasi akumulasi broker.

---

### D. Skill: `company-research` & `financial-health`
* **Tujuan**: Menguji kesehatan struktur keuangan emiten untuk membantah rumor gagal bayar atau mengidentifikasi risiko laten.
* **Input**: `symbol`.
* **Alur Analisa**:
  1. Ambil metrik neraca dan laba rugi dari `/v2/company/report/{symbol}/?sections=valuation,financials`.
  2. Hitung rasio kas lancar terhadap kewajiban jangka pendek (Quick Ratio & Current Ratio).
  3. Evaluasi tren Debt-to-Equity Ratio (DER) dan Interest Coverage Ratio.
  4. Jika rumor pasar menyatakan emiten gagal bayar tetapi rasio kas > 2.0x dan DER sehat, hasilkan temuan berlabel `CONTRADICTED`.
* **Keluaran Terstruktur**: Kartu kesehatan finansial (*Financial Health Matrix*) dan verifikasi klaim pasar.
