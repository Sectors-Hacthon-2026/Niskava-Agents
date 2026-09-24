# Niskava Agent

**Autonomous Financial Market Intelligence Orchestration Platform for the Indonesia Stock Exchange (IDX)**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Track](https://img.shields.io/badge/Sectors%20Hackathon%202026-Track%201%3A%20AI%20Agents%20%26%20Assistants-0969da.svg)](https://hackathon.sectors.app/)
[![Target Market](https://img.shields.io/badge/Market-IDX%20%28Indonesia%20Stock%20Exchange%29-1a7f37.svg)](#)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev/)
[![Python Version](https://img.shields.io/badge/Python-3.11+-3776AB.svg)](https://www.python.org/)
[![Storage](https://img.shields.io/badge/Storage-Local--First%20SQLite%20WAL-lightgrey.svg)](#)
[![Documentation](https://img.shields.io/badge/Docs-public%2Fdocs-purple.svg)](public/docs/README.md)

---

## Overview

**Niskava Agent** is an autonomous market intelligence and equity research orchestration platform engineered specifically for the Indonesia Stock Exchange (IDX / Bursa Efek Indonesia).

It bridges the critical operational gap between structured quantitative exchange facts (powered by the **Sectors Financial API v2**) and unstructured qualitative market intelligence (official IDXnet regulatory disclosures, corporate announcements, and syndicated business news).

When unusual market activity occurs—such as an unexplained trading volume surge, abrupt price breakout, or aggressive foreign capital accumulation—Niskava does not rely on passive chart visualization or speculative chatbot commentary. Instead, it formulates investigative hypotheses, executes deterministic mathematical anomaly detection, harvests contemporaneous external disclosures within a strict temporal window ($T_{\text{anomaly}} \pm 2\text{ days}$), and compiles an empirical evidence audit trail classified into `SUPPORTED`, `UNCERTAIN`, or `CONTRADICTED` findings.

> **Core Motto:** *"Don't just answer questions. Investigate them."*

---

## The Anti-Wrapper Manifesto: Why Generic AI Fails in Capital Markets

Most commercial "financial AI" tools are thin wrappers around general-purpose Large Language Models (LLMs). Deploying thin wrappers in capital markets introduces severe operational risks:

| Thin AI Wrapper Anti-Pattern | Niskava Architectural Defense |
|---|---|
| **Raw JSON Prompt Stuffing:** Dumping hundreds of raw candlestick rows exhausts token limits and degrades reasoning quality. | **Deterministic Compute Gate:** Raw time series data is processed locally by deterministic algorithms; only verified anomaly indicators enter model context. |
| **Mental Math Hallucinations:** Asking an LLM to calculate moving averages or Z-scores produces fabricated numbers. | **Law 1 (Deterministic Before Generative):** LLMs are strictly forbidden from performing mathematical calculations. All statistics are computed via NumPy. |
| **Atemporal Search (Causality Inversion):** Standard semantic search retrieves articles without date constraints, attributing price spikes to news published days *after* the event. | **Temporal-Aware News Anchoring:** Web and disclosure harvesting is locked strictly around the anomaly event date ($T_{\text{anomaly}} \pm 2\text{ days}$) to verify chronological precedence. |
| **Monolithic Prompts:** Single monolithic prompts fail to isolate analytical methodologies or support structured backtracking. | **4-Layer Cognitive Hierarchy:** Clean separation between MCP primitives, compute gates, modular domain skills (SOPs), and the cognitive ReAct loop. |
| **Unregulated Speculative Advice:** Thin wrappers often generate illegal buy/sell recommendations or price targets. | **Law 2 (Strict Non-Advisory Boundary):** Outputs an objective evidence audit trail. Zero buy/sell recommendations or price targets. |

---

## System Architecture

Niskava Agent is implemented as a **Tripartite Hybrid Stack** combining Go Core, a Python Agent Engine, and a React SPA Web Workspace:

```
┌─────────────────────────────────────────────────────────────────┐
│                            GO CORE                              │
│  - Gateway CLI & Daemon Process (`cmd/niskava`, `clients/cli`)  │
│  - Interactive Terminal HUD & REPL (charmbracelet/bubbletea)    │
│  - High-Throughput REST API & SSE Streaming Server (:20128)     │
│  - Static Web UI Bundler (//go:embed)                           │
│  - Zero-CGO SQLite WAL Persistence (modernc.org/sqlite)         │
│  - Subprocess IPC Broker (JSON-Lines over STDIN/STDOUT)         │
└────────────────────────────────┬────────────────────────────────┘
                                 │ Inter-Process Communication
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                      PYTHON AGENT ENGINE                        │
│                                                                 │
│  [Layer 4: ReAct Cognitive Orchestrator & Memory Engine]        │
│  - Autonomous ReAct Agent Loop (Reasoning + Tool Action)        │
│  - Local Associative Graph Memory (NetworkX + SQLite)           │
│  - Temporal Precedence & Causality Inference Engine             │
│                                                                 │
│  [Layer 3: Modular Skills Registry (Domain SOP Modules)]        │
│  - Market Anomaly Reconnaissance (`market_anomaly_recon`)       │
│  - Event Causality Audit (`event_causality_audit`)              │
│  - Insider & Foreign Flow Forensics (`insider_bandarmology`)    │
│  - Financial Health Stress Testing (`financial_health_stress`)  │
│  - Commodity Divergence (`mining_commodity_divergence`)         │
│  - Peer Valuation Benchmark (`peer_valuation_benchmark`)       │
│                                                                 │
│  [Layer 2: Deterministic Compute Gate (NumPy Firewall)]         │
│  - Volume Z-Scores (Vz), Abnormal Returns (Rt), Sector Beta     │
│  - Net Foreign Flow Z-Scores (Fz), Altman Z-Score Ratios        │
│                                                                 │
│  [Layer 1: Sectors MCP & News Engine Primitives]                │
│  - Sectors Financial API v2 MCP Server Adapter                  │
│  - Curated Sectors News & Corporate Filings Engine              │
│  - Content Sanitization via Trafilatura (<evidence_context>)    │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                   REACT SPA WEB WORKSPACE                       │
│  - Market Intelligence / Bloomberg Terminal Interface (Vite + Tailwind) │
│  - TradingView / Recharts Candlestick Anomaly Overlays          │
│  - Real-Time Thinking Stream via Server-Sent Events (SSE)       │
│  - Interactive Evidence Matrix & Causality Timeline Graph       │
└─────────────────────────────────────────────────────────────────┘
```

---

## The 6 Architectural Invariant Laws

All development on Niskava Agent is strictly governed by six foundational architectural invariants (codified in `AGENTS.md`):

1. **Law 1: Deterministic Before Generative**  
   LLMs must never compute statistics, volume moving averages, standard deviations, Z-scores, abnormal returns, or sector divergence. All quantitative metrics are computed deterministically via Python NumPy before prompting an LLM.
2. **Law 2: Strict Financial Non-Advisory Boundary (Sectors Hackathon Rule 12)**  
   Niskava is an investigative intelligence platform, **NOT an investment advisor**. The platform never issues BUY/SELL recommendations, price targets, or financial advice. All findings fit a 3-tier verification taxonomy (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
3. **Law 3: Prohibition of Automated Trade Execution (Sectors Hackathon Rule 06)**  
   Zero brokerage trading APIs, order routing logic, or account execution capabilities. Niskava is strictly read-only market intelligence.
4. **Law 4: Local-First Data Sovereignty**  
   Zero centralized cloud database. All user investigation sessions, findings, evidence graphs, chat histories, and caches reside locally in SQLite with Write-Ahead Logging (`~/.niskava/niskava.db`).
5. **Law 5: Credit Budget Discipline & Local Caching**  
   Strictly protects the 1,000 Sectors API credit grant. All historical daily candlestick data ($T < \text{today}$) is permanently cached (`expires_at = NULL`), incurring zero credit cost on repeat queries.
6. **Law 6: Local Conversational Graph Memory Engine**  
   Maintains multi-session associative context via local SQLite graph tables (`memory_nodes` and `memory_edges`) loaded into in-memory `networkx.DiGraph` queried via Ego-Graph traversal ($k \le 2$ hops) with exponential recency decay ($e^{-\lambda \Delta t}$).

---

## Quick Start

### 1. Prerequisites
- **Go**: Version 1.22 or higher
- **Python**: Version 3.11 or higher
- **Git**: Version 2.30 or higher

### 2. Clone and Setup Environment
```bash
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents

# Run the interactive setup wizard (configures API keys & creates Python virtual environment)
go run ./cmd/niskava setup
```

The wizard prompts for your **Sectors Financial API v2 Key** ([sectors.app](https://sectors.app/)) and your AI provider credentials (Google Gemini, Ollama, OpenRouter, or vLLM), then generates `~/.niskava/config.yaml`.

### 3. Build Executable Binary
```bash
# On Linux / macOS:
go build -o niskava ./cmd/niskava
chmod +x niskava

# On Windows:
go build -o niskava.exe ./cmd/niskava
```

---

## Interaction Modes & Usage

Niskava provides multiple interaction surfaces for different workflows:

### 1. Interactive Terminal UI (TUI) HUD Launcher
Running `niskava` without arguments launches the terminal HUD:
```bash
./niskava
```
- Interactive HUD launcher with diagnostics, session resume, and setup wizard.
- Prompt-driven interactive REPL with **Up/Down arrow prompt history** navigation.
- Live animated **Braille progress spinner** (`⠋`) showing real-time ReAct phase transitions.
- Autocomplete slash commands: `/investigate <TICKER>`, `/screen`, `/health <TICKER>`, `/memory`, `/lang`, `/clear`, `/exit`.

### 2. Autonomous Headless Investigation CLI
Execute a full 7-stage investigation directly from the shell:
```bash
./niskava investigate ANTM --days 30
```
Flags:
- `--days <N>`: Trading sessions to analyze (default: 30).
- `--offline`: Runs in offline mode using local mock fixtures without issuing live API requests.
- `--lang <en|id>`: Output language (`en` for English, `id` for Indonesian).
- `--verbose, -v`: Prints detailed debug logs and IPC payload messages.

### 3. Local Web Workspace
Launch the background REST/SSE server and interactive visual canvas:
```bash
./niskava serve --port 20128 --open
```
- Interactive candlestick chart with volume anomaly badges ($V_z \ge 2.5$) and breakout tags ($|R_t| \ge 5\%$).
- Real-time Server-Sent Events (SSE) streaming of agent reasoning and tool execution.
- Interactive Evidence Matrix and chronological causality graph.

### 4. Interactive Knowledge Graph Export
Export the local associative knowledge graph into a standalone HTML file:
```bash
./niskava graph --open
```

### 5. Model Context Protocol (MCP) Server
Expose Niskava tools, resources, and prompts to external agent environments (Claude Desktop, Cursor, Antigravity):
```bash
./niskava mcp
```
Claude Desktop configuration (`claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "niskava": {
      "command": "/usr/local/bin/niskava",
      "args": ["mcp"]
    }
  }
}
```

### 6. Telegram Bot Runner
Deploy Niskava as a personal market intelligence Telegram bot:
```bash
export TELEGRAM_BOT_TOKEN="your_token"
./niskava telegram
```

---

## Quantitative Anomaly Formulas

All quantitative indicators are calculated deterministically by NumPy before LLM activation:

- **Volume Anomaly Z-Score ($V_z$):**
  $$V_z = \frac{V_t - \mu_{20}}{\sigma_{20}}$$
  $V_z \ge 2.5$ triggers a `VOLUME_SPIKE` event (probability $< 0.6\%$).
- **Abnormal Price Return ($R_t$):**
  $$R_t = \frac{P_{\text{close}, t} - P_{\text{close}, t-1}}{P_{\text{close}, t-1}} \times 100\%$$
  $|R_t| \ge 5.0\%$ triggers a `PRICE_BREAKOUT` anomaly.
- **Sector Divergence ($D_t$):**
  $$D_t = R_{\text{stock}, t} - R_{\text{sector}, t}$$
  $|D_t| \ge 4.0\%$ indicates an **Idiosyncratic Catalyst** specific to the issuer.
- **Foreign Flow Inflow Z-Score ($F_z$):**
  $$F_z = \frac{F_t - \mu_{F, 20}}{\sigma_{F, 20}}$$
  $|F_z| \ge 2.5$ flags abnormal foreign capital movement.

---

## Evidence Verification Taxonomy & Confidence Rubric

Every finding is corroborated with a discrete confidence score:

| Score | Verification Level | Source Validation Criteria |
|:---:|:---|:---|
| **`1.00`** | **EXTRACTED** | Directly backed by official Sectors API quantitative records or formal IDXnet regulatory filings. |
| **`0.95`** | **Direct Structural Evidence** | Explicit timestamp correlation with official corporate press releases or regulatory disclosures. |
| **`0.85`** | **Strong Inference** | High temporal correlation with major national business media reporting (Kontan, Bisnis Indonesia, CNBC). |
| **`0.75`** | **Reasonable Inference** | Plausible catalyst from industry-wide trends corroborated by matching sector divergence metrics. |
| **`0.65`** | **Weak Inference** | Unverified market commentary, social media sentiment, or unconfirmed financial forum discussions. |
| **`0.55`** | **Speculative** | Distant co-occurrence without temporal causality or formal corroboration. |

---

## Configuration Reference

Configuration can be provided via `~/.niskava/config.yaml` or environment variables:

| Environment Variable | YAML Setting | Default Value | Description |
|---|---|---|---|
| `SECTORS_API_KEY` | `auth.sectors_api_key` | `""` | Sectors Financial API v2 key. |
| `GEMINI_API_KEY` | `auth.gemini_api_key` | `""` | Google Gemini API key. |
| `OPENAI_API_KEY` | `auth.openai_api_key` | `""` | OpenAI / OpenRouter API key. |
| `NISKAVA_AI_PROVIDER` | `ai.provider` | `"gemini"` | Inference backend (`gemini`, `ollama`, `openrouter`, `vllm`). |
| `NISKAVA_AI_MODEL` | `ai.model` | `"gemini-2.5-flash"` | Target language model name. |
| `NISKAVA_AI_ENDPOINT` | `ai.endpoint` | `""` | Custom API base URL (for Ollama or vLLM). |
| `NISKAVA_DB_PATH` | `storage.db_path` | `"~/.niskava/niskava.db"` | Local SQLite database file path. |
| `NISKAVA_PORT` | `server.port` | `20128` | Daemon REST & SSE server port. |
| `NISKAVA_LANG` | `preferences.language` | `"en"` | Default language (`"en"` or `"id"`). |
| `MOCK_SECTORS` | — | `"0"` | Set to `1` to run offline using static mock fixtures. |

---

## Detailed Public Documentation

Comprehensive guides and architectural specifications are available in the [`public/docs/`](public/docs/README.md) directory:

- **[Documentation Hub](public/docs/README.md)**: Sitemap and document index.
- **[Project Concept & Vision](public/docs/project-concept.md)**: In-depth thesis, IDX market dynamics, and Anti-Wrapper manifesto.
- **[Installation & Setup Guide](public/docs/installation.md)**: Prerequisites, cross-platform installation, virtualenv, and configuration.
- **[User Guide & Interfaces](public/docs/user-guide.md)**: Comprehensive manual for TUI REPL, Web Workspace, CLI flags, MCP, and Telegram.
- **[Features & System Architecture](public/docs/features-and-architecture.md)**: Tripartite stack, 6 Invariant Laws, 7-Stage pipeline, and mathematical formulas.
- **[Troubleshooting & FAQ](public/docs/troubleshooting-and-faq.md)**: Error diagnostics, SQLite WAL recovery, credit conservation, and FAQ.

---

## Repository Structure

```text
Niskava-Agents/
├── AGENTS.md                  # Single Source of Truth (SSoT) & Architectural Invariants
├── Makefile                   # Build and test automation
├── README.md                  # Main repository overview and quick start
├── public/                    # Public documentation assets & static files
│   └── docs/                  # Detailed User Guides & Technical Manuals
│       ├── README.md          # Documentation sitemap & index hub
│       ├── project-concept.md # Project thesis, IDX market context, & Anti-Wrapper manifesto
│       ├── installation.md    # Cross-platform installation & setup guide
│       ├── user-guide.md      # TUI REPL, CLI commands, Web Canvas, & MCP manual
│       ├── features-and-architecture.md # 7-stage pipeline & NumPy math specs
│       └── troubleshooting-and-faq.md   # Error diagnostics & FAQ
├── cmd/                       # Application binary entry points
│   └── niskava/
│       └── main.go            # Primary CLI binary entry point
├── clients/                   # User Surfaces (Presentation & Interfaces)
│   ├── cli/                   # Terminal Client (Cobra CLI subcommands & TUI)
│   │   └── tui/               # Bubbletea TUI, HUD launcher, & REPL components
│   └── web/                   # React SPA Web Workspace (Vite + Tailwind + shadcn/ui)
├── backend/                   # Core Backend & Cognitive Computation Subsystems
│   ├── core/                  # Go Core Daemon, REST/SSE Server, & SQLite WAL Persistence
│   └── engine/                # Python Agent Engine (ReAct Agent, Quant Math, Sectors News, MCP)
└── docs/                      # Internal Architecture Decision Records (ADR 01 - 11)
```

---

## Testing & Quality Verification

Run the automated test suites across Go and Python subsystems:

```bash
# Run Python Engine unit tests (200+ unit tests)
pytest backend/engine -v

# Run Go Core unit tests (IPC, SQLite WAL persistence, TUI)
go test ./...
```

For offline verification without network requests or API credit consumption:
```bash
export MOCK_SECTORS=1
./niskava investigate ANTM --days 30 --offline
```

---

## Contributors

This project is authored and maintained by:

| [<img src="https://github.com/Zyrexnn.png?size=120" width="120px;" alt="Zyrexnn"/><br /><sub><b>Zyrexnn</b></sub>](https://github.com/Zyrexnn)<br /><sub><b>Author & Lead Architect</b></sub><br />[![GitHub](https://img.shields.io/badge/GitHub-Zyrexnn-181717?style=flat&logo=github)](https://github.com/Zyrexnn) | [<img src="https://github.com/Lutfi1i.png?size=120" width="120px;" alt="Lutfi1i"/><br /><sub><b>Lutfi1i</b></sub>](https://github.com/Lutfi1i)<br /><sub><b>Core Maintainer</b></sub><br />[![GitHub](https://img.shields.io/badge/GitHub-Lutfi1i-181717?style=flat&logo=github)](https://github.com/Lutfi1i) | [<img src="https://github.com/Sazhumaa.png?size=120" width="120px;" alt="Sazhumaa"/><br /><sub><b>Sazhumaa</b></sub>](https://github.com/Sazhumaa)<br /><sub><b>Core Maintainer</b></sub><br />[![GitHub](https://img.shields.io/badge/GitHub-Sazhumaa-181717?style=flat&logo=github)](https://github.com/Sazhumaa) |
| :---: | :---: | :---: |

---

## Financial Non-Advisory Disclaimer

> **IMPORTANT DISCLAIMER**  
> Niskava Agent is an automated market intelligence and empirical research platform. All findings, anomaly alerts, and correlated evidence generated by the platform are derived from historical market data, public regulatory disclosures, and news media.
>
> Niskava Agent **DOES NOT** provide financial advice, investment recommendations, price targets, or solicitations to purchase or sell any security. Niskava operates under a strict non-advisory policy in compliance with Capital Market regulations (POJK / IDX) and Sectors Hackathon Rule 12. Users are solely responsible for their independent investment evaluations and risk assessments.

---

## Competition & License

- **Hackathon:** [Sectors Hackathon Indonesia 2026](https://hackathon.sectors.app/)
- **Category:** Track 1: AI Agents & Assistants
- **License:** MIT License — see the [LICENSE](LICENSE) file for details.
