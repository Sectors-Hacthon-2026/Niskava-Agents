# Installation and Setup Guide

This guide provides step-by-step instructions for installing, configuring, and verifying Niskava Agent across Linux, macOS, Windows, and Docker environments.

---

## 1. System Prerequisites

Before installation, verify the software installed on your host machine:

| Component | Minimum Version | Required For | Notes |
|---|---|---|---|
| **Node.js** | `18.0.0+` | NPM / NPX Launcher | Recommended for instant, zero-clone installation. |
| **Python** | `3.11+` | Quantitative Engine | Required for NumPy math, news harvesting, and local graph memory. |
| **Go** | `1.22+` | Source Build Only | Only required if compiling Go Core from source code. |
| **Git** | `2.30+` | Source Build | Required for cloning the repository. |
| **Docker** | `20.10+` | Container Mode | Optional zero-install alternative. |

---

## 2. Installation Methods

Choose the installation method that fits your environment:

### Method A: Instant Zero-Clone via NPX / NPM (Recommended)

If you have Node.js installed, you can launch Niskava Agent immediately without cloning the git repository or manually compiling binaries:

```bash
# 1. Run the interactive setup wizard (configures API keys)
npx @zyrexnns/niskava-agent setup

# 2. Run system doctor diagnostics to verify your setup
npx @zyrexnns/niskava-agent doctor

# 3. Launch the interactive REPL research terminal
npx @zyrexnns/niskava-agent

# 4. Or launch the local Web Workspace (:20128)
npx @zyrexnns/niskava-agent serve
```

#### Global Installation (System-Wide Command)
To install `niskava` globally on your machine so it is accessible from any terminal window:

```bash
npm install -g @zyrexnns/niskava-agent

# Check installed version
niskava version

# Run anywhere
niskava setup
niskava doctor
niskava investigate ANTM --days 30
```

#### How the NPM Launcher Works Under the Hood:
1. **User-Space Isolation:** Precompiled Go binaries are automatically downloaded and cached strictly in `~/.niskava/bin/` (Windows: `%USERPROFILE%\.niskava\bin`). No files are written into root-protected system directories, completely preventing `EACCES` permission errors.
2. **Native Architecture Detection:** Detects the exact operating system and CPU architecture:
   - Linux: `x64` (`amd64`), `arm64`
   - macOS: Apple Silicon (`arm64`), Intel (`x64`)
   - Windows: `x64` (`amd64`)
3. **Automatic Engine Synchronization:** Python analytical skills and agent scripts are automatically mirrored to `~/.niskava/engine`, resolving module paths seamlessly.

---

### Method B: Prebuilt Standalone Binaries (GitHub Releases)

Download precompiled standalone binaries directly from [GitHub Releases](https://github.com/Sectors-Hacthon-2026/Niskava-Agents/releases):

#### 1. Linux (x86_64 / ARM64)
```bash
# Download binary (replace with desired architecture)
curl -LO https://github.com/Sectors-Hacthon-2026/Niskava-Agents/releases/latest/download/niskava-linux-amd64

# Grant execution permissions
chmod +x niskava-linux-amd64
sudo mv niskava-linux-amd64 /usr/local/bin/niskava

# Verify
niskava version
```

#### 2. macOS (Apple Silicon / Intel)
```bash
# For Apple Silicon (M1/M2/M3/M4):
curl -LO https://github.com/Sectors-Hacthon-2026/Niskava-Agents/releases/latest/download/niskava-darwin-arm64
chmod +x niskava-darwin-arm64
sudo mv niskava-darwin-arm64 /usr/local/bin/niskava

# For Intel Macs:
curl -LO https://github.com/Sectors-Hacthon-2026/Niskava-Agents/releases/latest/download/niskava-darwin-amd64
chmod +x niskava-darwin-amd64
sudo mv niskava-darwin-amd64 /usr/local/bin/niskava
```

#### 3. Windows (x64)
1. Download `niskava-windows-amd64.exe` from GitHub Releases.
2. Rename to `niskava.exe`.
3. Add the directory to your user `PATH` environment variable.
4. Run `niskava.exe setup` in PowerShell or Windows Terminal.

---

### Method C: Build from Source Code (Developer Mode)

For contributors and developers who wish to modify the source code:

#### 1. Clone Repository
```bash
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents
```

#### 2. Run Automated Setup Script
* **Linux & macOS:**
  ```bash
  chmod +x install.sh && ./install.sh
  ```
* **Windows (PowerShell):**
  ```powershell
  Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
  .\install.ps1
  ```

#### 3. Manual Build Steps (Alternative)
If you prefer building step-by-step manually:

```bash
# Setup Python virtual environment
python3 -m venv backend/engine/.venv
source backend/engine/.venv/bin/activate
pip install -r backend/engine/requirements.txt

# Compile Go Core binary
go build -o bin/niskava ./cmd/niskava
```

---

### Method D: Docker Container (Zero-Install)

Run Niskava Agent inside an isolated Docker container:

```bash
# 1. Clone repository
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents

# 2. Configure environment
cp .env.example .env
# Edit .env with your SECTORS_API_KEY and GEMINI_API_KEY

# 3. Start container with Docker Compose
docker compose up -d

# 4. Access the Web Workspace
# Open http://localhost:20128 in your browser
```

---

## 3. Initial Configuration (`niskava setup`)

Run the interactive setup wizard to configure your credentials:

```bash
niskava setup
```

The wizard guides you through:

### Step 1: AI Provider Selection
Choose your preferred inference backend:
* **Google Gemini (Recommended):** Uses `gemini-2.0-flash` or `gemini-1.5-pro`. Fastest inference speed and large context window.
* **OpenAI / OpenRouter:** Compatible with OpenAI, OpenRouter, Groq, or DeepSeek models.
* **Local Ollama:** Completely offline private inference (e.g. `deepseek-r1:8b`, `qwen2.5:7b`).
* **Offline Mock Mode:** Deterministic execution without any LLM calls (ideal for CI/CD).

### Step 2: Sectors Financial API Key (**Required for Live Data**)
* Enter your **Sectors Financial API v2 key** (obtain free from [sectors.app](https://sectors.app/)).
* **Live Mode Auto-Toggle:** When a valid Sectors API key is entered, Niskava automatically activates **Live Mode** (`MOCK_SECTORS=0`).
* *Note:* If you press ENTER without entering a key, Niskava will alert you that it is running in Mock Simulation Mode (synthetic fixtures).

### Step 3: Local Storage & Preferences
* Database path defaults to `~/.niskava/niskava.db`.
* Language preference (`id` for Indonesian, `en` for English).

---

## 4. Verification & Diagnostics (`niskava doctor`)

Always run the built-in system doctor to verify environment readiness before your first research session:

```bash
niskava doctor
```

Example successful diagnostic output:
```text
  ██████╗  ██████╗  ██████╗████████╗ ██████╗ ██████╗
  ██╔══██╗██╔═══██╗██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗
  ██║  ██║██║   ██║██║        ██║   ██║   ██║██████╔╝
  ██║  ██║██║   ██║██║        ██║   ██║   ██║██╔══██╗
  ██████╔╝╚██████╔╝╚██████╗   ██║   ╚██████╔╝██║  ██║
  ╚═════╝  ╚═════╝  ╚═════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝
  Niskava System Health & Environmental Verification

  [✓] Go Runtime: go1.24.0 (PASS)
  [✓] Python Binary: /usr/bin/python3 (v3.11.8) (PASS)
  [✓] Python Packages: numpy, pandas, networkx, trafilatura, fpdf2 (PASS)
  [✓] Database: ~/.niskava/niskava.db (WAL Mode Enabled) (PASS)
  [✓] Sectors API v2: Connection verified (LIVE) (PASS)
  [✓] AI Provider: Gemini (gemini-2.0-flash) (READY) (PASS)

  Doctor check passed! Niskava Agent is fully operational.
```

If any check fails, `niskava doctor` provides immediate actionable recommendations. Refer to the [Troubleshooting & FAQ Guide](troubleshooting-and-faq.md) for detailed error resolutions.
