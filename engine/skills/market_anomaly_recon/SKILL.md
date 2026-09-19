---
name: market-anomaly-recon
description: Detect quantitative trading anomalies including statistical volume surges (MA20 Z-score), extreme abnormal returns, and foreign capital inflow divergence on IDX stocks.
triggers:
  - User inquiries about unusual price swings or massive volume surges on an IDX stock
  - Daily Unusual Market Activity (UMA) reconnaissance and screening
tools:
  - sectors_get_daily_candles
  - sectors_get_foreign_flow
  - sectors_get_subsector_peers
---

# Market Anomaly Reconnaissance Skill

## 1. Overview
The `market-anomaly-recon` skill performs systematic, deterministic quantitative screening on Indonesia Stock Exchange (IDX) tickers. It computes 20-day rolling baseline statistics using pure NumPy arithmetic, eliminating hallucination of price movements and statistical thresholds.

## 2. Analytical Procedure (SOP)
1. **Retrieve Time Series Data**: Pull 30 to 90 days of daily OHLCV candlesticks and net foreign institutional flow records via Sectors Financial API v2 (served from local SQLite cache whenever available).
2. **Deterministic Compute Gate (NumPy Firewall)**:
   - Calculate rolling 20-day mean volume ($\mu_{20}$) and standard deviation ($\sigma_{20}$).
   - Compute Volume Z-Score: $V_z = \frac{V_t - \mu_{20}}{\sigma_{20}}$.
   - Compute Daily Price Return: $R_t = \frac{P_t - P_{t-1}}{P_{t-1}} \times 100\%$.
   - Compute Foreign Flow Z-Score: $F_z = \frac{F_t - \mu_{F,20}}{\sigma_{F,20}}$.
   - Compute Subsector Divergence: $D_t = R_{t,\text{stock}} - R_{t,\text{subsector}}$.
3. **Threshold Evaluation**:
   - Anomaly Flag Triggered when: $(V_z \ge 2.5) \lor (|R_t| \ge 5.0\%) \lor (|F_z| \ge 2.5)$.
4. **Handoff**:
   - When an anomaly date $T_{\text{anomaly}}$ is isolated, forward the finding to `event-causality-audit` for external news and regulatory disclosure verification.
5. **Taxonomy & Confidence Rubric**:
   - Status: `SUPPORTED` (backed 100% by official exchange trade records).
   - Confidence Score: `1.00` (official Sectors API quantitative data).
