# Features & System Architecture

This document provides the definitive technical specification of Niskava Agent: the Tripartite Hybrid Stack, the 6 Invariant Laws, the 7-Stage Investigation Pipeline, the deterministic mathematical formulas, the 6 modular domain skills, and the Unified Model Context Protocol (MCP) server.

---

## 1. Tripartite Hybrid Stack Architecture

Niskava Agent combines the low-latency systems capabilities of Go, the scientific computing and AI ecosystem of Python, and the visual ergonomics of React:

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
│  - Institutional PDF Exporter (`investigation_report_pdf`)      │
│                                                                 │
│  [Layer 2: Deterministic Compute Gate (NumPy Firewall)]         │
│  - Volume Z-Scores, Abnormal Returns, Sector Divergence         │
│  - Foreign Inflow Z-Scores, Altman Z-Score Ratios               │
│                                                                 │
│  [Layer 1: Sectors MCP & News Engine Primitives]                │
│  - Sectors Financial API v2 MCP Server Adapter                  │
│  - Curated Sectors News & Corporate Filings Engine              │
│  - Content Sanitization via Trafilatura (<evidence_context>)    │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                   REACT SPA WEB WORKSPACE                       │
│  - Bloomberg Terminal-Inspired Dark & Light Themes              │
│  - Interactive Candlestick Charts & Anomaly Overlays            │
│  - Real-Time Thinking Stream via Server-Sent Events (SSE)       │
│  - Interactive Evidence Matrix & Causality Timeline Graph       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. The 6 Non-Negotiable Invariant Architectural Laws

Every subsystem within Niskava Agent is governed by six foundational architectural invariants (codified in `AGENTS.md`):

### Law 1: Deterministic Before Generative
* **Rule:** Large Language Models are strictly prohibited from calculating time-series statistics, moving averages, standard deviations, Z-scores, abnormal returns, or sector divergence.
* **Mechanism:** All numerical operations are computed deterministically via Python NumPy and Pandas before any LLM prompt is constructed. The LLM receives verified statistical summaries, never raw data arrays.
* **Rationale:** Completely eliminates numerical hallucinations and preserves token budget.

### Law 2: Strict Financial Non-Advisory Boundary
* **Rule:** Niskava is an investigative market intelligence platform, **NOT an investment advisor**. The system must never emit direct BUY/SELL recommendations, price targets, or portfolio allocation advice.
* **Mechanism:** All findings are strictly classified into a 3-tier verification taxonomy:
  - `SUPPORTED`: Validated by official Sectors API records or formal IDXnet regulatory disclosures.
  - `UNCERTAIN`: Plausible correlation observed, but direct causal proof is unverified.
  - `CONTRADICTED`: Market speculation refuted by corporate filings or audited statements.
* **Compliance:** Enforces OJK / IDX securities regulations across all CLI and web output surfaces.

### Law 3: Prohibition of Automated Trade Execution
* **Rule:** The system may analyze, screen, score, alert, and correlate evidence, but may never place, execute, or automate trade orders.
* **Mechanism:** The codebase contains zero trading execution APIs, broker connection libraries, or order-routing dependencies. It is strictly read-only market intelligence.

### Law 4: Local-First Data Sovereignty
* **Rule:** No centralized cloud database. All user investigation sessions, findings, evidence graphs, chat transcripts, and API caches must reside locally on the host machine.
* **Mechanism:** Go Core uses pure-Go zero-CGO SQLite (`modernc.org/sqlite`); Python uses standard library `sqlite3`. Database files are stored at `~/.niskava/niskava.db` with Write-Ahead Logging (`PRAGMA journal_mode = WAL;`) for concurrent read/write access.

### Law 5: Credit Budget Discipline & Local Caching
* **Rule:** Strictly protect the 1,000 Sectors API credit allocation. Zero redundant HTTP requests.
* **Mechanism:** All requests check the local `sectors_cache` table before issuing HTTP requests to `https://api.sectors.app/v2`. Historical daily candlestick data ($T < \text{today}$) is permanently cached (`expires_at = NULL`), incurring zero credit cost on repeat queries.
* **Offline Support:** Supports `MOCK_SECTORS=1` using static JSON fixtures strictly for CI/CD pipelines and unit testing.

### Law 6: Local Conversational Graph Memory Engine
* **Rule:** Agents must maintain associative context across multi-turn sessions without relying on expensive cloud vector databases or heavy external graph services.
* **Mechanism:** Entities and relationships are persisted in SQLite (`memory_nodes` and `memory_edges`), loaded into an in-memory `networkx.DiGraph`, and queried via Ego-Graph traversal ($k \le 2$ hops) with exponential recency decay ($e^{-\lambda \Delta t}$).

---

## 3. Deterministic Quantitative Anomaly Formulas

All quantitative indicators are calculated deterministically by the Python Engine before LLM activation.

### A. Volume Anomaly Z-Score ($V_z$)
Measures the statistical significance of trading volume on day $t$ relative to the 20-day trading volume baseline:

$$\mu_{20} = \frac{1}{20} \sum_{i=1}^{20} V_{t-i}$$

$$\sigma_{20} = \sqrt{\frac{1}{20} \sum_{i=1}^{20} (V_{t-i} - \mu_{20})^2}$$

$$V_z = \frac{V_t - \mu_{20}}{\sigma_{20}}$$

* **Trigger Threshold:** $V_z \ge 2.5\sigma$ triggers a `VOLUME_SPIKE` event (probability of random occurrence in a normal distribution is $< 0.6\%$).

---

### B. Abnormal Price Return ($R_t$)
Measures the percentage price movement on day $t$ compared to the previous trading session's close:

$$R_t = \left( \frac{P_{\text{close}, t} - P_{\text{close}, t-1}}{P_{\text{close}, t-1}} \right) \times 100\%$$

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

* **Trigger Threshold:** $|F_z| \ge 2.5\sigma$ signifies an abnormal foreign capital allocation event.

---

### E. Altman Z-Score Distress Index ($Z$)
Used by the `financial_health_stress` skill to evaluate balance sheet bankruptcy risk:

$$Z = 1.2 X_1 + 1.4 X_2 + 3.3 X_3 + 0.6 X_4 + 1.0 X_5$$

Where:
* $X_1 = \text{Working Capital} / \text{Total Assets}$ (Liquidity)
* $X_2 = \text{Retained Earnings} / \text{Total Assets}$ (Cumulative Profitability)
* $X_3 = \text{EBIT} / \text{Total Assets}$ (Operating Efficiency)
* $X_4 = \text{Market Value of Equity} / \text{Total Liabilities}$ (Leverage)
* $X_5 = \text{Sales} / \text{Total Assets}$ (Asset Turnover)

**Zone Classification:**
* $Z > 2.99$: **Safe Zone** (Low probability of insolvency).
* $1.81 \le Z \le 2.99$: **Grey Zone** (Moderate distress risk).
* $Z < 1.81$: **Distress Zone** (High probability of financial distress).

---

### Anomaly Decision & Routing Matrix

| Condition $V_z$ | Condition $|R_t|$ | Condition $|D_t|$ | Classification | Engine Action |
|:---:|:---:|:---:|:---|:---|
| $\ge 2.5\sigma$ | $\ge 5.0\%$ | $\ge 4.0\%$ | `IDIOSYNCRATIC_CATALYST` | Triggers high-priority `event_causality_audit` news investigation. |
| $\ge 2.5\sigma$ | $< 5.0\%$ | Any | `VOLUME_ACCUMULATION` | Activates `insider_bandarmology_forensic` for foreign/domestic flow tracking. |
| $< 2.5\sigma$ | $\ge 5.0\%$ | $< 4.0\%$ | `SECTOR_BETA_RALLY` | Attributes movement to broader sector macro trends; suppresses false-alarm company alarms. |
| $< 2.5\sigma$ | $< 5.0\%$ | Any | `NORMAL_VARIANCE` | Routes to fundamental health and valuation baseline screening. |

---

## 4. The 7-Stage Investigation Pipeline Protocol

Whenever a ticker investigation is executed (`niskava investigate <TICKER> --days 30`), Niskava runs a sequential 7-stage protocol:

```
[1. INITIATION] ──> [2. SECTORS_BASELINE] ──> [3. QUANT_ANOMALY]
                                                        │
                      ┌─────────────────────────────────┴─────────────────────────────────┐
                      ▼                                                                   ▼
             (Anomaly Detected)                                                   (No Anomaly)
                      │                                                                   │
              [4. GAP_DETECTION]                                                 [4b. FUNDAMENTAL]
                      │                                                                   │
             [5. NEWS_HARVEST]                                                   [6b. PEER_VALUATION]
                      │                                                                   │
          [6. EVIDENCE_CORRELATION]                                                       │
                      │                                                                   │
                      └───────────────────────────────┬───────────────────────────────────┘
                                                      ▼
                                         [7. SYNTHESIS_AND_STREAMING]
```

1. **Stage 1: INITIATION:** Creates investigation session in SQLite (`status = 'PENDING'`), initializes IPC stream.
2. **Stage 2: SECTORS_BASELINE:** Checks `sectors_cache`, pulls 30–90 days daily OHLCV, company profile, foreign flow, corporate actions, and suspensions.
3. **Stage 3: QUANT_ANOMALY:** Deterministic NumPy math calculates $V_z$, $R_t$, $D_t$, and $F_z$.
4. **Stage 4: GAP_DETECTION:** Formulates targeted temporal investigation hypotheses centered on $T_{\text{anomaly}} \pm 2\text{ days}$.
5. **Stage 5: NEWS_HARVEST:** Retrieves curated news and IDXnet disclosures from Sectors API v2, sanitizes text via `trafilatura`, and isolates content in `<evidence_context>` delimiters to neutralize prompt injection.
6. **Stage 6: EVIDENCE_CORRELATION:** Assesses temporal precedence:
   - News before volume surge $\to$ `LIKELY_CATALYST`
   - Volume surge before news release $\to$ `PRECEDED_ANNOUNCEMENT` (possible information leakage)
   - No explanatory news $\to$ `UNEXPLAINED_BY_NEWS`
7. **Stage 7: SYNTHESIS_AND_STREAMING:** Compiles findings and timeline into local SQLite, streams progress via SSE, and produces optional PDF reports.

---

## 5. The 6 Modular Domain Skills

Niskava organizes specialized equity analysis into reusable domain modules:

| Skill Identifier | Category | Input Signals | Primary Analysis & Output |
|---|---|---|---|
| **`market_anomaly_recon`** | Technical & Momentum | 30–90d OHLCV, Sector Index | MA20, Z-Scores ($V_z$), Abnormal Returns ($R_t$), Sector Divergence ($D_t$). |
| **`event_causality_audit`** | Catalysts & Filings | Anomaly date, Sectors News, Disclosures | Chronological sequence audit, 3-tier evidence classification (`SUPPORTED`). |
| **`insider_bandarmology_forensic`** | Flow & Ownership | Top Broker summary, Foreign flow, Insider filings | Broker accumulation concentration, top buyer/seller delta, insider trades. |
| **`financial_health_stress`** | Solvency & Balance Sheet | Balance Sheet, Income, Cash Flow | Altman Z-Score, Debt-to-Equity (DER), Current Ratio, Interest Coverage. |
| **`mining_commodity_divergence`** | Commodities & Macro | Subsector data, Global commodity prices | Correlation between commodity price swings (Nickel, Coal, Gold, CPO) and stock return. |
| **`peer_valuation_benchmark`** | Relative Valuation | Subsector peers, P/E, P/B, EV/EBITDA | Peer percentile ranking, discount/premium vs subsector median. |

---

## 6. Unified Model Context Protocol (MCP) Server

Niskava provides a native **Model Context Protocol (MCP)** server (`niskava mcp` or `python -m engine.mcp.server`), allowing external IDEs and agents (Claude Desktop, Cursor, Antigravity) to call Niskava tools natively:

### Exposed MCP Tools:
* **`get_daily_candles`**: Retrieve 30–90 days OHLCV candlestick records for any IDX ticker.
* **`compute_quant_anomalies`**: Execute deterministic NumPy volume and price anomaly screening.
* **`harvest_market_news`**: Fetch accredited financial news and regulatory filings with temporal filtering.
* **`query_sectors`**: Unified gateway for Sectors Financial API v2 (company reports, foreign flow, broker summaries, corporate actions).
* **`execute_skill`**: Execute any of the 6 institutional domain skills on demand.
* **`recall_graph_memory`**: Query local associative memory for historical investigation findings.

---

## 7. Institutional PDF Dossier Generation

When requested (`niskava investigate <TICKER> --pdf` or in the Web Workspace), Niskava compiles an institutional-grade PDF research dossier directly to `~/.niskava/reports/<TICKER>_investigation_<SESSION>.pdf`:
* Structured executive summary with anomaly alert headers.
* Complete quantitative indicators table ($V_z$, $R_t$, $D_t$, $F_z$).
* Corroborated evidence matrix with source URLs and publication timestamps.
* Formal Capital Market non-advisory disclaimer.

---

## 8. Local Conversational Graph Memory Engine Architecture

Traditional financial chat interfaces suffer from stateless session amnesia or rely on crude sliding-window chat history that bloats token expenditure and atemporally mixes unrelated discussion points. Furthermore, sending proprietary research notes or portfolio holdings to cloud vector services violates data privacy.

Niskava implements a **Local-First Associative Graph Memory Engine** that converts multi-turn research findings into a structured entity relationship graph stored locally in SQLite and traversed via NetworkX in Python.

<p align="center">
  <img src="../../docs/assets/memory-graph.png" alt="Niskava Local Conversational Graph Memory Engine" width="100%">
</p>

### A. Mathematical Formulation

#### 1. Temporal Recency Decay
Edge weights decay exponentially across elapsed trading sessions to ensure recent market catalysts take precedence over historical observations:

$$W_{\text{effective}} = W_0 \times e^{-\lambda \Delta t}$$

Where:
* $W_0$: Initial edge weight assigned during ingestion (default: $1.0$).
* $\lambda$: Decay rate coefficient ($\lambda \approx 0.10$ per day).
* $\Delta t$: Elapsed time in days between $T_{\text{observed}}$ and the current investigation timestamp.

#### 2. Bounded Ego-Graph Traversal
To eliminate context window dilution, retrieval is bounded to a local ego-network around the active subject entity $v$:

$$G_{\text{ego}}(v, k) = \{ u \in V \mid d(v, u) \le k \}$$

Where $k \le 2$ hops. This guarantees that traversal captures immediate relationships (e.g., `ANTM -> Smelter Catalyst`) and second-degree relationships (e.g., `Smelter Catalyst -> Nickel Sector`), while pruning unrelated market chatter. Traversal executes in memory via NetworkX in under 5 milliseconds.

---

### B. Graph Entity & Relation Taxonomy

The engine enforces a standardized schema persisted in SQLite tables `memory_nodes` and `memory_edges`:

| Node Classification | Identifier Code | Color Marker | Semantic Description |
|---|---|---|---|
| **User Research Profile** | `USER` | Light Blue | Central user anchor node tracking individual inquiries, watchlists, and price anchors. |
| **Stock Issuer** | `TICKER` | Yellow | IDX equity symbols (e.g., `ANTM`, `BBRI`, `BMRI`). |
| **Exchange Member** | `BROKER` | Purple | Licensed brokerage participants tracked during accumulation/distribution audits. |
| **Industry Sector** | `SECTOR` | Cyan | Official IDX sector and sub-sector classifications (e.g., `Basic Materials`, `Energy`). |
| **Disclosures & Events** | `CATALYST_EVENT` | Green | Official corporate disclosures, dividend announcements, and accredited news stories. |
| **Volume Outlier & Flow** | `VOLUME_OUTLIER` | Red | Quantitative anomaly events flagged by $V_z \ge 2.5\sigma$ or foreign flow streaks. |

#### Directed Relation Types:
* `(USER) ──[INVESTIGATED]──> (TICKER)`
* `(USER) ──[HOLDS_AT]──> (PRICE_LEVEL)`
* `(TICKER) ──[BELONGS_TO]──> (SECTOR)`
* `(TICKER) ──[TRIGGERED_ANOMALY]──> (VOLUME_OUTLIER)`
* `(TICKER) ──[CATALYZED_BY]──> (CATALYST_EVENT)`
* `(BROKER) ──[ACCUMULATED]──> (TICKER)`

---

### C. The 4-Phase Memory Lifecycle

```
[1. INGESTION] ──> [2. STORAGE & NORMALIZATION] ──> [3. EGO-GRAPH RETRIEVAL] ──> [4. PROMPT AUGMENTATION]
```

1. **Ingestion (Dual-Mode):**
   - **Deterministic (Zero-Token Cost):** Automatically registers entity relations upon completion of quantitative investigation pipelines (`niskava investigate`).
   - **Lightweight Extraction:** Extracts user holdings, target prices, and research notes from conversational prompts via compact JSON schema extraction.
2. **Storage & Normalization:**
   - Normalizes issuer references (e.g., *"Aneka Tambang"*, *"PT ANTM"*, *"Antam"* map to canonical `TICKER:ANTM`).
   - Writes to local SQLite (`~/.niskava/niskava.db`) tables `memory_nodes` and `memory_edges` with foreign key enforcement and Write-Ahead Logging (`WAL`).
3. **Ego-Graph Retrieval:**
   - Detects mentioned entities in the current query.
   - Loads the active subgraph into an in-memory `networkx.DiGraph`.
   - Computes PageRank and degree centrality to rank the most relevant hubs.
   - Extracts top $k \le 2$ hop neighbors with $W_{\text{effective}}$ above threshold.
4. **Prompt Augmentation:**
   - Formats extracted subgraphs into an isolated XML structure (`<investigative_memory>`) bounded under 300 tokens:
   ```xml
   <investigative_memory>
   - User investigated ANTM on 2026-09-22 (Session INV-2026-0042).
   - Anomaly: Volume spike 3.84σ coincided with Halmahera smelter completion.
   - Active position note: Entry recorded at IDR 1,450.
   </investigative_memory>
   ```

---

### D. Centrality & Visual Inspection Canvas

The engine calculates network centrality metrics in real time:
* **Degree Centrality:** Identifies entities connected to the highest number of anomalies or user investigations.
* **PageRank Scoring:** Pinpoints structural hubs that bridge distinct market clusters (e.g., a holding company linking multiple sector subsidiaries).
* **Visual Canvas:** Rendered via interactive force-directed physics in the Web Workspace (`/graph`) or exported to standalone HTML via `niskava graph --open`.

