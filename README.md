# Niskava Agent

**Autonomous Financial Market Intelligence & Quantitative Research Orchestration Platform for the Indonesia Stock Exchange (IDX)**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![NPM Version](https://img.shields.io/npm/v/@zyrexnns/niskava-agent.svg?style=flat&color=CB3837)](https://www.npmjs.com/package/@zyrexnns/niskava-agent)
[![Powered by Sectors API](https://img.shields.io/badge/Data%20Source-Sectors%20Financial%20API%20v2-0969da.svg)](https://sectors.app/)
[![Target Market](https://img.shields.io/badge/Market-IDX%20%28Indonesia%20Stock%20Exchange%29-1a7f37.svg)](#)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev/)
[![Python Version](https://img.shields.io/badge/Python-3.11+-3776AB.svg)](https://www.python.org/)
[![Storage](https://img.shields.io/badge/Storage-Local--First%20SQLite%20WAL-lightgrey.svg)](#)
[![Documentation](https://img.shields.io/badge/Docs-public%2Fdocs-purple.svg)](public/docs/README.md)

---

## Overview

**Niskava Agent** is an autonomous market intelligence and equity research platform designed specifically for the **Indonesia Stock Exchange (IDX / Bursa Efek Indonesia)**.

It bridges the critical gap between structured quantitative market facts (powered by official **Sectors Financial API v2** endpoints) and qualitative external intelligence (official IDXnet corporate disclosures, regulatory filings, and vetted financial news).

When unusual market activity occurs—such as a volume surge, price breakout, or aggressive foreign accumulation—Niskava does not output speculative chatbot commentary. Instead, it computes deterministic statistics via NumPy/Pandas, formulates investigative hypotheses, harvests external disclosures within a strict chronological window ($T_{\text{anomaly}} \pm 2\text{ days}$), and compiles an empirical evidence audit trail classified into `SUPPORTED`, `UNCERTAIN`, or `CONTRADICTED` findings.

> **Core Philosophy:** *"Don't just answer questions. Investigate them."*

---

## Key Capabilities

* **Deterministic NumPy Firewall (Law 1):** Zero mathematical hallucinations. Moving averages, Z-scores, abnormal returns, and sector divergence are computed deterministically before any LLM is invoked.
* **6 Modular Domain Skills:** Pre-packaged institutional SOP modules (Insider Bandarmology, Market Anomaly Recon, Financial Health Stress Testing, Commodity Divergence, Peer Valuation Benchmarks, and Event Causality Audits).
* **Multi-Surface Architecture:**
  - **Interactive Research REPL & Terminal:** Fast, keyboard-driven terminal with interactive prompt history, Braille spinners, and streaming Glamour markdown rendering.
  - **Web Workspace Canvas (`:20128`):** Bloomberg Terminal-inspired dark/light workspace with interactive TradingView candlestick charts, anomaly badges, and live SSE streaming.
  - **Autonomous Headless Pipeline:** Single CLI command (`niskava investigate <TICKER> --days 30`) producing structured JSON-Lines and Markdown audit dossiers.
  - **Telegram Bot Runner:** Personal research assistant on Telegram with whitelist access security.
* **Unified Model Context Protocol (MCP) Server:** Native MCP integration exposing tools and resources to external agents (Claude Desktop, Cursor, Antigravity).
* **Local-First Data Sovereignty:** Zero cloud telemetry. All sessions, memory graphs, and caches live locally in `~/.niskava/niskava.db` via pure-Go SQLite with Write-Ahead Logging (WAL).

---

## Quantitative Mathematical Engine

Niskava strictly enforces **Law 1 (Deterministic Before Generative)**. The quantitative engine computes mathematical indicators locally using NumPy before constructing any model prompts.

### 1. Volume Anomaly Z-Score ($V_z$)
Measures the statistical significance of trading volume on day $t$ relative to its 20-day historical baseline:

$$\mu_{20} = \frac{1}{20} \sum_{i=1}^{20} V_{t-i}$$

$$\sigma_{20} = \sqrt{\frac{1}{20} \sum_{i=1}^{20} (V_{t-i} - \mu_{20})^2}$$

$$V_z = \frac{V_t - \mu_{20}}{\sigma_{20}}$$

* **Trigger Condition:** $V_z \ge 2.5\sigma$ triggers an abnormal volume surge (probability of random occurrence in a normal distribution is $< 0.6\%$).

---

### 2. Abnormal Price Return ($R_t$)
Measures the single-session price breakout magnitude relative to the previous trading day's close:

$$R_t = \left( \frac{P_{\text{close}, t} - P_{\text{close}, t-1}}{P_{\text{close}, t-1}} \right) \times 100\%$$

* **Trigger Condition:** $|R_t| \ge 5.0\%$ flags a price breakout anomaly.

---

### 3. Sector Divergence Index ($D_t$)
Differentiates whether a stock's movement is driven by broad industry momentum or idiosyncratic company catalysts:

$$D_t = R_{\text{stock}, t} - R_{\text{sector}, t}$$

* **Trigger Condition:** $|D_t| \ge 4.0\%$ indicates an **Idiosyncratic Catalyst** specific to the issuer rather than broader market beta.

---

### 4. Foreign Flow Accumulation Z-Score ($F_z$)
Quantifies abnormal institutional or foreign capital movement (Net Foreign Buy/Sell in IDR):

$$F_z = \frac{F_t - \mu_{F, 20}}{\sigma_{F, 20}}$$

* **Trigger Condition:** $|F_z| \ge 2.5\sigma$ indicates significant institutional accumulation or distribution.

---

### Anomaly Decision & Routing Matrix

| Condition $V_z$ | Condition $|R_t|$ | Condition $|D_t|$ | Classification | Engine Routing Action |
|:---:|:---:|:---:|:---|:---|
| $\ge 2.5\sigma$ | $\ge 5.0\%$ | $\ge 4.0\%$ | `IDIOSYNCRATIC_CATALYST` | Triggers high-priority `event_causality_audit` and disclosure harvest. |
| $\ge 2.5\sigma$ | $< 5.0\%$ | Any | `VOLUME_ACCUMULATION` | Activates `insider_bandarmology_forensic` for foreign vs broker flow analysis. |
| $< 2.5\sigma$ | $\ge 5.0\%$ | $< 4.0\%$ | `SECTOR_BETA_RALLY` | Classifies as macro/sector momentum; suppresses false-alarm company alerts. |
| $< 2.5\sigma$ | $< 5.0\%$ | Any | `NORMAL_VARIANCE` | Routes to fundamental health and peer valuation screening. |

---

## Evidence Verification Taxonomy & Confidence Scoring

All qualitative evidence collected from official corporate filings (IDXnet) and news outlets is evaluated through a strict temporal and structural verification taxonomy:

```
┌────────────────────────────────────────────────────────────────────────┐
│                        3-TIER VERIFICATION TAXONOMY                    │
├────────────────────────────────────────────────────────────────────────┤
│  SUPPORTED   │ Directly validated by Sectors API records or official   │
│              │ IDXnet regulatory filings (causal link proven).         │
├──────────────┼─────────────────────────────────────────────────────────┤
│  UNCERTAIN    │ Plausible correlation observed, but direct causality    │
│              │ remains unconfirmed (e.g. market rumors, social buzz).   │
├──────────────┼─────────────────────────────────────────────────────────┤
│ CONTRADICTED │ Market claims refuted by formal corporate disclosures    │
│              │ or official financial reports.                          │
└──────────────┴─────────────────────────────────────────────────────────┘
```

### Standardized Confidence Rubric
Never assign continuous random floats. Niskava strictly adheres to discrete confidence tiers:
* `1.00`: **EXTRACTED** — Directly backed by official exchange records or regulatory filings.
* `0.95`: **Direct Structural Evidence** — Explicit timestamp correlation with official corporate press releases.
* `0.85`: **Strong Inference** — High temporal correlation with major verified financial media reporting.
* `0.75`: **Reasonable Inference** — Plausible catalyst from industry trends matching sector divergence.
* `0.65`: **Weak Inference** — Unverified market commentary, social media sentiment, or unconfirmed rumors.
* `0.55`: **Speculative** — Distant co-occurrence without temporal causality confirmation.

---

## 6 Modular Domain Skills

Niskava organizes analytical expertise into domain-specific Standard Operating Procedures (SOPs):

1. **`market_anomaly_recon`**: Deterministic statistical screening over 30–90 trading days to detect volume spikes, price breakouts, and sector divergence.
2. **`event_causality_audit`**: Anchors external news and IDXnet disclosures within $T_{\text{anomaly}} \pm 2\text{ days}$ to prove chronological precedence (`LIKELY_CATALYST` vs `PRECEDED_ANNOUNCEMENT`).
3. **`insider_bandarmology_forensic`**: Tracks top broker accumulation/distribution, foreign net inflow streaks, and major shareholder insider transactions.
4. **`financial_health_stress`**: Evaluates solvency, liquidity, Altman Z-score distress risk, debt-to-equity (DER), and debt coverage.
5. **`mining_commodity_divergence`**: Correlates commodity export benchmarks (Nickel, Coal, Gold, CPO) with mining and agribusiness equity performance.
6. **`peer_valuation_benchmark`**: Compares Price-to-Earnings (P/E), Price-to-Book (P/B), and EV/EBITDA against sub-sector peer medians.

---

## Quickstart & Installation

Niskava provides three installation paths based on your workflow:

### Option 1: Instant Zero-Clone via NPX / NPM (Recommended)

Requires Node.js (version 18 or higher):

```bash
# 1. Run interactive configuration wizard (sets API keys & preferences)
npx @zyrexnns/niskava-agent setup

# 2. Verify environment readiness and system health
npx @zyrexnns/niskava-agent doctor

# 3. Launch the interactive REPL research terminal
npx @zyrexnns/niskava-agent

# 4. Or launch the local Web Workspace (:20128)
npx @zyrexnns/niskava-agent serve
```

#### Global Installation:
```bash
npm install -g @zyrexnns/niskava-agent

# Run anywhere:
niskava setup
niskava doctor
niskava investigate ANTM --days 30
niskava serve
```

---

### Option 2: Standalone Precompiled Binaries

Download native prebuilt executables directly from [GitHub Releases](https://github.com/Sectors-Hacthon-2026/Niskava-Agents/releases):

* **Linux**: `niskava-linux-amd64` / `niskava-linux-arm64`
* **macOS**: `niskava-darwin-arm64` (Apple Silicon) / `niskava-darwin-amd64` (Intel)
* **Windows**: `niskava-windows-amd64.exe`

```bash
# Example for Linux/macOS:
chmod +x niskava-linux-amd64
./niskava-linux-amd64 setup
./niskava-linux-amd64
```

---

### Option 3: Build from Source Code (Developer Mode)

Requires **Go 1.22+** and **Python 3.11+**:

```bash
# 1. Clone repository
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents

# 2. Automated setup script
# On Linux/macOS:
chmod +x install.sh && ./install.sh

# On Windows (PowerShell):
.\install.ps1

# 3. Run setup wizard
./bin/niskava setup

# 4. Verify system health
./bin/niskava doctor
```

---

## Interaction Surfaces

### 1. Interactive Terminal REPL & HUD
```bash
niskava terminal    # or: niskava repl / niskava chat
```
* **Autocomplete Slash Commands:** `/help`, `/chats`, `/model`, `/setup`, `/config`, `/compact`, `/export`, `/graph`, `/doctor`, `/exit`.
* **Prompt History:** Navigate previous queries using `Up` / `Down` arrows.
* **Braille Progress Spinner:** Live indicator of ReAct phase transitions (NumPy math, news harvest, synthesis).

### 2. Autonomous Headless Investigation
```bash
# Run 30-day investigation on any IDX ticker:
niskava investigate BBCA --days 30

# Export structured PDF report:
niskava investigate ANTM --days 30 --pdf
```

### 3. Web Workspace Canvas
```bash
niskava serve --port 20128 --open
```
* Institutional terminal aesthetics (Bloomberg Dark and Warm Matte Light themes).
* Interactive TradingView candlestick charts with anomaly markers.
* Real-time Server-Sent Events (SSE) reasoning stream.
* Interactive Evidence Matrix and temporal causality graph.

### 4. Model Context Protocol (MCP) Server
```bash
niskava mcp
```
Connects Niskava tools and resources to Claude Desktop, Cursor, and Antigravity. Example Claude Desktop config (`claude_desktop_config.json`):
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

### 5. Telegram Bot
```bash
niskava telegram
```
Runs a private Telegram bot research assistant with whitelist authorization.

---

## Configuration Reference

Niskava reads configuration from environment variables or `~/.niskava/config.yaml` / `~/.niskava/.env`:

| Environment Variable | YAML Key | Default | Description |
|---|---|---|---|
| `SECTORS_API_KEY` | `auth.sectors_api_key` | `""` | Sectors Financial API v2 key (**required for live IDX data**). |
| `GEMINI_API_KEY` | `auth.gemini_api_key` | `""` | Google Gemini API key. |
| `OPENAI_API_KEY` | `auth.openai_api_key` | `""` | OpenAI / OpenRouter API key. |
| `AI_PROVIDER` | `ai.provider` | `"gemini"` | LLM provider (`gemini`, `openai`, `openrouter`, `ollama`, `offline`). |
| `OPENAI_MODEL` | `ai.model` | `"gemini-2.0-flash"` | Target model name. |
| `NISKAVA_DB_PATH` | `storage.db_path` | `"~/.niskava/niskava.db"` | Local SQLite database file path. |
| `NISKAVA_PORT` | `server.port` | `20128` | Local Web Workspace daemon port. |
| `NISKAVA_LANG` | `preferences.language` | `"id"` | Interface language (`id` or `en`). |
| `MOCK_SECTORS` | — | `"0"` | Set to `1` for offline testing fixtures (CI/CD only). |
| `NISKAVA_OFFLINE` | `preferences.offline_mode` | `"0"` | Set to `1` to run completely offline without LLM calls. |

> **Live Mode Auto-Toggle:** When `SECTORS_API_KEY` is present and valid, Niskava automatically activates **Live Mode** (`MOCK_SECTORS=0`), ensuring you always analyze real exchange data.

---

## Testing & Quality Verification

Run the full automated test suites across Go and Python:

```bash
# Run all Python unit tests (320+ tests covering math, ReAct loop, skills, MCP)
pytest backend/engine/tests/ -v

# Run all Go Core unit tests (IPC, SQLite WAL, config, TUI)
go test ./clients/cli/... ./backend/core/... -v
```

---

## Documentation Hub

Detailed documentation is available in the [`public/docs/`](public/docs/README.md) directory:

* **[Documentation Index](public/docs/README.md)**: Navigation sitemap.
* **[Installation Guide](public/docs/installation.md)**: In-depth setup for Linux, macOS, Windows, and Docker.
* **[User Guide & Manual](public/docs/user-guide.md)**: Complete guide to REPL commands, Web Workspace, and Telegram bot.
* **[Features & System Architecture](public/docs/features-and-architecture.md)**: Mathematical formulas, 6 Laws, and 7-stage pipeline.
* **[Troubleshooting & FAQ](public/docs/troubleshooting-and-faq.md)**: Common error resolutions, Python environments, and FAQ.
* **[Project Concept & Thesis](public/docs/project-concept.md)**: Market inefficiency thesis and the Anti-Wrapper manifesto.

---

## Financial Non-Advisory Disclaimer

> **IMPORTANT NOTICE:**  
> Niskava Agent is an automated financial market intelligence and empirical research platform. All findings, anomaly detections, and evidence correlations are generated from historical market facts, official regulatory disclosures, and accredited news media.
>
> Niskava Agent **DOES NOT** provide financial advice, investment recommendations, price targets, or solicitations to buy or sell any security. Niskava operates under a strict non-advisory policy in compliance with Capital Market regulations (POJK / IDX). Users are solely responsible for their own investment evaluations and risk assessments.

---

## License & Acknowledgments

* **License:** [MIT License](LICENSE)
* **Market Data Source:** Powered by official [Sectors Financial API v2](https://sectors.app/).
* **Participating Project:** Developed for [Sectors Hackathon Indonesia 2026](https://hackathon.sectors.app/) (Track 1: AI Agents & Assistants).
