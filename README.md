# Niskava Agent — Financial OSINT & Market Intelligence Orchestration Platform

> **"Don't just answer questions. Investigate them."**

[![Track](https://img.shields.io/badge/Sectors%20Hackathon%202026-Track%201%3A%20AI%20Agents%20%26%20Assistants-blue.svg)](https://hackathon.sectors.app/)
[![Target Market](https://img.shields.io/badge/Market-IDX%20%28Indonesia%20Stock%20Exchange%29-green.svg)](#)
[![Architecture](https://img.shields.io/badge/Architecture-Hybrid%20%28Go%20%2B%20Python%20%2B%20React%29-orange.svg)](#)
[![Storage](https://img.shields.io/badge/Storage-Local--First%20SQLite-lightgrey.svg)](#)
[![Documentation](https://img.shields.io/badge/Docs-public%2Fdocs-purple.svg)](public/docs/README.md)

---

## 📌 Executive Summary

**Niskava Agent** is an autonomous financial OSINT (*Open Source Intelligence*) and market intelligence orchestration platform engineered specifically for the **Indonesia Stock Exchange (IDX)**. Niskava bridges the critical gap between quantitative market facts (powered by **Sectors Financial API v2**) and external qualitative intelligence (IDXnet regulatory disclosures, corporate announcements, financial media, and market signals).

Niskava is **not a generic financial chatbot nor an API wrapper**. Its core engine functions as an *Investigative Orchestration Platform* that dynamically:
1. Formulates investigation hypotheses and domain SOPs (*Skill Routing*).
2. Retrieves quantitative baseline facts with zero numerical hallucinations (*Sectors API v2*).
3. Evaluates volume spikes, price abnormal returns, and net foreign inflows deterministically (*NumPy Z-Score Engine*).
4. Conducts targeted contextual OSINT harvesting anchored around anomaly timestamps ($T_{\text{anomaly}} \pm 2\text{ days}$).
5. Correlates findings, verifies temporal causality, and classifies evidence within a strict 3-tier verification taxonomy (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
6. Delivers verified intelligence across two unified interfaces: an **Interactive CLI Terminal (Go/Bubbletea)** and a **Local Web Workspace (React/Vite)**.

---

## 📚 Detailed Public Documentation & User Guides

Complete, in-depth documentation is available in the [`public/docs/`](public/docs/README.md) directory:

| Document | Direct Link | Content Highlights |
|---|---|---|
| 📖 **Documentation Hub** | **[public/docs/README.md](public/docs/README.md)** | Sitemap, quick links, and documentation overview. |
| 📥 **Installation Guide** | **[public/docs/installation.md](public/docs/installation.md)** | Step-by-step setup (Go 1.22+, Python 3.11+), virtual environment, API Keys, setup wizard, & binary compilation. |
| 🎮 **User Guide** | **[public/docs/user-guide.md](public/docs/user-guide.md)** | Main HUD Launcher, Interactive TUI REPL, **Up/Down arrow prompt history**, **Live Braille spinner**, slash commands, Web Canvas, & Telegram bot. |
| 🏛️ **Features & Architecture** | **[public/docs/features-and-architecture.md](public/docs/features-and-architecture.md)** | Tripartite Hybrid Stack, The 6 Invariant Laws, 7-Stage Pipeline Protocol, NumPy math, & verification confidence rubric. |
| ❓ **Troubleshooting & FAQ** | **[public/docs/troubleshooting-and-faq.md](public/docs/troubleshooting-and-faq.md)** | Solving AI connection errors, SQLite WAL locks, Sectors credit conservation, offline mode (`--offline`), & FAQ. |

---

## 🚀 Quick Start Guide

### 1. Clone & Run Interactive Setup Wizard
```bash
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents

# Run interactive setup wizard (configures API keys & creates Python virtual environment)
go run ./cmd/niskava setup
```

### 2. Launch Interactive Terminal UI (TUI)
```bash
# Launch interactive TUI HUD Launcher Menu
go run ./cmd/niskava
```

Or compile a single standalone binary:
```bash
go build -o niskava.exe ./cmd/niskava
.\niskava.exe
```

### 3. Launch Local Web Workspace (React/Vite)
```bash
.\niskava.exe serve --port 8080 --open
```

---

## 🏛️ System Architecture: Tripartite Hybrid Stack

```text
┌─────────────────────────────────────────────────────────────┐
│                          GO CORE                            │
│  - Primary Gateway & Single Binary Executable CLI (`niskava`)│
│  - Interactive TUI Terminal (charmbracelet/bubbletea)       │
│  - High-Performance REST & SSE Server (Real-time Streaming) │
│  - Pure-Go SQLite Persistence (modernc.org/sqlite)         │
│  - Static Web UI Bundler (//go:embed)                      │
└─────────────┬───────────────────────────────────────────────┘
              │ Local IPC (JSON Lines / STDIN-STDOUT)
              ▼
┌─────────────────────────────────────────────────────────────┐
│                    PYTHON AGENT ENGINE                      │
│  [Layer 4] ReAct Cognitive Orchestrator & Memory Engine     │
│  [Layer 3] Modular Skills Registry (Domain SOP Modules)     │
│  [Layer 2] Deterministic Compute Gate (NumPy Anomaly Math)  │
│  [Layer 1] Sectors MCP Server & Dual-Engine OSINT Harvester │
└─────────────┬───────────────────────────────────────────────┘
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

## 📜 The 6 Architectural Invariant Laws

Niskava's development is strictly governed by 6 non-negotiable architectural laws (codified in `AGENTS.md`):

1. **Law 1: Deterministic Before Generative** — LLMs are prohibited from calculating statistics, Z-scores, or moving averages. All quantitative indicators are computed deterministically via NumPy/Pandas before prompting an LLM.
2. **Law 2: Strict Financial Non-Advisory Boundary (Sectors Hackathon Rule 12)** — Niskava is an investigative intelligence platform, **NOT an investment advisor**. The system MUST NEVER output direct BUY/SELL recommendations, price targets, or financial advice. All findings fit into a 3-tier verification taxonomy (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
3. **Law 3: Prohibition of Automated Trade Execution (Sectors Hackathon Rule 06)** — Zero brokerage execution APIs or order routing logic. Strictly read-only market intelligence.
4. **Law 4: Local-First Data Sovereignty** — All user sessions, evidence graphs, and caches reside locally at `~/.niskava/niskava.db` using SQLite Write-Ahead Logging (`modernc.org/sqlite`).
5. **Law 5: Credit Budget Discipline & Local Caching** — Strictly protects the 1,000 Sectors API credit grant. Historical daily candlestick data ($T < \text{today}$) is permanently cached locally (`expires_at = NULL`).
6. **Law 6: Local Conversational Graph Memory Engine** — Maintains associative context across sessions using local SQLite graph storage (`memory_nodes` & `memory_edges`) loaded into in-memory Python `NetworkX.DiGraph`.

---

## 📁 Repository Structure

```text
Niskava-Agents/
├── AGENTS.md                  # Single Source of Truth (SSoT) & Architectural Invariants
├── Makefile                   # Build & test automation scripts
├── README.md                  # Main repository overview & quick start
├── public/                    # Public documentation assets & static files
│   └── docs/                  # Detailed User Guides & Technical Manuals
│       ├── README.md          # Documentation sitemap & index hub
│       ├── installation.md    # Detailed installation & setup guide
│       ├── user-guide.md      # TUI REPL, CLI commands, & Web Canvas manual
│       ├── features-and-architecture.md # 7-stage pipeline & NumPy math specs
│       └── troubleshooting-and-faq.md # Error resolution & FAQ
├── cmd/                       # Application binary entry points
│   └── niskava/
│       └── main.go            # Primary CLI binary entry point
├── clients/                   # User Surfaces (Presentation & Interfaces)
│   ├── cli/                   # Terminal Client (Cobra CLI subcommands)
│   │   └── tui/               # Interactive Bubbletea Terminal UI components
│   └── web/                   # React SPA Web Workspace (Vite + Tailwind + shadcn/ui)
├── backend/                   # Core Backend & Cognitive Computation Subsystems
│   ├── core/                  # Go Core Daemon & Persistence Layer (modernc.org/sqlite)
│   └── engine/                # Python Agent Engine (AI, Quant Math, OSINT)
└── docs/                      # Technical Architecture Decision Records (ADR 01 - 11)
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
