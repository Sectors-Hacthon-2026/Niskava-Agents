# Troubleshooting & Frequently Asked Questions (FAQ)

This guide provides diagnostic procedures, common error resolutions, operational best practices, and answers to frequently asked questions about Niskava Agent.

---

## 1. Quick Diagnostic Check

Whenever you encounter an unexpected error, start by running the built-in system doctor:

```bash
# If using NPM/NPX:
npx @zyrexnns/niskava-agent doctor

# If using local binary:
niskava doctor
```

The doctor command verifies the Go runtime, Python binary, Python quantitative packages, SQLite WAL database, Sectors API connectivity, and AI provider credentials in under 2 seconds.

---

## 2. Common Issues & Solutions

### Issue 1: Mock Mode vs Live Mode (System Stays in Mock Simulation)

**Symptoms:**
* Every stock investigated triggers an identical synthetic volume surge with a Z-Score of **`35.71σ`** and 125,000,000 shares.
* The terminal logs indicate `Mock Simulation Mode active` or `Offline Mode`.

**Root Cause:**
* During initial setup, the Sectors Financial API key was left blank, which writes `MOCK_SECTORS=1` and `NISKAVA_OFFLINE=1` into `.env`.
* If you manually added `SECTORS_API_KEY` later but left `MOCK_SECTORS=1` or `NISKAVA_OFFLINE=1` in your configuration, the system remained locked in mock mode.

**Resolution:**
1. **The Easiest Fix — Run `niskava setup`:**
   Run the interactive setup wizard and enter your valid Sectors API key:
   ```bash
   niskava setup
   ```
   *Niskava automatically activates **Live Mode** (`MOCK_SECTORS=0`) as soon as a valid key is provided.*
2. **Manual Configuration (`.env` or `~/.niskava/.env`):**
   Open your `.env` file and set the flags to live mode:
   ```ini
   SECTORS_API_KEY=your_actual_sectors_api_key_here
   MOCK_SECTORS=0
   NISKAVA_OFFLINE=0
   ```
3. **Verify Live Connectivity:**
   Run `niskava doctor` and confirm:
   ```text
   [✓] Sectors API v2: Connection verified (LIVE) (PASS)
   ```

---

### Issue 2: Windows Python "WindowsApps" Stub Error

**Symptoms:**
On Windows, when running `niskava` or `niskava.exe`, the terminal reports:
```text
Python was not found; run without arguments to install from the Microsoft Store...
# or
engine subprocess error: The system cannot find the file specified.
```

**Root Cause:**
Windows ships with a zero-byte placeholder shortcut at `C:\Users\<User>\AppData\Local\Microsoft\WindowsApps\python.exe` that opens the Microsoft Store instead of running Python.

**Resolution:**
1. Install official Python 3.11 or higher from [python.org](https://www.python.org/downloads/). During installation, **check the box: "Add python.exe to PATH"**.
2. Alternatively, disable the Windows app execution aliases:
   * Open **Windows Settings** $\to$ **Apps** $\to$ **Advanced app settings** $\to$ **App execution aliases**.
   * Turn **OFF** "App Installer (python.exe)" and "App Installer (python3.exe)".
3. Point Niskava directly to your real Python binary:
   ```powershell
   # In PowerShell:
   $env:NISKAVA_PYTHON_BIN = "C:\Program Files\Python311\python.exe"
   ```
   Or set `NISKAVA_PYTHON_BIN=C:\Program Files\Python311\python.exe` in `~/.niskava/.env`.

---

### Issue 3: `ModuleNotFoundError: No module named 'engine'` or Missing Packages

**Symptoms:**
When running an investigation, the terminal outputs:
```text
ModuleNotFoundError: No module named 'engine'
# or
ModuleNotFoundError: No module named 'numpy' / 'pandas' / 'fpdf2'
```

**Root Cause:**
* When using NPX/NPM, the local Python virtualenv did not complete its initial synchronization.
* When running from source, dependencies were not installed into the active Python environment.

**Resolution:**
1. **If using NPX / Global NPM:**
   The NPM launcher automatically synchronizes the engine to `~/.niskava/engine`. If interrupted, manually reinstall requirements:
   ```bash
   pip install -r ~/.niskava/engine/requirements.txt
   ```
2. **If running from source:**
   Ensure you activate the virtual environment before running:
   ```bash
   source backend/engine/.venv/bin/activate
   pip install -r backend/engine/requirements.txt
   ```
3. Run `niskava doctor` to verify that all 5 critical packages (`numpy`, `pandas`, `networkx`, `trafilatura`, `fpdf2`) show `[PASS]`.

---

### Issue 4: AI Provider Connection / API Key Missing

**Symptoms:**
```text
### ⚠️ Konfigurasi Gemini API Key Tidak Ditemukan
GEMINI_API_KEY tidak ditemukan di environment atau konfigurasi.
```

**Resolution:**
1. **For Google Gemini:**
   Get a free Gemini API key from [Google AI Studio](https://aistudio.google.com/). Add it to `~/.niskava/.env`:
   ```ini
   GEMINI_API_KEY=AIzaSy...
   AI_PROVIDER=gemini
   ```
2. **For OpenAI / OpenRouter:**
   ```ini
   OPENAI_API_KEY=sk-...
   AI_PROVIDER=openai
   OPENAI_MODEL=gpt-4o-mini
   ```
3. **For Local Offline LLM (Ollama):**
   ```ini
   AI_PROVIDER=ollama
   OLLAMA_BASE_URL=http://localhost:11434
   OLLAMA_MODEL=deepseek-r1:8b
   ```
   *Make sure Ollama is running (`ollama serve`) and the model is pulled (`ollama pull deepseek-r1:8b`).*

---

### Issue 5: Inference Timeout During Deep Analysis

**Symptoms:**
```text
Koneksi timeout setelah 25 detik ke AI provider
# or
Request timed out waiting for AI response
```

**Resolution:**
Deep multi-tool financial reasoning requires sufficient token generation time, especially for local Ollama models on CPU.
* In the REPL, adjust the timeout immediately:
  ```text
  /timeout balanced   # 60s (Recommended baseline)
  /timeout deep       # 120s (For complex multi-tool analysis)
  /timeout local      # 180s (For local CPU inference)
  ```
* Or set `NISKAVA_LLM_TIMEOUT=60.00` in `~/.niskava/.env`.

---

### Issue 6: Port 20128 Already in Use (`listen tcp :20128: bind: address already in use`)

**Symptoms:**
When launching `niskava serve`, the server fails to start because port 20128 is occupied.

**Resolution:**
1. Specify an alternative port:
   ```bash
   niskava serve --port 20130 --open
   ```
2. Or set `NISKAVA_PORT=20130` in `~/.niskava/.env`.
3. To terminate the lingering process on port 20128:
   * Linux/macOS: `lsof -i :20128 | awk 'NR>1 {print $2}' | xargs kill -9`
   * Windows: `netstat -ano | findstr :20128` then `taskkill /PID <PID> /F`

---

### Issue 7: SQLite Database Locked (`sqlite3.OperationalError: database is locked`)

**Symptoms:**
```text
sqlite3.OperationalError: database is locked
```

**Resolution:**
Niskava uses SQLite with Write-Ahead Logging (WAL). A database lock occurs if an earlier process terminated abruptly while holding an uncommitted transaction.
1. Ensure no other instance of `niskava` is running in the background:
   ```bash
   killall niskava
   ```
2. Check for stale WAL files in `~/.niskava/`:
   ```bash
   # Run SQLite checkpoint to flush WAL:
   sqlite3 ~/.niskava/niskava.db "PRAGMA wal_checkpoint(TRUNCATE);"
   ```

---

### Issue 8: Telegram Bot Ignores Messages

**Symptoms:**
The Telegram bot connects and runs, but does not reply when you send messages in Telegram.

**Root Cause:**
For security, Niskava enforces a **User Whitelist**. If your Telegram username or user ID is not in the whitelist, the bot quietly ignores requests to prevent unauthorized API credit consumption.

**Resolution:**
1. Add your Telegram username (without `@`) to `~/.niskava/.env`:
   ```ini
   NISKAVA_TELEGRAM_ALLOWED_USERS=YourTelegramUsername
   ```
2. Or adjust the whitelist directly from the Web Workspace settings (`http://localhost:20128`).

---

## 3. Frequently Asked Questions (FAQ)

### Q1: Is my financial research data sent to the cloud?
**No.** Niskava strictly follows **Law 4 (Local-First Data Sovereignty)**. All investigation records, chat histories, memory graphs, and cached candlestick data are saved locally on your computer at `~/.niskava/niskava.db`. There is zero centralized database or cloud analytics telemetry.

### Q2: Why doesn't Niskava provide automated buy or sell order buttons?
**Law 3 and securities regulations strictly prohibit automated order execution.** Niskava is a pure read-only market intelligence and empirical verification platform, not an execution broker. It helps you analyze evidence and verify facts so that you can make informed, independent investment decisions.

### Q3: How does Niskava protect my Sectors API credit allocation?
**Law 5 (Credit Conservation)** ensures that zero redundant API calls are made. Candlestick data for past trading sessions ($T < \text{today}$) is permanent and never expires (`expires_at = NULL`), so repeat analyses on the same stock consume zero API credits. Fundamental reports are cached locally for 24 hours.

### Q4: Can I use Niskava with free local LLMs without paying for API keys?
**Yes.** You can install [Ollama](https://ollama.com/) locally and use models such as `deepseek-r1:8b` or `qwen2.5:7b`. You only need a free Sectors Financial API key from [sectors.app](https://sectors.app/) for the exchange data.

### Q5: How do I completely reset Niskava to factory defaults?
To clear all local settings, database records, and caches:
```bash
# Backup first if desired, then remove the .niskava folder:
rm -rf ~/.niskava
```
Running `niskava setup` will recreate a fresh directory and database.
