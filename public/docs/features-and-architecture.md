# 🏛️ Niskava Agent — Features & System Architecture

This document outlines the technical architecture, **The 6 Invariant Laws**, the **7-Stage Investigation Pipeline Protocol**, and the deterministic mathematical formulas powering Niskava Agent.

---

## 🧱 1. Tripartite Hybrid Stack Architecture

Niskava Agent is engineered as a 3-tier hybrid stack combining binary execution performance (Go Core), quantitative data processing (Python NumPy/Pandas), and interactive web UI rendering (React/Vite).

```text
┌─────────────────────────────────────────────────────────────┐
│                          GO CORE                            │
│  - Gateway & Single Binary Executable CLI (`niskava`)       │
│  - Interactive TUI Terminal (charmbracelet/bubbletea)       │
│  - REST Server & Server-Sent Events (SSE Real-Time Stream)  │
│  - Pure-Go Zero-CGO SQLite WAL (modernc.org/sqlite)         │
└──────────────────────────────┬──────────────────────────────┘
                               │ Local IPC (JSON Lines STDIN/STDOUT)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    PYTHON AGENT ENGINE                      │
│  - [Layer 4] ReAct Cognitive Engine & Local Graph Memory   │
│  - [Layer 3] Modular Skills Registry (Domain SOP Modules)   │
│  - [Layer 2] Deterministic Compute Gate (NumPy Anomaly Math)│
│  - [Layer 1] Sectors MCP Server Adapter & Dual OSINT Engine │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│               REACT SPA WEB WORKSPACE (Vite)                │
│  - Bloomberg Terminal / Cyber-OSINT Visual Aesthetic        │
│  - Candlestick Chart & Volume Surge Anomaly Markers         │
│  - Real-Time Thinking Stream via Server-Sent Events (SSE)   │
│  - Interactive Evidence Matrix & Timeline Causality Graph   │
└──────────────────────────────┴──────────────────────────────┘
```

---

## 📜 2. The 6 Architectural Invariant Laws

Every module in Niskava Agent strictly adheres to 6 non-negotiable architectural laws (as codified in `AGENTS.md`):

### Law 1: Deterministic Before Generative (P2 / 02-deterministic-quant-pre-llm)
LLMs are **STRICTLY PROHIBITED** from calculating statistics, moving averages, Z-scores, abnormal returns, or sector divergence. All quantitative metrics MUST be computed deterministically via **NumPy/Pandas** before prompting an LLM.

### Law 2: Strict Financial Non-Advisory Boundary (P3 / Hackathon Rule 12)
Niskava is an investigative market intelligence platform, **NOT an investment advisor**. The system **MUST NEVER** issue direct BUY/SELL recommendations, price targets, or personalized financial advice. All findings are categorized into a 3-Tier Verification Taxonomy (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).

### Law 3: Prohibition of Automated Trade Execution (Hackathon Rule 06)
Zero brokerage execution APIs or order routing logic. The platform is strictly *read-only market intelligence*.

### Law 4: Local-First Data Sovereignty (P5 / 04-local-first-sqlite-storage)
No centralized cloud database. All user investigation sessions, evidence graphs, and local caches reside locally at `~/.niskava/niskava.db` using SQLite with Write-Ahead Logging (`PRAGMA journal_mode = WAL;`).

### Law 5: Credit Budget Discipline & Local Caching (P6 / 03-sectors-v2)
Strictly protects the 1,000 Sectors API credit grant. Historical daily candlestick data ($T < \text{today}$) is permanently cached locally (`expires_at = NULL`), incurring 0 credit cost on repeat queries.

### Law 6: Local Conversational Graph Memory Engine (06-local-conversational-graph-memory)
Maintains associative context across multi-day research sessions without cloud vector databases. Memory nodes & edges are stored in local SQLite (`memory_nodes` & `memory_edges`), loaded into in-memory Python `NetworkX.DiGraph`, and queried via *Ego-Graph* traversal ($k \le 2$ hops) with exponential recency decay ($e^{-\lambda \Delta t}$).

---

## 🔄 3. The 7-Stage Investigation Pipeline Protocol

Every investigation follows a strict 7-stage domain protocol:

1. **Stage 1: INITIATION** — Accepts target ticker symbol (e.g., `ANTM`), creates SQLite session (`PENDING` $\to$ `RUNNING`).
2. **Stage 2: SECTORS_BASELINE** — Checks local cache, retrieves 30–90 days daily OHLCV from Sectors API v2 (`/v2/daily/{symbol}/`), financial reports, Net Foreign Flow, and corporate actions.
3. **Stage 3: QUANT_ANOMALY** — Deterministic NumPy evaluation:
   - 20-day Volume Moving Average ($MA_{20}$)
   - Volume Z-Score ($V_z = \frac{V_t - \mu_{20}}{\sigma_{20}}$)
   - Abnormal Return ($R_t$) & Sector Divergence ($D_t$)
   - Foreign Flow Inflow Z-Score ($F_z$)
4. **Stage 4: GAP_DETECTION** — Formulates temporal causality investigation hypotheses centered around anomaly dates ($T_{\text{anomaly}} \pm 2\text{ days}$).
5. **Stage 5: OSINT_HARVEST** — Dual-Engine OSINT gathering: pulls curated news from Sectors v2 Unified News API (`/v2/news/`) and executes targeted dorking on Google News RSS for official IDXnet filings and financial media (Kontan, Bisnis, CNBC Indonesia), sanitized via `trafilatura`.
6. **Stage 6: EVIDENCE_CORRELATION** — Evaluates temporal precedence:
   - News release precedes volume surge $\to$ `LIKELY_CATALYST`
   - Volume surge precedes news release $\to$ `PRECEDED_ANNOUNCEMENT`
   - No explanatory news $\to$ `UNEXPLAINED_BY_NEWS`
7. **Stage 7: SYNTHESIS_AND_STREAMING** — Synthesizes structured findings, event timelines, and narrative summary; persists to local SQLite; and streams progress events in JSON Lines via IPC to Go Core for real-time SSE broadcast.

---

## 🎯 4. Evidence Verification Taxonomy & Confidence Rubric

| Score | Verification Level | Evidence Description / Source |
|:---:|:---|:---|
| **`1.00`** | **EXTRACTED** | Directly backed by official Sectors API quantitative records or formal IDXnet regulatory filings. |
| **`0.95`** | **Direct Structural Evidence** | Explicit timestamp correlation with official corporate press releases or regulatory disclosures. |
| **`0.85`** | **Strong Inference** | High temporal correlation with major mainstream national financial media coverage (Kontan, Bisnis, CNBC). |
| **`0.75`** | **Reasonable Inference** | Plausible catalyst from industry trends matching sector divergence patterns. |
| **`0.65`** | **Weak Inference** | Unverified market commentary, social media sentiment, or unconfirmed forum rumors. |
| **`0.55`** | **Speculative** | Co-occurrence without verified temporal causality. |
