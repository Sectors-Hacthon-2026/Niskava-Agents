# Features & System Architecture

This document provides a technical specification of the Niskava Agent architecture, the Tripartite Hybrid Stack, the 4-Layer Cognitive Hierarchy, the 6 Invariant Laws, the 7-Stage Investigation Pipeline, and the deterministic mathematical formulas powering the anomaly detection engine.

---

## 1. Tripartite Hybrid Stack Architecture

Niskava Agent combines the low-latency systems capabilities of Go, the scientific computing and AI ecosystem of Python, and the responsive user experience of React:

```
┌─────────────────────────────────────────────────────────────────┐
│                            GO CORE                              │
│  - Gateway CLI & Daemon Process (`cmd/niskava`, `clients/cli`)  │
│  - Interactive Terminal HUD & REPL (charmbracelet/bubbletea)    │
│  - High-Throughput REST API & SSE Streaming Server (:20128)     │
│  - Embedded Static Web Workspace Distribution (//go:embed)      │
│  - Zero-CGO SQLite WAL Persistence (modernc.org/sqlite)         │
│  - IPC Process Supervisor (JSON-Lines over STDIN/STDOUT)        │
└────────────────────────────────┬────────────────────────────────┘
                                 │ Inter-Process Communication
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                      PYTHON AGENT ENGINE                        │
│                                                                 │
│  [Layer 4: ReAct Cognitive Orchestrator & Memory Engine]        │
│  - Autonomous ReAct Agent Loop (Reasoning + Action)             │
│  - Local Associative Graph Memory (NetworkX + SQLite)           │
│  - Temporal Precedence & Causality Inference Engine             │
│                                                                 │
│  [Layer 3: Modular Skills Registry (Domain SOP Modules)]        │
│  - Market Anomaly Reconnaissance (`market_anomaly_recon`)       │
│  - Event Causality Audit (`event_causality_audit`)              │
│  - Insider & Foreign Flow Forensics (`insider_bandarmology`)    │
│  - Financial Health Stress Testing (`financial_health_stress`)  │
│  - Commodity Divergence (`mining_commodity_divergence`)         │
│  - Peer Valuation Benchmark (`peer_valuation_benchmark`)       │
│                                                                 │
│  [Layer 2: Deterministic Compute Gate (NumPy Firewall)]         │
│  - Volume Z-Scores, Abnormal Returns, Sector Divergence         │
│  - Foreign Inflow Z-Scores, Altman Z-Score Ratios               │
│                                                                 │
│  [Layer 1: MCP & News Data Primitives]                         │
│  - Sectors Financial API v2 MCP Server Adapter                  │
│  - Dual-Engine Targeted News Harvest (Sectors News + Google RSS Dorks) │
│  - Content Extraction & HTML Sanitization via Trafilatura       │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                   REACT SPA WEB WORKSPACE                       │
│  - Terminal / Market Intelligence Design Language                       │
│  - Interactive Candlestick Charts & Anomaly Overlays            │
│  - Real-Time Thinking Stream via Server-Sent Events (SSE)       │
│  - Interactive Evidence Matrix & Causality Timeline Graph       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. The 6 Non-Negotiable Invariant Architectural Laws

Every subsystem within Niskava Agent is strictly governed by six foundational architectural invariants:

### Law 1: Deterministic Before Generative
- **Rule:** Large Language Models are strictly prohibited from calculating time-series statistics, moving averages, standard deviations, Z-scores, abnormal returns, or sector divergence.
- **Mechanism:** All numerical operations are computed deterministically via Python NumPy and Pandas before any LLM prompt is constructed. The LLM receives verified statistical summaries, never raw data arrays.
- **Rationale:** Completely eliminates numerical hallucinations and preserves token budget.

### Law 2: Strict Financial Non-Advisory Boundary
- **Rule:** Niskava is an investigative market intelligence platform, **NOT an investment advisor**. The system must never emit direct BUY/SELL recommendations, price targets, or portfolio allocation advice.
- **Mechanism:** All findings are strictly classified into a 3-tier verification taxonomy:
  - `SUPPORTED`: Validated by official Sectors API records or formal IDXnet regulatory disclosures.
  - `UNCERTAIN`: Plausible correlation observed, but direct causal proof is unverified.
  - `CONTRADICTED`: Market speculation refuted by corporate filings or audited statements.
- **Compliance:** Enforces OJK / IDX securities regulations and Sectors Hackathon Rule 12 across all interfaces.

### Law 3: Prohibition of Automated Trade Execution
- **Rule:** The system may analyze, screen, score, alert, and correlate evidence, but may never place, execute, or automate trade orders.
- **Mechanism:** The codebase contains zero trading execution APIs, broker connection libraries, or order-routing dependencies. It is strictly read-only market intelligence.

### Law 4: Local-First Data Sovereignty
- **Rule:** No centralized cloud database. All user investigation sessions, findings, evidence graphs, chat transcripts, and API caches must reside locally on the host machine.
- **Mechanism:** Go Core uses pure-Go zero-CGO SQLite (`modernc.org/sqlite`); Python uses standard library `sqlite3`. Database files are stored at `~/.niskava/niskava.db` with Write-Ahead Logging (`PRAGMA journal_mode = WAL;`) for concurrent read/write access.

### Law 5: Credit Budget Discipline & Local Caching
- **Rule:** Strictly protect the 1,000 Sectors API credit allocation. Zero redundant HTTP requests.
- **Mechanism:** All requests check the local `sectors_cache` table before issuing HTTP requests to `https://api.sectors.app/v2`. Historical daily candlestick data ($T < \text{today}$) is permanently cached (`expires_at = NULL`), incurring zero credit cost on repeat queries.
- **Offline Support:** Supports `MOCK_SECTORS=1` using static JSON fixtures for CI/CD pipelines and unit testing.

### Law 6: Local Conversational Graph Memory Engine
- **Rule:** Agents must maintain associative context across multi-turn sessions without relying on expensive cloud vector databases or heavy external graph services.
- **Mechanism:** Entities and relationships are persisted in SQLite (`memory_nodes` and `memory_edges`), loaded into an in-memory `networkx.DiGraph`, and queried via Ego-Graph traversal ($k \le 2$ hops) with exponential recency decay ($e^{-\lambda \Delta t}$).

---

## 3. The 7-Stage Investigation Pipeline Protocol

Whenever a ticker investigation is initiated (e.g., `niskava investigate ANTM --days 30`), Niskava executes a sequential 7-stage protocol:

```
[1. INITIATION]
       │
       ▼
[2. SECTORS_BASELINE]
       │
       ▼
[3. QUANT_ANOMALY] ──(No Anomaly)──▶ [Fundamental Screening]
       │ (Anomaly Detected)                       │
       ▼                                          │
[4. GAP_DETECTION]                                │
       │                                          │
       ▼                                          │
[5. NEWS_HARVEST]                                │
       │                                          │
       ▼                                          │
[6. EVIDENCE_CORRELATION] ◀───────────────────────┘
       │
       ▼
[7. SYNTHESIS_AND_STREAMING]
```

1. **Stage 1: INITIATION**
   - Receives target ticker symbol (e.g., `ANTM`).
   - Creates an investigation session record in SQLite (`status = 'PENDING'`).
   - Transitions state to `RUNNING` and initializes IPC stream.

2. **Stage 2: SECTORS_BASELINE**
   - Queries `sectors_cache` to retrieve or fetch:
     - 30 to 90 days of daily OHLCV candlesticks (`/v2/daily/{symbol}/`).
     - Company overview, management, and major shareholders (`/v2/company/report/{symbol}/`).
     - Daily Net Foreign Flow records (`/v2/foreign-flow/{symbol}/`).
     - Scheduled and historical corporate actions (`/v2/corporate-actions/{symbol}/`).
     - Trading suspension history (`/v2/suspensions/`).

3. **Stage 3: QUANT_ANOMALY**
   - Executes deterministic NumPy anomaly formulas across price series and foreign flow.
   - Evaluates Volume Z-Score ($V_z$), Abnormal Return ($R_t$), Sector Divergence ($D_t$), and Foreign Flow Z-Score ($F_z$).
   - If statistical anomalies are identified, flags the anomaly date $T_{\text{anomaly}}$ and transitions to Stage 4. If no anomalies exist, transitions directly to fundamental screening.

4. **Stage 4: GAP_DETECTION**
   - Formulates targeted temporal investigation hypotheses centered tightly around the anomaly window: $T_{\text{anomaly}} \pm 2\text{ days}$.
   - Generates structured search dork queries combining the company name, ticker, and exchange-specific disclosure terminology.

5. **Stage 5: NEWS_HARVEST**
   - Executes parallel Dual-Engine intelligence harvesting:
     - **Curated News:** Fetches categorized market news from Sectors v2 Unified News API (`/v2/news/`).
     - **Targeted Media Dorking:** Queries Google News RSS with boolean operators for major Indonesian financial media (Kontan, Bisnis Indonesia, CNBC Indonesia, Investor Daily, IDXnet disclosures).
   - Sanitizes and extracts article text using `trafilatura`.
   - Wraps content in strict boundary delimiters (`<evidence_context>`) to neutralize prompt injection attacks from untrusted external web pages.

6. **Stage 6: EVIDENCE_CORRELATION**
   - Performs temporal sequence verification:
     - If verified disclosure precedes volume spike $\to$ `LIKELY_CATALYST`.
     - If volume spike precedes corporate disclosure $\to$ `PRECEDED_ANNOUNCEMENT` (possible information leakage).
     - If volume spike occurs without public news $\to$ `UNEXPLAINED_BY_NEWS`.
   - Assigns verification tags (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`) and discrete confidence ratings.

7. **Stage 7: SYNTHESIS_AND_STREAMING**
   - Synthesizes findings, event timelines, and narrative conclusions.
   - Persists all results to local SQLite database tables (`investigations`, `findings`, `timeline_events`, `evidence_sources`).
   - Streams progress and structured findings via JSON-Lines over IPC to Go Core for real-time SSE delivery.

---

## 4. Deterministic Quantitative Anomaly Formulas

All quantitative indicators are calculated deterministically by the Python Engine before LLM activation.

### A. Volume Anomaly Z-Score ($V_z$)
Measures the statistical significance of trading volume on day $t$ relative to the 20-day trading volume baseline:

$$\mu_{20} = \frac{1}{20} \sum_{i=1}^{20} V_{t-i}$$

$$\sigma_{20} = \sqrt{\frac{1}{20} \sum_{i=1}^{20} (V_{t-i} - \mu_{20})^2}$$

$$V_z = \frac{V_t - \mu_{20}}{\sigma_{20}}$$

* **Trigger Threshold:** $V_z \ge 2.5$ triggers a `VOLUME_SPIKE` event (probability of random occurrence in a normal distribution is $< 0.6\%$).

---

### B. Abnormal Price Return ($R_t$)
Measures the percentage price movement on day $t$ compared to the previous trading session's close:

$$R_t = \frac{P_{\text{close}, t} - P_{\text{close}, t-1}}{P_{\text{close}, t-1}} \times 100\%$$

* **Trigger Threshold:** $|R_t| \ge 5.0\%$ triggers a `PRICE_BREAKOUT` anomaly.

---

### C. Sector Divergence ($D_t$)
Disentangles whether an asset's price movement is driven by broad industry momentum or idiosyncratic corporate catalysts:

$$D_t = R_{\text{stock}, t} - R_{\text{sector}, t}$$

* **Trigger Threshold:** $|D_t| \ge 4.0\%$ indicates an **Idiosyncratic Catalyst** specific to the issuer rather than systematic sector movement.

---

### D. Foreign Flow Inflow Z-Score ($F_z$)
Quantifies abnormal institutional or foreign capital accumulation:

$$F_z = \frac{F_t - \mu_{F, 20}}{\sigma_{F, 20}}$$

Where $F_t$ represents the Net Foreign Flow (in IDR) on trading session $t$.

* **Trigger Threshold:** $|F_z| \ge 2.5$ signifies an abnormal foreign capital allocation event.

---

### Anomaly Decision & Routing Matrix

| Condition $V_z$ | Condition $|R_t|$ | Condition $|D_t|$ | Classification | Engine Action |
|:---:|:---:|:---:|:---|:---|
| $\ge 2.5$ | $\ge 5.0\%$ | $\ge 4.0\%$ | `IDIOSYNCRATIC_CATALYST` | Triggers high-priority `event_causality_audit` news investigation. |
| $\ge 2.5$ | $< 5.0\%$ | Any | `VOLUME_ACCUMULATION` | Activates `insider_bandarmology_forensic` for foreign/domestic flow tracking. |
| $< 2.5$ | $\ge 5.0\%$ | $< 4.0\%$ | `SECTOR_BETA_RALLY` | Attributes movement to broader sector macro trends; suppresses false-alarm company alarms. |
| $< 2.5$ | $< 5.0\%$ | Any | `NORMAL_VARIANCE` | Routes to fundamental health and valuation baseline screening. |

---

## 5. Local Associative Graph Memory Engine

Niskava maintains multi-turn contextual memory locally using an associative knowledge graph backed by SQLite:

- **Data Model:**
  - `memory_nodes`: Represents unique market entities (e.g., `TICKER:ANTM`, `PERSON:CEO`, `EVENT:DIVIDEND`, `SECTOR:BASIC_MATERIALS`).
  - `memory_edges`: Represents typed directional relationships between entities (e.g., `ACCUMULATED_BY`, `SUBSIDIARY_OF`, `DIVERTED_FROM`) with metadata, confidence score, and observation timestamps.
  - `chat_messages`: Full multi-turn conversation logs linked to session IDs.

- **Ego-Graph Traversal:**
  When a query is received, the memory engine identifies focal seed entities and extracts an ego-subgraph of radius $k \le 2$ hops.

- **Exponential Recency Decay:**
  Edge weights decay dynamically based on the elapsed time since the relationship was last corroborated:

  $$w(t) = w_0 \cdot \exp(-\lambda \Delta t)$$

  Where $\Delta t$ is elapsed days and $\lambda = 0.05$ (half-life of approximately 14 days), ensuring recent market events carry higher relevance while preserving long-term structural links.

---

## 6. Discrete Verification Taxonomy & Confidence Rubric

To prevent subjective continuous probability scores (e.g. 0.50), Niskava enforces a standardized discrete verification rubric:

| Score | Verification Level | Source Validation Criteria |
|:---:|:---|:---|
| **`1.00`** | **EXTRACTED** | Directly backed by official Sectors API quantitative records or formal IDXnet regulatory disclosures. |
| **`0.95`** | **Direct Structural Evidence** | Explicit timestamp correlation with official company press releases or exchange disclosure announcements. |
| **`0.85`** | **Strong Inference** | High temporal correlation with major national business media reporting (Kontan, Bisnis Indonesia, CNBC Indonesia). |
| **`0.75`** | **Reasonable Inference** | Plausible catalyst from industry-wide trends corroborated by matching sector divergence metrics. |
| **`0.65`** | **Weak Inference** | Unverified market commentary, social media sentiment, or unconfirmed financial forum discussions. |
| **`0.55`** | **Speculative** | Distant co-occurrence without temporal causality or formal corroboration. |

---

## 7. Adaptive Inference Timeout Architecture

To eliminate socket disconnects and premature timeout failures during complex multi-tool research turns without sacrificing responsiveness on simple prompts, Niskava employs an observation-aware **Adaptive Inference Timeout Engine**:

### Mathematical Formula:
$$\text{Timeout}_{\text{call}} = \min\Big(\text{Base Timeout} + \max(0, N_{\text{obs}}) \times 10.0\text{s},\; 4 \times \text{Base Timeout},\; 300.0\text{s}\Big)$$

Where:
- $\text{Base Timeout}$: Configured baseline (default `60.0s`, customizable via `NISKAVA_LLM_TIMEOUT`).
- $N_{\text{obs}}$: Cumulative count of deterministic tool observations ingested during the current chat cycle.
- **Soft Cap ($4 \times \text{Base}$)**: Prevents runaway execution during massive batch observations.
- **Hard Cap ($300.0\text{s}$)**: Absolute upper limit guaranteeing system liveness.

### Prompt Complexity Classification:
Before tool execution begins, the agent classifies query intent deterministically:
1. **Simple** ($\le 5$ iterations, timeout capped at $15\text{s}$): Identity queries, greetings, help commands.
2. **General** ($\le 10$ iterations, timeout capped at $20\text{s}$): Macro reviews, sector overviews.
3. **Deep** ($\le 30$ iterations, full adaptive timeout): Ticker investigations, forensic causality audits, peer stress tests.

### Tri-Surface Synchronization:
User preferences are synchronized bidirectionally across three interfaces:
1. **Interactive Terminal Wizard (`niskava setup`)**: Configures predefined profiles (`Fast 25s`, `Balanced 60s`, `Deep 120s`, `Local LLM 180s`, or `Custom`).
2. **Terminal REPL (`/timeout [val]`)**: Modifies session timeout and immediately persists to `~/.niskava/config.yaml`.
3. **Web Workspace Canvas**: Interactive UI slider reading from `GET /api/settings` and updating via `PATCH /api/settings`.

