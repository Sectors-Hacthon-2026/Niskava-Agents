# Niskava CLI Command Reference Manual

Dokumentasi referensi lengkap antarmuka baris perintah (*Command Line Interface* - CLI) untuk **Niskava Agent**. CLI dibangun menggunakan framework Cobra dan Charmbracelet Bubble Tea untuk interaksi Terminal User Interface (TUI), melayani eksekusi investigasi otonom headless, REPL interaktif, server daemon lokal, dan penyedia Model Context Protocol (MCP).

---

## 1. Tata Kelola Konfigurasi & Resolusi Hierarki

Perilaku CLI dikendalikan melalui sistem konfigurasi hierarkis 4-tingkat (*Precedence Hierarchy*):

```text
CLI Flags > Environment Variables (NISKAVA_*) > Config File (~/.niskava/config.yaml) > Default Fallbacks
```

### 1.1 Berkas Konfigurasi Lokal
Lokasi default berkas konfigurasi: `~/.niskava/config.yaml` dengan izin file POSIX `0600`.

```yaml
auth:
  ai_provider: "gemini"               # gemini | openai | ollama
  sectors_api_key: "sec_live_..."     # Kunci API Sectors v2
  sectors_base_url: "https://api.sectors.app/v2"
  gemini_api_key: "AIza..."          # Kunci API Google Gemini
  gemini_model: "gemini-2.0-flash"
  openai_api_key: ""
  openai_base_url: "http://localhost:20128/v1"
  openai_model: "gpt-4o-mini"
  anthropic_api_key: ""
  ollama_base_url: "http://localhost:11434"
  ollama_model: "qwen2.5:7b"

storage:
  db_path: "~/.niskava/niskava.db"    # Lokasi database SQLite WAL

engine:
  python_bin: "python3"               # Path biner Python (mendukung .venv)
  entrypoint: "backend.engine.runner"

server:
  port: 20128                         # Port daemon REST/SSE

preferences:
  default_market: "IDX"               # Pasar default (Bursa Efek Indonesia)
  offline_mode: false                 # Mode mock offline (Law 5)
  language: "id"                      # id (Bahasa Indonesia) | en (English)
  llm_timeout_secs: 60.0              # Batas waktu tunggu inferensi AI (10 - 300 detik)

memory:
  enabled: true
  decay_lambda: 0.05                  # Faktor peluruhan waktu relasi asosiatif
  ego_radius: 2                       # Radius batas hop ego-network (Law 6)
  max_context_tokens: 4096

telegram:
  bot_token: ""                       # Token bot Telegram dari @BotFather
  enabled: false
  allowed_users: []                   # Whitelist user ID atau username
```

### 1.2 Environment Variables

| Variable | Subdomain | Deskripsi |
|---|---|---|
| `NISKAVA_CONFIG` | Core | Path absolut berkas `config.yaml` kustom |
| `NISKAVA_DB_PATH` | Storage | Lokasi berkas SQLite `niskava.db` (WAL Mode) |
| `NISKAVA_PORT` | Server | Port listen HTTP daemon (default: `20128` atau `8080`) |
| `NISKAVA_PYTHON_BIN` | Engine | Biner Python 3.11+ yang digunakan runner (.venv) |
| `NISKAVA_ENGINE_PATH`| Engine | Path direktori Python engine (default: `./backend/engine`) |
| `NISKAVA_LANG` | Preferences | Bahasa antarmuka (`id` atau `en`) |
| `NISKAVA_LLM_TIMEOUT`| Preferences | Batas waktu inferensi LLM dalam detik (10 - 300, default: 60.00) |
| `NISKAVA_MAX_REACT_ITERATIONS`| Engine | Batas maksimum siklus ReAct reasoning loop (default: 30) |
| `NISKAVA_OFFLINE` | Preferences | Mengaktifkan mode pengujian offline |
| `MOCK_SECTORS` | Preferences | Menggunakan fixture statis lokal untuk data Sectors |
| `SECTORS_API_KEY` | Auth | Kunci otentikasi Sectors Financial API v2 |
| `SECTORS_BASE_URL` | Auth | Base URL endpoint Sectors API (default: https://api.sectors.app/v2) |
| `GEMINI_API_KEY` | Auth | Kunci API Google Gemini Cloud |
| `OPENAI_API_KEY` | Auth | Kunci API OpenAI / OpenRouter / Gateway |
| `OPENAI_BASE_URL` | Auth | Endpoint OpenAI-compatible (contoh: http://localhost:20128/v1) |
| `OPENAI_MODEL` | Auth | Nama model inference (contoh: gpt-4o-mini, deepseek-chat) |
| `NISKAVA_TELEGRAM_TOKEN` | Telegram | Token bot Telegram dari @BotFather (fallback: `TELEGRAM_BOT_TOKEN`) |
| `NISKAVA_TELEGRAM_ENABLED`| Telegram | Mengaktifkan autostart bot saat `niskava serve` (`1` atau `0`) |
| `NISKAVA_TELEGRAM_ALLOWED_USERS`| Telegram | Daftar pengguna terotorisasi (dipisah koma, fallback: `TELEGRAM_ALLOWED_USERS`) |

---

## 2. Global Flags

Flags berikut berlaku untuk seluruh perintah dan subperintah Niskava:

* `--config <path>`: Menentukan berkas konfigurasi kustom selain `~/.niskava/config.yaml`.
* `--lang <id|en>`: Memaksa bahasa antarmuka TUI menjadi Bahasa Indonesia (`id`) atau Bahasa Inggris (`en`).
* `--session <session_id>`: Mengaitkan eksekusi langsung ke ID sesi percakapan/investigasi tertentu.
* `-v, --verbose`: Mengaktifkan log diagnosa rinci ke terminal.

---

## 3. Katalog Perintah

### 3.1 Perintah Utama: `niskava` (Bare Command)
Menjalankan daemon latar belakang lokal secara otomatis dan menampilkan **Interactive Launcher HUD (Terminal UI)** pada terminal.

```bash
niskava [flags]
```

#### Pilihan Menu Launcher:
1. `terminal`: Membuka terminal interaktif REPL untuk analisis percakapan langsung.
2. `web`: Menjalankan browser dan membuka antarmuka Web Workspace Dashboard.
3. `sessions`: Membuka *Session Selector* interaktif untuk memilih dan melanjutkan sesi sebelumnya.
4. `setup`: Menjalankan Setup Wizard konfigurasi awal.
5. `lang`: Beralih bahasa antarmuka (`id` $\leftrightarrow$ `en`).
6. `help`: Menampilkan ringkasan bantuan dan dokumentasi non-advisory.

#### Perintah Slash di Terminal REPL:
Ketik `/` saat berada di input REPL untuk membuka popup autocomplete:

| Perintah Slash | Kategori | Penjelasan & Contoh |
|---|---|---|
| `/help` | `[SYSTEM]` | Menampilkan panduan tombol pintas dan daftar perintah slash. |
| `/chats` | `[NAV]` | Membuka selektor sesi Bubbletea untuk menelusuri dan mengganti riwayat obrolan. |
| `/resume <ID>` | `[INTEL]` | Melanjutkan sesi obrolan tertentu berdasarkan ID (`/resume CHAT-20260925-0001`). |
| `/timeout [profil\|detik]` | `[SYSTEM]` | Mengatur batas waktu inferensi LLM: `fast` (25s), `balanced` (60s), `deep` (120s), `local` (180s), atau nilai kustom `10-300` detik. |
| `/export [md\|json]` | `[INTEL]` | Mengekspor transkrip obrolan sesi aktif ke berkas laporan `.md` atau `.json` lokal. |
| `/fork [title]` | `[INTEL]` | Mencabangkan sesi aktif ke ID sesi baru untuk menguji skenario pasar alternatif (*session forking*). |
| `/search <keyword>` | `[INTEL]` | Mencari teks pesan dari riwayat obrolan SQLite langsung dari prompt REPL. |
| `/anomalies` | `[INTEL]` | Menampilkan tabel anomali kuantitatif yang terdeteksi selama sesi aktif. |
| `/skills` | `[INTEL]` | Menampilkan katalog 6 domain SOP terdaftar (`market_anomaly_recon`, `event_causality_audit`, dll). |
| `/doctor` | `[SYSTEM]` | Menjalankan pemeriksaan diagnostik sistem komprehensif langsung dari dalam REPL. |
| `/cache [stats\|clean]` | `[SYSTEM]` | Menampilkan statistik efisiensi kuota Sectors API v2 & membersihkan entri cache kadaluarsa. |
| `/graph` | `[INTEL]` | Membuka visualisasi graf memori asosiatif langsung di browser. |
| `/web` | `[NAV]` | Membuka antarmuka visual Web Workspace di peramban web default. |
| `/sessions` | `[INTEL]` | Mencetak daftar sesi investigasi dan obrolan dari SQLite lokal. |
| `/health` | `[SYSTEM]` | Menampilkan status konektivitas daemon, database, dan latensi model AI. |
| `/lang [en\|id]` | `[SYSTEM]` | Beralih bahasa antarmuka seketika (`/lang id` atau `/lang en`). |
| `/reset` | `[SYSTEM]` | Membersihkan memori graf kerja sesi aktif dan memulai percakapan baru. |
| `/clear` | `[SYSTEM]` | Membersihkan layar terminal dan merender ulang banner. |
| `/back` atau `/exit` | `[NAV]` | Keluar dari REPL dan kembali ke menu utama Launcher HUD. |

#### Contoh Penggunaan:
```bash
# Menjalankan launcher interaktif
niskava

# Langsung melompat ke REPL terminal interaktif
niskava terminal
# atau via alias singkat:
niskava repl
niskava chat

# Langsung melompat ke REPL terminal untuk sesi tertentu
niskava --session SES-20260924-ANTM

# Menjalankan antarmuka dalam Bahasa Inggris
niskava --lang en
```

---

### 3.2 `niskava terminal` (Alias: `niskava repl`, `niskava chat`)
Meluncurkan antarmuka *Live Conversational REPL* secara langsung tanpa perlu melewati menu Launcher HUD.

```bash
niskava terminal [flags]
niskava repl [flags]
niskava chat [flags]
```

#### Flags Khusus:
* `-s, --session <id>`: Langsung melanjutkan ID sesi percakapan tertentu.
* `-l, --lang <id|en>`: Memaksa preferensi bahasa antarmuka (`id` atau `en`).

---

### 3.3 `niskava investigate`
Menjalankan alur kerja investigasi terstruktur 7-Stage pipeline secara otonom (*headless execution*) atau interaktif pada emiten Bursa Efek Indonesia (IDX).

```bash
niskava investigate <TICKER> [flags]
```

#### Argumen:
* `<TICKER>` (Wajib): Kode ticker 4-5 huruf resmi IDX (contoh: `ANTM`, `BBCA`, `BBRI`, `ASII`).

#### Flags Khusus:
* `-i, --interactive`: Menjalankan audit lalu **langsung membuka Live REPL** yang terfokus pada emiten target.
* `-d, --days <n>`: Rentang hari analisis historis ke belakang (`30`, `60`, atau `90`, default: `30`).
* `-f, --export-format <md|json|pdf>`: Format ekspor dossier audit hasil investigasi (`md`, `json`, atau `pdf`).
* `-o, --export-out <path>`: Path tujuan berkas laporan ekspor (contoh: `./ANTM_Report.pdf` atau `./laporan.md`).
* `--offline`: Menjalankan investigasi menggunakan data lokal *cached* tanpa mengurangi kuota API (Law 5).
* `--python-bin <path>`: Menentukan path biner Python khusus untuk proses engine.
* `--engine-path <path>`: Menentukan direktori Python engine khusus.

#### Contoh Penggunaan:
```bash
# Investigasi default 30 hari untuk saham ANTM
niskava investigate ANTM

# Investigasi saham ANTM dan langsung mengekspor laporan resmi ke format PDF
niskava investigate ANTM --days 30 --export-format pdf

# Ekspor PDF dengan path dan nama berkas khusus
niskava investigate BBRI --export-format pdf --export-out ~/Documents/Audit_BBRI.pdf

# Investigasi saham ANTM dan langsung membuka Live REPL interaktif
niskava investigate ANTM -i

# Investigasi 90 hari saham BBRI dalam mode offline
niskava investigate BBRI --days 90 --offline
```

#### Tahapan Output Pipeline (7-Stage SOP):
1. `INITIATION`: Pembuatan session ID dan alokasi sesi di SQLite.
2. `SECTORS_BASELINE`: Pengambilan data historis OHLCV, foreign flow, laporan emiten, dan aksi korporasi.
3. `QUANT_ANOMALY`: Perhitungan statistik deterministik NumPy ($V_z$, $F_z$, abnormal return $R_t$, dispersi sektor $D_t$).
4. `GAP_DETECTION`: Formulasi hipotesis temporal di sekitar tanggal $T_{\text{anomaly}} \pm 2\text{ hari}$.
5. `NEWS_HARVEST`: Penarikan berita bursa terkurasi dan keterbukaan informasi emiten via Sectors API v2.
6. `EVIDENCE_CORRELATION`: Penentuan urutan temporal (`LIKELY_CATALYST`, `PRECEDED_ANNOUNCEMENT`, `UNEXPLAINED_BY_NEWS`) dan status 3-tier (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
7. `SYNTHESIS`: Ringkasan dossier terstruktur dan penyimpanan permanen ke basis data.

---

### 3.3 `niskava serve`
Menjalankan daemon REST server dan Server-Sent Events (SSE) streaming di latar belakang untuk melayani Web Workspace Dashboard dan klien eksternal.

```bash
niskava serve [flags]
```

#### Flags Khusus:
* `-p, --port <int>`: Port HTTP daemon lokal (default: `20128` atau dari konfigurasi).
* `--open`: Membuka browser secara otomatis menuju Web Workspace setelah server online.
* `--telegram`: Mengaktifkan layanan bot Telegram secara bersamaan di background.

#### Contoh Penggunaan:
```bash
# Menjalankan daemon standar pada port 20128
niskava serve

# Menjalankan daemon pada port 8080 dan langsung membuka browser
niskava serve --port 8080 --open

# Menjalankan daemon sekaligus mengaktifkan poller bot Telegram
niskava serve --telegram
```

---

### 3.4 `niskava setup`
Menjalankan *Interactive Setup Wizard* berbasis TUI untuk konfigurasi awal kredensial, pilihan model AI, dan preferensi sistem tanpa perlu menyunting file YAML secara manual.

```bash
niskava setup [flags]
```

#### Langkah yang Dikonfigurasi Wizard:
1. **Kategori & Provider AI**: Memilih External Router (OpenRouter / 9router), Cloud Direct (Gemini, OpenAI, DeepSeek, Groq), atau Local Offline (Ollama, LM Studio).
2. **Kunci API & Endpoint**: Input API Key dan Base URL sesuai provider yang dipilih.
3. **Kunci Sectors Financial API v2**: Input token `SECTORS_API_KEY` atau tekan ENTER untuk mengaktifkan 100% Offline Mock Mode gratis (Law 5).
4. **Diagnostik Koneksi Live**: Ping otomatis ke endpoint AI dan Sectors v2 untuk memverifikasi latensi dan autentikasi.
5. **Runtime Python Engine**: Pengecekan otomatis dependensi `.venv`, dengan opsi *auto-bootstrap* instalasi jika belum siap.
6. **Profil Timeout Inferensi AI**:
   - `[1]` Fast (25s) — untuk Cloud API cepat
   - `[2]` Balanced / Adaptive (60s) — *[Rekomendasi]* untuk multi-tool IDX
   - `[3]` Deep Reasoning (120s) — untuk model penalaran besar (o3, Gemini 2.5 Pro)
   - `[4]` Local LLM (180s) — untuk Ollama CPU / hardware terbatas
   - `[5]` Custom (10–300s)
7. **Penyimpanan Sinkron**: Menyimpan konfigurasi ke `.env` lokal (`0600`) dan disinkronkan ke `~/.niskava/config.yaml`.

---

### 3.5 `niskava telegram` (Alias: `niskava bot`)
Menjalankan worker percakapan bot Telegram berbasis *long-polling* terdedikasi serta mengelola diagnostik dan whitelist pengguna. Pesan pengguna diarahkan langsung ke ReAct investigation pipeline dan disinkronisasikan ke database lokal.

```bash
niskava telegram [subcommand] [flags]
niskava bot [subcommand] [flags]
```

#### Flags Khusus Perintah Utama:
* `--token <string>`: Token bot Telegram (override token di `config.yaml` atau env).

#### Subperintah:
* `niskava telegram status` : Memeriksa konektivitas langsung ke API Telegram (`getMe`), identitas bot, status autostart, dan whitelist aktif.
* `niskava telegram user list` : Menampilkan seluruh User ID numerik dan @username Telegram yang terotorisasi.
* `niskava telegram user add <id_or_username> [more...]` : Menambahkan pengguna ke daftar whitelist (otomatis tersimpan ke `config.yaml` dan `.env`).
* `niskava telegram user remove <id_or_username> [more...]` : Menghapus pengguna dari whitelist.
* `niskava telegram test --chat-id <ID> [--message <string>]` : Mengirim pesan ping uji coba ke chat ID target untuk memverifikasi pengiriman pesan.

#### Flags Subperintah `test`:
* `--chat-id <string>` : ID chat Telegram target (wajib).
* `--message <string>` : Pesan kustom pengujian.

#### Contoh Penggunaan:
```bash
# Menjalankan bot dengan token dari config.yaml / .env
niskava telegram

# Menjalankan bot dengan token eksplisit
niskava bot --token "7123456789:AAHxyz..."

# Memeriksa diagnostik dan status koneksi bot
niskava telegram status

# Menambahkan analis ke whitelist
niskava telegram user add @analyst_idx 123456789

# Melihat daftar whitelist saat ini
niskava telegram user list

# Mengirim pesan uji coba koneksi
niskava telegram test --chat-id 123456789
```

---

### 3.6 `niskava sessions`
Menampilkan, mencari, menghapus, dan mengekspor riwayat sesi percakapan dan sesi investigasi yang tersimpan di basis data SQLite lokal.

```bash
niskava sessions [subcommand] [flags]
```

#### Subperintah:
* `niskava sessions [list]` : Menampilkan daftar sesi percakapan AI dan investigasi audit.
* `niskava sessions delete <SESSION_ID>` : Menghapus sesi percakapan dan seluruh riwayat perannya secara permanen dari SQLite.
* `niskava sessions search <KEYWORD>` : Mencari kata kunci tertentu di seluruh riwayat obrolan pesan SQLite.
* `niskava sessions export <SESSION_ID> [--format md|json] [--out path]` : Mengekspor transkrip laporan audit sesi ke berkas Markdown atau JSON.

#### Flags Khusus:
* `-n, --limit <int>`: Jumlah maksimum sesi yang ditampilkan (default: `20`).
* `-t, --type <string>`: Filter tipe sesi: `'chat'`, `'investigation'`, atau `'all'` (default: `'all'`).
* `-f, --format <string>`: Format ekspor pada `export`: `'md'` atau `'json'` (default: `'md'`).
* `-o, --out <path>`: Path berkas output pada `export`.

#### Contoh Penggunaan:
```bash
# Menampilkan 10 sesi terakhir
niskava sessions list -n 10

# Mencari percakapan bertema ANTM
niskava sessions search ANTM

# Mengekspor sesi ke format Markdown
niskava sessions export CHAT-20260926-0001 --format md --out ./laporan_antm.md

# Menghapus sesi
niskava sessions delete CHAT-20260926-0001
```

---

### 3.7 `niskava graph`
Mengekspor dan memvisualisasikan Knowledge Graph intelijen pasar (relasi emiten, anomali, katalis, broker) ke dalam berkas HTML interaktif mandiri (*standalone PyVis*).

```bash
niskava graph [flags]
```

#### Flags Khusus:
* `-o, --output <path>`: Jalur penyimpanan berkas HTML keluaran (default: `~/.niskava/graph.html`).
* `-s, --session <id>`: Memfilter visualisasi khusus untuk satu ID sesi.
* `--open`: Membuka berkas HTML visualisasi langsung di peramban web default.

#### Contoh Penggunaan:
```bash
# Render seluruh graf memori global dan buka di browser
niskava graph --open

# Render graf khusus sesi tertentu ke file kustom
niskava graph --session INV-20260924-ANTM -o ./audit_antm.html --open
```

---

### 3.8 `niskava mcp`
Menjalankan server **Model Context Protocol (MCP)** resmi Niskava melalui antarmuka *Standard Input/Output* (stdio JSON-RPC 2.0). Digunakan untuk menghubungkan Niskava sebagai *tool engine* ke AI Desktop clients seperti **Claude Desktop**, **Cursor IDE**, dan coding assistants.

```bash
niskava mcp [flags]
```

#### Flags Khusus:
* `--stdio`: Mengaktifkan mode transportasi stdio (default: `true`).
* `--py-bin <path>`: Menentukan executable Python yang mengeksekusi server MCP Python.

#### Kemampuan yang Diekspos Server MCP:
* **Tools Deterministic (14 tools):**
  * `quant_compute_anomalies`: Perhitungan Z-score, MA20 volume, dan abnormal return tanpa halusinasi numerik.
  * `quant_analyze_bandarmology`: Evaluasi konsentrasi broker top 1/3/5 dan volume akumulasi/distribusi.
  * `sectors_get_daily_candles`: Pengambilan data candlestick OHLCV historis dari Sectors API v2.
  * `sectors_get_foreign_flow`: Pelacakan arus dana investor asing net inflow/outflow.
  * `sectors_get_company_report`: Laporan fundamental dan ikhtisar keuangan emiten.
  * `news_harvest_market_news`: Panen berita pasar terkurasi dan keterbukaan informasi.
  * `memory_recall_context`: Pengambilan memori graf asosiatif masa lalu emiten.
* **Resources:** Metrik cache lokal dan diagnostik kesehatan sistem.
* **Prompts:** Template investigasi formal IDX 7-stage SOP.

---

### 3.9 `niskava doctor`
Menjalankan inspeksi diagnostik komprehensif pada seluruh subsistem Niskava untuk memvalidasi kesiapan lingkungan kerja sebelum melakukan analisis pasar.

```bash
niskava doctor [flags]
```

#### Komponen yang Diperiksa oleh Doctor:
1. **Sistem Operasi & Arsitektur**: Kompatibilitas OS (Linux, macOS, Windows).
2. **Basis Data SQLite**: Status Write-Ahead Logging (`PRAGMA journal_mode = WAL`), integritas skema tabel, dan permission.
3. **Runtime Python Engine**: Keberadaan biner Python 3.11+, virtual environment (`.venv`), dan pustaka kuantitatif (`numpy`, `pandas`, `networkx`).
4. **Provider AI & Konektivitas**: Uji koneksi live (ping HTTP) dan latensi respon ke endpoint model inferensi yang terkonfigurasi.
5. **Sectors Financial API v2**: Validasi status autentikasi token atau mode pengujian offline mock.

#### Contoh Penggunaan:
```bash
# Menjalankan pemeriksaan kesehatan lengkap
niskava doctor
```

---

### 3.10 `niskava completion`
Menghasilkan skrip autokompresi shell (*shell autocompletion script*) untuk Bash, Zsh, Fish, dan PowerShell.

```bash
niskava completion [bash|zsh|fish|powershell]
```

#### Contoh Penggunaan:
```bash
# Bash
source <(niskava completion bash)

# Zsh
niskava completion zsh > "${fpath[1]}/_niskava"

# PowerShell
niskava completion powershell | Out-String | Invoke-Expression
```

---

## 4. Exit Codes Standar

| Exit Code | Nama Status | Penjelasan |
|---|---|---|
| `0` | `SUCCESS` | Perintah selesai dieksekusi tanpa kesalahan |
| `1` | `GENERAL_ERROR` | Kesalahan umum aplikasi, kegagalan runner, atau argumen tidak valid |
| `2` | `USAGE_ERROR` | Format perintah salah, flag tidak dikenal, atau argumen kurang |
| `130` | `INTERRUPT` | Proses dihentikan oleh pengguna melalui `Ctrl + C` (SIGINT) |
