# 📚 Niskava Agent — Public Documentation & User Guide Hub

Welcome to the official documentation hub for **Niskava Agent** — an autonomous financial OSINT (*Open Source Intelligence*) and market intelligence orchestration platform for the **Indonesia Stock Exchange (IDX)**.

This documentation suite provides comprehensive installation instructions, system architecture overviews, CLI & TUI user guides, and troubleshooting manuals for financial analysts, equity researchers, and developers.

---

## 🗂️ Documentation Sitemap

| Documentation File | Description & Core Content |
|---|---|
| 📖 **[Installation Guide](installation.md)** | System prerequisites (Go 1.22+, Python 3.11+), virtual environment creation, API Key setup (`SECTORS_API_KEY`, `GEMINI_API_KEY`), interactive CLI wizard, and standalone binary compilation. |
| 🎮 **[User Guide](user-guide.md)** | Complete interaction guide: **Terminal UI (TUI) HUD Launcher**, **Interactive REPL**, prompt history navigation (Up/Down arrows), slash commands (`/investigate`, `/screen`, etc.), headless CLI mode, **Web Canvas Workspace**, and **Telegram Bot**. |
| 🏛️ **[Features & System Architecture](features-and-architecture.md)** | In-depth breakdown of the *Tripartite Hybrid Stack* (Go + Python + React), **The 6 Invariant Laws**, the **7-Stage Investigation Pipeline**, *NumPy Anomaly Engine*, and Evidence Verification Taxonomy. |
| ❓ **[Troubleshooting & FAQ](troubleshooting-and-faq.md)** | Resolving common errors, Sectors API v2 credit budget conservation, offline deterministic mode (`--offline`), SQLite WAL database locks, and frequently asked questions. |

---

## 🚀 Quick Start (3-Step Setup)

Want to try Niskava Agent immediately?

### 1. Clone Repository & Setup Environment
```bash
git clone https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git
cd Niskava-Agents
```

### 2. Run Interactive Setup Wizard
```bash
go run ./cmd/niskava setup
```
> *The interactive wizard will guide you through setting up API keys, creating a Python virtual environment (`venv`), and verifying all dependencies automatically.*

### 3. Launch Interactive Terminal UI (TUI)
```bash
go run ./cmd/niskava
```
Or compile into a single executable binary:
```bash
go build -o niskava.exe ./cmd/niskava
.\niskava.exe
```

---

## ⚖️ Strict Financial Non-Advisory Disclaimer (Law 2)

> **⚠️ IMPORTANT DISCLAIMER:**  
> Niskava Agent is an investigative market intelligence platform and is **NOT a licensed investment advisor**. Niskava **NEVER** issues direct BUY/SELL recommendations, price targets, or personalized financial advice. All findings are strictly objective, evidence-based OSINT findings classified within a 3-tier verification taxonomy (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).

---

## 📬 Support & Repository

- **GitHub Repository**: [Sectors-Hackathon-2026/Niskava-Agents](https://github.com/Sectors-Hacthon-2026/Niskava-Agents)
- **Competition Target**: [Sectors Hackathon Indonesia 2026 — Track 1: AI Agents & Assistants](https://hackathon.sectors.app/)
