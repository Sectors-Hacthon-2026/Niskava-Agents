---
name: insider-bandarmology-forensic
description: Forensic audit of broker accumulation/distribution patterns, Top-3 Buyer Concentration (C3), and insider trading disclosures on the Indonesia Stock Exchange.
triggers:
  - Inquiries about who is accumulating or distributing an IDX stock (e.g. "Who bought ANTM?", "Is bandar accumulating BBRI?")
  - Anomaly precedes news release (PRECEDED_ANNOUNCEMENT) or extreme foreign inflow is detected
tools:
  - sectors_get_broker_summary
  - sectors_get_broker_registry
  - sectors_get_filings
---

# Insider & Bandarmology Forensic Skill

## 1. Overview
The `insider-bandarmology-forensic` skill analyzes market microstructure on the Indonesia Stock Exchange (IDX). It investigates broker concentration and insider transaction filings (directors, commissioners, and controlling shareholders) using deterministic concentration ratios and broker cohort mappings.

## 2. Analytical Procedure (SOP)
1. **Retrieve Broker Transaction Data**:
   - Query `/v2/broker-summary-top/{symbol}/` to obtain net buyer and seller volumes by broker code.
   - Query `/v2/broker-registry/` to classify participating brokers by domicile (`FOREIGN` vs `DOMESTIC`) and cohort (`INSTITUTION` vs `RETAIL`).
2. **Compute Top-3 Concentration Ratio ($C_3$)**:
   - Calculate: $C_3 = \frac{\sum_{j=1}^3 \text{Volume Buyer}_j}{\text{Total Market Volume}}$.
   - If $C_3 \ge 65.0\%$, classify market structure as `INSTITUTIONAL_ACCUMULATION`.
   - If top sellers dominate, classify as `INSTITUTIONAL_DISTRIBUTION`.
3. **Audit Insider Filings (Tier 1 Evidence)**:
   - Query `/v2/filings/?symbol={symbol}` for formal OJK/IDX insider ownership change reports.
   - Correlate filing dates and insider transaction directions (BUY/SELL) with price anomalies.
4. **Taxonomy & Confidence Rubric**:
   - Official OJK/IDX insider filings: `SUPPORTED`, Confidence `1.00`.
   - Broker summary concentration: `SUPPORTED`, Confidence `0.95`.
   - Inferences regarding undeclared market makers: `UNCERTAIN`, Confidence `0.65`.
