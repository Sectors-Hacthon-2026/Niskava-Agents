# Niskava Agent — Financial OSINT & Market Intelligence Orchestration Platform

> **"Don't just answer questions. Investigate them."**

[![Track](https://img.shields.io/badge/Sectors%20Hackathon%202026-Track%201%3A%20AI%20Agents%20%26%20Assistants-blue.svg)](https://hackathon.sectors.app/)
[![Target Market](https://img.shields.io/badge/Market-IDX%20%28Indonesia%20Stock%20Exchange%29-green.svg)](#)
[![Architecture](https://img.shields.io/badge/Architecture-Hybrid%20%28Go%20%2B%20Python%20%2B%20React%29-orange.svg)](#)
[![Storage](https://img.shields.io/badge/Storage-Local--First%20SQLite-lightgrey.svg)](#)

---

## 📌 Executive Summary

**Niskava Agent** is an autonomous financial OSINT (*Open Source Intelligence*) and market intelligence orchestration platform engineered specifically for the **Indonesia Stock Exchange (IDX)**. Niskava bridges the critical gap between quantitative market facts (powered by **Sectors Financial API v2**) and external qualitative intelligence (IDXnet regulatory disclosures, corporate announcements, financial media, and market signals).

Niskava is **not a generic financial chatbot nor an API wrapper**. Its core engine functions as an *Investigative Orchestration Platform* that dynamically:
1. Formulates investigation hypotheses and SOPs (*Skill Routing*).
2. Retrieves quantitative baseline facts with zero hallucinations (*Sectors API v2*).
3. Evaluates volume spikes, price abnormal returns, and net foreign inflows deterministically (*NumPy Z-Score Engine*).
4. Conducts targeted contextual OSINT harvesting anchored around anomaly timestamps ($T_{\text{anomaly}} \pm 2\text{ days}$).
5. Correlates findings, verifies temporal causality, and classifies evidence within a strict 3-tier verification taxonomy (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
6. Delivers verified intelligence across two unified interfaces: an **Interactive CLI Terminal (Go/Bubbletea)** and a **Local Web Workspace (React/Vite)**.

---

## 🏛️ System Architecture: Tripartite Hybrid Stack

Niskava adopts a 3-tier hybrid stack (*Go + Python + React*) designed for execution speed, quantitative rigor, and high-fidelity user experience:

```
┌─────────────────────────────────────────────────────────────┐
│                          GO CORE                            │
│  - Primary Gateway & Single Binary Executable CLI (`niskava`)│
│  - Interactive TUI Terminal (charmbracelet/bubbletea)       │
│  - High-Performance REST & SSE Server (Real-time Streaming) │
│  - Pure-Go SQLite Persistence (modernc.org/sqlite)         │
│  - Static Web UI Bundler (//go:embed)                      │
└──────────────────────────────┬──────────────────────────────┘
                               │ Local IPC (JSON Lines / STDIN-STDOUT)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    PYTHON AGENT ENGINE                      │
│  [Layer 4] ReAct Cognitive Orchestrator & Memory Engine     │
│  [Layer 3] Modular Skills Registry (Domain SOP Modules)     │
│  [Layer 2] Deterministic Compute Gate (NumPy Anomaly Math)  │
│  [Layer 1] Sectors MCP Server & Dual-Engine OSINT Harvester │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│               REACT SPA WEB WORKSPACE (Vite)                │
│  - Cyber-OSINT / Bloomberg Terminal Visual Aesthetic        │
│  - Interactive Candlestick Chart & Volume Anomaly Markers   │
│  - Real-Time Thinking Stream via Server-Sent Events (SSE)   │
│  - Interactive Evidence Matrix & Timeline Graph             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔄 7-Stage Investigation Pipeline (SOP)

Every investigation follows a strict 7-stage domain protocol modeled after professional equity research workflows:

```
┌─────────────┐     ┌──────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ 1. INITIATION│ ──> │2. SECTORS_BASELINE│ ──> │ 3. QUANT_ANOMALY │ ──> │4. GAP_DETECTION │
└─────────────┘     └──────────────────┘     └──────────────────┘     └─────────────────┘
                                                                               │
┌────────────────────────┐     ┌────────────────────────┐                      │
│7. SYNTHESIS & STREAMING│ <── │ 6. EVIDENCE_CORRELATION│ <── ┌─────────────────┴─┐
└────────────────────────┘     └────────────────────────┘     │ 5. OSINT_HARVEST  │
                                                              └───────────────────┘
```

1. **Stage 1: INITIATION** — Accepts target ticker symbol (e.g., `ANTM`), initializes local investigation session (`PENDING` $\to$ `RUNNING`).
2. **Stage 2: SECTORS_BASELINE** — Pulls 30–90 days daily OHLCV candlestick data, financial statements, Net Foreign Flow, corporate actions, and special notations from Sectors v2 API (utilizing local caching).
3. **Stage 3: QUANT_ANOMALY** — Executes non-generative NumPy evaluation: computes 20-day Volume Moving Average (MA20), Volume Z-Score ($V_z$), Abnormal Return ($R_t$), Sector Divergence ($D_t$), and Foreign Flow Z-Score ($F_z$).
4. **Stage 4: GAP_DETECTION** — Formulates targeted investigation hypotheses centered around anomaly dates ($T_{\text{anomaly}} \pm 2\text{ days}$).
5. **Stage 5: OSINT_HARVEST** — Fetches news via Sectors Unified News API and executes unblocked boolean dorking on official IDXnet filings and mainstream Indonesian financial press (Kontan, Bisnis, CNBC Indonesia), sanitized via `trafilatura`.
6. **Stage 6: EVIDENCE_CORRELATION** — Evaluates temporal precedence and assigns verification status (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
7. **Stage 7: SYNTHESIS & STREAMING** — Generates structured findings and event timelines, persists data to local SQLite, and broadcasts agent thinking steps via Server-Sent Events (SSE) to the Web UI and CLI.

---

## 📜 Architectural Invariants (The 6 Laws)

Niskava's development is strictly governed by 6 non-negotiable architectural laws:

1. **Law 1: Deterministic Before Generative**  
   LLMs are prohibited from calculating statistics, Z-scores, or sector divergences. All quantitative metrics are computed deterministically via NumPy/Pandas before prompting an LLM, eliminating numerical hallucinations entirely.
2. **Law 2: Strict Financial Non-Advisory Boundary (Sectors Hackathon Rule 12)**  
   Niskava is an investigative intelligence platform, **NOT an investment advisor**. The system **MUST NEVER** output direct BUY/SELL recommendations, price targets, or personalized financial advice. All findings must fit into the 3-Tier Verification Taxonomy (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
3. **Law 3: Prohibition of Automated Trade Execution (Sectors Hackathon Rule 06)**  
   Zero brokerage execution APIs or order routing logic. The system is strictly read-only market intelligence.
4. **Law 4: Local-First Data Sovereignty**  
   All user investigation sessions, evidence graphs, and local caches reside locally at `~/.niskava/niskava.db` using SQLite with Write-Ahead Logging (`modernc.org/sqlite` pure-Go zero CGO).
5. **Law 5: Credit Budget Discipline & Local Caching**  
   Strictly protects the 1,000 Sectors API credit grant. Historical daily candlestick data ($T < \text{today}$) is permanently cached (`expires_at = NULL`), incurring 0 credit cost on repeated queries.
6. **Law 6: Local Conversational Graph Memory Engine**  
   Maintains associative context across multi-day sessions using local SQLite graph storage (`memory_nodes` & `memory_edges`) loaded into in-memory Python `NetworkX.DiGraph` with $k \le 2$ hop Ego-Graph traversal.

---

## 🎯 Verification Taxonomy & Confidence Rubric

Every piece of evidence is assigned a standardized discrete confidence score:

| Score | Verification Level | Evidence Description / Source |
|:---:|:---|:---|
| **`1.00`** | **EXTRACTED** | Directly backed by official Sectors API quantitative records or formal IDXnet regulatory filings. |
| **`0.95`** | **Direct Structural Evidence** | Explicit timestamp correlation with official corporate press releases or regulatory disclosures. |
| **`0.85`** | **Strong Inference** | High temporal correlation with major mainstream national financial media coverage (Kontan, Bisnis, CNBC). |
| **`0.75`** | **Reasonable Inference** | Plausible catalyst from industry trends matching sector divergence patterns. |
| **`0.65`** | **Weak Inference** | Unverified market commentary, social media sentiment, or unconfirmed forum rumors. |
| **`0.55`** | **Speculative** | Co-occurrence without verified temporal causality. |

---

## 🚀 Installation & Quick Start

### 📋 Prerequisites
* **Go**: v1.22 or higher
* **Python**: v3.11 or higher (with `numpy`, `pandas`, `networkx`, `trafilatura`)
* **Node.js & npm**: v18+ (for frontend development)

### 🔨 Installation Steps

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
   cd Niskava-Agents
   ```

2. **Set Up Python Engine**:
   ```bash
   cd engine
   python -m venv venv
   # Linux/macOS:
   source venv/bin/activate
   # Windows PowerShell:
   .\venv\Scripts\Activate.ps1

   pip install -r requirements.txt
   cd ..
   ```

3. **Configure API Keys**:
   Create a configuration file at `~/.niskava/config.yaml` or set environment variables:
   ```bash
   export SECTORS_API_KEY="your_sectors_api_key_here"
   export GEMINI_API_KEY="your_gemini_api_key_here"
   ```

4. **Build the Go Binary**:
   ```bash
   go build -o niskava cmd/niskava/main.go
   ```

---

## 💻 CLI & Local Web Workspace Usage

### 1. Launch Investigation via CLI
```bash
# Investigate unusual activity on ticker ANTM (Aneka Tambang Tbk)
./niskava investigate ANTM --days 30

# Example Interactive TUI Output:
# [●] Initializing Investigation: ANTM
#  ├── [1/4] Sectors Baseline Data .......... [OK] 30 trading days retrieved
#  ├── [2/4] Quantitative Anomaly Detection . [ALERT] Volume surge (3.8σ) on Sep 12
#  ├── [3/4] OSINT Contextual Gathering ..... [OK] 4 corporate filings & news found
#  └── [4/4] Cross-Verification & Correlation [OK] 3 validated findings generated
```

### 2. Launch Local Web Workspace
```bash
# Start local server and open browser automatically
./niskava serve --port 8080 --open
```
Navigate to `http://localhost:8080` to access the interactive candlestick chart, investigation timeline, real-time SSE thinking stream, and evidence matrix.

### 3. Manage Sessions
```bash
# List all historical investigation sessions stored in SQLite
./niskava sessions

# Resume a specific investigation session
./niskava resume INV-2026-0042
```

---

## 📁 Repository Structure

```text
Niskava-Agents/
├── AGENTS.md                  # Single Source of Truth (SSoT) & Architectural Invariants
├── Makefile                   # Build & test automation scripts
├── README.md                  # Main repository documentation
├── cmd/                       # Application binary entry points
│   └── niskava/
│       └── main.go            # Primary CLI binary entry point
├── clients/                   # User Surfaces (Presentation & Interfaces)
│   ├── cli/                   # Terminal Client (Cobra CLI subcommands)
│   │   └── tui/               # Interactive Bubbletea Terminal UI components
│   └── web/                   # React SPA Web Workspace (Vite + Tailwind + shadcn/ui)
│       ├── src/               # React source code (App.tsx, components)
│       ├── package.json       # Node.js dependencies
│       └── vite.config.ts     # Vite bundler configuration
├── backend/                   # Core Backend & Cognitive Computation Subsystems
│   ├── core/                  # Go Core Daemon & Persistence Layer
│   │   ├── config/            # Configuration loader & parser
│   │   ├── db/                # Local-first SQLite database driver (modernc.org/sqlite)
│   │   ├── ipc/               # Go-to-Python IPC communication bridge
│   │   ├── security/          # Security bounds & data sanitization
│   │   └── server/            # REST API & Server-Sent Events (SSE) server
│   └── engine/                # Python Agent Engine (AI, Quant Math, OSINT)
│       ├── agent/             # Universal ReAct loop (react_agent.py) & pipeline
│       ├── memory/            # Graph memory engine (graph_memory.py)
│       ├── osint/             # Resilient dual-engine scraper (harvester.py)
│       ├── quant/             # Deterministic NumPy Z-Score math (anomaly.py)
│       ├── sectors/           # Sectors API v2 REST/MCP client (client.py)
│       ├── skills/            # Modular SOP domain skills
│       ├── tests/             # Consolidated Python unit test suite (74 tests)
│       ├── runner.py          # IPC execution script
│       └── requirements.txt   # Python dependencies
└── docs/                      # Technical Documentation & Architecture Decision Records (ADR)
    ├── 00-foundations/        # Vision, principles, and glossary
    ├── 10-product/            # Product scope, personas, and hackathon strategy
    ├── 20-architecture/       # DB schema, Sectors API specs, OSINT, IPC, security
    ├── 30-agent/              # Agent skills catalog, graph memory, & pipeline SOP
    └── 90-decisions/          # Architecture Decision Records (ADR 01 - 08)
```

---

## ⚠️ Financial Non-Advisory Disclaimer

> **IMPORTANT DISCLAIMER:**  
> **Niskava Agent** is an automated market intelligence, OSINT, and data research platform. All findings, anomalies, and correlated evidence generated by Niskava Agent are derived from historical market data, public regulatory filings, and news media. They **DO NOT** constitute financial advice, investment recommendations, price targets, or solicitations to buy or sell any securities. All analyses are strictly for informational and investigative research purposes. Users are solely responsible for their own investment decisions.

---

## 🏆 Competition & License

* **Hackathon**: [Sectors Hackathon Indonesia 2026](https://hackathon.sectors.app/)
* **Category**: **Track 1: AI Agents & Assistants**
* **License**: MIT License
