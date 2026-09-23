# ❓ Niskava Agent — Troubleshooting & FAQ

This document provides troubleshooting guidance for common issues, error messages, and frequently asked questions when operating Niskava Agent.

---

## 🛠️ 1. Troubleshooting Common Issues

### ❌ Problem 1: `Unable to Connect to AI Provider`
**Symptom:**
When launching an investigation or entering prompt in REPL, an error displays `Unable to Connect to AI Provider`.

**Solution:**
1. Ensure your `GEMINI_API_KEY` (or chosen AI provider key) is correctly configured in `~/.niskava/config.yaml` or set via the interactive setup wizard (`niskava setup`).
2. If using a local LLM gateway (**Ollama** or **vLLM**), verify that the service is running at its local endpoint (e.g., `http://localhost:11434` for Ollama).
3. You can also run in deterministic offline mode without an LLM:
   ```bash
   niskava investigate ANTM --offline
   ```

---

### ❌ Problem 2: `SQLite Database Locked` / `database is locked`
**Symptom:**
An error `sqlite3.OperationalError: database is locked` occurs when accessing `~/.niskava/niskava.db`.

**Solution:**
1. Niskava Agent operates using **SQLite WAL (Write-Ahead Logging)** mode to support concurrent reads and writes.
2. Ensure no orphaned or stuck `niskava` processes are running in the background.
3. On Windows PowerShell, force-terminate stuck instances:
   ```powershell
   Get-Process -Name niskava | Stop-Process -Force
   ```
4. On Linux / macOS:
   ```bash
   pkill -f niskava
   ```

---

### ❌ Problem 3: `Python Virtual Environment Not Found`
**Symptom:**
Go Core CLI emits a log stating `Python executable not found`.

**Solution:**
1. Verify that the Python virtual environment in `backend/engine/venv` is created and required dependencies are installed:
   ```bash
   cd backend/engine
   python -m venv venv
   # Windows:
   .\venv\Scripts\pip.exe install -r requirements.txt
   # Linux/macOS:
   ./venv/bin/pip install -r requirements.txt
   ```
2. Alternatively, run the interactive setup wizard to fix paths automatically:
   ```bash
   niskava setup
   ```

---

### ❌ Problem 4: Sectors API Credit Exceeded (`429 Too Many Requests`)
**Symptom:**
Sectors API v2 returns HTTP 429 or credit limit exhausted error.

**Solution:**
1. Niskava Agent enforces **Law 5 (Credit Budget Discipline)** by caching all historical OHLCV candlestick data permanently in local SQLite (`expires_at = NULL`).
2. Re-investigating the same ticker will fetch data from local cache without consuming Sectors API credits.
3. Check local cache statistics via CLI:
   ```bash
   niskava status --cache
   ```

---

## ❓ 2. Frequently Asked Questions (FAQ)

### Q1: Does Niskava Agent provide BUY or SELL stock recommendations?
**Answer:**  
**NO.** In accordance with **Architectural Law 2** and **Sectors Hackathon Rule 12**, Niskava Agent **NEVER** issues investment advice, BUY/SELL signals, or target prices. Niskava is an objective market intelligence platform that gathers and verifies quantitative facts and OSINT evidence.

---

### Q2: Can Niskava Agent automatically execute stock trades on broker accounts?
**Answer:**  
**NO.** In accordance with **Architectural Law 3** and **Sectors Hackathon Rule 06**, Niskava Agent is strictly read-only. There are zero brokerage execution APIs or order routing logic.

---

### Q3: How do I run the Web Canvas Workspace on a custom port?
**Answer:**  
Use the `--port` flag:
```bash
niskava serve --port 9090 --open
```

---

### Q4: Where is my investigation history and chat data stored?
**Answer:**  
In accordance with **Architectural Law 4 (Local-First Data Sovereignty)**, all data is stored locally on your machine at:
- **Windows**: `C:\Users\<Username>\.niskava\niskava.db`
- **Linux/macOS**: `~/.niskava/niskava.db`

---

### Q5: How do I recompile the binary after pulling updates?
**Answer:**  
Execute Go build:
```bash
git pull origin dev
go build -o niskava.exe ./cmd/niskava
```
