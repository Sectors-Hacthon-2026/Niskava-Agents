# Installation and Setup Guide

This guide provides instructions for installing, configuring, and verifying Niskava Agent on Linux, macOS, Windows, and Docker.

---

## 1. System Requirements

Ensure your host environment meets the minimum software requirements before proceeding:

| Component | Minimum Version | Required By | Purpose |
|---|---|---|---|
| **Go** | `1.22` or higher | Go Core | CLI entry points, REST/SSE server, SQLite WAL persistence, IPC broker. |
| **Python** | `3.11` or higher | Python Engine | Deterministic NumPy math, News harvesting, ReAct reasoning agent loop. |
| **Git** | `2.30` or higher | Source control | Cloning and updating repository files. |
| **Docker** *(Optional)* | `20.10` or higher | Containerization | Zero-install alternative running everything in container. |

---

## 2. Fast Installation (Recommended)

### Option 0: Zero-Clone via NPX / NPM (Instant Run)

If you have Node.js (>= 18) installed, you can launch Niskava Agent immediately without cloning the git repository:

```bash
# Run interactive setup wizard
npx @zyrexnns/niskava-agent setup

# Check environment readiness
npx @zyrexnns/niskava-agent doctor

# Run conversational terminal (REPL)
npx @zyrexnns/niskava-agent

# Start local web workspace (:20128)
npx @zyrexnns/niskava-agent serve
```

To install globally as a system-wide command:
```bash
npm install -g @zyrexnns/niskava-agent
niskava setup
```

---

### Option A: From Source via Git Clone

Clone the repository first:
```bash
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents
```

### Option A: One-Liner Script Installers

#### On Linux & macOS:
```bash
chmod +x install.sh
./install.sh
```
*The installer automatically verifies Go and Python, builds the `.venv` in the repository root, installs quantitative packages, and compiles `bin/niskava`.*

#### On Windows (PowerShell):
```powershell
.\install.ps1
```
> **Note on PowerShell Script Execution:** If script execution is restricted on Windows, run:
> `Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass` in your current PowerShell window before executing `.\install.ps1`.

---

### Option B: Docker Container (Zero-Install)

If you have Docker and Docker Compose installed:
```bash
# Start Web Workspace daemon in the background on http://localhost:8080
docker compose up -d

# Check live logs
docker compose logs -f
```

---

### Option C: Manual Step-by-Step Installation

If you prefer to configure everything manually:

#### Step 1: Create Python Virtual Environment (.venv) at Project Root

**On Linux / macOS:**
```bash
python3 -m venv .venv
.venv/bin/pip install --upgrade pip
.venv/bin/pip install -r backend/engine/requirements.txt
```

**On Windows (PowerShell / CMD):**
```powershell
python -m venv .venv
.\.venv\Scripts\pip.exe install --upgrade pip
.\.venv\Scripts\pip.exe install -r backend\engine\requirements.txt
```

#### Step 2: Compile Standalone Go Core Binary

**On Linux / macOS:**
```bash
mkdir -p bin
go build -o bin/niskava ./cmd/niskava
chmod +x bin/niskava
```

**On Windows:**
```powershell
if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" }
go build -o bin\niskava.exe .\cmd\niskava
```

---

## 3. Configuration & Setup Wizard

### Method A: Dynamic Setup Wizard (Recommended)
Run the dynamic setup wizard:

```bash
# On Linux / macOS:
./bin/niskava setup

# On Windows:
.\bin\niskava.exe setup
```

The wizard guides you through:
1. Selecting your AI inference provider (OpenRouter, Google Gemini, Ollama, DeepSeek, Groq, or OpenAI).
2. Entering your **Sectors Financial API v2 Key** ([sectors.app](https://sectors.app/)) or pressing Enter for 100% Offline Mock Mode.
3. Automatically detecting and bootstrapping the Python `.venv` environment if missing.
4. Testing live connectivity against endpoints.
5. Saving configuration synchronously to both `.env` and `~/.niskava/config.yaml`.

---

### Method B: Manual Configuration

You can manually edit or create `~/.niskava/config.yaml` (Windows: `C:\Users\<Username>\.niskava\config.yaml`):

```yaml
version: "1.0.0"

auth:
  sectors_api_key: "YOUR_SECTORS_API_KEY"
  gemini_api_key: "YOUR_GEMINI_API_KEY"
  openai_api_key: ""
  openai_base_url: "https://openrouter.ai/api/v1"
  openai_model: "deepseek/deepseek-chat"

ai:
  provider: "openai"               # Options: openai, gemini, ollama
  model: "deepseek/deepseek-chat"
  temperature: 0.1

storage:
  db_path: "~/.niskava/niskava.db" # Local SQLite database location (Law 4)
  journal_mode: "WAL"

server:
  host: "127.0.0.1"
  port: 20128

preferences:
  language: "id"                   # "id" (Indonesian) or "en" (English)
  default_market: "IDX"
  default_timeframe_days: 30
```

---

## 4. Verification & Diagnostics

### 1. Run System Health Doctor
Verify that all system components, quantitative libraries, and database permissions are ready:

```bash
# Linux / macOS:
./bin/niskava doctor

# Windows:
.\bin\niskava.exe doctor
```

The visual diagnostic HUD checks:
- Operating system and architecture
- SQLite database WAL mode status
- Python quantitative engine (`numpy`, `pandas`, `networkx`)
- AI provider endpoint reachability & latency
- Sectors Financial API quota and mock status
- Engine directory mobility

---

### 2. Launch Web Workspace (Dashboard)
Start the REST/SSE daemon and open the interactive dashboard:

```bash
# Opens http://localhost:20128 automatically in your default browser:
./bin/niskava serve

# Run headlessly (without opening browser):
./bin/niskava serve --open=false
```

---

### 3. Launch Interactive Terminal UI (TUI)
Launch the interactive terminal research terminal:

```bash
./bin/niskava
```

---

### 4. Enable Shell Autocompletion (Optional)

Generate tab-completion for subcommands and popular IDX ticker suggestions (`BBCA`, `BBRI`, `ANTM`, etc.):

```bash
# Bash:
source <(./bin/niskava completion bash)

# Zsh:
source <(./bin/niskava completion zsh)

# PowerShell (Windows):
.\bin\niskava.exe completion powershell | Out-String | Invoke-Expression
```
