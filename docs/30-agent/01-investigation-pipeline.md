# 01 — Siklus Investigasi Otonom (7-Stage Pipeline)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.1.0  
**Terakhir Diperbarui:** 2026-09-16  
**Kepatuhan Lintasan:** Track 1 · AI Agents & Assistants (*Multi-Step Reasoning & Custom Orchestration*)  

Setiap sesi investigasi Niskava Agent dieksekusi melalui 7 tahapan terstruktur yang berurutan, menjamin penalaran berbasis hipotesis dan bukti empiris (*hypothesis-driven investigation*).

---

## 1. Diagram Alur Investigasi

```text
1. INITIATION ──▶ 2. SECTORS_BASELINE ──▶ 3. QUANT_ANOMALY
                                                 │
                         ┌───────────────────────┴───────────────────────┐
                         ▼                                               ▼
             [Anomali Terdeteksi]                              [Tidak Ada Anomali]
           4. GAP_DETECTION                                   4b. FUNDAMENTAL_ONLY
                         │                                               │
                         ▼                                               │
              5. OSINT_HARVEST                                           │
                         │                                               │
                         ▼                                               ▼
             6. EVIDENCE_CORRELATION ◀───────────────────────────────────┘
                         │
                         ▼
             7. SYNTHESIS_AND_STREAMING (Audit Trail & Reporting)
```

---

## 2. Keselarasan Pipeline dengan Syarat Kualifikasi Track 1

Arsitektur 7-Stage Pipeline dirancang secara khusus untuk memenuhi kriteria evaluasi Track 1 Sectors Hackathon:

| Kriteria Resmi Track 1 | Peran dalam 7-Stage Pipeline Niskava |
|---|---|
| **Multi-step reasoning flows** | Pipeline mengeksekusi urutan logis berjenjang: Inisiasi $\to$ Baseline $\to$ Anomali $\to$ Kesenjangan Bukti $\to$ Panen OSINT $\to$ Korelasi Kausalitas $\to$ Sintesis Temuan. |
| **Custom tool-use pipelines** | Memadukan modul komputasi kuantitatif deterministik NumPy, klien Sectors v2 API, dan crawler OSINT bertarget dalam satu alur terkoordinasi. |
| **Routing between data sources** | Membedakan secara tegas antara data angka primer (*Sectors Historical OHLCV & Financials*) dan konteks naratif eksternal (*IDXnet, Media Berita*). |
| **Memory & state management** | Database SQLite lokal menyimpan status sesi, tabel anomali, dokumen OSINT, serta relasi graf bukti (*evidence nodes*). |
| **Autonomous task execution** | Kueri pencarian berita dirumuskan secara mandiri oleh agen berdasarkan jendela tanggal anomali tanpa panduan manual pengguna. |
| **Purpose-built interface** | Streaming progres investigasi tahap demi tahap secara langsung ke terminal TUI (`Bubbletea`) dan Web UI (`SSE`). |

---

## 3. Rincian 7 Tahapan Pipeline

### Stage 1: Inisiasi Sesi (`INITIATION`)
* Pengguna memicu investigasi via CLI (`niskava investigate ANTM --days 30`) atau tombol di Web Workspace.
* Go Core membuat entitas sesi baru di database SQLite dengan status `PENDING`, kemudian memanggil Subprocess Python dengan parameter terkait.
* Status beralih ke `RUNNING`.

### Stage 2: Penarikan Data Dasar (`SECTORS_BASELINE`)
* Python client memeriksa tabel `sectors_cache` di SQLite lokal.
* Melakukan penarikan data komprehensif dari Sectors API v2:
  1. `GET /v2/daily/{symbol}/`: Deret waktu OHLCV 30–90 hari.
  2. `GET /v2/company/report/{symbol}/?sections=valuation,financials,peers`: Rasio fundamental inti dan valuasi industri.
  3. `GET /v2/foreign-flow/{symbol}/`: Deret aliran modal investor asing (*Net Foreign Inflow* IDR).
  4. `GET /v2/corporate-actions/{symbol}/`: Jadwal cum-date dividen, pemecahan saham (*stock split*), dan *rights issue*.
  5. `GET /v2/suspensions/?symbol={symbol}`: Riwayat suspensi bursa beserta tautan dokumen PDF resmi BEI.
* Seluruh data disimpan ke dalam cache lokal dan disiapkan untuk pemrosesan deterministik.

### Stage 3: Deteksi Anomali Deterministik (`QUANT_ANOMALY`)
* Skrip matematika mengevaluasi deret waktu secara deterministik menggunakan NumPy (tanpa LLM):
  * **Volume Z-Score ($V_z$):** Deviasi volume terhadap rata-rata bergerak 20 hari ($\mu_{20}$, $\sigma_{20}$).
  * **Abnormal Return ($R_t$) & Sector Divergence ($D_t$):** Deviasi imbal hasil saham terhadap rata-rata subsektor.
  * **Foreign Flow Anomaly ($F_z$):** Deteksi lonjakan akumulasi/distribusi modal asing tidak wajar ($\ge 2.5\sigma$).
* **Kriteria Pemicu Anomali:**
  $$\text{Trigger} \iff (V_z \ge 2.5) \lor (|R_t| \ge 5.0\%) \lor (|F_z| \ge 2.5)$$
* Jika anomali terdeteksi, catat record anomali di tabel `anomalies` dan lanjutkan ke Stage 4. Jika tidak, arahkan ke profiling fundamental standar (Stage 4b).

### Stage 4: Perumusan Kesenjangan Bukti (`GAP_DETECTION`)
* Agen merumuskan hipotesis investigasi spesifik pada jendela waktu temporal $T_{\text{anomaly}} \pm 2\text{ hari}$:
  > *"Terjadi anomali volume 3.84σ dan net buy asing +Rp111,3 Miliar pada 12 September 2026. Apakah terdapat aksi korporasi, transaksi insider, atau pengumuman resmi BEI di jendela waktu 10–13 September?"*

### Stage 5: Penelusuran Intelijen Eksternal (`OSINT_HARVEST`)
* Agen membentuk kueri pencarian bertarget secara deterministik pada jendela waktu temporal $T_{\text{anomaly}} \pm 2\text{ hari}$.
* Mengeksekusi penarikan bukti paralel via **Dual-Engine Harvester** ([`07-resilient-dual-engine-osint-architecture.md`](../90-decisions/07-resilient-dual-engine-osint-architecture.md)):
  1. **Sectors API v2 Suite**:
     * `GET /v2/news/?ticker={symbol}`: Berita pasar modal terkurasi.
     * `GET /v2/filings/?symbol={symbol}`: Catatan transaksi kepemilikan orang dalam (*insiders*) di sekitar tanggal anomali.
     * `GET /v2/suspensions/?symbol={symbol}`: Konfirmasi surat keputusan suspensi atau UMA bursa.
  2. **Google News RSS Engine**: Mengeksekusi dorking terarah untuk menangkap pelaporan keterbukaan informasi BEI tersindikasi (*IDX Channel, Kontan, Bisnis*) dan isu komoditas/hukum tanpa risiko pemblokiran ISP lokal.
* Mengekstrak isi teks artikel secara bersih menggunakan `trafilatura` (membuang elemen HTML bising/iklan).
* Membungkus hasil ekstraksi ke dalam blok terisolasi `<evidence_context>` guna mencegah *Indirect Prompt Injection* sebelum disajikan ke LLM.
* Menyimpan kandidat item bukti ke tabel `osint_cache` dan `evidence_items` pada database SQLite lokal.

### Stage 6: Korelasi Bukti & Evaluasi Kausalitas (`EVIDENCE_CORRELATION`)
* Agen mencocokkan stempel waktu (*timestamp*) dokumen berita, pengumuman bursa, dan transaksi insider terhadap waktu terjadinya lonjakan transaksi:
  * Berita/keterbukaan mendahului lonjakan volume $\rightarrow$ `LIKELY_CATALYST`.
  * Volume melonjak mendahului rilis berita $\rightarrow$ `PRECEDED_ANNOUNCEMENT`.
  * Tidak ada berita yang sesuai namun foreign flow melonjak $\rightarrow$ `FOREIGN_DRIVEN_UNEXPLAINED`.
  * Tidak ada berita maupun flow pendukung $\rightarrow$ `UNEXPLAINED_BY_NEWS`.
* Memberikan status verifikasi pada setiap temuan: `SUPPORTED`, `UNCERTAIN`, atau `CONTRADICTED`.

### Stage 7: Sintesis, Audit Trail & Streaming (`SYNTHESIS_AND_STREAMING`)
* Menyusun kartu temuan (*findings*), ringkasan naratif, dan timeline kronologis ke dalam database SQLite.
* Memancarkan event streaming via STDOUT (JSONL) ke Go Core, yang selanjutnya diteruskan ke terminal CLI dan Web Dashboard via Server-Sent Events (SSE).
* Status sesi diperbarui menjadi `COMPLETED`.
