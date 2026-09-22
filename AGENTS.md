# AGENTS.md — Autonomous Developer & Agent Operational Constitution

> **Single Source of Truth (SSoT) for AI Coding Agents, Subagents, and Human Developers working on Niskava Agent.**

---

## 1. System Identity & Mission

**Niskava Agent** is an autonomous financial OSINT and market intelligence orchestration platform designed specifically for the **Indonesia Stock Exchange (IDX)**. It bridges the critical gap between quantitative market facts (provided by Sectors Financial API v2) and external qualitative intelligence (company disclosures, regulatory filings, news, and market signals).

* **Core Motto:** *"Don't just answer questions. Investigate them."*
* **Competition Target:** [Sectors Hackathon Indonesia 2026](https://hackathon.sectors.app/) — **Track 1: AI Agents & Assistants**.
* **Primary Persona:** Professional equity analysts, financial journalists, and serious retail swing traders who require rigorous, verifiable evidence rather than speculative commentary.
* **Dual Interaction Modes:**
  1. **Interactive Conversational AI Assistant (Hermes-Style REPL & Web Canvas):** Prompt-driven natural language financial research terminal (`niskava` interactive REPL) and web canvas (`niskava serve`). Powered by an autonomous ReAct loop calling deterministic tools with streaming Glamour markdown rendering.
  2. **Autonomous Headless Pipeline:** Single-command structured audit trail (`niskava investigate <TICKER> --days 30`).

---

## 2. Non-Negotiable Architectural Invariants (The 6 Laws)

Every AI agent modifying, generating, or refactoring code in this repository MUST strictly abide by the following six architectural laws. Violating these invariants breaks project compliance and invalidates competition eligibility:

### Law 1: Deterministic Before Generative (P2 / 02-deterministic-quant-pre-llm)
* **Rule:** Never allow an LLM to calculate statistics, volume moving averages, Z-scores, abnormal returns, or sector divergence.
* **Mechanism:** All quantitative indicators MUST be computed deterministically via Python NumPy/Pandas before calling any LLM. In conversational prompt mode, the agent MUST call the `compute_quant_anomalies` tool rather than estimating numbers or performing mental math.
* **Rationale:** Eliminates numerical hallucination entirely and preserves precious token budget.

### Law 2: Strict Financial Non-Advisory Boundary (P3 / 05-strict-financial-non-advisory-boundary / Hackathon Rule 12)
* **Rule:** Niskava is an investigative intelligence platform, NOT an investment advisor. The system MUST NEVER output direct BUY/SELL recommendations, price targets, or personalized financial advice.
* **Mechanism:** All findings must be strictly classified into the Three-Tier Verification Taxonomy:
  - `SUPPORTED`: Proven by official Sectors quantitative data or formal IDXnet disclosures.
  - `UNCERTAIN`: Correlation observed but direct causality unverified (e.g. social media rumors).
  - `CONTRADICTED`: Market claims refuted by official corporate disclosures or financial statements.
* **Disclaimer:** Every output surface (CLI and Web Workspace) MUST display the standard non-advisory disclaimer.

### Law 3: Prohibition of Automated Trade Execution (Hackathon Rule 06)
* **Rule:** Products may analyze, screen, score, alert, and correlate evidence, but may NEVER place, execute, or automate buy or sell orders on brokerage accounts.
* **Mechanism:** The codebase has ZERO trading execution APIs, broker connection libraries, or order-routing dependencies. It is strictly read-only market intelligence.

### Law 4: Local-First Data Sovereignty (P5 / 04-local-first-sqlite-storage)
* **Rule:** No centralized cloud database. All user investigation sessions, findings, evidence graphs, chat histories (`chat_messages`), and local caches MUST reside locally in `~/.niskava/niskava.db` using SQLite with Write-Ahead Logging (`PRAGMA journal_mode = WAL;`).
* **Mechanism:** Go Core uses pure-Go zero-CGO SQLite (`modernc.org/sqlite`); Python uses standard `sqlite3`.

### Law 5: Credit Budget Discipline & Local Caching (P6 / 03-sectors-v2-and-credit-conservation)
* **Rule:** Strictly protect the 1,000 Sectors API credit grant. Zero redundant calls.
* **Mechanism:** Always check `sectors_cache` before firing any HTTP request to `https://api.sectors.app/v2`. Historical daily OHLCV candlestick data ($T < \text{today}$) is permanently cached (`expires_at = NULL`), incurring 0 credit cost on repeat queries.
* **Offline Testing:** Always support `MOCK_SECTORS=1` with static fixture JSONs for CI/CD and unit tests.

### Law 6: Local Conversational Graph Memory Engine (06-local-conversational-graph-memory)
* **Rule:** Agents must maintain context across multi-day sessions without heavy external graph databases (e.g. Neo4j) or paid cloud vector services.
* **Mechanism:** Persist associative graph relations in local SQLite (`memory_nodes`, `memory_edges`, and `chat_messages`), load into Python `NetworkX.DiGraph` in-memory, and perform Ego-Graph traversal ($k \le 2$ hops) with exponential recency decay ($e^{-\lambda \Delta t}$).

---

## 3. Hybrid Architecture & Component Responsibilities

The codebase follows the Tripartite Hybrid Stack (01-hybrid-stack-go-python-react) powered by a **4-Layer Agentic Hierarchy** in the Python Engine (08-modular-skills-and-mcp-architecture):

```
┌─────────────────────────────────────────────────────────────┐
│                          GO CORE                            │
│  - Gateway, Single Executable CLI, REST/SSE Server          │
│  - Interactive TUI & REPL (charmbracelet/bubbletea+glamour) │
│  - Interactive Setup Wizard (`niskava setup`)               │
│  - Pure-Go SQLite Persistence (modernc.org/sqlite)          │
│  - Static Web UI Bundler (//go:embed)                       │
└──────────────────────────────┬──────────────────────────────┘
                               │ Local IPC (JSON Lines / STDIN/STDOUT)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    PYTHON AGENT ENGINE                      │
│                                                             │
│  [Layer 4: Cognitive ReAct Loop & Memory Engine]            │
│  - Autonomous ReAct Agent Loop (Prompt-driven Hermes-style) │
│  - Local Graph Memory Engine (NetworkX + SQLite)            │
│  - Evidence Correlation & Temporal Causality Reasoning      │
│                              │                              │
│  [Layer 3: Modular Skills Registry (Domain SOPs)]           │
│  - market-anomaly-recon, event-causality-audit              │
│  - insider-bandarmology-forensic, financial-health-stress   │
│                              │                              │
│  [Layer 2: Deterministic Compute Gate (NumPy Firewall)]     │
│  - Anomaly Math: MA20, Z-Scores (Vz, Fz), Abnormal Returns  │
│                              │                              │
│  [Layer 1: MCP & OSINT Primitives]                          │
│  - Sectors Financial API v2 Client & MCP Server Adapter     │
│  - Dual-Engine Targeted OSINT (Unified News + Google RSS)   │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│               REACT SPA WEB WORKSPACE (Vite)                │
│  - Visual Cyber-OSINT / Bloomberg Terminal Aesthetic        │
│  - Conversational AI Assistant Canvas with Live SSE Stream  │
│  - TradingView / Recharts Candlestick Anomaly Markers       │
│  - Interactive Evidence Matrix & Timeline Graph             │
└─────────────────────────────────────────────────────────────┘
```

### Decoupled Repositories & Configurable Monorepo Pattern
1. **Repository Decoupling:**
   * **Docs Repository (`niskava-docs`):** Single Source of Truth (SSoT) for specifications, mathematical proofs, skills, and architecture decisions.
   * **Codebase Repository (`niskava-codebase`):** Target implementation repository structured as a polyglot monorepo (`cmd/`, `internal/`, `engine/`, `web/`).
2. **Dynamic Path Resolution:** Paths must NEVER be hardcoded. Go Core resolves Python runner, web dist, and database paths via:
   `CLI Flags > Environment Variables (NISKAVA_*) > Config File (~/.niskava/config.yaml) > Dynamic Fallback`.

---

## 4. The 7-Stage Investigation Pipeline Protocol

Whenever writing agent workflow logic, strictly adhere to the sequential 7-Stage SOP (documented in `docs/30-agent/01-investigation-pipeline.md`):

1. **Stage 1: INITIATION** — Receive target ticker (e.g. `ANTM`), create investigation session in SQLite with status `PENDING`, transition to `RUNNING`.
2. **Stage 2: SECTORS_BASELINE** — Check `sectors_cache`, pull 30–90 days daily OHLCV from `/v2/daily/{symbol}/`, fetch company report with sections from `/v2/company/report/{symbol}/`, retrieve Net Foreign Inflow from `/v2/foreign-flow/{symbol}/`, corporate actions from `/v2/corporate-actions/{symbol}/`, and suspension notices from `/v2/suspensions/`.
3. **Stage 3: QUANT_ANOMALY** — NumPy deterministic evaluation: calculate MA20 volume, Volume Z-Score ($V_z$), Abnormal Return ($R_t$), Sector Divergence ($D_t$), and Foreign Flow Inflow Z-Score ($F_z$). If $V_z \ge 2.5$, $|R_t| \ge 5\%$, or $|F_z| \ge 2.5$, trigger anomaly record. If not, route to fundamental baseline (Stage 4b).
4. **Stage 4: GAP_DETECTION** — Formulate targeted temporal investigation hypothesis around $T_{\text{anomaly}} \pm 2\text{ days}$.
5. **Stage 5: OSINT_HARVEST** — Execute parallel Dual-Engine collection (07-resilient-dual-engine-osint-architecture): pull curated news from Sectors v2 Unified News API (`/v2/news/`) and run unblocked targeted boolean dorking on Google News RSS for syndicated IDX disclosures and financial media (Kontan, Bisnis, CNBC). Sanitize via `trafilatura`, isolate context in XML tags (`<evidence_context>`), and extract candidate evidence items.
6. **Stage 6: EVIDENCE_CORRELATION** — Assess temporal precedence:
   - News before volume surge $\to$ `LIKELY_CATALYST`
   - Volume surge before news release $\to$ `PRECEDED_ANNOUNCEMENT`
   - No explanatory news $\to$ `UNEXPLAINED_BY_NEWS`
   Assign verification status (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
7. **Stage 7: SYNTHESIS_AND_STREAMING** — Generate structured findings, timeline events, and narrative summary; persist to SQLite; stream progress events in JSON Lines via IPC to Go Core for real-time SSE broadcast.

---

## 5. Confidence Scoring & Rubric Invariant

Never assign arbitrary continuous confidence floats (such as 0.50). Always use the standardized discrete confidence rubric:

* `1.00`: **EXTRACTED** — Directly backed by official Sectors API records or official IDXnet regulatory disclosures.
* `0.95`: **Direct Structural Evidence** — Explicit timestamp correlation with official corporate press releases.
* `0.85`: **Strong Inference** — High temporal correlation with major mainstream financial press reporting.
* `0.75`: **Reasonable Inference** — Plausible catalyst from industry-wide trends with matching sector divergence.
* `0.65`: **Weak Inference** — Unverified market commentary, social media sentiment, or unconfirmed rumors.
* `0.55`: **Speculative** — Distant co-occurrence without temporal causality confirmation.

---

## 6. Coding Standards & Naming Conventions

* **Primary Code Language:** All source code (variable names, function signatures, structs, classes, docstrings, and commit messages) MUST be written in idiomatic English.
* **LLM Prompts & Extraction Schemas:** All system prompts, few-shot examples, and Pydantic/JSON schemas MUST be in English for maximum inference precision.
* **Documentation (`docs/`):** Written in clear Bahasa Indonesia with standardized English technical terms (e.g. *Z-score, daily candlestick, Local-First SQLite*), ensuring full alignment with Indonesian team members.
* **Go Code:** Follow official Go standards (`gofmt`, explicit error handling `if err != nil`, zero CGO, `modernc.org/sqlite`).
* **Python Code:** Python 3.11+, strict type annotations (`typing`), Pydantic models for data interchange, deterministic execution.

---

## 7. Secrets Custody & Pre-Submission Quality Gates

* **Zero Hardcoded Secrets:** NEVER commit API keys (`SECTORS_API_KEY`, `GEMINI_API_KEY`) into git. Use environment variables or local `~/.niskava/config.yaml` (`0600` permissions).
* **Git Branch Protection Invariant (dev -> main):** Direct push to `main` is strictly prohibited. All feature development (`feat/*`), bugfixes (`fix/*`), and documentation (`docs/*`) MUST branch from `dev` and merge into `dev` first. `main` only accepts stabilized, verified merges from `dev` after all automated tests, builds, and quality gates pass.
* **Submission Freeze Rule:** The project freezes permanently upon submission or on **30 September 2026 at 23:59 WIB**. No commits or bug fixes are allowed after freeze under penalty of disqualification.
* **Public Repository Retention:** The repository MUST remain public for at least 90 days after winners are announced (through January 2027).

---

## 8. Agent Navigation Guide (Directory Mapping)

When asked to work on specific aspects of the system, navigate directly to these canonical references:

| System Domain | Technical Documentation | Target Codebase Location (Configurable) |
|---|---|---|
| **Hackathon Strategy & Rules** | `docs/10-product/05-hackathon-strategy.md` | `README.md`, `AGENTS.md` |
| **System Architecture & CLI** | `docs/20-architecture/01-system-overview.md` | `cmd/niskava/`, `clients/cli/` |
| **Conversational REPL & TUI** | `docs/10-product/03-product-scope-and-surfaces.md` | `clients/cli/tui/`, `backend/engine/agent/react_agent.py` |
| **Setup Wizard & Configuration** | `docs/20-architecture/01-system-overview.md` | `clients/cli/setup.go` |
| **SQLite Schema & Persistence** | `docs/20-architecture/02-database-schema.md` | `backend/core/db/`, `backend/engine/memory/` |
| **Sectors v2 API Integration** | `docs/20-architecture/03-sectors-v2-api.md` | `backend/engine/sectors/` |
| **Quantitative Anomaly Math** | `docs/20-architecture/05-anomaly-detection-math.md` | `backend/engine/quant/` |
| **OSINT Engine & Disclosures** | `docs/20-architecture/06-osint-engine.md` | `backend/engine/osint/` |
| **7-Stage Investigation Pipeline** | `docs/30-agent/01-investigation-pipeline.md` | `backend/engine/agent/pipeline.py` |
| **Conversational Graph Memory** | `docs/30-agent/05-conversational-memory-engine.md` | `backend/engine/memory/graph_memory.py` |
| **Web Workspace & Visual UI** | `docs/10-product/03-product-scope-and-surfaces.md` | `clients/web/src/` |
| **Security & Regulatory Compliance** | `docs/20-architecture/08-security-and-compliance.md` | `backend/core/security/` |
| **Git Workflow & Branch Protection** | `docs/20-architecture/09-git-workflow-and-branching-strategy.md` | Git topology (`dev` -> `main`) |
| **Modular Skills & MCP Registry** | `docs/30-agent/02-skills-catalog.md`, `docs/90-decisions/08-modular-skills-and-mcp-architecture.md` | `backend/engine/agent/tools.py` |
| **Architecture Decisions** | `docs/90-decisions/*.md` | Root and sub-packages |

