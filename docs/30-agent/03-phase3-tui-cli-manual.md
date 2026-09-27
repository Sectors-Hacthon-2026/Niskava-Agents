# Manual Penggunaan & Arsitektur Phase 3: Advanced TUI & Headless Pipeline

> **Panduan Teknis Operasional CLI, TUI REPL, Visualisasi Anomali ASCII, dan Headless Pipeline Niskava Agent.**

---

## 1. Ringkasan Modul Phase 3

Phase 3 menambahkan kapabilitas eksekusi tingkat lanjut pada **Terminal User Interface (TUI)** dan **Command Line Interface (CLI)** Niskava Agent:

1. **Headless Investigation Pipeline (`niskava investigate <TICKER>`)**:
   - Menjalankan 7-Stage SOP audit trail secara otomatis dari terminal.
   - Opsi ekspor laporan audit ke berkas Markdown (`.md`) atau JSON (`.json`) via flag `--export-format` (`-f`) dan `--export-out` (`-o`).
2. **Terminal ASCII Anomaly Visualizer (`clients/cli/tui/ascii_chart.go`)**:
   - Visualisasi grafis batang ASCII sparkline untuk anomali volume ($V_z \ge 2.5$) dan lonjakan harga ($R_t \ge 5\%$).
   - Terintegrasi di dalam perintah `/anomalies` TUI REPL.
3. **Daemon Telegram Bot CLI (`niskava telegram`)**:
   - Menjalankan daemon bot Telegram riset pasar modal langsung dari terminal.

---

## 2. Perintah & Opsi CLI Baru

### A. Sub-perintah `niskava investigate`
```bash
# 1. Investigasi headless 30 hari standar:
niskava investigate ANTM

# 2. Investigasi 30 hari dengan ekspor otomatis laporan ke file Markdown:
niskava investigate ANTM --days 30 --export-format md --export-out ANTM_Report.md

# 3. Investigasi 90 hari dengan ekspor JSON:
niskava investigate BBRI --days 90 --export-format json --export-out BBRI_Audit.json

# 4. Mode interaktif REPL langsung fokus ke emiten:
niskava investigate ANTM -i
```

#### Parameter Flag:
- `--days, -d`: Jendela pengamatan harian (`30`, `60`, atau `90` hari). Default: `30`.
- `--export-format, -f`: Format ekspor laporan (`md` atau `json`).
- `--export-out, -o`: Jalur berkas hasil ekspor laporan.
- `--interactive, -i`: Buka TUI REPL interaktif yang dipreferensikan ke emiten target.
- `--offline`: Jalankan dalam mode offline mock data.

---

### B. Perintah Slash Command `/anomalies` di TUI REPL
Ketika berada di dalam mode percakapan interaktif (`niskava` / `niskava terminal`), ketik `/anomalies` untuk menampilkan visualisasi grafis ASCII anomali kuantitatif:

```text
📊 ANTM — Quant Anomaly & Volatility Visualizer (30-Day Observation Window)
─────────────────────────────────────────────────────────────────────────────
  #1 [2026-09-12] Ticker: ANTM | Volume Z-Score: 3.82σ | Abnormal Return: +8.50%
     Price Bar  : [▄▄▅▄▃▄▄▄] (Volume Z-Score Spike)
     Volume Bar : [█▓▓█████] (Val: 14500000.00 | Baseline: 3800000.00)
     ↳ Lonjakan volume anomali 3.82x di atas rata-rata pergerakan MA20
```

---

### C. Sub-perintah `niskava telegram`
```bash
# Menjalankan bot Telegram riset pasar modal dari CLI:
niskava telegram --token <BOT_TOKEN_FROM_BOTFATHER>
```

---

## 3. Arsitektur Komponen Baru

```
clients/cli/
├── investigate.go           # CLI entrypoint headless 7-stage SOP + export flags
├── investigate_test.go      # Unit test flag ekspor & registrasi subcommand
├── telegram.go              # Daemon bot runner Telegram CLI
└── tui/
    ├── ascii_chart.go       # Visualizer grafis batang ASCII sparkline pergerakan anomali
    ├── ascii_chart_test.go  # Unit test rendering grafis ASCII
    └── repl.go              # REPL handler terintegrasi dengan ASCII anomaly chart
```

---

## 4. Pengujian & Verifikasi Kualitas

Seluruh komponen Phase 3 lulus uji unit 100% dan terkompilasi tanpa CGO:

```bash
# 1. Eksekusi Unit Test Suite:
go test -v ./clients/cli/...

# 2. Kompilasi Executable Tunggal:
go build -o bin/niskava.exe ./cmd/niskava
```
