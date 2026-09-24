# Panduan Alur Kerja CLI & Integrasi Terminal — Niskava Agent

Panduan praktis pengoperasian dan integrasi antarmuka terminal (*Command Line Interface* & *Terminal User Interface*) untuk **Niskava Agent**. Dokumen ini mencakup panduan interaktif Hermes-style REPL, navigasi popup slash commands, integrasi Model Context Protocol (MCP) dengan AI desktop clients (Claude Desktop & Cursor), serta skrip otomasi headless.

---

## 1. Arsitektur Terminal & UX Design System

Antarmuka CLI Niskava dirancang dengan prinsip **Institutional Cyber-OSINT / Bloomberg Terminal Dark Aesthetic**:
* **Framework TUI:** [Charmbracelet Bubble Tea](https://github.com/charmbracelet/bubbletea) (The Elm Architecture untuk Go).
* **Styling & Layout:** [Charmbracelet Lipgloss](https://github.com/charmbracelet/lipgloss).
* **Markdown Renderer:** [Charmbracelet Glamour](https://github.com/charmbracelet/glamour) dengan tema dark terminal teroptimasi.
* **Dual Language (i18n):** Dukungan penuh Bahasa Indonesia (`id`) dan English (`en`) secara dinamis.

---

## 2. Tiga Mode Interaksi Terminal

Niskava mendukung 3 pola interaksi yang disesuaikan dengan kebutuhan analisis:

```text
┌─────────────────────────────────────────────────────────────┐
│                    MODE INTERAKSI CLI                       │
├──────────────────────────────┬──────────────────────────────┤
│ 1. Interactive Launcher HUD  │ Menu visual 9router-style    │
│    (niskava)                 │ navigasi antar-permukaan     │
├──────────────────────────────┼──────────────────────────────┤
│ 2. Live Conversational REPL  │ Terminal riset tanya-jawab   │
│    (niskava terminal)        │ streaming ReAct loop         │
├──────────────────────────────┼──────────────────────────────┤
│ 3. Headless Pipeline         │ Skrip satu perintah audit    │
│    (niskava investigate)     │ trail emiten IDX (SOP 7-tahap)│
└──────────────────────────────┴──────────────────────────────┘
```

---

## 3. Panduan Penggunaan Interactive REPL Terminal

Mode REPL (*Read-Eval-Print Loop*) memberikan pengalaman terminal cerdas untuk berdialog dengan agen investigasi, mengevaluasi anomali kuantitatif, dan memanggil bukti OSINT.

### 3.1 Memulai Sesi REPL
```bash
# Melalui launcher utama:
niskava -> Pilih 'Terminal (REPL)'

# Atau langsung membuka REPL untuk sesi tertentu:
niskava --session SES-20260924-ANTM
```

### 3.2 Navigasi & Pintasan Keyboard

| Tombol Keyboard | Aksi / Fungsi |
|---|---|
| `Enter` | Mengirim prompt atau memilih item menu slash yang aktif |
| `Up` / `Down` ($\uparrow$ / $\downarrow$) | Menelusuri riwayat prompt sebelumnya (*Prompt History Navigation*) |
| `/` | Memicu jendela popup filter *Slash Commands* otomatis |
| `Esc` | Menutup jendela popup slash command tanpa keluar dari REPL |
| `Tab` | Melengkapi auto-complete perintah slash yang sedang diketik |
| `Ctrl + C` (Saat streaming) | Mengirim sinyal pembatalan turn (*Abort ReAct Turn*) tanpa mematikan aplikasi |
| `Ctrl + C` (Saat idle, 2x) | Keluar dari aplikasi dengan aman (*Double-press Exit Guard*) |

### 3.3 Daftar Slash Commands

Ketik karakter garis miring (`/`) pada baris input untuk menampilkan daftar perintah interaktif:

```text
/───────────────────────────── SLASH COMMANDS ─────────────────────────────\
│ /help      [SYSTEM] Panduan & dokumentasi perintah REPL                   │
│ /back      [NAV]    Kembali ke menu launcher utama                        │
│ /chats     [NAV]    Buka daftar riwayat sesi obrolan interaktif           │
│ /resume    [INTEL]  Melanjutkan investigasi sesi sebelumnya               │
│ /reset     [SYSTEM] Reset konteks percakapan sesi saat ini                │
│ /graph     [INTEL]  Buka visualisasi Knowledge Graph di browser           │
│ /clear     [SYSTEM] Bersihkan layar terminal                              │
│ /web       [NAV]    Luncurkan Web Workspace Dashboard di browser          │
│ /sessions  [INTEL]  Daftar riwayat sesi investigasi & obrolan             │
\──────────────────────────────────────────────────────────────────────────/
```

### 3.4 Alur Kerja Investigasi Percakapan (Contoh Kasus)

1. **Inisiasi Pertanyaan Riset:**
   ```text
   niskava [IDX] > Analisis pergerakan volume anomali ANTM dalam 1 minggu terakhir
   ```
2. **Visualisasi ReAct Stream:**
   Terminal akan menampilkan indikator pemikiran agen (*Agent Reasoning*), diikuti eksekusi tool kuantitatif deterministik:
   ```text
   💭 Penalaran Agen:
      Mendeteksi permintaan audit volume saham ANTM. Memeriksa anomali deterministik
      menggunakan quant_compute_anomalies sebelum menarik berita pasar.

   🛠️ Aktivitas Eksekusi Alat:
      [CALL] quant_compute_anomalies(ticker="ANTM", timeframe_days=7)
      [RESULT] Volume Z-Score = 3.84 (AMBANG BATAS 2.5 TERLAMPAUI)

   🚨 ANOMALI KUANTITATIF TERDETEKSI:
      Tanggal: 2026-09-21 | Metrik: VOLUME | Nilai: 84.5M lembar (Baseline: 22.1M)
   ```
3. **Sintesis Bukti dengan Taksonomi 3-Tier (Law 2):**
   Agen menyajikan kesimpulan verifikasi faktual tanpa memberikan saran finansial:
   ```text
   RINGKASAN TEMUAN (AUDIT TRAIL 3-TIER):
   ● [SUPPORTED] Lonjakan volume didahului lonjakan harga komoditas nikel global (+4.2%).
     Sumber: Sectors v2 Market Feeds (Confidence: 0.95 | LIKELY_CATALYST)
   ● [UNCERTAIN] Rumor akuisisi konsorsium baterai EV di media sosial belum diverifikasi
     melalui keterbukaan informasi resmi IDXnet.
     Sumber: Kompilasi OSINT (Confidence: 0.65 | UNEXPLAINED_BY_NEWS)
   ```

---

## 4. Integrasi Model Context Protocol (MCP)

Niskava menyediakan server resmi **Model Context Protocol (MCP)** berbasis stdio JSON-RPC 2.0 (`niskava mcp`). Fitur ini memungkinkan Niskava digunakan secara langsung sebagai penyedia data pasar IDX dan kalkulasi kuantitatif oleh AI agent eksternal seperti **Claude Desktop** dan **Cursor IDE**.

### 4.1 Konfigurasi Claude Desktop

Tambahkan konfigurasi server Niskava ke dalam berkas konfigurasi Claude Desktop:

* **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
* **Linux:** `~/.config/Claude/claude_desktop_config.json`
* **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "niskava-idx": {
      "command": "niskava",
      "args": ["mcp"],
      "env": {
        "SECTORS_API_KEY": "sec_live_anda_di_sini",
        "NISKAVA_LANG": "id"
      }
    }
  }
}
```

### 4.2 Konfigurasi Cursor IDE

Untuk menggunakan Niskava sebagai tool provider di Cursor:
1. Buka **Cursor Settings** $\to$ **Features** $\to$ **MCP**.
2. Klik tombol **+ Add New MCP Server**.
3. Isi kolom konfigurasi:
   * **Name:** `niskava`
   * **Type:** `command`
   * **Command:** `niskava mcp`

Atau tambahkan langsung ke berkas `.cursor/mcp.json` di root workspace proyek:
```json
{
  "mcpServers": {
    "niskava": {
      "command": "niskava",
      "args": ["mcp"]
    }
  }
}
```

### 4.3 Menguji Server MCP Melalui Baris Perintah
Pastikan komunikasi JSON-RPC berfungsi normal menggunakan pipe standar:

```bash
# Uji Handshake Inisialisasi:
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}' | niskava mcp

# Respon yang Diharapkan:
# {"jsonrpc":"2.0","id":1,"result":{"serverInfo":{"name":"niskava-mcp-engine","version":"1.0.0"},...}}

# Uji Panggilan Tool Kuantitatif:
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"quant_compute_anomalies","arguments":{"ticker":"ANTM"}}}' | niskava mcp
```

---

## 5. Otomasi Headless Scripting & CI/CD

Perintah `niskava investigate` dirancang untuk diintegrasikan ke dalam cron jobs, skrip shell, atau pipeline otomatisasi harian.

### 5.1 Skrip Audit Emiten Portofolio Harian (`audit_portfolio.sh`)
Contoh skrip bash untuk melakukan skrining anomali pada daftar saham pilihan secara otomatis setiap penutupan bursa (16:30 WIB):

```bash
#!/usr/bin/env bash
set -euo pipefail

WATCHLIST=("ANTM" "BBCA" "BBRI" "ASII" "TLKM")
REPORT_DIR="$HOME/niskava_reports/$(date +%Y%m%d)"
mkdir -p "$REPORT_DIR"

echo "=== MEMULAI AUDIT HARIAN IDX: $(date) ==="

for TICKER in "${WATCHLIST[@]}"; do
  echo "[-] Mengaudit $TICKER..."
  
  # Eksekusi pipeline 7-tahap
  if niskava investigate "$TICKER" --days 30; then
    echo "[✓] Audit $TICKER selesai."
  else
    echo "[✗] Audit $TICKER gagal." >&2
  fi
done

# Ekspor grafik memori gabungan hari ini
niskava graph -o "$REPORT_DIR/market_graph.html"

echo "=== SELURUH AUDIT SELESAI. LAPORAN TERSIMPAN DI $REPORT_DIR ==="
```

### 5.2 Menjalankan Daemon di Background (Systemd Service)
Untuk menjalankan Niskava daemon secara persisten pada server lokal atau VPS:

```ini
# /etc/systemd/system/niskava.service
[Unit]
Description=Niskava Agent Market Intelligence Daemon
After=network.target

[Service]
Type=simple
User=user
WorkingDirectory=/home/user
ExecStart=/usr/local/bin/niskava serve --port 20128
Restart=always
RestartSec=5
Environment=NISKAVA_LANG=id
Environment=SECTORS_API_KEY=sec_live_anda

[Install]
WantedBy=multi-user.target
```

Aktivasi layanan:
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now niskava
```

---

## 6. Standar Bahasa Ganda (Dual-Language i18n)

Niskava mendukung perpindahan bahasa dinamis tanpa perlu instalasi ulang:
* **Pengaturan Permanen:** Jalankan `niskava setup` atau atur `preferences.language: "en"` di `~/.niskava/config.yaml`.
* **Pengaturan Per Sesi CLI:** Berikan flag `--lang en` saat menjalankan perintah CLI apa pun.
* **Pengaturan Cepat di REPL:** Ketik perintah slash `/lang` untuk membuka selector bahasa secara interaktif.
