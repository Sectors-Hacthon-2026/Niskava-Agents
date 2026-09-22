# Niskava Agent — Financial OSINT & Intelligence Orchestration Platform

> **"Don't just answer questions. Investigate them."**  
> Platform orkestrator investigasi finansial dan intelijen pasar otonom berbasis *evidence-first*, menggabungkan data fundamental/pasar **Sectors MCP** dengan sinyal eksternal **OSINT (News, Public Filings, Corporate Actions)**.

---

## 1. Executive Summary & Value Proposition

**Niskava Agent** adalah sistem orkestrasi AI otonom untuk investigasi finansial. Niskava bukan sekadar chatbot finansial dan bukan pembungkus (wrapper) API biasa. Core sistemnya adalah **Investigative Orchestrator** yang secara dinamis:
1. Menentukan metodologi investigasi (*Skills*).
2. Menarik data kuantitatif dasar (*Sectors MCP*).
3. Mendeteksi anomali pasar dan kesenjangan informasi (*Evidence Gap Detection*).
4. Melakukan pencarian intelijen eksternal kontekstual (*Targeted OSINT*).
5. Mengkorelasikan bukti, menilai kausalitas, dan memvalidasi temuan (*Evidence Layer*).
6. Menyajikan hasil terverifikasi dalam dua antarmuka terpadu: **CLI Terminal** dan **Local Web Workspace**.

```
                ┌─────────────────────────────────────────────────────────┐
                │                     USER INQUIRY                        │
                │        "Investigate unusual movement on ANTM"           │
                └────────────────────────────┬────────────────────────────┘
                                             │
                                             ▼
                ┌─────────────────────────────────────────────────────────┐
                │                   NISKAVA ORCHESTRATOR                  │
                │           Intent → Skill Router → Plan Pipeline         │
                └─────────────┬─────────────────────────────┬─────────────┘
                              │                             │
                     [1. Baseline Facts]           [2. Context & Gaps]
                              ▼                             ▼
                ┌───────────────────────────┐ ┌───────────────────────────┐
                │        SECTORS MCP        │ │   EXTERNAL OSINT ENGINE   │
                │ Price, Volume, Financials │ │ News, Filings, Disclosures│
                └─────────────┬─────────────┘ └─────────────┬─────────────┘
                              │                             │
                              └──────────────┬──────────────┘
                                             │
                                             ▼
                ┌─────────────────────────────────────────────────────────┐
                │                     EVIDENCE LAYER                      │
                │  Correlation • Audit Trail • Causality Verification     │
                │       [SUPPORTED | UNCERTAIN | CONTRADICTED]            │
                └────────────────────────────┬────────────────────────────┘
                                             │
                                             ▼
                       ┌─────────────────────┴─────────────────────┐
                       │                                           │
                       ▼                                           ▼
             ┌───────────────────┐                       ┌───────────────────┐
             │    NISKAVA CLI    │                       │   LOCAL WORKSPACE │
             │  Terminal/DevOps  │                       │  Vite/React + UI  │
             └───────────────────┘                       └───────────────────┘
```

---

## 2. Arsitektur Teknologi: Hybrid Stack (Go + Python + React)

Untuk memaksimalkan performa, *developer experience*, kapabilitas AI, dan estetika demo hackathon, Niskava menggunakan arsitektur hybrid:

```
┌────────────────────────────────────────────────────────────────────────┐
│                              GO CORE                                   │
│                     (Single Binary Distribution)                       │
│                                                                        │
│  ┌──────────────────────┐  ┌─────────────────┐  ┌───────────────────┐  │
│  │     CLI Engine       │  │  Local Server   │  │   Embedded UI     │  │
│  │ (Cobra + Bubbletea)  │  │ (REST / SSE API)│  │  (//go:embed)     │  │
│  └──────────┬───────────┘  └────────┬────────┘  └─────────┬─────────┘  │
│             │                       │                     │            │
│             └───────────────────────┼─────────────────────┘            │
│                                     │                                  │
│                             ┌───────▼────────┐                         │
│                             │ Session Storage│                         │
│                             │ (SQLite DB)    │                         │
│                             └───────┬────────┘                         │
└─────────────────────────────────────┼──────────────────────────────────┘
                                      │ IPC / Local REST
                                      ▼
┌────────────────────────────────────────────────────────────────────────┐
│                         PYTHON AGENT ENGINE                            │
│                                                                        │
│  ┌──────────────────────┐  ┌─────────────────┐  ┌───────────────────┐  │
│  │ Quant Anomaly Engine │  │ Sectors Client  │  │  Agent Reasoning  │  │
│  │ (Z-score, Bollinger) │  │  (MCP / REST)   │  │  & OSINT Scraper  │  │
│  └──────────────────────┘  └─────────────────┘  └───────────────────┘  │
└────────────────────────────────────────────────────────────────────────┘
```

### Komponen 1: Clients Surface (`clients/cli/` & `clients/web/`)
* **Peran**: Antarmuka pengguna visual dan terminal interaktif.
* **Fitur Utama**:
  * **Interactive Terminal UI (`clients/cli/`)**: Menggunakan `spf13/cobra`, `charmbracelet/bubbletea`, dan `glamour` untuk menghadirkan pengalaman REPL percakapan interaktif bernuansa Bloomberg Terminal dan setup wizard terpandu.
  * **Web Workspace (`clients/web/`)**: React 18 + Vite + Tailwind CSS + shadcn/ui. Menyajikan visual charting anomali candlestick, timeline kejadian kronologis, kartu temuan bukti terverifikasi, serta live SSE streaming.

### Komponen 2: Go Core Daemon (`backend/core/` & `cmd/niskava/`)
* **Peran**: Gateway utama, CLI entrypoint, server lokal, dan penyaji web dashboard mandiri.
* **Fitur Utama**:
  * **Single Executable**: Binary tunggal `niskava` yang meng-embed seluruh aset frontend statis menggunakan Go `//go:embed clients/web/dist`.
  * **Local HTTP & SSE Server**: Menyediakan REST API untuk session management (`/api/chat/sessions`), branching forking, transcript export, serta Server-Sent Events (SSE) untuk streaming *real-time ReAct thinking steps* agent.
  * **Local Persistence**: SQLite (`modernc.org/sqlite` murni Go tanpa CGO) dengan WAL mode, foreign keys, dan *self-healing zombie recovery* saat startup.

### Komponen 3: Python Agent Engine (`backend/engine/`)
* **Peran**: Otak analisis otonom ReAct universal, eksekutor SOP skills, dan integrasi data bursa.
* **Fitur Utama (Arsitektur 4-Layer)**:
  * **Layer 4: Cognitive ReAct Loop**: Mengelola alur penalaran bertahap (*Thought* $\to$ *Tool Call* $\to$ *Observation* $\to$ *Synthesis*) menggunakan Universal Model-Agnostic LLM endpoint (`NISKAVA_LLM_API_BASE`, kompatibel dengan 9router local proxy `http://localhost:20128/v1`, Ollama, OpenRouter, Gemini), pemetaan Ego-Graph memory, dan penyusunan temuan bukti.
  * **Layer 3: Modular Skills Registry**: Menyediakan SOP analisis terstandarisasi (`market-anomaly-recon`, `event-causality-audit`, `insider-bandarmology-forensic`, `financial-health-stress-test`).
  * **Layer 2: Deterministic Compute Gate**: Menghitung anomali teknikal & fundamental secara pasti via NumPy (Volume Z-Score, Abnormal Return, Foreign Flow Z-Score) sebelum LLM diaktifkan, memutus halusinasi angka secara total.
  * **Layer 1: Sectors MCP & Dual-Engine OSINT**: Adapter data bursa terstandarisasi via Sectors MCP dan panen berita/keterbukaan informasi resmi BEI secara terarah pada jendela $T_{\text{anomaly}} \pm 2\text{ hari}$ dengan disk cache lokal.

---

## 3. Workflow Investigasi & Evidence Gap Detection

Niskava tidak langsung menelusuri web secara acak. Investigasi dilakukan secara bertahap (*hypothesis-driven*):

```
1. INITIATION       User input: "Investigate ANTM price surge last week"
        │
2. BASELINE         Sectors MCP menarik data 30 hari:
                    - Harga, volume harian, rasio valuasi, perbandingan sektor.
        │
3. DETECT ANOMALY   Quant engine mendeteksi:
                    - Tanggal: 12 September.
                    - Anomali: Volume melonjak 3.8x di atas MA20, harga +8.2%.
                    - Status: Sektor tambang netral (+0.4%) -> Anomali spesifik emiten!
        │
4. EVIDENCE GAP     Agent bertanya: "Ada peristiwa apa seputar ANTM pada 10-12 September?"
        │
5. OSINT HARVEST    Agent menjalankan pencarian terarah:
                    - Query: "ANTM corporate action nickel smelter Sept 2026"
                    - Ditemukan: Pengumuman peresmian smelter baru dan kontrak ekspor.
        │
6. CORRELATION      Agent mengorelasikan timestamp berita dengan lonjakan volume di Sectors.
        │
7. AUDIT & REPORT   Findings dinilai:
                    - Finding 1: Volume Spike -> Status: SUPPORTED (Sectors Data)
                    - Finding 2: Catalyst Smelter -> Status: SUPPORTED (Official Disclosure)
                    - Causality: LIKELY CORRELATED (High Confidence)
```

---

## 4. Skills (Metodologi Investigasi)

Skills memandu agent agar tidak bertindak serampangan, melainkan mengikuti SOP analis finansial:

| Nama Skill | Trigger | Sumber Data Utama | Output Kunci |
|---|---|---|---|
| `market-anomaly` | Pergerakan harga/volume tidak biasa | Sectors historical data | Anomaly date, magnitude, sector divergence |
| `company-research` | Profil fundamental & valuasi emiten | Sectors company financials | Health score, revenue trend, peer ranking |
| `event-correlation` | Ada gap antara pergerakan & fundamental | News, IDX Filings, Web Search | Timeline kejadian, korelasi sentimen vs volume |
| `financial-health` | Penurunan kinerja atau potensi risiko | Sectors balance sheet & cash flow | Debt maturity, liquidity warning, red flags |
| `peer-comparison` | Analisis industri & komparasi saham | Sectors sector metrics | Relative valuation, market share shift |

---

## 5. Model Evidence & Boundary Finansial

### Taksonomi Status Bukti
Setiap kesimpulan agent wajib memiliki label verifikasi:
* **`SUPPORTED`**: Didukung penuh oleh data kuantitatif Sectors atau dokumen resmi keterbukaan informasi.
* **`UNCERTAIN`**: Terdapat indikasi atau korelasi waktu, tetapi belum ada bukti kausalitas langsung (misal: rumor media sosial/blog).
* **`CONTRADICTED`**: Klaim pasar bertolak belakang dengan fakta laporan keuangan Sectors.

### Financial Advice Boundary
Untuk kepatuhan hukum dan etika analisis:
* ✅ **Fokus**: Investigasi faktual, anomali data, deteksi sinyal, pemetaan bukti, dan korelasi waktu.
* ❌ **Dilarang**: Memberikan rekomendasi beli/jual langsung (*BUY/SELL*), target harga pasti, atau nasihat investasi personal.

---

## 6. Antarmuka: CLI & Web Dashboard

### A. Pengalaman CLI (Power User & Otomasi)
```bash
# Menjalankan investigasi
niskava investigate ANTM --days 30

# Menampilkan riwayat investigasi
niskava sessions

# Membuka kembali sesi tertentu
niskava resume INV-2026-001

# Menjalankan web server dashboard lokal
niskava serve --port 8080 --open
```

Contoh visualisasi terminal:
```text
$ niskava investigate ANTM

[●] Initializing Investigation: ANTM (Aneka Tambang Tbk)
 ├── [1/4] Sectors Baseline Data .................... [OK] 30 trading days retrieved
 ├── [2/4] Quantitative Anomaly Detection ........... [ALERT] Volume surge (3.8σ) on Sep 12
 ├── [3/4] OSINT Contextual Gathering ............... [OK] 4 corporate filings & news found
 └── [4/4] Cross-Verification & Correlation ......... [OK] 3 validated findings generated

─────────────────────────────────────────────────────────────────────────────
FINDINGS SUMMARY:
[SUPPORTED]   Unusual trading volume spike detected on Sep 12 (Volume: 184M vs Avg: 48M).
[SUPPORTED]   Official disclosure: Completion of nickel processing unit expansion.
[UNCERTAIN]   Market rumor regarding imminent acquisition (Unverified forum source).

Investigation saved as session: INV-2026-0042
Run 'niskava serve' to view interactive timeline & evidence graph.
```

### B. Pengalaman Local Web Dashboard (`localhost:8080`)
* **Live Investigation Stream**: Log proses investigasi agent secara real-time via SSE.
* **Interactive Candlestick + Volume Chart**: Titik anomali diberi penanda (*flagged markers*).
* **Investigation Timeline**: Kolom kronologis kejadian sebelum dan sesudah anomali.
* **Evidence Cards**: Rincian setiap bukti dengan tingkat keyakinan (*confidence score*) dan sumber data yang dapat diklik.

---

## 7. Rencana MVP untuk Hackathon

Untuk memastikan deliverable selesai tepat waktu dan memiliki daya pikat demo maksimal:

| Milestone | Deliverable | Keterangan |
|---|---|---|
| **Phase 1: Core Foundation** | • Skema SQLite lokal<br>• Integrasi Sectors API/MCP di Python<br>• Script deteksi anomali kuantitatif | Menghasilkan dataset anomali harga/volume yang valid untuk saham target (misal ANTM/BBRI). |
| **Phase 2: Agent Orchestration** | • Pipeline Orchestrator Python<br>• Targeted OSINT search (Google News/Tavily/DuckDuckGo)<br>• Korelasi evidence & output JSON terstruktur | Agent mampu menjalankan satu siklus investigasi lengkap dan menyimpannya ke session DB. |
| **Phase 3: Go CLI & Daemon** | • Binary CLI Go (`niskava investigate`, `niskava sessions`)<br>• REST API & SSE handler lokal di Go | CLI dapat memanggil engine Python dan menampilkan output streaming. |
| **Phase 4: Web Dashboard** | • UI Vite + React + Tailwind + shadcn/ui<br>• Chart pergerakan saham & visual timeline bukti<br>• Bundle UI ke dalam binary Go (`//go:embed`) | Dashboard interaktif berjalan mulus di browser lokal. |
| **Phase 5: Pitch & Demo Prep** | • Skenario demo "Golden Path" (Saham ANTM)<br>• Dokumentasi README & video demo | Presentasi fokus pada demo perbandingan: Chatbot biasa vs Niskava Investigator. |
