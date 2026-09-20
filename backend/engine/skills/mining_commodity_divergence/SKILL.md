---
name: mining-commodity-divergence
description: Measure Pearson correlation between IDX mineral/energy mining stocks and global benchmark spot commodity prices (Nickel, Coal, Gold) to isolate company-specific idiosyncratic alpha.
triggers:
  - Significant movements in IDX mining/energy stocks (ANTM, PTBA, ADRO, MEDC, TINS, MBMA, INCO)
  - Inquiries regarding whether a stock price surge was driven by global commodity prices or internal catalysts
tools:
  - sectors_get_mining_detail
  - sectors_get_commodity_price
  - sectors_get_daily_candles
---

# Mining Commodity Divergence Skill

## 1. Overview
The `mining-commodity-divergence` skill assesses whether an IDX mining issuer's price action is merely beta to global commodity price trends or reflects company-specific idiosyncratic catalysts. It deterministically computes 30-day Pearson correlation coefficients ($r$) against global spot prices.

## 2. Analytical Procedure (SOP)
1. **Identify Primary Commodity**:
   - Query `/v2/mining-company-detail/{slug}/` to determine the primary mineral or energy product (e.g. Nickel for ANTM, Coal for PTBA/ADRO, Gold for MDKA).
2. **Retrieve Time Series**:
   - Pull 30 daily OHLCV candlesticks for the target equity.
   - Pull 30 benchmark price observations for the relevant commodity from `/v2/commodity-price/{commodity}/`.
3. **Deterministic Compute Gate (Pearson Correlation & Returns)**:
   - Compute Pearson correlation $r$ between daily equity return and commodity spot price delta.
   - Calculate cumulative returns: $\Delta_{\text{stock}}$ and $\Delta_{\text{commodity}}$.
4. **Divergence Classification**:
   - If $r \ge 0.60$: Classify as `COMMODITY_DRIVEN`. Movement is driven by macro commodity trends.
   - If $r < 0.60$: Classify as `IDIOSYNCRATIC_COMPANY_ALPHA`. Movement is driven by company-specific catalysts (e.g. smelter commissioning, corporate acquisitions).
5. **Taxonomy & Confidence Rubric**:
   - High correlation confirmed: `SUPPORTED`, Confidence `0.85`–`0.95`.
   - Idiosyncratic alpha isolated: `SUPPORTED`, Confidence `0.85`.
