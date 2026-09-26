# Panduan Integrasi Frontend & CLI — Niskava Agent

> **Single Source of Truth (SSoT) untuk Pengembang Klien: Nabil (Web Workspace) & Agung (CLI/TUI)**  
> Dokumen ini memuat spesifikasi endpoint Backend (BE), kontrak data, urutan event streaming SSE/IPC, aturan non-negotiable hackathon, serta tata kelola state input (*thinking lock & abort*).

---

## 📌 Daftar Isi
1. [Daftar Aturan Wajib (Non-Negotiable Invariants)](#1-daftar-aturan-wajib-non-negotiable-invariants)
2. [Panduan Khusus Nabil (Web Workspace — `clients/web/`)](#2-panduan-khusus-nabil-web-workspace--clientsweb)
   - [A. Stack & Desain Estetika](#a-stack--desain-estetika)
   - [B. Daftar Endpoint REST API](#b-daftar-endpoint-rest-api)
   - [C. Protokol SSE Streaming (`POST /api/chat`)](#c-protokol-sse-streaming-post-apichat)
   - [D. Aturan UX: Input Locking & Tombol Abort](#d-aturan-ux-input-locking--tombol-abort)
   - [E. Komponen UI Wajib](#e-komponen-ui-wajib)
3. [Panduan Khusus Agung (CLI & TUI Terminal — `clients/cli/`)](#3-panduan-khusus-agung-cli--tui-terminal--clientscli)
   - [A. Stack & Arsitektur CLI](#a-stack--arsitektur-cli)
   - [B. Perintah CLI yang Tersedia](#b-perintah-cli-yang-tersedia)
   - [C. Penyesuaian REPL Interaktif (`repl.go`)](#c-penyesuaian-repl-interaktif-replgo)
   - [D. Tata Kelola Interupsi (`Ctrl + C`)](#d-tata-kelola-interupsi-ctrl--c)
4. [Tabel Taksonomi Bukti & Warna UI](#4-tabel-taksonomi-bukti--warna-ui)
5. [Checklist Kesiapan Sebelum Demo/Submit](#5-checklist-kesiapan-sebelum-demosubmit)

---

## 1. Daftar Aturan Wajib (Non-Negotiable Invariants)

Semua komponen UI/CLI yang Anda buat **WAJIB** tunduk pada 6 Hukum Niskava & Aturan Resmi Hackathon:

* **Law 1: Deterministic Before Generative (P2)**: Jangan pernah menghitung persentase return, moving average, atau Z-Score di Frontend. Tampilkan murni angka dari backend (`z_score`, `baseline_value`, `metric_value`).
* **Law 2: Strict Financial Non-Advisory Boundary (P3 / Rule 12)**:
  * **DILARANG KERAS** menampilkan tombol, tag, atau label rekomendasi **BUY / SELL / HOLD** atau target harga.
  * Setiap bukti investigasi **WAJIB** diklasifikasikan ke dalam 3 tier: `[SUPPORTED]`, `[UNCERTAIN]`, atau `[CONTRADICTED]`.
  * Setiap layar (Web footer & CLI exit) **WAJIB** menampilkan teks disclaimer resmi:
    > *"Niskava Agent adalah platform intelijen pasar modal otonom untuk Bursa Efek Indonesia (IDX), BUKAN penasihat investasi atau broker berizin. Seluruh temuan disajikan secara deskriptif untuk tujuan riset verifikasi fakta dan BUKAN rekomendasi investasi."*
* **Law 3: Zero Broker Execution (Rule 06)**: Tidak ada fitur beli/jual atau integrasi akun sekuritas.
* **Law 4: Local-First Sovereignty (P5)**: Backend berjalan lokal di mesin pengguna (`http://localhost:20128` secara default), database lokal di `~/.niskava/niskava.db`.

---

## 2. Panduan Khusus Nabil (Web Workspace — `clients/web/`)

Status saat ini: `clients/web/src/App.tsx` masih berupa placeholder kosong (`return null;`). Tugas Nabil adalah membangun antarmuka web interaktif berbasis React.

### A. Stack & Desain Estetika
* **Framework**: React 18 + Vite + TypeScript.
* **Styling**: Tailwind CSS + `shadcn/ui`.
* **Theme / Palette**: **Market Intelligence / Bloomberg Terminal Dark Mode**:
  * Background Utama: `#090D16` / `#0B0F17` (Deep Cyber Black)
  * Card / Panel: `#111827` (Dark Slate) dengan border `#1E293B`
  * Accent Utama (Niskava Green): `#22C55E` / `#4ADE80` (Terminal Matrix)
  * Warning / Uncertain: `#FACC15` (Amber Gold)
  * Danger / Anomaly / Contradicted: `#EF4444` / `#F87171` (Crimson Red)
  * Font: Monospace (`JetBrains Mono` / `Fira Code`) untuk metrik/ticker, Inter / Sans-serif untuk narasi.

---

### B. Daftar Endpoint REST API
Base URL Backend: `http://localhost:20128` (atau dynamic dari window location).

| Method | Endpoint Path | Kegunaan | Request Body | Format Response |
|---|---|---|---|---|
| `GET` | `/api/health` | Status daemon, SQLite, & provider AI | `-` | `{"status": "ok", "version": "1.0.0", "market": "IDX"}` |
| `GET` | `/api/chat/sessions` | Mengambil daftar sesi chat | `?limit=50&offset=0&q={search}` | `{"sessions": [...], "total": N}` |
| `POST` | `/api/chat/sessions` | Membuat sesi baru | `{"title": "Analisis ANTM", "model": "hermes"}` | `{"status": "created", "session": {...}}` (201) |
| `GET` | `/api/chat/sessions/:id` | Detail metadata sesi | `-` | `{"session": {...}}` |
| `PATCH` | `/api/chat/sessions/:id` | Rename / Pin / Ubah status sesi | `{"title": "Nama Baru", "is_pinned": true}` | `{"status": "updated", "session": {...}}` |
| `DELETE`| `/api/chat/sessions/:id` | Hapus sesi & riwayat pesan | `-` | `{"status": "deleted", "session_id": "..."}` |
| `GET` | `/api/chat/sessions/:id/messages` | Ambil riwayat chat lengkap dalam sesi | `?limit=100` | `{"messages": [...], "total": N}` |
| `POST` | `/api/chat/sessions/:id/fork` | Cabangkan percakapan ke sesi baru | `{"title": "Fork Sesi", "up_to_message_id": "..."}` | `{"status": "forked", "session": {...}}` (201) |
| `POST` | `/api/chat/sessions/:id/reset` | Kosongkan pesan dalam sesi | `-` | `{"status": "reset", "session_id": "..."}` |
| `POST` | `/api/chat/sessions/:id/abort` | **Batalkan ReAct yang sedang berjalan** | `-` | `{"status": "aborted", "session_id": "..."}` |
| `GET` | `/api/chat/sessions/:id/export` | Unduh transkrip laporan | `?format=markdown` atau `?format=json` | File attachment (.md atau .json) |
| `GET` | `/api/chat/search` | Pencarian teks pesan global | `?q={keyword}&limit=20` | `{"results": [...], "count": N}` |
| `GET` | `/api/graph/data` | Node & edge graf memori lokal | `?session_id={id}&radius=2` | `{"nodes": [...], "edges": [...]}` |

---

### C. Protokol SSE Streaming (`POST /api/chat`)
Kirim request chat via HTTP POST:
```http
POST /api/chat
Content-Type: application/json

{
  "prompt": "apakah ada anomali volume pada saham ANTM?",
  "session_id": "CHAT-20260920-0001"
}
```
Backend akan merespons dengan header `Content-Type: text/event-stream`. 

#### Urutan Event yang Dialirkan:
1. **`event: thought`** $\to$ Pemikiran analitik / reasoning step agent (tampilkan di laci/collapsible "Reasoning Steps"):
   ```json
   data: {"thought": "Menganalisis prompt pengguna. Memeriksa apakah ANTM mengalami anomali kuantitatif..."}
   ```
2. **`event: tool_call`** $\to$ Eksekusi tool deterministik / skill:
   ```json
   data: {"tool": "compute_quant_anomalies", "args": {"ticker": "ANTM", "volume_z_threshold": 2.5}}
   ```
3. **`event: observation`** $\to$ Hasil eksekusi data mentah tool:
   ```json
   data: {"tool": "compute_quant_anomalies", "summary": "Ditemukan lonjakan volume Z=3.84σ pada 2026-09-10"}
   ```
4. **`event: anomaly_detected`** $\to$ Deteksi anomali numerik:
   ```json
   data: {"ticker": "ANTM", "anomaly_date": "2026-09-10", "metric_type": "VOLUME_AND_PRICE_SURGE", "z_score": 3.84, "metric_value": 145000000.0, "baseline_value": 32000000.0}
   ```
5. **`event: finding_emitted`** $\to$ Temuan bukti terverifikasi (tampilkan sebagai kartu temuan):
   ```json
   data: {"title": "Katalis Uji Coba Smelter Baru", "claim": "Pengumuman smelter mendahului lonjakan volume", "verification_status": "SUPPORTED", "confidence_score": 0.95}
   ```
6. **`event: message_chunk`** $\to$ Streaming per kata/token narasi jawaban:
   ```json
   data: {"chunk": "Berdasarkan "}
   ```
7. **`event: message_complete`** $\to$ Pesan utuh selesai disimpan di SQLite:
   ```json
   data: {"content": "...", "session_id": "..."}
   ```
8. **`event: done`** $\to$ Akhir dari stream, status sesi kembali ke `IDLE`:
   ```json
   data: {"status": "COMPLETED", "session_id": "..."}
   ```
9. **`event: error`** $\to$ Jika terjadi kegagalan jaringan atau provider:
   ```json
   data: {"error": "Koneksi ke LLM timeout"}
   ```

---

### D. Aturan UX: Input Locking & Tombol Abort

> [!CAUTION]
> **Aturan Wajib State Management Chat:**
> 1. **Disable Input Saat Streaming**:
>    Ketika SSE sedang aktif (`isStreaming === true`), Textarea prompt dan tombol Send **HARUS DI-DISABLE**.
> 2. **Tombol Stop / Batal (Abort)**:
>    Ubah tombol *Send* menjadi tombol *Stop* (kotak merah / ikon silang). Jika diklik, kirim:
>    ```javascript
>    fetch(`/api/chat/sessions/${currentSessionId}/abort`, { method: 'POST' });
>    ```
> 3. **Handling HTTP 409 Conflict**:
>    Jika pengguna mencoba mengirim prompt saat backend masih sibuk, BE akan mengembalikan status `HTTP 409 Conflict`. Tangkap error ini di frontend dan tampilkan toast:
>    *"Agen masih memproses pesan sebelumnya. Silakan tunggu atau klik tombol Stop."*

---

### E. Komponen UI Wajib yang Harus Dibangun Nabil
1. **Sidebar Sesi**:
   * Daftar riwayat percakapan (`GET /api/chat/sessions`).
   * Tombol *New Chat* (`POST /api/chat/sessions`).
   * Fitur Search bar (`GET /api/chat/search?q=...`).
   * Action menu per sesi: Rename, Pin, Export (.md/.json), Delete.
2. **Interactive Chat Canvas**:
   * Bubble User & Bubble Assistant.
   * Collapsible `<details>` untuk Thought Process (`event: thought` & `event: tool_call`).
   * Kartu Bukti (*Evidence Cards*) berlabel `[SUPPORTED]`, `[UNCERTAIN]`, `[CONTRADICTED]`.
   * Disclaimer non-advisory di bawah input box.
3. **Anomaly & Chart Visualizer** (Tab / Drawer Kanan):
   * Chart candlestick harga & bar volume (bisa pakai Lightweight Charts atau Recharts).
   * Marker anomali merah saat $V_z \ge 2.5$.
4. **Knowledge Graph Modal / Tab**:
   * Embed iframe ke `/graph` (visualisasi memory graph Vis.js bawaan Go Server) atau visualisasi node-link kustom via data `/api/graph/data`.

---

## 3. Panduan Khusus Agung (CLI & TUI Terminal — `clients/cli/`)

Status saat ini: CLI sudah memiliki launcher (`clients/cli/tui/launcher.go`), headless investigator (`clients/cli/investigate.go`), dan interactive REPL (`clients/cli/tui/repl.go`).

### A. Stack & Arsitektur CLI
* **Framework**: Go 1.24+ murni (Zero-CGO).
* **TUI Libraries**: `github.com/charmbracelet/bubbletea`, `bubbles`, `lipgloss`, `glamour`.
* **Theme**: Matrix / Market Intelligence Green (`#4ADE80`, `#22C55E`), Amber (`#FACC15`), Red (`#EF4444`).

---

### B. Perintah CLI yang Tersedia
Biner utama dikompilasi ke `bin/niskava` (atau diinstall global via `go install ./cmd/niskava`):

```bash
# 1. Menjalankan Interactive REPL TUI (Hermes-Style)
niskava

# 2. Menjalankan Single-Turn Headless Investigation
niskava investigate ANTM --days 30

# 3. Menjalankan Background Daemon & REST/SSE Server
niskava serve --port 20128

# 4. Membuka Visualisasi Memory Knowledge Graph di Browser
niskava graph

# 5. Menjalankan Setup Wizard Konfigurasi Interaktif
niskava setup

# 6. Menjalankan Mode MCP Server (JSON-RPC stdio)
niskava mcp

# 7. Menampilkan Riwayat Sesi Investigasi & Chat dari SQLite
niskava sessions

# 8. Menjalankan Diagnostik Kesehatan Lengkap Lingkungan Kerja
niskava doctor
```

---

### C. Penyesuaian REPL Interaktif (`repl.go`)
1. **Slash Commands**:
   * Agung sudah memiliki popup slash command di Bubbletea lengkap dengan 13 perintah (`/help`, `/chats`, `/resume`, `/timeout`, `/reset`, `/graph`, `/clear`, `/web`, `/sessions`, `/health`, `/lang`, `/back`, `/exit`).
   * Perintah `/timeout` mendukung profil instan: `fast` (25s), `balanced` (60s), `deep` (120s), `local` (180s), atau nilai kustom `10-300` detik, langsung tersimpan ke `~/.niskava/config.yaml`.
   * Pastikan perintah `/reset` membersihkan memori graf di database:
     ```go
     appDB.ClearMemoryGraph()
     ```
2. **Koneksi Dual-Mode (Subprocess IPC vs HTTP Daemon)**:
   * Saat ini `repl.go` memanggil Python runner lokal via `ipc.RunSubprocess()`.
   * Jika daemon `niskava serve` aktif di background (`http://localhost:20128`), berikan opsi fallback untuk mengalirkan stream via SSE HTTP agar sesi terminal dan web tersinkronisasi.
3. **Locking Input Terminal Saat Agen Berpikir**:
   * Selama `executeChatTurn()` berlangsung, input teks terminal harus diblokir (*read-only*) hingga event `session_complete` atau error diterima agar pengguna tidak mengetik di atas stream rendering Glamour.

---

### D. Tata Kelola Interupsi (`Ctrl + C`)
* Jangan biarkan `Ctrl + C` langsung mematikan aplikasi CLI jika agen sedang berpikir!
* Tangkap sinyal `SIGINT`:
  * Batalkan konteks subprocess Python (`cancel()`).
  * Cetak pesan: `\n[!] Eksekusi dibatalkan oleh pengguna.\n`
  * Kembalikan kursor prompt terminal `USER > ` ke layar agar pengguna bisa memasukkan perintah berikutnya tanpa perlu membuka ulang aplikasi.

---

## 4. Tabel Taksonomi Bukti & Warna UI

Ketika menampilkan kartu temuan (*findings*) di Web Canvas (Nabil) maupun Terminal TUI (Agung), gunakan palet warna dan definisi baku berikut:

| Status Verifikasi | Warna Badge Web / Hex | Warna Lipgloss TUI | Definisi & Kriteria Bukti |
|---|---|---|---|
| **`SUPPORTED`** | Emerald Green (`#22C55E`) | `Foreground("#052E16"), Background("#22C55E")` | Terbukti secara faktual oleh laporan resmi Sectors API atau keterbukaan informasi formal IDXnet. |
| **`UNCERTAIN`** | Amber Yellow (`#FACC15`) | `Foreground("#0F172A"), Background("#FACC15")` | Ada korelasi pergerakan, namun bukti kausalitas belum terverifikasi (rumor media sosial, kabar burung). |
| **`CONTRADICTED`** | Crimson Red (`#EF4444`) | `Foreground("#FFFFFF"), Background("#EF4444")` | Klaim atau rumor pasar **terbukti salah / dibantah** oleh laporan keuangan atau surat resmi emiten. |

---

## 5. Checklist Kesiapan Sebelum Demo/Submit

### Untuk Nabil (Web):
- [ ] Menjalankan `npm run dev` di `clients/web/` dan terhubung ke backend `localhost:20128`.
- [ ] SSE streaming berjalan lancar (token teks mengalir halus).
- [ ] Textarea disabled & tombol *Send* berubah menjadi tombol *Stop* saat streaming berlangsung.
- [ ] Menguji pembatalan: Klik *Stop* $\to$ BE menerima `/abort` $\to$ status sesi kembali ke `IDLE`.
- [ ] Slider *Inference Timeout* (10–300s) di Settings modal tersinkronisasi dua arah via `GET/PATCH /api/settings`.
- [ ] Mengetik prompt typo *"cek berita hari ini domg"* $\to$ menampilkan berita pasar umum tanpa memunculkan emiten fiktif `DOMG`.
- [ ] Disclaimer finansial non-advisory tampil jelas di footer.

### Untuk Agung (CLI):
- [ ] `niskava` (interactive REPL) berjalan mulus dengan banner HUD Market Intelligence.
- [ ] Mengetik `/help`, `/chats`, `/resume`, `/timeout`, `/reset`, `/graph`, `/web`, `/clear` berfungsi normal.
- [ ] Tekan `Ctrl + C` saat agen berpikir $\to$ proses berhenti seketika dan prompt `USER > ` kembali aktif.
- [ ] Output rendering markdown menggunakan Glamour rapi dan tidak merusak layout terminal.
- [ ] Kompilasi biner Go bersih (`go build -o bin/niskava ./cmd/niskava`) dan tes `go test -v -race ./...` lulus.

---

> **Ada Pertanyaan atau Kendala Integrasi?**  
> Seluruh kode backend Go (`backend/core/`) dan Python Engine (`backend/engine/`) telah siap 100% dan lulus 272/272 unit test Python serta seluruh suite Go test. Jalankan `go test ./backend/... -v` atau `PYTHONPATH=backend pytest backend/engine/tests/ -q` untuk memeriksa integritas backend kapan saja.
