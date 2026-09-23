# 📥 Niskava Agent — Installation & Setup Guide

This document provides a comprehensive step-by-step guide on system prerequisites, environment setup, API key configuration, and building the Niskava Agent executable across Windows, Linux, and macOS platforms.

---

## 📋 1. System Prerequisites

Before installing Niskava Agent, ensure your system satisfies the following software requirements:

| Software | Minimum Version | Description / Purpose |
|---|---|---|
| **Go** | `v1.22` or higher | Required for compiling Go Core, CLI, REST/SSE Server, and SQLite persistence. |
| **Python** | `v3.11` or higher | Required for the Python Agent Engine (ReAct loop, NumPy Anomaly Engine, OSINT harvester). |
| **Git** | `v2.30` or higher | Required for cloning the repository. |
| **Node.js & npm** *(Optional)* | `v18.0` or higher | Required only if you intend to customize or build the Web Workspace React/Vite frontend. |

---

## 🛠️ 2. Step-by-Step Installation

### Step 1: Clone the Repository
Open your shell / terminal and execute:

```bash
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents
```

---

### Step 2: Set Up Python Engine Virtual Environment

The Python Engine handles non-generative anomaly calculations, OSINT dorking, and ReAct agent loops.

#### On Windows (PowerShell / CMD):
```powershell
cd backend/engine
python -m venv venv
.\venv\Scripts\Activate.ps1
pip install -r requirements.txt
cd ..\..
```

#### On Linux / macOS:
```bash
cd backend/engine
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
cd ../..
```

> **💡 Core Dependencies:** The key Python packages include `numpy`, `pandas`, `networkx`, `trafilatura`, `requests`, `pydantic`, and `beautifulsoup4`.

---

### Step 3: Configure API Keys & Environment Variables

Niskava Agent requires a **Sectors Financial API v2 Key** to access official market data (OHLCV candles, financial statements, Net Foreign Flow, corporate actions). For qualitative ReAct reasoning, Niskava supports **Gemini API** (default) or any OpenAI-compatible provider (Ollama / OpenRouter / vLLM).

#### Option A: Using the Interactive Setup Wizard (Recommended)

Niskava provides an interactive CLI wizard that automatically creates your configuration at `~/.niskava/config.yaml` and environment file at `~/.niskava/.env`:

```bash
go run ./cmd/niskava setup
```

The setup wizard will prompt you for:
1. **Sectors API v2 Key** (`SECTORS_API_KEY`) — Get a free key at [Sectors Financial API](https://sectors.app/).
2. **AI Provider Choice** — Select `gemini` (default), `ollama`, `openrouter`, or `vllm`.
3. **Gemini API Key / OpenRouter Key** (`GEMINI_API_KEY`).

#### Option B: Manual Configuration (`~/.niskava/config.yaml`)

Create the configuration file at `~/.niskava/config.yaml` (Windows: `C:\Users\<Username>\.niskava\config.yaml`):

```yaml
sectors:
  api_key: "YOUR_SECTORS_API_KEY_HERE"
  base_url: "https://api.sectors.app/v2"

ai:
  provider: "gemini"
  model: "gemini-2.5-flash"
  api_key: "YOUR_GEMINI_API_KEY_HERE"

storage:
  db_path: "~/.niskava/niskava.db"
  journal_mode: "WAL"

server:
  port: 8080
  host: "127.0.0.1"
```

---

### Step 4: Build Standalone Executable Binary (Go Core)

To compile a single executable binary with zero external runtime dependencies:

#### On Windows:
```powershell
go build -o niskava.exe ./cmd/niskava
```

#### On Linux / macOS:
```bash
go build -o niskava ./cmd/niskava
chmod +x niskava
```

Binaries can also be built into the `bin/` directory for convenience:
```powershell
go build -o bin/niskava.exe ./cmd/niskava
```

---

## ⚡ 3. Installation Verification

Run the automated test suites to verify that all components are set up properly:

### Test CLI Binary:
```bash
# Launch binary to view HUD launcher menu
./niskava.exe
```

### Test Python Engine Unit Tests:
```bash
pytest backend/engine
```
*(Verifies 207+ Python unit tests pass clean).*

### Test Go Core Unit Tests:
```bash
go test ./...
```
*(Verifies IPC, SQLite WAL persistence, and TUI pass 100%).*

---

## 🌐 4. Offline Mode (Zero Credit Consumption)

If you want to run deterministic quantitative analysis without invoking LLMs or consuming Sectors API credits:

```bash
# Uses pre-packaged static JSON mock fixtures
$env:MOCK_SECTORS="1"
.\niskava.exe investigate ANTM --offline
```
