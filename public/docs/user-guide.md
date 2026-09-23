# 🎮 Niskava Agent — User Guide & Feature Manual

This document provides a comprehensive guide on all interaction modes, CLI commands, interactive REPL features, Web Canvas Workspace, and bot integrations available in Niskava Agent.

---

## 🖥️ 1. Dual Interaction Modes

Niskava Agent offers two distinct interaction surfaces tailored for interactive exploration and automated research pipelines:

```text
┌──────────────────────────────────────────────────────────────────┐
│                   NISKAVA INTERACTION SURFACES                   │
├──────────────────────────────────────────────────────────────────┤
│ 1. Terminal UI (TUI) REPL & HUD Launcher                         │
│    -> Run `niskava` without arguments                           │
│    -> Interactive HUD menu, prompt-driven REPL, slash commands  │
│                                                                  │
│ 2. Autonomous Headless Pipeline (Single Command)                 │
│    -> Run `niskava investigate <TICKER> --days 30`               │
│    -> Single-command structured audit trail & direct summary     │
│                                                                  │
│ 3. Web Canvas Workspace (Visual Browser Terminal)                │
│    -> Run `niskava serve --port 8080 --open`                     │
│    -> TradingView/Recharts candles, SSE stream & evidence graph  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 🚀 2. TUI HUD Launcher Menu & Interactive REPL

Running the `niskava` executable without subcommands launches the **HUD Launcher Menu** featuring the ASCII `NISKAVA` branding:

```bash
.\niskava.exe
```

```text
  ███╗   ██╗██╗███████╗██╗  ██╗██████╗  ██╗   ██╗██████╗ 
  ████╗  ██║██║██╔════╝██║ ██╔╝██╔══██╗ ██║   ██║██╔══██╗
  ██╔██╗ ██║██║███████╗█████═╝ ███████║ ██║   ██║███████║
  ██║╚██╗██║██║╚════██║██╔═██╗ ██╔══██║ ╚██╗ ██╔╝██╔══██║
  ██║ ╚████║██║███████║██║  ██ ██║  ██║  ╚████╔╝ ██║  ██║
  ╚═╝  ╚═══╝╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═══╝  ╚═╝  ╚═╝
```

### Main HUD Options:
- **`[T]` Terminal UI (Interactive Live CLI)** — Enters the prompt-driven interactive REPL.
- **`[W]` Web Workspace (Serve React UI)** — Starts the local REST/SSE server and opens your web browser.
- **`[I]` Investigate Stock Ticker** — Promptly initiates an investigation on any IDX ticker (e.g., `ANTM`, `BBCA`).
- **`[H]` Help & Commands Guide** — Displays interactive help guide and slash commands menu.
- **`[Q]` Exit Niskava** — Exits the application.

---

## ⌨️ 3. Interactive REPL Features & Navigation

Inside the **Terminal UI (Option `[T]`)**, you can query market intelligence using natural language prompts.

### A. Prompt History Navigation (Up & Down Arrow Keys)
- Press **Up Arrow (↑)** to recall previously entered research queries or prompts.
- Press **Down Arrow (↓)** to return to newer prompt entries.

### B. Live Animated Braille Spinner (`⠋`)
While the ReAct agent evaluates market data or queries Sectors API tools, an animated **Braille Spinner** provides live status updates (80ms tick frame):
```text
⠋ [hermes] Initializing analysis & planning investigation...
```

### C. Autocomplete Slash Commands (`/`)
Type `/` inside the REPL to display an autocomplete popup menu styled with category badges (`[SYSTEM]`, `[NAV]`, `[INTEL]`):

| Slash Command | Category | Purpose & Description |
|---|---|---|
| `/investigate <TICKER>` | `[INTEL]` | Triggers a full 7-stage investigation pipeline on an IDX ticker. E.g., `/investigate ANTM` |
| `/screen <CRITERIA>` | `[INTEL]` | Screens IDX stocks based on volume spikes or abnormal return metrics. |
| `/health <TICKER>` | `[INTEL]` | Analyzes financial health, Altman Z-Score, and solvency ratios. |
| `/memory` | `[NAV]` | Visualizes local conversational graph memory (*associative ego-graph*). |
| `/clear` | `[SYSTEM]` | Clears the terminal screen buffer. |
| `/exit` | `[SYSTEM]` | Exits the REPL and **returns to the Main HUD Launcher Menu**. |

---

## 📊 4. Autonomous Headless Investigation CLI

For automated scripts or headless execution:

### Investigate Single Ticker:
```bash
niskava investigate ANTM --days 30
```

### Options & Flags:
- `--days <N>`: Number of daily candlestick trading sessions to analyze (default: 30).
- `--offline`: Runs in offline mode using local mock fixtures without LLM inference.
- `--output <json|markdown>`: Output format for investigation findings.

---

## 🌐 5. Web Canvas Workspace (`niskava serve`)

To access TradingView candlestick charts, volume anomaly markers, and interactive evidence matrices:

```bash
niskava serve --port 8080 --open
```

Open your browser at `http://localhost:8080` to access:
1. **Interactive Candlestick & Anomaly Chart**: Visual markers for volume surges ($V_z \ge 2.5\sigma$) and abnormal price movements.
2. **Real-time SSE Thinking Stream**: Streaming ReAct agent reasoning steps.
3. **Evidence Matrix & Timeline Graph**: Temporal relationship graph linking official IDXnet disclosures and financial media coverage.

---

## 🤖 6. Telegram Bot Integration (`niskava bot`)

Niskava Agent supports Telegram Bot integration to answer market research queries directly from Telegram.

### Launching Telegram Bot:
```bash
export TELEGRAM_BOT_TOKEN="YOUR_TELEGRAM_BOT_TOKEN"
niskava bot
```
