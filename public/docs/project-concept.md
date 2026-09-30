# Project Concept & Vision

## 1. Executive Summary

Niskava Agent is an autonomous financial market intelligence and equity research orchestration platform designed specifically for the Indonesia Stock Exchange (IDX / Bursa Efek Indonesia).

Unlike conventional financial analytics dashboards that only present raw charts, or generic chatbot wrappers that hallucinate calculations, Niskava operates under an **evidence-first investigative paradigm**. It automatically bridges the gap between structured quantitative facts (provided by the Sectors Financial API v2) and unstructured qualitative disclosures (IDX regulatory filings, corporate announcements, and syndicated financial news).

When an abnormal market event occurs—such as an unprecedented trading volume spike, sharp price breakout, or aggressive foreign capital inflow—Niskava does not merely display the anomaly. It formulates investigative hypotheses, executes deterministic mathematical verification, harvests contemporaneous external disclosures within a strict temporal window, and establishes causal relationships backed by a verifiable audit trail.

---

## 2. Real-World Usability & Immediate Value Proposition

Niskava Agent directly resolves the **Real-World Usability (40% Weight)** criterion established by the Sectors Hackathon Indonesia 2026: *"How well does the project address a real-world problem? Can someone use it today and benefit from it?"*

### Real-World Problems Solved
1. **The 45-Minute Manual Reconnaissance Bottle-Neck**:
   When an IDX stock experiences an abrupt breakout or abnormal volume spike, an equity analyst or retail trader spends 30 to 60 minutes cross-checking broker summaries, scanning IDXnet corporate disclosures, and searching financial media. Niskava automates this entire pipeline into an evidence audit trail generated in under 6 seconds.
2. **Combating Speculative "Pom-Pom" & Market Rumors**:
   Retail investor communities are flooded with unsubstantiated rumors and Telegram channel pumps. Niskava's 3-Tier Verification Taxonomy objectively proves whether rumors are corroborated by official filings (`SUPPORTED`), unproven correlation (`UNCERTAIN`), or explicitly debunked by company disclosures (`CONTRADICTED`).
3. **Information Leakage Detection (Pre-Announcement Volatility)**:
   By computing temporal precedence between volume spikes ($T_{\text{surge}}$) and formal public releases ($T_{\text{news}}$), Niskava detects unusual accumulation occurring *before* public disclosures (`PRECEDED_ANNOUNCEMENT`), revealing insider activity and flow asymmetry.
4. **Immediate Usability Today**:
   No complex cloud deployments or database setups are required. A user can run `npx @zyrexnns/niskava-agent` or `./niskava investigate ANTM --days 30` right now from any terminal and obtain immediate empirical insights.

---

## 3. Market Inefficiency on the Indonesia Stock Exchange (IDX)

The Indonesia Stock Exchange presents distinct structural characteristics that create severe information asymmetries for market participants:

### Fragmented Information Landscape
Quantitative trading data and qualitative disclosures reside in disconnected silos. Daily price series and broker summary statistics are available via trading terminals or data APIs, while corporate disclosures are posted to the IDXnet reporting portal, and explanatory reporting is dispersed across national business media (e.g., Kontan, Bisnis Indonesia, CNBC Indonesia). Reconciling a single unusual trading session typically requires an analyst to perform 30 to 45 minutes of manual cross-referencing.

### Pre-Announcement Volume Surges (Information Leakage)
In emerging markets, unusual volume accumulation or price appreciation frequently precedes official public announcements by 24 to 72 hours. Traditional market data tools cannot correlate whether a corporate disclosure was the catalyst that triggered a surge, or whether the surge occurred before any public information was disclosed.

### Retail Sentiment vs. Institutional Flow
The IDX retail investor base has grown rapidly, often driven by social media commentary, influencer channels, and unverified rumors. Disentangling speculative retail chatter from genuine institutional accumulation (*bandarmology* / foreign flow) requires rigorous quantitative cross-examination against official exchange flow metrics.

---

## 3. The Anti-Wrapper Manifesto: Why Generic AI Wrappers Fail in Capital Markets

Many tools marketed as "Financial AI" are simple thin wrappers built on top of commercial Large Language Models (LLMs). Deploying such wrappers in equity research introduces fundamental operational hazards:

| The 5 Anti-Patterns of Financial AI Wrappers | Niskava Agent Architectural Defense |
|---|---|
| **1. Raw JSON Dump & Prompt Stuffing**<br>Dumping hundreds of raw candlestick rows directly into a prompt exhausts token limits and introduces noise that degrades reasoning. | **Two-Phase Compute Gate:** Raw time series data is processed locally by deterministic NumPy algorithms; only summarized anomaly indicators are exposed to the reasoning model. |
| **2. Mental Math & Numerical Hallucination**<br>Allowing an LLM to calculate percentage returns, moving averages, or standard deviations leads to confident numerical hallucinations. | **Law 1 (Deterministic Before Generative):** LLMs are strictly forbidden from performing mathematical calculations. All statistics, Z-scores, and divergence metrics are computed by NumPy. |
| **3. Atemporal Semantic Search (Causality Inversion)**<br>Standard vector similarity search retrieves articles based on topical similarity without temporal constraints, frequently attributing price surges to news published days *after* the event. | **Temporal-Aware News Anchoring:** Web harvesting is anchored strictly around the anomaly event date ($T_{\text{anomaly}} \pm 2\text{ days}$) to preserve chronological causality. |
| **4. Monolithic Prompt Architecture**<br>Stuffing all instructions into a single massive system prompt prevents modular problem solving, auditable steps, and skill-specific reasoning. | **4-Layer Agent Hierarchy:** Strict separation between MCP I/O primitives, deterministic math, modular analytical skills (SOPs), and the cognitive ReAct loop. |
| **5. Unregulated Financial Advice**<br>Generic models frequently produce unsolicited buy/sell recommendations or speculative price targets, violating securities regulations. | **Law 2 (Strict Non-Advisory Boundary):** Niskava produces only evidence-backed intelligence classified into a 3-tier verification taxonomy (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`). |

---

## 4. The Evidence-First Investigation Paradigm

Niskava adopts an evidence-first philosophy inspired by open investigative workflows and formal intelligence analysis:

```
[ Market Observation / User Prompt ]
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 4: ReAct Cognitive Orchestrator                       │
│ - Analyzes query intent                                     │
│ - Selects appropriate analytical skill (Domain SOP)         │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: Modular Skills Registry                            │
│ - Market Anomaly Reconnaissance                             │
│ - Event Causality Audit                                     │
│ - Insider & Foreign Flow Forensics                          │
│ - Financial Health Stress Testing                           │
│ - Mining & Commodity Divergence                             │
│ - Peer Valuation Benchmarking                               │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Deterministic Compute Gate (NumPy Firewall)        │
│ - Volume Z-Score (Vz >= 2.5)                                │
│ - Abnormal Return (|Rt| >= 5.0%)                            │
│ - Sector Divergence (|Dt| >= 4.0%)                          │
│ - Foreign Inflow Z-Score (|Fz| >= 2.5)                      │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: MCP & News Data Primitives                        │
│ - Sectors Financial API v2 (OHLCV, Financials, Flow)        │
│ - Sectors News Engine (Curated Sectors News & Filings)      │
│ - Trafilatura HTML Sanitization                             │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ Verification Engine: 3-Tier Classification & Confidence     │
│ - SUPPORTED    : Corroborated by official exchange filings  │
│ - UNCERTAIN    : Plausible correlation, direct proof absent │
│ - CONTRADICTED : Refuted by official disclosures or filings │
└─────────────────────────────────────────────────────────────┘
```

### Deterministic Ground Truth
Official exchange records provided by the Sectors Financial API v2 serve as the unassailable baseline. Historical price series, financial ratios, net foreign flow figures, and corporate action schedules are cached permanently in local storage and treated as deterministic facts.

### Temporal Precedence Verification
Evidence correlation enforces strict chronological order:
- **`LIKELY_CATALYST`**: An official filing or verified news report was published prior to or simultaneously with the trading volume spike.
- **`PRECEDED_ANNOUNCEMENT`**: The trading volume surge occurred before any public corporate announcement was released, flagging possible early market awareness or information leakage.
- **`UNEXPLAINED_BY_NEWS`**: A statistically significant anomaly occurred with no corresponding public disclosures or media reporting, indicating purely technical, order-driven, or unannounced market dynamics.

---

## 5. Target User Personas & Use Cases

### Equity Research Analysts & Associates
- Automate first-pass background checks on morning price or volume outliers.
- Cross-examine company disclosures against financial statements and peer valuation multiples without manually opening multiple websites.
- Generate structured Markdown research briefs with verifiable citations.

### Financial Journalists & Market Reporters
- Rapidly identify the exact catalyst behind sudden index-moving stocks.
- Verify whether market rumors circulating on social channels are contradicted or supported by official IDXnet disclosures.
- Review historical event timelines to establish whether price movement preceded official press releases.

### Serious Retail Swing Traders & Independent Investors
- Screen for stocks experiencing institutional accumulation (foreign flow Z-scores $\ge 2.5$) while avoiding retail sentiment traps.
- Perform financial solvency stress tests (Altman Z-Score, current ratios, debt maturity schedules) before entering positions.
- Maintain full data privacy by keeping all research notes, session transcripts, and memory graphs locally on their personal machines.

---

## 6. Ethical Guardrails & Regulatory Compliance

Niskava Agent is strictly an informational and investigative research system. It is engineered with hardcoded architectural constraints ensuring full compliance with capital market regulations (including OJK and IDX regulations, as well as Sectors Hackathon guidelines):

1. **No Buy/Sell Recommendations**: The system never outputs trade directives, price targets, or portfolio allocation advice.
2. **No Automated Order Routing**: The codebase contains zero broker connectivity, order submission APIs, or automated trading integrations.
3. **Evidence-Based Transparency**: Every conclusion references specific dates, official publication links, and discrete confidence scores according to a standardized verification rubric.
