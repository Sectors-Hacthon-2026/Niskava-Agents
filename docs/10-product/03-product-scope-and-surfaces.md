# 03 — Cakupan Produk & Antarmuka Pengguna (Surfaces)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.2.0  
**Terakhir Diperbarui:** 2026-09-22  

Niskava Agent mengusung filosofi arsitektur **"Dual Surfaces, Single Engine"**:
1. **Terminal CLI & TUI Launcher** untuk kecepatan eksekusi, otomasi skrip, investigasi mendalam, dan pengguna teknis (*power users*).
2. **Local Web Workspace** untuk eksplorasi interaktif visual, visualisasi grafik candlestick, timeline kejadian, kartu bukti, dan memory knowledge graph.

Kedua antarmuka ditenagai oleh satu binary biner Go yang sama dan membaca database SQLite lokal yang sama (`~/.niskava/niskava.db`).

---

## 1. Surface A: Terminal CLI & Conversational REPL (`niskava`)

Dibangun menggunakan Go (`spf13/cobra`, `charmbracelet/bubbletea` untuk TUI interaktif, `charmbracelet/lipgloss`, dan `charmbracelet/glamour` untuk rendering Markdown bergaya Bloomberg Terminal / Market Intelligence):

### Perintah Utama (CLI Commands)
```bash
# 1. Mode Asisten Percakapan Finansial (Hermes-Style REPL & Interactive Launcher)
# Menjalankan interactive launcher, HUD real-time, dan REPL tanya-jawab bahasa alami
niskava

# 2. Pemilihan Bahasa Antarmuka (English default / Indonesian)
niskava --lang id
niskava --lang en

# 3. Wizard Konfigurasi Interaktif (Setup Onboarding)
# Menuntun konfigurasi .env (API Keys, provider AI, Sectors API) dengan live connection ping
niskava setup

# 4. Investigasi Langsung Emiten (Headless 7-Stage Pipeline)
niskava investigate ANTM --days 30

# 5. Ekspor & Visualisasi Knowledge Graph Memori Lokal di Browser
niskava graph

# 6. Menjalankan Model Context Protocol (MCP) Server via JSON-RPC Stdio
# Siap dihubungkan ke Claude Desktop, Cursor, atau Antigravity
niskava mcp

# 7. Manajemen Sesi & Riwayat Interaktif
niskava sessions

# 8. Menjalankan Server Web Dashboard & AI Assistant Canvas
niskava serve --port 20128 --open

# 9. Mode Offline / Testing (Fixture JSON lokal)
niskava investigate ANTM --offline
```

### Mockup Pengalaman Interaktif Terminal REPL (Hermes Mode):
```text
┌─────────────────────────────────────────────────────────────┐
│               NISKAVA FINANCIAL AGENT v1.0.0               │
│      Autonomous IDX Market Intelligence REPL        │
│       Provider: 9router (hermes) | Storage: Local SQLite    │
└─────────────────────────────────────────────────────────────┘

niskava [hermes] > Kenapa saham ANTM volumenya melonjak tinggi baru-baru ini?

  ● Thought: Pengguna menanyakan anomali lonjakan volume saham ANTM.
    Sesuai Hukum 1 (Deterministic Before Generative), saya memanggil tool
    compute_quant_anomalies untuk menghitung statistik deterministik terlebih dahulu.
  ▶ Tool Call: compute_quant_anomalies(symbol="ANTM", days=30)
  ✔ Observation: Volume Z-Score 3.84σ (184.5M lembar) terdeteksi pada 12 Sep 2026.
  ▶ Tool Call: harvest_market_news(symbol="ANTM", query="ANTM lonjakan volume")
  ✔ Observation: Ditemukan 4 berita dan keterbukaan informasi smelter Halmahera Timur.

# Ringkasan Intelijen: Lonjakan Volume ANTM

Berdasarkan investigasi kuantitatif deterministik dan penelusuran berita bursa:

* **Deteksi Kuantitatif (Sectors v2):**
  Pada 12 September 2026, terjadi lonjakan volume abnormal sebesar **184.5M lembar**
  (rata-rata 20 hari: 48.2M lembar), menghasilkan **Volume Z-Score +3.84σ**
  dan pergerakan harga abnormal **+8.25%** dengan net buy investor asing **Rp111,3 Miliar**.

* **Korelasi Kausalitas Berita (News):**
  - `[SUPPORTED]` (Conf: 0.95): Keterbukaan informasi resmi BEI terkait peresmian ekspansi
    smelter nikel Halmahera Timur dirilis pada 12 September pagi.
  - `[UNCERTAIN]` (Conf: 0.65): Beredar rumor akuisisi konsesi tambang tambahan di forum ritel.

* **Audit Trail:** Sesi percakapan tersimpan di `~/.niskava/niskava.db`.
```

---

## 2. Surface B: Local Web Workspace (`localhost:8080`)

Frontend Single Page Application (SPA) modern yang dibangun dengan **Vite + React 18 + Tailwind CSS + shadcn/ui**, dikompilasi ke dalam biner Go menggunakan directive `//go:embed web/dist`.

### Komponen Kunci Web Dashboard:
1. **Interactive AI Assistant Canvas (`/api/chat`)**:
   * Antarmuka percakapan interaktif dengan dukungan Server-Sent Events (SSE) real-time.
   * Menampilkan kartu *Thinking Step*, *Tool Invocation* (Quant / News), dan *Evidence Synthesis*.
   * Mempertahankan riwayat multi-turn chat secara persisten melalui SQLite (`chat_messages`).
2. **Header & Status Banner**: Menampilkan status koneksi agent/model provider, waktu investigasi, ticker aktif, dan tombol ekspor laporan (Markdown / JSON).
3. **Metrics & Anomaly Strip**: Kartu ringkasan cepat:
   * *Volume $Z$-Score* (misal: `+3.84σ` — Abnormal High)
   * *Abnormal Return* (misal: `+8.25%` vs Sektor `+0.45%`)
   * *Broker Dominance* (misal: Asing Net Buy IDR 84 Miliar)
4. **Interactive Financial Chart**:
   * Menampilkan grafik candlestick harga dan bar volume.
   * Pin anomali visual (*anomaly flag marker*) pada tanggal $T_{anomaly}$ yang dapat diklik untuk menyorot bukti terkait.
5. **Chronological Event Timeline**:
   * Urutan vertikal peristiwa: dari pergerakan volume awal $\rightarrow$ rilis pengumuman bursa $\rightarrow$ pemberitaan media massa $\rightarrow$ penutupan harga.
6. **Evidence Cards Matrix**:
   * Kartu temuan dengan penanda warna taksonomi:
     * Hijau (`SUPPORTED`): Data resmi Sectors / Pengumuman BEI.
     * Kuning (`UNCERTAIN`): Berita media pihak ketiga atau rumor pasar belum terverifikasi.
     * Merah (`CONTRADICTED`): Narasi yang terbantahkan oleh fakta laporan keuangan.
   * Setiap kartu memiliki tombol *"Lihat Sumber Asli"* dan indikator confidence score.
7. **Live SSE Terminal Drawer**:
   * Laci terminal interaktif di bagian bawah yang menampilkan stream log penalaran (*thinking steps*) agent secara real-time saat investigasi sedang diproses.
