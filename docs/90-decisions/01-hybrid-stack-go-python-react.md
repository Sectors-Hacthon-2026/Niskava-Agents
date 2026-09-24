# 01 — Arsitektur Hybrid Tripartit (Go + Python + React)

**Status:** ACCEPTED  
**Tanggal:** 2026-09-16  
**Pengambil Keputusan:** Core Team  
**Menggantikan:** `docs/decisions/tech-stack-hybrid.md`  

---

## 1. Konteks & Permasalahan

Niskava Agent dirancang sebagai platform intelijen pasar otonom yang harus memenuhi tiga kebutuhan kontradiktif:
1. **Developer Experience & CLI Portabilitas**: Membutuhkan binary tunggal yang cepat, ringan, tanpa overhead instalasi runtime rumit, serta mendukung rendering TUI interaktif di terminal.
2. **Kekuatan Ekosistem Data Science & AI**: Membutuhkan ekosistem Python yang kaya untuk kalkulasi statistik (`numpy`), pemanggilan LLM, dan manipulasi teks scraping.
3. **Antarmuka Visual Modern**: Membutuhkan dashboard visual interaktif berbasis web untuk charting finansial (candlestick + markers), timeline kejadian, dan kartu bukti.

Jika dibangun 100% Python, distribusi binary tunggal (PyInstaller) lambat, berat (>100MB), dan performa server HTTP/SSE konkuren terbatas. Jika dibangun 100% Go, integrasi library prompt AI dan manipulasi data numerik kurang fleksibel dibandingkan ekosistem Python.

---

## 2. Keputusan

Menerapkan **arsitektur hybrid terpadu (Tripartite Hybrid Architecture)** dengan strategi repositori dan resolusi path yang dinamis:
* **Pemisahan Repositori & Struktur Monorepo**:
  * Repositori Dokumentasi & Spesifikasi (`niskava-docs`) berdiri sendiri sebagai Single Source of Truth (SSoT).
  * Repositori Implementasi Kode (`niskava-codebase`) dikelola dalam repositori polyglot monorepo dengan segregasi bersih antara antarmuka klien (`clients/`) dan layanan backend (`backend/`).
* **Clients Surface (`clients/`)**:
  * **Interactive CLI & TUI (`clients/cli/`)**: Menangani antarmuka terminal interaktif berbasis Bubbletea & Glamour, perintah Cobra (`investigate`, `serve`, `sessions`), dan wizard setup.
  * **Web Workspace (`clients/web/`)**: Dibangun dengan Vite + React 18 + Tailwind CSS + shadcn/ui. Dikompilasi menjadi aset statis dan di-embed ke dalam biner Go (`//go:embed clients/web/dist`) atau di-serve terpisah selama perancangan UI.
* **Backend Architecture (`backend/`)**:
  * **Go Core Daemon (`backend/core/` & `cmd/niskava/`)**: Berperan sebagai pintu masuk utama (*gateway*), server REST & SSE (`/api/chat/*`), persistensi SQLite murni Go (`modernc.org/sqlite` zero CGO), dan manajemen subprocess IPC.
  * **Python Agent Engine (`backend/engine/`)**: Dijalankan sebagai stateless child process on-demand via Subprocess IPC (JSON Lines). Menangani Universal Model-Agnostic ReAct loop, komputasi deterministik anomali kuantitatif (NumPy), Sectors v2 API client, Sectors News & Disclosure Engine, dan Local Graph Memory (`NetworkX`). Lokasi biner dan modul dapat dikonfigurasi dinamis via flag, env var, atau config file.

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **Full Python Monolith (CLI via Typer + Streamlit UI)** | Binary PyInstaller lambat startup-nya (>3 detik), konsumsi memori tinggi, Streamlit kurang fleksibel untuk custom dark-mode theme dan charting interaktif custom. |
| **Full Go Stack (CLI + Web + AI SDK Go)** | Ekosistem quant data science di Go terbatas; prompt tooling dan evaluasi AI di Python jauh lebih matang untuk kebutuhan hackathon yang dinamis. |
| **Electron Desktop App** | Terlalu berat (>150MB memory & disk), tidak ramah CLI bagi developer, berlawanan dengan filosofi single lightweight binary. |
| **Strict Hardcoded Paths Monorepo** | Mengunci lokasi modul secara kaku menyebabkan broken build saat direktori direfaktor atau diuji di lingkungan OS/CI yang berbeda. |

---

## 4. Konsekuensi

### Positif
* **Kemudahan Distribusi Pengguna**: Pengguna cukup menjalankan file biner `niskava`, langsung mendapatkan antarmuka CLI dan web dashboard lokal di `localhost:8080`.
* **Performa Tinggi**: Go menangani networking konkuren, goroutine streaming, dan persistensi database lokal dengan latensi mikrodetik.
* **Fleksibilitas Analitik**: Python dapat mengeksploitasi seluruh ekosistem data science terkini tanpa membebani biner utama.
* **Modularitas & Kemudahan Refaktor**: Struktur folder implementasi dapat diubah atau disesuaikan tanpa merusak kontrak data IPC atau arsitektur sistem.
* **Kebersihan Dokumentasi**: Repositori dokumen bebas dari polusi biner hasil kompilasi, file `.db`, dan dependensi vendor (`node_modules/`, `venv/`).

### Negatif / Kompromi yang Diterima
* **Persyaratan Python Runtime**: Mesin pengembang membutuhkan runtime Python 3.11+ yang terpasang di sistem host.
* **Manajemen IPC & Konfigurasi**: Harus merawat protokol komunikasi data JSON Lines antar proses (diatur formal di `20-architecture/04-ipc-and-api-contract.md`) dan mekanisme resolusi path dinamis.
