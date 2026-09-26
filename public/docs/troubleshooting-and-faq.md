# Troubleshooting & Frequently Asked Questions (FAQ)

This guide provides diagnostic procedures, common error resolutions, operational best practices, and answers to frequently asked questions.

---

## 1. Diagnostic Procedures & Error Resolutions

### Issue 1: AI Provider Connection Error
**Symptoms:**
When initiating an investigation or executing a prompt in the REPL, the interface reports:
```text
Error: unable to connect to AI provider (gemini / ollama / openrouter)
```

**Diagnostic Steps & Resolutions:**
1. **Verify API Key:** Ensure your API key is correctly defined in `~/.niskava/config.yaml` or exported in your environment:
   ```bash
   echo $GEMINI_API_KEY
   ```
2. **Interactive Setup:** Re-run the interactive setup wizard to validate keys:
   ```bash
   ./niskava setup
   ```
3. **Local LLM Endpoint (Ollama / vLLM):** If using Ollama, ensure the service is running and listening:
   ```bash
   curl http://localhost:11434/api/tags
   ```
   Confirm that the model specified in `config.yaml` (e.g. `llama3.1:latest`) has been pulled locally:
   ```bash
   ollama pull llama3.1
   ```
4. **Deterministic Fallback:** You can always run investigations in offline mode without invoking an AI provider:
   ```bash
   ./niskava investigate ANTM --offline
   ```

---

### Issue 1B: LLM Inference Timeout During Deep Multi-Tool Analysis
**Symptoms:**
During deep ticker analysis, peer comparisons, or multi-tool reasoning, the agent reports:
```text
Koneksi timeout setelah 25 detik ke http://localhost:20128/v1/chat/completions
# atau
Request timed out waiting for AI response
```

**Cause:**
Complex financial investigations require the LLM to inspect multiple quantitative metrics and news releases. If running a local model (Ollama on CPU) or a deep reasoning model, 25 seconds may not be enough for the model to synthesize observations.

**Resolutions:**
1. **Switch Timeout Profile in REPL:**
   Adjust the active timeout immediately with the `/timeout` slash command:
   ```text
   /timeout balanced   # 60s (Recommended baseline)
   /timeout deep       # 120s (For complex analysis)
   /timeout local      # 180s (For CPU/Ollama inference)
   /timeout 90         # Custom value in seconds (10 - 300)
   ```
2. **Configure in Setup Wizard:**
   Run `./bin/niskava setup` and select **Step 5: AI Inference Timeout Profile**.
3. **Adjust in Web Workspace Canvas:**
   Open `http://localhost:20128`, click the **Settings** icon (top right), and move the **Inference Timeout** slider to your desired value. Click **Save Configuration**.
4. **Environment Variable Override:**
   Add `NISKAVA_LLM_TIMEOUT=60.00` to your `.env` file or export it in your shell environment.

---

### Issue 2: SQLite Database Locked (`database is locked`)
**Symptoms:**
The terminal or engine emits an error message:
```text
sqlite3.OperationalError: database is locked
```

**Cause:**
Niskava uses SQLite with Write-Ahead Logging (`PRAGMA journal_mode = WAL;`) for concurrent read/write access. A database lock occurs if an earlier process terminated unexpectedly while holding an exclusive write transaction.

**Resolutions:**
1. **Terminate Orphaned Processes:** Ensure no previous `niskava` background processes are still running.
   - **Linux / macOS:**
     ```bash
     pkill -f niskava
     ```
   - **Windows (PowerShell):**
     ```powershell
     Get-Process -Name niskava -ErrorAction SilentlyContinue | Stop-Process -Force
     ```
2. **Check WAL Journal Files:** Inspect your database directory (`~/.niskava/`). If temporary lock files (`niskava.db-shm` or `niskava.db-wal`) persist after all processes have exited, run an integrity check:
   ```bash
   sqlite3 ~/.niskava/niskava.db "PRAGMA integrity_check;"
   sqlite3 ~/.niskava/niskava.db "PRAGMA wal_checkpoint(TRUNCATE);"
   ```

---

### Issue 3: Python Virtual Environment or Quantitative Dependencies Missing
**Symptoms:**
Go Core reports:
```text
Error: quantitative dependencies missing or python executable not found
```

**Resolutions:**
1. **Run Auto-Bootstrap via Setup Wizard (Recommended):**
   ```bash
   ./bin/niskava setup
   ```
   When prompted, choose `Y` to allow Niskava to automatically configure `.venv` and install `requirements.txt`.
2. **Manual Virtual Environment Setup:**
   Ensure the Python virtual environment exists in the root directory and dependencies are installed:
   - **Linux / macOS:**
     ```bash
     python3 -m venv .venv
     .venv/bin/pip install --upgrade pip
     .venv/bin/pip install -r backend/engine/requirements.txt
     ```
   - **Windows (PowerShell):**
     ```powershell
     python -m venv .venv
     .\.venv\Scripts\pip.exe install --upgrade pip
     .\.venv\Scripts\pip.exe install -r backend\engine\requirements.txt
     ```
3. If using a custom Python installation, specify the binary path explicitly in `~/.niskava/config.yaml`:
   ```yaml
   engine:
     python_bin: "/usr/bin/python3"
     engine_path: "./backend/engine"
   ```
   Or set the environment variable:
   ```bash
   export NISKAVA_PYTHON_BIN="/usr/bin/python3"
   ```

---

### Issue 4: Windows PowerShell Script Execution Blocked
**Symptoms:**
Running `.\install.ps1` produces the following error:
```text
File install.ps1 cannot be loaded because running scripts is disabled on this system.
```

**Resolution:**
PowerShell by default blocks script execution. Run the following command in your current PowerShell session:
```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
```
Then re-run `.\install.ps1`.

---

### Issue 5: Python Not Found on Windows (`'python' is not recognized`)
**Symptoms:**
Command Prompt or PowerShell reports:
```text
'python' is not recognized as an internal or external command
```

**Resolution:**
Re-install Python 3.11+ from [python.org](https://www.python.org/downloads/) and make sure to check the checkbox **"Add python.exe to PATH"** on the first installation screen. If already installed, add `C:\Users\<Username>\AppData\Local\Programs\Python\Python312` and its `Scripts` directory to your system Environment Variables.

---

### Issue 4: Sectors API Rate Limits or Credit Depletion (`429 Too Many Requests`)
**Symptoms:**
Requests to the Sectors Financial API fail with HTTP status code 429.

**Cause & Architectural Safeguards:**
Under **Law 5 (Credit Budget Discipline)**, Niskava actively protects your Sectors API credit allocation. Historical daily candlestick data ($T < \text{today}$) is permanently cached in local SQLite storage (`expires_at = NULL`), meaning repeated queries for historical data incur zero credit cost.

**Resolutions:**
1. **Verify Local Cache:** Check whether data for the target ticker already exists in local storage:
   ```bash
   sqlite3 ~/.niskava/niskava.db "SELECT cache_key, endpoint, expires_at FROM sectors_cache;"
   ```
2. **Use Offline Mock Mode:** When testing or developing, activate mock fixtures to bypass the remote API entirely:
   ```bash
   export MOCK_SECTORS=1
   ./niskava investigate ANTM --offline
   ```

---

### Issue 5: Port Binding Conflict on Web Workspace Server
**Symptoms:**
Starting the web workspace yields:
```text
Error: listen tcp 127.0.0.1:20128: bind: address already in use
```

**Resolutions:**
1. Specify an alternate port using the `--port` flag:
   ```bash
   ./niskava serve --port 20130 --open
   ```
2. Alternatively, identify and terminate the process holding the port:
   - **Linux / macOS:**
     ```bash
     lsof -i :20128
     kill -9 <PID>
     ```
   - **Windows:**
     ```powershell
     netstat -ano | findstr :20128
     Stop-Process -Id <PID> -Force
     ```

---

## 2. Operational Best Practices

### Local Database Backup & Restore
All investigation sessions, findings, and memory graphs are stored in a single SQLite database file. To back up your research data:

```bash
# Create a hot backup
sqlite3 ~/.niskava/niskava.db ".backup ~/.niskava/niskava_backup.db"
```

To restore from a backup:
```bash
cp ~/.niskava/niskava_backup.db ~/.niskava/niskava.db
```

### Resetting Cache
To purge expired Sectors API cache entries while preserving permanent historical daily candlestick data:

```bash
sqlite3 ~/.niskava/niskava.db "DELETE FROM sectors_cache WHERE expires_at IS NOT NULL AND expires_at < datetime('now');"
```

---

## 3. Frequently Asked Questions (FAQ)

### Does Niskava Agent provide BUY, SELL, or HOLD stock recommendations?
**No.** Under **Law 2 (Strict Financial Non-Advisory Boundary)** and Sectors Hackathon Rule 12, Niskava Agent strictly refrains from providing investment advice, price targets, or trade recommendations. All findings are presented as an objective, verified intelligence audit trail categorized into `SUPPORTED`, `UNCERTAIN`, or `CONTRADICTED` evidence.

---

### Can Niskava Agent execute trades directly through my brokerage account?
**No.** Under **Law 3 (Prohibition of Automated Trade Execution)**, Niskava is strictly a read-only market intelligence and research tool. The codebase contains no broker connection libraries, order routing APIs, or trade execution capabilities.

---

### Is my research data sent to external cloud servers?
**No.** Under **Law 4 (Local-First Data Sovereignty)**, all chat transcripts, investigation logs, evidence graphs, and API caches reside locally on your machine in `~/.niskava/niskava.db`. If you use a local AI provider like Ollama, zero data leaves your local network. When using cloud AI providers (e.g., Google Gemini), only sanitized statistical summaries and analytical context are sent for reasoning; raw database files remain strictly local.

---

### Which stock markets and instruments are supported?
Niskava Agent is currently optimized for equities listed on the **Indonesia Stock Exchange (IDX / Bursa Efek Indonesia)**. It supports all 4-letter and 5-letter IDX tickers (e.g., `ANTM`, `BBCA`, `BBRI`, `GOTO`, `ADRO`). It also calculates broader sector divergence against official IDX sector indices (e.g., IDX Finance, IDX Basic Materials, IDX Energy).

---

### How does Niskava ensure that news is the actual catalyst of a price move?
Niskava enforces **temporal causality verification**. In Stage 6 of the investigation pipeline, the engine cross-references the exact publication timestamp of an article or regulatory filing against the timestamp of the trading volume surge:
- If the news was published before the volume surge $\to$ classified as `LIKELY_CATALYST`.
- If the volume surge occurred before any news announcement $\to$ classified as `PRECEDED_ANNOUNCEMENT`, highlighting potential information leakage.
- If no news exists within the window $\to$ classified as `UNEXPLAINED_BY_NEWS`.

---

### Can I run Niskava entirely offline without internet access?
**Yes.** By setting `export MOCK_SECTORS=1` and using a locally hosted LLM via Ollama (or running `--offline` mode), Niskava can execute investigations and test suites completely disconnected from the internet.
