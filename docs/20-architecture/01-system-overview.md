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
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ Layer 4: ReAct Cognitive Orchestrator (Multi-Turn Reasoning Loop)    │  │
│  └───────────────────────────────────┬───────────────────────────────────┘  │
│                                      │ Activates Skill SOPs                 │
│                                      ▼                                      │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ Layer 3: Modular Skills Registry (market-anomaly, causality, insider) │  │
│  └───────────────────────────────────┬───────────────────────────────────┘  │
│                                      │ Enforces Deterministic Math          │
│                                      ▼                                      │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ Layer 2: Deterministic Compute Gate (NumPy Anomaly & Flow Firewall)  │  │
│  └───────────────────────────────────┬───────────────────────────────────┘  │
│                                      │ Standardized Tool Calls              │
│                                      ▼                                      │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ Layer 1: Sectors MCP & News Engine Primitives (Sectors API v2)        │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Pembagian Peran & Tanggung Jawab Komponen

### A. Clients Surface (`clients/`)
* **Interactive CLI & TUI (`clients/cli/`)**:
  * Menggunakan `spf13/cobra`, `charmbracelet/bubbletea`, dan `charmbracelet/glamour` untuk menghadirkan terminal REPL percakapan interaktif (Hermes-style) dengan rendering Markdown ala Bloomberg Terminal.
  * Menyediakan interactive setup wizard (`niskava setup`) untuk konfigurasi kredensial (LLM API key/proxy base, Sectors API) dengan validasi ping koneksi live.
  * Mendukung mode investigasi langsung (`niskava investigate <TICKER> --days 30`) dan manajemen sesi (`niskava sessions`).
* **Local Web Workspace (`clients/web/`)**:
  * **Framework**: React 18 + Vite + TypeScript + Tailwind CSS + shadcn/ui.
  * **Financial Charting**: Visualisasi candlestick dan bar volume dengan penanda khusus titik anomali kuantitatif.
  * **Reactive Real-time Canvas**: Menerima streaming token, thought steps, tool calls, dan temuan investigasi secara langsung melalui SSE tanpa refresh halaman.

### B. Backend Architecture (`backend/`)
* **Go Core Daemon (`backend/core/` & `cmd/niskava/`)**:
  * **Single Executable Distributor**: Mengompilasi seluruh aplikasi menjadi 1 file biner mandiri. Seluruh file frontend statis di-embed langsung ke dalam biner menggunakan `//go:embed clients/web/dist`.
  * **Local Web Server (`backend/core/server/`)**: Menyediakan REST API manajemen sesi multi-turn (`/api/chat/sessions`), forking, export, abort, serta Server-Sent Events (SSE) streaming (`/api/chat`) dengan ketahanan zero write timeout (`WriteTimeout: 0`) dan 15-second heartbeat keep-alive ping.
  * **Session Persistence Manager (`backend/core/db/`)**: Berkomunikasi dengan database SQLite lokal (`chat_sessions`, `chat_messages`, `investigations`, `anomalies`, dll.) menggunakan driver murni Go (`modernc.org/sqlite`) tanpa kebutuhan compiler CGO. Dilengkapi *self-healing zombie recovery* saat inisialisasi.
  * **Subprocess IPC Runner (`backend/core/ipc/`)**: Mengelola eksekusi child process Python secara aman dengan scanning streaming JSON Lines menggunakan buffer 1 MB.
* **Python Agent Engine (`backend/engine/`)**:
  * **Universal Model-Agnostic ReAct Loop (`backend/engine/agent/`)**: Mengelola dialog multi-turn, pemanggilan tool deterministik otonom, dan sintesis bukti menggunakan antarmuka standar OpenAI-compatible (`/chat/completions`) tanpa vendor lock-in. Dilengkapi **Progressive Skill Disclosure (ADR-11)** dengan 4 gateway primitives (`execute_skill`, `query_sectors`, `search_news`, `query_memory`) yang memangkas ukuran prompt sistem hingga ~716 token serta menyematkan rekomendasi langkah lanjutan proaktif.
  * **Deterministic Quant Anomaly (`backend/engine/quant/`)**: Menghitung $Z$-score volume ($V_z$), abnormal return ($R_t$), divergensi sektor ($D_t$), dan foreign flow $Z$-score ($F_z$) menggunakan library NumPy murni sesuai Hukum 1. LLM dilarang berhitung mandiri.
  * **Sectors v2 API Client & News Engine (`backend/engine/sectors/`)**: Melakukan request terstruktur ke API Sectors untuk data candle, broker flow, mining extension, serta penarikan berita bursa & keterbukaan informasi emiten terkurasi (`/v2/news/`) dengan cache lokal disk.
  * **Local Graph Memory (`backend/engine/memory/`)**: In-memory NetworkX DiGraph yang disinkronkan ke tabel SQLite `memory_nodes` & `memory_edges` dengan decay temporal.
  * **Streaming JSONL Emitter**: Mengirimkan update berkala (`agent_thought`, `agent_tool_call`, `agent_observation`, `agent_message_chunk`, `agent_message_complete`) dalam format JSON Lines ke STDOUT.

---

## 3. Pola Struktur Folder Repositori Proyek (Polyglot Monorepo)

Repositori kode implementasi disusun dengan pola **Polyglot Monorepo** standar industri dengan pemisahan tegas antara antarmuka klien (*clients*) dan layanan backend (*backend*):

```text
niskava/                         # Root direktori repositori implementasi
├── cmd/
│   └── niskava/                 # Entrypoint Go utama (CLI & daemon runner)
│       └── main.go
│
├── clients/                     # Segregated Client Surfaces
│   ├── cli/                     # CLI commands & Bubble Tea TUI
│   │   ├── cmd/                 # Cobra subcommands (root, investigate, serve, sessions, setup)
│   │   ├── tui/                 # Bubbletea interactive terminal app & Glamour renderer
│   │   └── setup.go             # Interactive Setup Wizard
│   └── web/                     # React 18 + Vite + Tailwind CSS Workspace
│       ├── src/                 # Komponen UI (Market Intelligence aesthetic, Recharts/Lightweight)
│       ├── package.json         # Dependensi frontend npm/vite
│       ├── vite.config.ts       # Konfigurasi bundler (output dist/)
│       └── tsconfig.json
│
├── backend/                     # Segregated Backend Architecture
│   ├── core/                    # Go Core Daemon
│   │   ├── config/              # Pengaturan dinamis & parser (~/.niskava/config.yaml / env)
│   │   ├── db/                  # SQLite persistence (modernc.org/sqlite, zero CGO, WAL)
│   │   ├── ipc/                 # Subprocess IPC runner & scanner streaming JSONL
│   │   ├── security/            # Sanitasi input, credential masking & path validation
│   │   └── server/              # HTTP REST (/api/chat/*) & SSE streaming server
│   │
│   └── engine/                  # Python Agent Engine (3.11+)
│       ├── agent/               # ReAct Agent loop, prompt templates, tools registry
│       ├── quant/               # Anomali kuantitatif deterministik (NumPy/Pandas Z-Scores)
│       ├── sectors/             # Sectors v2 API client, news engine & disk cache lokal
│       ├── memory/              # Local Graph Memory Engine (NetworkX DiGraph)
│       ├── tests/               # 272 unit tests komprehensif engine Python (100% green)
│       ├── runner.py            # Entrypoint IPC headless investigation pipeline
│       └── pyproject.toml       # Dependensi modern Python (uv / pip)
│
├── docs/                        # SSoT Documentation Hub
├── scripts/                     # Operational scripts (database seed, health check)
├── Makefile                     # Root build orchestrator & multi-language task runner
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
| **Python Engine Path** | `./backend/engine` (atau `backend.engine.runner`) | `NISKAVA_ENGINE_PATH` | `--engine-path` / `engine.entrypoint` | Runner menemukan entrypoint baik dalam root repo maupun lingkungan terpasang. |
| **Web UI Assets** | `//go:embed clients/web/dist` | `NISKAVA_WEB_DIR` | `--web-dir` / `web.dist_path` | Mode produksi memakai embedded binary. Mode development dapat mengarahkan ke `clients/web/dist` lokal. |
| **SQLite DB Path** | `~/.niskava/niskava.db` | `NISKAVA_DB_PATH` | `--db-path` / `storage.db_path` | Memungkinkan database diletakkan di lokasi kustom atau memori (`:memory:`) untuk testing isolasi. |
| **HTTP Port** | `20128` (atau `8080`) | `NISKAVA_PORT` | `--port` / `server.port` | Port lokal REST/SSE dapat dipindah sesuai ketersediaan port. |
| **LLM Inference Timeout** | `60.0` detik | `NISKAVA_LLM_TIMEOUT` | `/timeout` / `preferences.llm_timeout_secs` | Skala adaptif otomatis per iterasi ReAct: $\text{Base} + (N_{\text{obs}} \times 10\text{s})$, batas 10–300 detik. |
| **LLM API Base** | `http://localhost:20128/v1` | `OPENAI_BASE_URL` | `--llm-api-base` / `auth.openai_base_url` | Endpoint OpenAI-compatible (9router local proxy, Ollama, OpenRouter, vLLM). |
| **LLM Model** | `hermes` | `OPENAI_MODEL` | `--llm-model` / `auth.openai_model` | Model id universal tanpa vendor lock-in. |

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

