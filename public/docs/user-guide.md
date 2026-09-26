# User Guide & Interface Manual

Niskava Agent provides multiple interaction surfaces suited for ad-hoc terminal research, browser-based visual investigation, automated CI/CD scripting, and external agent integrations.

---

## Overview of Interaction Surfaces

```
┌───────────────────────────────────────────────────────────────────┐
│                    NISKAVA INTERACTION SURFACES                   │
├───────────────────────────────────────────────────────────────────┤
│ 1. Terminal UI (TUI) REPL & HUD Launcher                          │
│    Command : `niskava` (without arguments)                        │
│    Use Case: Interactive prompt-driven research, slash commands.  │
│                                                                   │
│ 2. Autonomous Headless Pipeline CLI                               │
│    Command : `niskava investigate <TICKER> --days 30`             │
│    Use Case: Scripted, non-interactive execution & audit trails.  │
│                                                                   │
│ 3. Web Workspace (Browser Visual Terminal)                        │
│    Command : `niskava serve --port 20128 --open`                  │
│    Use Case: Candlestick charting, SSE stream, evidence matrices. │
│                                                                   │
│ 4. Session History & Management                                   │
│    Command : `niskava sessions`                                   │
│    Use Case: Reviewing past investigations and resuming chats.    │
│                                                                   │
│ 5. Interactive Knowledge Graph Visualization                     │
│    Command : `niskava graph --open`                               │
│    Use Case: Visualizing associative memory graphs in browser.    │
│                                                                   │
│ 6. Model Context Protocol (MCP) Server                            │
│    Command : `niskava mcp`                                        │
│    Use Case: Connecting Niskava to Claude Desktop, Cursor, etc.   │
│                                                                   │
│ 7. Telegram Bot Runner                                            │
│    Command : `niskava telegram`                                   │
│    Use Case: Mobile market alerts and conversational queries.     │
└───────────────────────────────────────────────────────────────────┘
```

---

## 1. Terminal UI (TUI) HUD Launcher

Executing `niskava` without arguments starts the interactive Terminal UI launcher:

```bash
./niskava
```

The launcher displays application health status, configured API keys, and a menu navigable with arrow keys or shortcut letters:

- **`[T]` Terminal UI (Interactive Live CLI)**: Launches the natural language REPL.
- **`[W]` Web Workspace (Serve React UI)**: Starts the background server and opens your web browser.
- **`[S]` Saved Sessions & History**: Browse and resume previous chat or investigation sessions.
- **`[H]` Health & Diagnostics**: Displays database status, virtual environment paths, and provider connectivity.
- **`[L]` Language Selector**: Toggle between English (`en`) and Indonesian (`id`).
- **`[U]` Setup Wizard**: Re-run the interactive configuration setup.
- **`[?]` Help & Commands Guide**: Complete overview of terminal hotkeys and slash commands.
- **`[Q]` Exit Niskava**: Cleanly closes daemon processes and returns to shell.

---

## 2. Interactive Terminal REPL

The interactive REPL provides a prompt-driven environment with real-time reasoning feedback and rich markdown rendering.

### Prompt History Navigation
- Press **Up Arrow (↑)** to navigate backwards through previous prompt history.
- Press **Down Arrow (↓)** to navigate forward through newer prompts.

### Real-Time Braille Progress Indicator
During tool execution, Sectors API queries, and web harvesting, an animated Braille spinner displays live progress:

```text
⠋ [hermes] Running market anomaly reconnaissance on ANTM...
```

### Natural Language Prompt-Driven Research
The REPL is powered by an autonomous Hermes-style ReAct loop with progressive skill disclosure. You do not need to memorize rigid syntax—simply ask questions in plain Indonesian or English:
- *"Analisis saham BBCA: apakah foreign flow 5 hari terakhir searah dengan IHSG?"*
- *"Mengapa saham BUMI mengalami lonjakan volume kemarin? Cek keterbukaan informasi IDX."*
- *"Bandingkan valuasi perbankan big-4 (BBCA, BBRI, BMRI, BBNI) dengan Altman Z-Score."*

### Autocomplete Slash Commands
Type `/` in the prompt input to open the interactive autocomplete popup:

| Command | Category | Description | Example |
|---|---|---|---|
| `/help` | `[SYSTEM]` | Displays available keyboard shortcuts and slash commands. | `/help` |
| `/chats` | `[NAV]` | Opens interactive Bubbletea session selector to browse & switch chats. | `/chats` |
| `/resume <ID>` | `[INTEL]` | Resumes a specific chat session by its unique ID. | `/resume CHAT-20260925-0001` |
| `/timeout [val]` | `[SYSTEM]` | Sets LLM inference timeout (`fast`, `balanced`, `deep`, `local`, or seconds `10-300`). | `/timeout balanced` |
| `/graph` | `[INTEL]` | Opens the associative knowledge graph visualization directly in browser. | `/graph` |
| `/web` | `[NAV]` | Launches/opens the Web Workspace canvas in your default browser. | `/web` |
| `/sessions` | `[INTEL]` | Displays recent investigation and chat sessions stored in local SQLite. | `/sessions` |
| `/health` | `[SYSTEM]` | Prints daemon status, database connection, and AI provider latency check. | `/health` |
| `/lang [en\|id]` | `[SYSTEM]` | Switches interface and response language (`en` or `id`). | `/lang id` |
| `/reset` | `[SYSTEM]` | Resets working memory graph for the current session and starts fresh. | `/reset` |
| `/clear` | `[SYSTEM]` | Clears the terminal screen and redraws the banner. | `/clear` |
| `/back` or `/exit` | `[NAV]` | Returns cleanly to the Main HUD Launcher menu. | `/back` |

---

### Inference Timeout Configuration
Niskava features an **Adaptive Inference Timeout Engine** that automatically scales LLM reasoning time based on how many tool observations have been collected:
$$\text{Timeout} = \text{Base Timeout} + (\text{Tool Observations} \times 10\text{ seconds})$$

You can customize the base timeout across three convenient interfaces:
1. **Interactive Setup Wizard**: Run `niskava setup` and choose Step 5 (Fast 25s, Balanced 60s, Deep 120s, Local 180s, Custom).
2. **Interactive REPL**: Use `/timeout fast`, `/timeout balanced`, `/timeout deep`, `/timeout local`, or `/timeout 90`.
3. **Web Workspace Settings**: Open the Settings modal and adjust the **Inference Timeout** slider (10s – 300s). Changes are synchronized immediately.

---

## 3. Autonomous Headless Investigation CLI

For automated scripts, scheduled cron jobs, or batch processing, run investigations directly from the command line:

```bash
./niskava investigate <TICKER> [flags]
```

### Available Flags:
- `--days <N>`: Number of daily trading sessions to analyze (default: `30`).
- `--offline`: Executes using local mock fixtures without issuing live Sectors API requests or consuming credits.
- `--lang <en|id>`: Output language (`en` for English, `id` for Indonesian).
- `--verbose, -v`: Prints detailed debug logs and IPC payload messages.
- `--config, -c <path>`: Specifies a custom configuration file path.

### Example:
```bash
./niskava investigate ANTM --days 60
```

The command outputs a structured terminal report detailing:
1. Identified quantitative anomalies (Volume Z-Scores, Abnormal Returns, Sector Divergence).
2. Harvested corporate filings and financial news.
3. Chronological causality assessment (`LIKELY_CATALYST`, `PRECEDED_ANNOUNCEMENT`, etc.).
4. Structured evidence matrix with discrete confidence ratings.

---

## 4. Web Workspace (`niskava serve`)

Niskava includes a self-contained local web application featuring TradingView/Recharts candlestick charts, real-time Server-Sent Events (SSE) streaming, and interactive evidence causality maps:

```bash
./niskava serve --port 20128 --open
```

### Web Workspace Features:
- **Interactive Candlestick Charting**: Visual candlestick price history overlaid with volume surge markers ($V_z \ge 2.5$) and price breakout tags ($|R_t| \ge 5\%$).
- **Live SSE Streaming**: Watch the ReAct agent's thoughts, tool calls, and evidence collection unfold in real time.
- **Interactive Evidence Matrix**: Filter findings by verification status (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
- **Causality Timeline Graph**: Interactive visual timeline correlating news publication timestamps against trading volume spikes.

To run the web server in the background alongside the CLI, specify `--port` as needed (default is `20128`).

---

## 5. Session Management (`niskava sessions`)

All chat conversations and investigation runs are persisted in local SQLite storage.

### List Saved Sessions:
```bash
./niskava sessions --type all --limit 20
```

Options for `--type`:
- `all`: Displays both conversational chats and structured ticker investigations.
- `chat`: Filters for interactive REPL conversational sessions.
- `investigation`: Filters for standalone ticker investigation runs.

### Resume a Specific Session in REPL:
```bash
./niskava --session CHAT-20260921-0001
```

---

## 6. Interactive Knowledge Graph Export (`niskava graph`)

Niskava can render its local associative knowledge graph into a standalone, interactive HTML visualization:

```bash
# Render the entire local associative memory graph
./niskava graph --open

# Render a graph scoped to a specific investigation
./niskava graph --session INV-20260921-ANTM -o ./antm_graph.html --open
```

The generated HTML file uses PyVis and NetworkX to provide physics-based node clustering, entity filtering (tickers, catalysts, executives, sectors), and inspectable edge relationships.

---

## 7. Model Context Protocol (MCP) Server (`niskava mcp`)

Niskava implements the standard Model Context Protocol (MCP) over standard input/output (`stdio`), allowing external agent environments (such as Claude Desktop, Cursor, and Antigravity) to directly utilize Niskava tools and data sources.

### Starting the MCP Server:
```bash
./niskava mcp
```

### Claude Desktop Integration Configuration
Add Niskava to your Claude Desktop configuration file:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

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

### Exposed MCP Primitives:
- **Tools**:
  - `compute_quant_anomalies`: Calculates Volume Z-scores, abnormal returns, and sector divergence.
  - `get_sectors_daily`: Retrieves cached daily candlestick time-series data.
  - `get_sectors_company_report`: Fetches comprehensive company profiles and financial metrics.
  - `harvest_market_news`: Executes temporal-aware news and regulatory filing dorking.
  - `query_graph_memory`: Queries local associative memory nodes and relationships.
- **Resources**: System cache statistics and local database health.
- **Prompts**: Standardized multi-step investigative research workflows.

---

## 8. Telegram Bot Integration (`niskava telegram`)

Niskava can operate as a Telegram bot using long-polling, delivering market intelligence and anomaly alerts directly to mobile devices:

```bash
# Set your Telegram bot token
export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"

# Launch bot runner
./niskava telegram
# or: ./niskava bot
```

Users can query stock tickers, request financial health assessments, or receive automated anomaly alerts directly in their Telegram chat.
