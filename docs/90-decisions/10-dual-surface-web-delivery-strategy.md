# 10 — Strategi Web Delivery Ganda: Embedded Go HTML Single-Binary Fallback vs Vite React SPA

**Status:** ACCEPTED  
**Tanggal:** 2026-09-22  
**Pengambil Keputusan:** Core Team  
**Dokumen Terkait:** [`01-hybrid-stack-go-python-react.md`](01-hybrid-stack-go-python-react.md), [`04-local-first-sqlite-storage.md`](04-local-first-sqlite-storage.md), [`../10-product/05-hackathon-strategy.md`](../10-product/05-hackathon-strategy.md), [`../docs/TEAM_CLIENTS_INTEGRATION_GUIDE.md`](../docs/TEAM_CLIENTS_INTEGRATION_GUIDE.md)

---

## 1. Konteks & Permasalahan

Dalam evaluasi Sectors Hackathon Indonesia 2026, dewan juri akan mengunduh dan menjalankan repositori di lingkungan lokal mereka. Kondisi lingkungan dewan juri bervariasi secara signifikan:
* **Risiko Dependensi Node/NPM**: Jika Web Dashboard mewajibkan instalasi `node`, `npm install`, dan `vite build` secara manual sebelum biner Go dapat berjalan, probabilitas kegagalan evaluasi akibat perbedaan versi Node.js atau kegagalan unduhan dependensi npm di mesin juri sangat tinggi.
* **Rubrik Usability 40% & Technical Depth 30%**: Juri mengharapkan pengalaman *"it just works"* dalam hitungan detik setelah mengkloning repositori.
* **Kebutuhan Visualisasi Kompleks**: Di sisi lain, visualisasi finansial tingkat lanjut (candlestick interaktif TradingView/Recharts, eksplorasi node graf) membutuhkan ekosistem frontend yang kaya seperti React dan Vite.

---

## 2. Keputusan Arsitektur

Niskava Agent mengadopsi strategi **Dual-Surface Web Delivery (Zero-Dependency Embedded Fallback + Decoupled Extended Vite SPA)**:

```
                                  [ Pengguna / Juri ]
                                           │
                                    niskava serve
                                           │
                                           ▼
                            ┌──────────────────────────────┐
                            │      Go Core Web Server      │
                            │      (localhost:20128)       │
                            └──────────────┬───────────────┘
                                           │
                    ┌──────────────────────┴──────────────────────┐
                    │                                             │
                    ▼                                             ▼
        [ Juri Tanpa Node.js ]                       [ Pengembang / Web Client ]
   Embedded Market Intelligence Dashboard                 Decoupled Vite + React 18 SPA
       (Built-in di server.go)                           (clients/web/)
  • Zero npm install / No build step            • Rich Interactive Candlestick Charts
  • Live SSE Chat Canvas & History              • Vis.js / Node Graph Exploration
  • Auto-launch via `niskava serve -o`          • Tailwind CSS + shadcn/ui components
```

### Karakteristik Desain:
1. **Tier 1: Embedded Single-Binary Market Intelligence Canvas (Bawaan `server.go`)**:
   * Server Go menyajikan Single-Page Web Canvas mandiri langsung dari memori biner pada root path `/`.
   * Mendukung koneksi live SSE (`/api/chat`), riwayat sesi multi-turn, preset pertanyaan cepat, tombol stop/abort, dan visualisasi graf interaktif Vis.js (`/graph`).
   * **Nol Instalasi Tambahan**: Berjalan seketika pada binary Go tanpa memerlukan Node.js atau file web eksternal apa pun di mesin pengguna.
2. **Tier 2: Decoupled Vite + React SPA (`clients/web/`)**:
   * Proyek frontend terpisah di `clients/web/` yang berkomunikasi ke daemon Go Core melalui kontrak REST & SSE (`/api/*`).
   * Dirancang untuk pengembangan antarmuka desktop kaya fitur secara independen oleh tim frontend tanpa mengganggu stabilitas biner Go Core.

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **Wajib `npm build` Sebelum Go Build** | Sangat berisiko bagi dewan juri yang tidak memiliki toolchain Node.js terkonfigurasi. Biner Go akan gagal menyajikan UI jika folder `dist` kosong. |
| **Pure CLI Only (Tanpa Web Workspace Sama Sekali)** | Kehilangan potensi nilai maksimal pada rubrik Usability 40% untuk visualisasi candlestick dan graf relasi bukti. |
| **Hosting Cloud Web Terpusat** | Melanggar Hukum 4 (Local-First Data Sovereignty) dan keterbatasan biaya operasional pasca-hackathon. |

---

## 4. Konsekuensi

### Positif
* **Keandalan 100% Saat Evaluasi Juri**: Juri cukup menjalankan satu biner `./bin/niskava serve --open` dan web dashboard langsung menyala seketika di browser.
* **Separation of Concerns**: Tim backend Go/Python dan tim frontend React dapat bekerja secara paralel tanpa saling memblokir build pipeline.
* **Kepatuhan Kriteria Track 1**: Menghadirkan antarmuka visual purpose-built yang estetik dan tangguh.

### Negatif / Kompromi
* Pembaruan skema UI perlu diselaraskan antara template embedded bawaan Go dengan komponen React mandiri.
