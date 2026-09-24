# Installation and Setup Guide

This guide provides instructions for installing, configuring, and verifying Niskava Agent on Linux, macOS, and Windows.

---

## 1. System Requirements

Ensure your host environment meets the minimum software requirements before proceeding:

| Component | Minimum Version | Required By | Purpose |
|---|---|---|---|
| **Go** | `1.22` or higher | Go Core | CLI entry points, REST/SSE server, SQLite WAL persistence, IPC broker. |
| **Python** | `3.11` or higher | Python Engine | Deterministic NumPy math, OSINT harvesting, ReAct reasoning agent loop. |
| **Git** | `2.30` or higher | Source control | Cloning and updating repository files. |
| **Node.js / npm** *(Optional)* | `18.0` or higher | Web Client | Only required if rebuilding or modifying the React SPA frontend. |

---

## 2. Step-by-Step Installation

### Step 1: Clone the Repository
Clone the repository and enter the project directory:

```bash
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents
```

---

### Step 2: Set Up the Python Engine Virtual Environment

The Python Engine executes deterministic quantitative algorithms, web scraping, and agent orchestration. It must be run inside a dedicated virtual environment.

#### On Linux / macOS:
```bash
cd backend/engine
python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt
cd ../..
```

#### On Windows (PowerShell):
```powershell
cd backend\engine
python -m venv venv
.\venv\Scripts\Activate.ps1
pip install --upgrade pip
pip install -r requirements.txt
cd ..\..
```

> **Note on PowerShell Script Execution:** If you encounter an execution policy error on Windows, run `Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass` in your PowerShell session before activating the virtual environment.

Core Python dependencies installed via `requirements.txt`:
- `numpy`, `pandas`: Deterministic time-series and anomaly calculations.
- `networkx`: In-memory directed graph modeling for conversational and market memory.
- `trafilatura`, `beautifulsoup4`, `feedparser`: OSINT harvesting, web article parsing, and content sanitization.
- `requests`, `urllib3`: HTTP client for external data ingestion.
- `pydantic`: Strict data validation and schema serialization.
- `pytest`: Unit testing framework.

---

### Step 3: Configure Authentication & Settings

Niskava Agent requires credentials for official market data access and an AI provider for conversational and investigative synthesis.

#### Required Credentials
1. **Sectors Financial API v2 Key**: Required for official IDX market data (daily candlestick prices, corporate actions, financial ratios, net foreign flow). Register for an API key at [sectors.app](https://sectors.app/).
2. **AI Provider Key**:
   - **Google Gemini** (Default): Fast, multimodal reasoning with `gemini-2.5-flash` or `gemini-1.5-pro`. Get a key from [Google AI Studio](https://aistudio.google.com/).
   - **Ollama** (Local): Fully local, private inference without external cloud API calls.
   - **OpenRouter / vLLM / OpenAI-compatible**: Any standard OpenAI-compatible API endpoint.

---

#### Method A: Interactive Setup Wizard (Recommended)

Run the interactive setup wizard via Go:

```bash
go run ./cmd/niskava setup
```

The wizard guides you through:
1. Detecting system tools (Go, Python, virtual environment path).
2. Entering your Sectors API key and verifying connection status.
3. Selecting your AI inference provider (Gemini, Ollama, OpenRouter, vLLM).
4. Generating `~/.niskava/config.yaml` and `~/.niskava/.env` with secure file permissions.

---

#### Method B: Manual Configuration

You can manually create the configuration file at `~/.niskava/config.yaml` (on Windows: `C:\Users\<Username>\.niskava\config.yaml`).

Sample `config.yaml`:

```yaml
version: "1.0.0"

auth:
  sectors_api_key: "YOUR_SECTORS_API_KEY"
  gemini_api_key: "YOUR_GEMINI_API_KEY"
  openai_api_key: ""

ai:
  provider: "gemini"               # Options: gemini, ollama, openrouter, vllm
  model: "gemini-2.5-flash"        # Target model identifier
  endpoint: ""                     # Required only for Ollama (e.g., http://localhost:11434/v1) or vLLM
  temperature: 0.1                 # Low temperature for analytical consistency

storage:
  db_path: "~/.niskava/niskava.db" # Local SQLite database location
  journal_mode: "WAL"              # Write-Ahead Logging for concurrency

server:
  host: "127.0.0.1"
  port: 20128                      # Background daemon and REST/SSE port

engine:
  python_bin: ""                   # Path to python executable (auto-detected if empty)
  engine_path: ""                  # Path to backend/engine (auto-detected if empty)

preferences:
  language: "en"                   # Interface language: "en" (English) or "id" (Indonesian)
  default_market: "IDX"
  default_timeframe_days: 30
```

You can also export environment variables directly:

```bash
export SECTORS_API_KEY="your_sectors_api_key_here"
export GEMINI_API_KEY="your_gemini_api_key_here"
export NISKAVA_DB_PATH="$HOME/.niskava/niskava.db"
```

---

### Step 4: Build the Standalone Binary

Compile the single Go Core executable:

#### On Linux / macOS:
```bash
go build -o niskava ./cmd/niskava
chmod +x niskava
```

#### On Windows (PowerShell / CMD):
```powershell
go build -o niskava.exe ./cmd/niskava
```

Optionally move the binary into your system `PATH` (e.g., `/usr/local/bin` on Linux/macOS, or a custom scripts directory on Windows) to execute `niskava` globally from any terminal.

---

## 3. Verifying the Installation

### 1. Verify Go Core Binary & Diagnostic HUD
Execute the compiled binary without subcommands:

```bash
./niskava
```

The terminal displays the HUD Launcher menu. Select `[H]` to view health diagnostics, confirming database connectivity, Python engine status, and API key validity.

### 2. Run Python Engine Test Suite
Verify that all quantitative algorithms, memory engines, and tool integrations pass automated tests:

```bash
# From project root with venv activated
pytest backend/engine -v
```

All 200+ unit tests should pass with zero errors.

### 3. Run Go Core Test Suite
Verify the Go Core daemon, IPC pipeline, and persistence layer:

```bash
go test ./...
```

---

## 4. Offline Mode (Zero Credit Consumption)

For CI/CD testing or evaluation without invoking external APIs or consuming Sectors API credits, run Niskava in offline mode using pre-packaged static mock fixtures:

```bash
# On Linux / macOS:
export MOCK_SECTORS=1
./niskava investigate ANTM --days 30 --offline

# On Windows (PowerShell):
$env:MOCK_SECTORS="1"
.\niskava.exe investigate ANTM --days 30 --offline
```

In offline mode:
- Quantitative data is served from local JSON fixtures.
- The deterministic compute gate evaluates formulas locally.
- Zero network requests are issued, ensuring offline reliability and testing determinism.
