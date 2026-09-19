---
name: peer-valuation-benchmark
description: Evaluate relative valuation metrics (P/E, P/B, EV/EBITDA, ROE) against subsector peers using median dispersion and Interquartile Range (IQR) percentiles without providing investment advice or price targets.
triggers:
  - User inquiries regarding valuation fairness, cheap/expensive status, or peer comparisons ("Is ANTM cheap?", "How does BBCA compare to peers?")
  - Fundamental review of valuation posture following significant rallies or selloffs
tools:
  - sectors_get_subsector_peers
  - sectors_get_company_report
---

# Peer Valuation Benchmark Skill

## 1. Overview
The `peer-valuation-benchmark` skill benchmarks an IDX company's valuation metrics against its immediate subsector peers. In strict compliance with **Law 2 (Strict Financial Non-Advisory Boundary)**, this skill maps valuation posture (discount vs premium relative to the peer group) without producing price targets or BUY/SELL recommendations.

## 2. Analytical Procedure (SOP)
1. **Retrieve Company & Subsector Fundamentals**:
   - Query `/v2/company/report/{symbol}/` to obtain company P/E, P/B, market cap, and subsector slug.
   - Query `/v2/subsector/{subsector}/` to obtain peer list and median industry multiples.
2. **Deterministic Compute Gate (Median & IQR Percentile)**:
   - Compute the median and Interquartile Range (IQR = $Q_3 - Q_1$) of peer valuation multiples.
   - Calculate valuation z-score: $z_{\text{val}} = \frac{\text{PER}_{\text{stock}} - \text{Median}_{\text{subsector}}}{\text{IQR}_{\text{subsector}}}$.
3. **Valuation Posture Classification**:
   - If $z_{\text{val}} \le -0.50$: Classify as `TRADING_AT_DISCOUNT` relative to peers.
   - If $z_{\text{val}} \ge 0.50$: Classify as `TRADING_AT_PREMIUM` relative to peers.
   - Otherwise: Classify as `FAIRLY_ALIGNED_WITH_PEERS`.
4. **Taxonomy & Confidence Rubric**:
   - Official Sectors API peer valuation metrics: `SUPPORTED`, Confidence `1.00`.
   - Clear non-advisory disclaimer mandatory on every evaluation.
