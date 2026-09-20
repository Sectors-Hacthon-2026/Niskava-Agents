---
name: event-causality-audit
description: Audit temporal causality between price/volume spikes and external events (IDX disclosures, suspensions, corporate actions, and accredited financial news).
triggers:
  - An anomaly date T_anomaly has been detected by market-anomaly-recon
  - User inquiries regarding catalysts behind price surges ("Why did ANTM move?", "Any news on BBRI?")
tools:
  - sectors_get_suspensions
  - sectors_get_corporate_actions
  - osint_harvest_dual_engine
---

# Event Causality Audit Skill

## 1. Overview
The `event-causality-audit` skill correlates quantitative anomaly timestamps with qualitative news releases, corporate disclosures, and exchange suspension notices. It enforces strict chronological ordering to prevent temporal causality inversion (news occurring after a price jump cannot be cited as its cause).

## 2. Analytical Procedure (SOP)
1. **Define Strict Temporal Window**:
   - Given anomaly date $T_{\text{anomaly}}$, set observation window to $[T_{\text{anomaly}} - 2\text{ days}, T_{\text{anomaly}} + 1\text{ day}]$.
   - Disregard news or rumors outside this window to maintain evidentiary integrity.
2. **Query Regulatory Disclosures (Tier 1 Evidence)**:
   - Check `/v2/suspensions/` for exchange Unusual Market Activity (UMA) or suspension notices, capturing official PDF links.
   - Check `/v2/corporate-actions/` for scheduled dividends, stock splits, or rights issues.
3. **Execute Dual-Engine OSINT Collection (Tier 2 Evidence)**:
   - Fetch curated news from Sectors API v2 `/v2/news/`.
   - Perform targeted Google News RSS dorking against accredited financial outlets (Kontan, Bisnis Indonesia, CNBC Indonesia).
4. **Evaluate Temporal Precedence**:
   - If news timestamp < volume surge timestamp $\to$ `LIKELY_CATALYST`.
   - If volume surge timestamp < news timestamp $\to$ `PRECEDED_ANNOUNCEMENT` (suspected selective disclosure / info leakage).
   - If no news or filing is detected in the window $\to$ `UNEXPLAINED_BY_NEWS`.
5. **Taxonomy & Confidence Rubric**:
   - Official IDX disclosures / PDF filings: `SUPPORTED`, Confidence `0.95`–`1.00`.
   - Accredited media news reporting: `SUPPORTED`, Confidence `0.85`.
   - Social media or unverified chatter: `UNCERTAIN`, Confidence `0.65`.
