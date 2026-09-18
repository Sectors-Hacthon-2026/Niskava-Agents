# 01 — Arsitektur Sistem (Hybrid Go + Python + React)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

---

## 1. Diagram Blok Arsitektur Hybrid

Niskava Agent dibangun di atas arsitektur tripartit hybrid yang memadukan keandalan sistem Go, kekuatan analitik Python, dan reaktivitas modern React SPA:

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                             USER SURFACES                                   │
│                                                                             │
│     Terminal REPL / CLI ($ niskava)        Local Browser (localhost:8080)   │
└───────────────────────┬─────────────────────────────────────┬───────────────┘
                        │                                     │
                        │                                     │ HTTP REST / SSE
                        ▼                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                               GO CORE DAEMON                                │
│                     (Single Binary Distribution: niskava)                   │
│                                                                             │
│  ┌───────────────────────┐  ┌───────────────────────┐  ┌─────────────────┐  │
│  │ Cobra Router & Setup  │  │   Local HTTP Router   │  │ Embedded Assets │  │
│  │ (niskava / setup)     │  │   (net/http / SSE)    │  │  (//go:embed)   │  │
│  └───────────┬───────────┘  └───────────┬───────────┘  └─────────────────┘  │
│              │                          │                                   │
│  ┌───────────▼───────────┐              │                                   │
│  │ Glamour / Bubbletea   │              │                                   │
│  │ Conversational REPL   │              │                                   │
│  └───────────┬───────────┘              │                                   │
│              └────────────┬─────────────┘                                   │
│                           │                                                 │
│               ┌───────────▼────────────┐     ┌───────────────────────────┐  │
│               │ Session & Store Engine │────▶│   SQLite Database File    │  │
│               │   (modernc.org/sqlite) │     │   (~/.niskava/niskava.db) │  │
│               │   (sessions & chats)   │     │   (WAL Mode)              │  │
│               └───────────┬────────────┘     └───────────────────────────┘  │
└───────────────────────────┼─────────────────────────────────────────────────┘
                            │ Subprocess IPC (JSON Lines via Stdin/Stdout)
                            ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            PYTHON AGENT ENGINE                              │
│                                                                             │
│  ┌───────────────────┐      ┌───────────────────┐     ┌──────────────────┐  │
│  │  ReAct Agent Loop │      │   Quant Anomaly   │     │ Sectors v2 Client│  │
│  │  (Multi-Turn LLM) │◀────▶│   Math Engine     │     │ & OSINT Harvester│  │
│  └───────────────────┘      └───────────────────┘     └──────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Pembagian Peran & Tanggung Jawab Komponen

### A. Go Core Daemon (`/cmd/niskava` & `/internal/`)
* **Single Executable Distributor**: Mengompilasi seluruh aplikasi menjadi 1 file biner mandiri. Seluruh file frontend statis di-embed langsung ke dalam biner menggunakan `//go:embed web/dist`.
* **Conversational REPL & CLI Engine**: Menggunakan `spf13/cobra`, `charmbracelet/bubbletea`, dan `charmbracelet/glamour` untuk menghadirkan terminal REPL percakapan interaktif (Hermes-style) dengan rendering Markdown ala Bloomberg Terminal.
* **Interactive Setup Wizard (`niskava setup`)**: Menuntun pengguna dalam konfigurasi kredensial `.env` (9router local proxy `http://localhost:20128/v1`, Google Gemini, Sectors API) dengan validasi ping koneksi live.
* **Local Web Server**: Menyediakan REST API untuk manajemen sesi dan percakapan (`/api/chat`), serta Server-Sent Events (SSE) untuk streaming log real-time ke web dashboard.
* **Session Persistence Manager**: Berkomunikasi dengan database SQLite lokal (`sessions`, `anomalies`, `chat_messages`) menggunakan driver murni Go (`modernc.org/sqlite`) tanpa kebutuhan compiler CGO.

### B. Python Agent Engine (`/engine/`)
* **Stateless Subprocess Runner**: Dijalankan oleh Go Core sebagai child process on-demand melalui `engine/runner.py`.
* **Autonomous ReAct Agent Loop (`engine/agent/react_agent.py`)**: Mengelola dialog percakapan multi-turn, pemanggilan tool deterministik secara otonom, dan sintesis bukti.
* **Deterministic Quant Anomaly (`engine/quant/`)**: Menghitung $Z$-score volume, abnormal return, dan divergensi sektor menggunakan library matematika Python murni (NumPy) sesuai Hukum 1. LLM dilarang berhitung secara mandiri.
* **Sectors v2 API Client & Dual OSINT Engine**: Melakukan request terstruktur ke API Sectors untuk data candle, broker flow, mining extension, dan harvesting berita RSS BEI terkurasi.
* **Streaming JSONL Emitter**: Mengirimkan update berkala (`agent_event`, `agent_message_chunk`, `agent_message_complete`) dalam format JSON Lines ke STDOUT untuk ditangkap secara streaming oleh Go daemon.

### C. Local Web Workspace (`/web/`)
* **Framework**: React 18 + Vite + TypeScript + Tailwind CSS + shadcn/ui.
* **Financial Charting**: Visualisasi candlestick dan bar volume dengan penanda khusus titik anomali.
* **Reactive Updates**: Menerima streaming status dan temuan investigasi secara langsung melalui SSE tanpa refresh halaman.

---

## 3. Pola Struktur Folder Repositori Proyek (Configurable Monorepo)

Repositori kode implementasi (*codebase repo*) disusun dengan pola **Polyglot Monorepo** yang modular, namun dengan jalur direktori dan nama modul yang **dinamis dan dapat dikonfigurasi (*customizable & adaptable*)**:

```text
niskava-codebase/                # Root direktori repositori implementasi
├── cmd/
│   └── niskava/                 # Entrypoint Go utama (CLI & daemon runner)
│       └── main.go
├── internal/                    # Modul privat Go Core (dapat diubah/direfaktor secara bebas)
│   ├── cli/                     # Handler perintah CLI (spf13/cobra: investigate, serve, sessions)
│   ├── config/                  # Pengaturan dinamis & parser (~/.niskava/config.yaml / env)
│   ├── db/                      # SQLite persistence (modernc.org/sqlite, zero CGO)
│   ├── ipc/                     # Subprocess IPC runner & scanner streaming JSONL
│   ├── server/                  # HTTP REST & SSE streaming server
│   └── tui/                     # Interactive Terminal UI (charmbracelet/bubbletea)
│
├── engine/                      # Python Agent Engine (nama direktori/package dapat disesuaikan)
│   ├── agent/                   # 7-Stage Pipeline orchestrator & prompt templates
│   ├── quant/                   # Anomali kuantitatif deterministik (NumPy/Pandas Z-Scores)
│   ├── sectors/                 # Sectors v2 API client & disk cache lokal
│   ├── osint/                   # Harvester berita RSS & keterbukaan IDX (trafilatura)
│   ├── memory/                  # Local Graph Memory Engine (NetworkX DiGraph)
│   ├── runner.py                # Entrypoint IPC modul yang dipanggil oleh Go Core
│   ├── pyproject.toml           # Konfigurasi dependensi Python modern
│   └── requirements.txt         # Fallback dependensi standar pip
│
├── web/                         # React SPA Frontend Workspace (dapat dipindah/di-repoint)
│   ├── src/                     # Komponen antarmuka (Cyber-OSINT aesthetic, Recharts/TradingView)
│   ├── package.json             # Dependensi frontend npm/vite
│   ├── vite.config.ts           # Konfigurasi bundler (output dist/)
│   └── tsconfig.json
│
├── Makefile                     # Root build orchestrator & task runner (opsional/fleksibel)
├── go.mod                       # Modul Go root
├── go.sum
├── .gitignore                   # Mengabaikan *.db, venv/, node_modules/, dist/, .env
└── README.md                    # Panduan build & kontribusi kode implementasi
```

---

## 4. Mekanisme Konfigurasi & Resolusi Path Dinamis (*Dynamic Path Resolution*)

Arsitektur Niskava dirancang agar **tidak mengikat mati (*anti-hardcoding*)** lokasi folder atau biner runtime. Setiap path penting di-resolusi secara dinamis dengan urutan hierarki (*precedence order*):

```text
CLI Flags (Override Tertinggi) 
  ──▶ Environment Variables 
    ──▶ File Konfigurasi (~/.niskava/config.yaml) 
      ──▶ Default Relative Path (Fallback)
```

### Tabel Matriks Resolusi Dinamis

| Komponen Sistem | Default Path | Environment Variable | CLI Flag / Config Key | Keterangan & Fleksibilitas |
|---|---|---|---|---|
| **Python Binary** | `python3` (dari `$PATH`) | `NISKAVA_PYTHON_BIN` | `--python-bin` / `engine.python_bin` | Mendukung venv kustom, pyenv, conda, atau path absolut biner Python. |
| **Python Engine Path** | `./engine` (atau module `engine.runner`) | `NISKAVA_ENGINE_PATH` | `--engine-path` / `engine.entrypoint` | Jika direktori diubah (misal menjadi `niskava/` atau paket terinstal `pip install -e .`), runner tetap menemukan entrypoint. |
| **Web UI Assets** | `//go:embed web/dist` | `NISKAVA_WEB_DIR` | `--web-dir` / `web.dist_path` | Mode produksi memakai embedded binary. Mode development dapat mengarahkan ke folder dist lokal atau reverse proxy Vite dev server. |
| **SQLite DB Path** | `~/.niskava/niskava.db` | `NISKAVA_DB_PATH` | `--db-path` / `storage.db_path` | Memungkinkan database diletakkan di lokasi kustom atau memori (`:memory:`) untuk testing isolasi. |
| **HTTP Port** | `8080` | `NISKAVA_PORT` | `--port` / `server.port` | Port lokal REST/SSE dapat dipindah jika 8080 sedang digunakan. |

---

## 5. Hubungan Pemisahan Repositori (*Decoupled Repositories Architecture*)

Proyek Niskava memisahkan antara **Repositori Dokumentasi & Spesifikasi** dengan **Repositori Implementasi Kode**:

1. **Repositori Dokumentasi (`niskava-docs` / repo saat ini):**
   * Berfungsi sebagai *Single Source of Truth (SSoT)*, berisi seluruh spesifikasi produk, fondasi, formula matematika quant, katalog skill agent, ADR, dan konstitusi `AGENTS.md`.
   * Berdiri mandiri, bersih dari artefak kompilasi biner, `node_modules/`, atau dependensi pustaka Python.
2. **Repositori Implementasi Kode (`niskava-codebase` / target repo terpisah):**
   * Menyimpan kode sumber implementasi murni mengikuti struktur folder modular di atas.
   * Mengacu pada spesifikasi di repositori dokumen sebagai panduan implementasi dan kontrak antarmuka (IPC JSONL, REST API, DDL Database).
   * Pengembang atau agen AI bebas melakukan refaktor struktur folder internal repositori kode selama kontrak komunikasi data (IPC) dan resolusi path dinamis tetap dipatuhi.

