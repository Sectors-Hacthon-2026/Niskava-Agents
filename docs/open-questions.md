# Daftar Pertanyaan Terbuka & Keputusan Tertunda (Open Questions)

**Status:** ACTIVE REGISTER  
**Versi Dokumen:** 1.2.0  
**Terakhir Diperbarui:** 2026-09-22  

Daftar isu desain, pertimbangan arsitektur, dan pertanyaan produk yang masih dalam tahap evaluasi. Dokumen ini diperbarui secara berkala hingga keputusan final dikunci ke dalam Architecture Decision Record (ADR).

---

## 1. Register Pertanyaan Terbuka

| ID | Topik & Pertanyaan | Status | Keputusan / Dampak Arsitektur | Target Keputusan |
|---|---|---|---|---|
| **OQ-01** | **Pilihan LLM Default untuk MVP**: Apakah mengunci Gemini 2.0 Flash (Cloud) sebagai default, atau menyediakan opsi fallback lokal via Ollama? | `RESOLVED` | **Keputusan:** Dikunci pada **Gemini 2.0 Flash** via API. Menjamin target latensi end-to-end < 6 detik (P95), kestabilan video demo, dan kemudahan evaluasi repositori oleh dewan juri tanpa perlu setup GPU lokal. | 16 Sep 2026 |
| **OQ-02** | **Parameter Ambang Batas Saham Lapis Tiga**: Apakah threshold $V_z \ge 2.5$ perlu dinaikkan untuk saham dengan market cap kecil (< IDR 1T) yang sering memiliki volatilitas liar harian? | `RESOLVED` | **Keputusan:** Threshold $V_z$ didukung secara adaptif via argumen deterministik `detect_historical_anomalies(..., volume_z_threshold=2.5)`. False positive saham lapis tiga disaring oleh verifikasi silang divergensi subsektor ($D_t$) dan korelasi komoditas global. | 22 Sep 2026 |
| **OQ-03** | **Integrasi Langsung PDF Keterbukaan Informasi BEI**: Apakah parser PDF langsung untuk dokumen pengumuman IDXnet perlu masuk ke scope MVP, atau cukup mengandalkan teks agregasi Sectors v2 News? | `RESOLVED` | **Keputusan:** Scope difokuskan pada **Sectors v2 News API (`/v2/news/`) & Corporate Filings** untuk menjamin data bersumber 100% dari Sectors Financial API sesuai aturan Rule 06 sebelum batas pembekuan kode 8 Oktober 2026. | 24 Sep 2026 |
| **OQ-04** | **Watchlist & Background Daemon Mode**: Apakah agen perlu menjalankan background loop untuk memonitor watchlist saham secara otomatis setiap penutupan bursa (16:00 WIB)? | `DEFERRED` | Ditaruh di roadmap pasca-hackathon (v2.0) agar tim fokus menyempurnakan alur investigasi on-demand per emiten di v1.0. | Pasca-Hackathon |
| **OQ-05** | **Strategi Antarmuka Web Utama (Embedded HTML vs React SPA)**: Bagaimana cara menjamin visual dashboard dapat dijalankan oleh juri tanpa hambatan instalasi Node/npm lokal? | `RESOLVED` | **Keputusan:** Mengadopsi strategi **Dual-Surface Web Delivery** (ADR-10): Go Core Daemon menyajikan dashboard interaktif Market Intelligence bawaan secara embedded di `http://localhost:20128` (zero npm dependency), didukung opsi extended React SPA di `clients/web/`. | 22 Sep 2026 |
| **OQ-06** | **Standardisasi Lokalisasi Dwibahasa (English & Indonesian)**: Apakah terminal dan output investigasi disajikan dalam Bahasa Indonesia atau Inggris untuk penjurian? | `RESOLVED` | **Keputusan:** Menerapkan arsitektur i18n dwibahasa penuh (ADR-09). TUI mendukung pengalihan bahasa instan via flag `--lang en/id` dan hotkey `L` di launcher menu. | 22 Sep 2026 |

---

## 2. Prosedur Penyelesaian

Ketika sebuah pertanyaan terbuka di atas disepakati solusinya:
1. Status diubah menjadi `RESOLVED`.
2. Keputusan dicatat ke dalam dokumen spesifikasi yang relevan atau dibuatkan ADR baru di `docs/90-decisions/`.
3. Tautan ADR dicantumkan pada tabel di atas sebagai riwayat keputusan.
