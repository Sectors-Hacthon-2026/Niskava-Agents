<p align="center">
  <a href="https://github.com/Sectors-Hacthon-2026/Niskava-Agents">
    <img src="docs/assets/branding/niskava_logo_white_256.png" alt="Niskava Agent Logo" width="128" style="border-radius: 20px; box-shadow: 0 8px 24px rgba(0,0,0,0.12);">
  </a>
</p>

<h1 align="center">Niskava Agent</h1>

<p align="center">
  <strong>Autonomous Financial Market Intelligence & Equity Research Agent for the Indonesia Stock Exchange (IDX)</strong>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License: Apache 2.0"></a>
  <a href="https://www.npmjs.com/package/@zyrexnns/niskava-agent"><img src="https://img.shields.io/npm/v/@zyrexnns/niskava-agent.svg?style=flat&color=CB3837" alt="NPM Version"></a>
  <a href="https://sectors.app/"><img src="https://img.shields.io/badge/Data%20Source-Sectors%20Financial%20API%20v2-0969da.svg" alt="Powered by Sectors API"></a>
  <a href="#"><img src="https://img.shields.io/badge/Market-IDX%20(Indonesia%20Stock%20Exchange)-1a7f37.svg" alt="Target Market"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8.svg" alt="Go Version"></a>
  <a href="https://www.python.org/"><img src="https://img.shields.io/badge/Python-3.11+-3776AB.svg" alt="Python Version"></a>
  <a href="#"><img src="https://img.shields.io/badge/Storage-Local--First%20SQLite%20WAL-lightgrey.svg" alt="Storage"></a>
  <a href="public/docs/README.md"><img src="https://img.shields.io/badge/Docs-public%2Fdocs-purple.svg" alt="Documentation"></a>
</p>

<p align="center">
  <a href="#-what-is-niskava-agent"><strong>What is Niskava?</strong></a> •
  <a href="#-quickstart-up-and-running-in-2-minutes"><strong>Quickstart</strong></a> •
  <a href="#-the-central-gateway-choose-your-experience"><strong>Gateway & Interfaces</strong></a> •
  <a href="#-core-highlights"><strong>Core Highlights</strong></a> •
  <a href="#-documentation-hub"><strong>Docs Hub</strong></a> •
  <a href="#-financial-non-advisory-disclaimer"><strong>Disclaimer</strong></a>
</p>

---

## ⚡ What is Niskava Agent?

**Niskava Agent** is an autonomous market intelligence agent designed specifically for the **Indonesia Stock Exchange (IDX / Bursa Efek Indonesia)**.

Instead of guessing or hallucinating numbers like conventional chat wrappers, Niskava pairs hard quantitative facts from the official **Sectors Financial API v2** with verifiable corporate filings (IDXnet disclosures) and curated news.

> *"Don't just answer questions. Investigate them."*

* **No Math Hallucinations:** Statistical indicators (moving averages, Volume Z-scores, abnormal returns, foreign flow) are computed deterministically with NumPy before any AI reasoning occurs.
* **Temporal Evidence Audits:** News and regulatory announcements are pinned to the exact anomaly date window ($T \pm 2\text{ days}$) to distinguish true catalysts from retrospective market gossip.
* **100% Local & Sovereign:** All sessions, chats, and research history are stored locally on your machine in SQLite WAL (`~/.niskava/niskava.db`). Zero cloud telemetry.
* **Strict Non-Advisory Guardrail:** Niskava is an investigative research assistant, not a financial advisor. It produces objective evidence trails instead of speculative buy/sell calls.

---

## 🚀 Quickstart: Up and Running in 2 Minutes

You can run Niskava immediately with **zero manual setup** using `npx` (requires Node.js 18+):

### Step 1: Run the Interactive Setup Wizard
Configure your Sectors API key and preferred AI model (Gemini, OpenAI, or local Ollama) in an easy step-by-step terminal prompt:

```bash
npx @zyrexnns/niskava-agent setup
```

### Step 2: Launch the Central Gateway
Launch the main Niskava gateway:

```bash
npx @zyrexnns/niskava-agent
```

> **Prefer a permanent global install?**  
> Simply run `npm install -g @zyrexnns/niskava-agent`, and you can use the `niskava` command directly anywhere:
> ```bash
> niskava setup    # One-time setup
> niskava          # Launch Central Gateway
> ```

---

## 🧭 The Central Gateway: Choose Your Experience

When you run `niskava`, you are greeted by the **Central Gateway**—your unified control center for all research modes:

<p align="center">
  <img src="docs/assets/niskava-cli-gateway.png" alt="Niskava Central CLI Gateway" width="90%" style="border-radius: 12px; box-shadow: 0 8px 32px rgba(0,0,0,0.25);">
</p>

From this single menu, press a single key to launch into whatever interface fits your workflow:

### 🌐 1. Web Workspace Canvas (`[W]`)
Press **`W`** (or run `niskava serve --open`) to launch the rich visual web dashboard in your browser (`http://localhost:20128`):

* **Interactive TradingView Candlesticks:** High-resolution charts with automated anomaly markers and volume indicators.
* **Live SSE Streaming Reasoning:** Watch the AI agent formulate hypotheses, query disclosures, and cross-reference data in real time.
* **Visual Evidence Matrix:** Color-coded verification taxonomy (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
* **Bloomberg-Inspired Design:** Clean dark terminal and warm matte light modes designed for deep reading.

<p align="center">
  <img src="docs/assets/niskava-web-dashboard.png" alt="Niskava Web Workspace Canvas" width="100%" style="border-radius: 12px; box-shadow: 0 8px 32px rgba(0,0,0,0.25);">
</p>

### 💻 2. Terminal UI & Interactive REPL (`[T]`)
Press **`T`** (or run `niskava terminal`) for a keyboard-driven conversational research terminal:

* **Conversational IDX Intelligence:** Ask natural language questions like *"Mengapa saham ANTM tiba-tiba melonjak minggu lalu?"* or *"Cek akumulasi foreign flow BBCA 30 hari terakhir"*.
* **Interactive Slash Commands:** `/help`, `/chats`, `/model`, `/doctor`, `/config`, and `/exit`.
* **Low-Latency & Distraction-Free:** Streaming Glamour markdown rendering with smooth progress spinners.

### 📜 3. Sessions & System Health (`[S]`, `[C]`)
* **`[S]` Research Sessions:** Instantly browse, review, and resume past research sessions and audit trails saved in your local SQLite database.
* **`[C]` Health Check & Doctor:** One-key diagnostic check of your API keys, local database connection, and runtime environment.

---

## 🛡️ Core Highlights

```
┌────────────────────────────────────────────────────────────────────────┐
│                        HOW NISKAVA PROTECTS YOUR WORKFLOW              │
├────────────────────────────────────────────────────────────────────────┤
│  Deterministic Math Firewall  │ Zero numerical hallucination. Volume   │
│                               │ Z-scores, abnormal returns, and flow   │
│                               │ metrics computed before LLM inference. │
├───────────────────────────────┼────────────────────────────────────────┤
│  Temporal Causality Audits    │ Verifies whether news preceded or      │
│                               │ followed volume surges (T ± 2 days).   │
├───────────────────────────────┼────────────────────────────────────────┤
│  Institutional Bandarmology   │ Tracks foreign flow accumulation (Fz)  │
│                               │ and broker distribution patterns.      │
├───────────────────────────────┼────────────────────────────────────────┤
│  100% Local-First Privacy     │ All research history stays on your     │
│                               │ machine in ~/.niskava/niskava.db.      │
├───────────────────────────────┼────────────────────────────────────────┤
│  Financial Non-Advisory Guard │ Pure empirical audit trails; zero      │
│                               │ speculative buy/sell price calls.      │
└───────────────────────────────┴────────────────────────────────────────┘
```

---

## 📦 Other Ways to Install

Besides `npx` / `npm`, Niskava is available as standalone binaries and from source:

<details>
<summary><strong>Option A: Standalone Precompiled Binaries (No Node/Go required)</strong></summary>

Download ready-to-run executables from [GitHub Releases](https://github.com/Sectors-Hacthon-2026/Niskava-Agents/releases):

* **Linux:** `niskava-linux-amd64` / `niskava-linux-arm64`
* **macOS:** `niskava-darwin-arm64` (Apple Silicon) / `niskava-darwin-amd64` (Intel)
* **Windows:** `niskava-windows-amd64.exe`

```bash
# Example for Linux/macOS:
chmod +x niskava-linux-amd64
./niskava-linux-amd64 setup
./niskava-linux-amd64
```
</details>

<details>
<summary><strong>Option B: Build from Source (Developer Mode)</strong></summary>

Requires **Go 1.22+** and **Python 3.11+**:

```bash
# 1. Clone repository
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents

# 2. Run automated installer
chmod +x install.sh && ./install.sh   # Linux/macOS
# or: .\install.ps1                   # Windows PowerShell

# 3. Launch
./bin/niskava setup
./bin/niskava
```
</details>

---

## 📚 Documentation Hub

Looking for deep architectural proofs, math formulas, or advanced integration guides? All in-depth technical documentation is organized in the [`public/docs/`](public/docs/README.md) hub:

| Guide | Description |
|---|---|
| 📖 **[Comprehensive User Guide](public/docs/user-guide.md)** | Full guide for REPL slash commands, Web Workspace, MCP server (Claude Desktop/Cursor), and Telegram bot. |
| 📐 **[Features & Architecture](public/docs/features-and-architecture.md)** | Mathematical proofs ($V_z, R_t, D_t, F_z$), the 6 Architectural Laws, and the 7-stage investigation pipeline. |
| ⚙️ **[Installation & Config Guide](public/docs/installation.md)** | Advanced platform installation, environment variables, offline testing mode, and Docker setups. |
| 🔧 **[Troubleshooting & FAQ](public/docs/troubleshooting-and-faq.md)** | Resolving SQLite WAL locks, API key validation, and common questions. |
| 💡 **[Project Concept & Thesis](public/docs/project-concept.md)** | Market inefficiency thesis, the Anti-Wrapper manifesto, and regulatory compliance. |

---

## ⚖️ Financial Non-Advisory Disclaimer

> **IMPORTANT REGULATORY NOTICE:**  
> Niskava Agent is an automated financial market intelligence and empirical research platform. All anomaly detections, quantitative metrics, and evidence correlations are generated from historical market facts, accredited news media, and official exchange disclosures.
>
> Niskava Agent **DOES NOT** provide financial advice, investment recommendations, price targets, or solicitations to buy or sell any security. Niskava operates under a strict non-advisory policy in compliance with Capital Market regulations (POJK / IDX). Users are solely responsible for their own investment evaluations and risk assessments.

---

## 📄 License & Acknowledgments

* **License:** [Apache License 2.0](LICENSE) — free and open for research and development.
* **Market Data Source:** Powered by official [Sectors Financial API v2](https://sectors.app/).
* **Participating Project:** Developed for [Sectors Hackathon Indonesia 2026](https://hackathon.sectors.app/) (Track 1: AI Agents & Assistants).
