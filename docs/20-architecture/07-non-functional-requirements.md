# 07 — Persyaratan Non-Fungsional (Non-Functional Requirements)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

Dokumen ini mendefinisikan batasan teknis (*engineering constraints*), anggaran latensi (*latency budgets*), serta standar kualitas performa sistem Niskava Agent.

---

## 1. Anggaran Latensi Eksekusi (Latency Budgets)

Untuk memberikan pengalaman terminal dan web yang responsif, setiap tahap investigasi memiliki batas waktu (*time budget*):

| Tahapan Operasi | Target P95 | Batas Maksimum | Keterangan |
|---|---|---|---|
| **Sectors API Baseline Fetch** | 300 ms | 1.500 ms | Dipercepat hingga < 10 ms jika terkena cache SQLite lokal. |
| **Kalkulasi Anomali Deterministik** | 15 ms | 50 ms | Eksekusi berbasis matematika murni `numpy`. |
| **Targeted OSINT & News Fetch** | 1.800 ms | 4.000 ms | Agregasi berita Sectors v2 & pencarian terarah. |
| **LLM Reasoning & Evidence Synthesis** | 2.500 ms | 5.000 ms | Menggunakan model inferensi cepat (Gemini 2.0 Flash). |
| **Total Waktu Investigasi End-to-End** | **< 6.000 ms** | **12.000 ms** | Dari penekanan Enter di CLI hingga hasil muncul lengkap. |

---

## 2. Jejak Komputasi & Efisiensi Sumber Daya (Footprint)

1. **Ukuran Biner Mandiri (Single Executable)**:
   * Target ukuran file biner Go (`niskava`): **$\le 35$ MB** (sudah mencakup web assets `web/dist` yang di-embed via `//go:embed`).
2. **Konsumsi Memori (RAM Footprint)**:
   * Status Siaga (*Idle Web Server*): $\le 30$ MB RAM.
   * Status Aktif Menjalankan Investigasi (*Peak Load*): $\le 150$ MB RAM.
3. **Penyimpanan Lokal (SQLite Database)**:
   * Estimasi konsumsi disk: $\approx 2$ MB per 50 sesi investigasi lengkap (termasuk timeline, anomali, dan cache respons Sectors).

---

## 3. Portabilitas & Ketergantungan Sistem (Zero External Dependencies)

* **Arsitektur CPU**: Dukungan penuh untuk `linux/amd64`, `linux/arm64`, dan `darwin/arm64` (Apple Silicon).
* **Bebas Kompiler CGO**: Database SQLite wajib menggunakan `modernc.org/sqlite` (murni Go) sehingga biner dapat dikompilasi silang (*cross-compiled*) tanpa membutuhkan `gcc` atau library C eksternal pada mesin host pengguna.
* **Python Runtime**: Engine Python berjalan dalam virtual environment lokal (`.venv`) dengan dependensi minimal: `requests`, `numpy`, `google-genai`.

---

## 4. Keandalan & Penanganan Kegagalan (Resilience & Fallback)

* **Sectors API Outage**: Jika koneksi ke Sectors terputus, sistem otomatis memeriksa cache SQLite lokal. Jika cache tidak tersedia, sistem memberikan pesan kesalahan ramah (*graceful degradation*) tanpa membuat program crash.
* **LLM Rate Limit**: Jika kuota inferensi LLM mencapai batas, sistem menyajikan temuan anomali kuantitatif terlebih dahulu disertai penanda bahwa analisis kualitatif tertunda.
