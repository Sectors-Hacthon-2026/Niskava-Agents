# 07 — Arsitektur OSINT Tangguh Berbasis Dual-Engine Precision Harvester

**Status:** ACCEPTED  
**Tanggal:** 2026-09-16  
**Pengambil Keputusan:** Core Team  
**Dokumen Terkait:** `docs/20-architecture/06-osint-engine.md`, `docs/30-agent/01-investigation-pipeline.md`, `docs/90-decisions/03-sectors-v2-and-credit-conservation.md`

---

## 1. Konteks & Permasalahan

Dalam menginvestigasi anomali transaksi pada Bursa Efek Indonesia (IDX), Niskava Agent bertugas menutup kesenjangan informasi (*Evidence Gap*) antara fakta kuantitatif (Sectors API) dan konteks kualitatif (berita, keterbukaan informasi emiten, dan pengumuman regulator).

Namun, audit empiris terhadap jaringan internet Indonesia dan infrastruktur pasar modal mengungkap kendala fatal:
1. **Pemblokiran ISP / Kominfo:** DuckDuckGo (`duckduckgo.com`), yang lazim digunakan sebagai *headless search library* (`duckduckgo_search` / `ddgs`), diblokir secara resmi oleh Kementerian Kominfo RI sejak Juli 2024 pada DNS Trust Positif. Script yang mengandalkannya akan langsung melempar error *Name or service not known*.
2. **Proteksi WAF Portal Bursa:** Situs resmi Bursa Efek Indonesia (`idx.co.id`) menggunakan Cloudflare/Akamai bot-management dengan TLS fingerprinting. Seluruh *direct HTTP client* Python (`requests`, `httpx`, `urllib`) ditolak dengan kode `HTTP 403 Forbidden`.
3. **Kerapuhan Sinyal Media Sosial:** Scraping unauthenticated ke Stockbit Stream dan X (Twitter) sangat tidak stabil, melanggar ToS, membutuhkan login akun pribadi, dan rentan terhadap pemblokiran IP mendadak pada saat demo langsung ke dewan juri.
4. **Kepatuhan Kompetisi:** Regulasi Hackathon Sectors 2026 (Rule 06) mewajibkan Sectors API v2 sebagai pilar data inti, bukan pemanggilan dekoratif.

---

## 2. Keputusan

Tim memutuskan untuk mengadopsi arsitektur **Dual-Engine Precision Harvester** yang tangguh, legal, bebas biaya langganan, dan kebal sensor lokal:

1. **Dual-Engine Harvesting Pattern:**
   * **Engine 1 (Inti Terstruktur):** Sectors Financial API v2 Unified News (`GET /v2/news/?ticker={ticker}`). Memenuhi syarat kepatuhan hackathon dan menyediakan arsip berita terkurasi tanpa scraping.
   * **Engine 2 (Pencarian Bertarget Terbuka):** Google News RSS Search Engine (`news.google.com/rss/search?q=...&hl=id&gl=ID&ceid=ID:id`). 100% aktif di Indonesia, gratis, tanpa API key, mengembalikan XML terstruktur dengan judul, sumber berita finansial terakreditasi (*Kontan, Bisnis, CNBC Indonesia, IDX Channel*), dan timestamp presisi detik.
2. **Strategi Secondary Disclosure Dorking:**
   * Alih-alih membongkar proteksi Cloudflare `idx.co.id` yang rapuh, agen mengumpulkan keterbukaan informasi bursa melalui pelaporan tersindikasi pada *IDX Channel*, *Kontan*, dan *EmitenNews* yang mempublikasikan nomor surat resmi BEI dan kutipan direksi dalam hitungan menit. Item bukti tetap diklasifikasikan sebagai **Tier 1 Official Disclosure**.
3. **Pembersihan Deterministik & Isolasi Prompt Injection:**
   * Menggunakan library Python `trafilatura` untuk mengekstrak isi teks murni dan membuang elemen web bising (iklan, form, banner).
   * Membungkus teks mentah ke dalam format terisolasi `<evidence_context>` dengan blok CDATA dan larangan eksplisit eksekusi instruksi bagi LLM (*Indirect Prompt Injection Defense*).
4. **Penyedia Fleksibel (Pluggable) & Mode Mock Offline:**
   * Menyediakan antarmuka penyedia pencarian (*search provider*) yang dapat beralih ke Tavily API atau SearXNG jika diinginkan pengguna.
   * Menyediakan mode `MOCK_OSINT=1` menggunakan *fixture* JSON lokal (`tests/fixtures/osint/`) untuk pengujian unit deterministik dan jaminan demo hackathon 100% sukses tanpa koneksi internet.
5. **Penyimpanan Lokal & Caching SQLite:**
   * Hasil ekstraksi bukti disimpan di tabel `osint_cache` dan `evidence_items` pada SQLite lokal (`~/.niskava/niskava.db`), memastikan kedaulatan data dan performa sub-milidetik untuk pencarian berulang.

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **DuckDuckGo Search (`ddgs`)** | Diblokir oleh Kominfo RI di Indonesia; langsung *crash* saat dijalankan di jaringan lokal. |
| **Direct Scraper `idx.co.id`** | Terblokir Cloudflare WAF (HTTP 403); memicu ketergantungan *captcha bypass* yang ilegal dan rapuh. |
| **Headless Browser (Selenium / Playwright)** | Menambah bobot instalasi ratusan megabyte (Chromium engine), startup lambat (5-10 detik), dan merusak prinsip *single executable / lightweight*. |
| **Scraping Stockbit / Twitter X** | Berisiko tinggi terkena IP ban, melanggar ToS, dan API resmi X mematok biaya mahal ($100+/bulan). |
| **Hanya Mengandalkan Sectors News v2** | Tidak cukup untuk menangkap detail nomor surat klarifikasi bursa spesifik atau rumor industri di luar basis data Sectors. |

---

## 4. Konsekuensi

### Positif
* **Resiliensi Ekstrem:** Mesin pencari OSINT 100% lolos sensor ISP Indonesia dan kebal pemblokiran bot Cloudflare.
* **Nol Biaya Tambahan:** Google News RSS dan Trafilatura sepenuhnya gratis tanpa membutuhkan kredit API eksternal.
* **Performa Cepat:** Penarikan feed XML berlangsung dalam tempo $< 1$ detik.
* **Aman dari Prompt Injection:** Struktur `<evidence_context>` mengisolasi data web dari instruksi penalaran agen.
* **Siap Demo Offline:** Mode `MOCK_OSINT=1` memberikan perlindungan total jika jaringan venue hackathon mengalami kendala.

### Negatif / Kompromi yang Diterima
* Membutuhkan normalisasi format tanggal RSS (RFC 822 / GMT) ke standar ISO 8601 WIB (+07:00).
* Keterbukaan informasi BEI bergantung pada pelaporan cepat media finansial tersindikasi (IDX Channel / Kontan) daripada akses dokumen PDF mentah langsung di server bursa.
