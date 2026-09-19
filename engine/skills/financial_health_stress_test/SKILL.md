---
name: financial-health-stress-test
description: Conduct balance sheet stress testing (liquidity, debt-to-equity, and interest coverage ratios) and fact-check bankruptcy/default market rumors against official quarterly financial filings.
triggers:
  - User inquiries regarding bankruptcy risk, debt distress, or insolvency rumors ("Is GOTO going bankrupt?", "Can BBRI pay its debts?")
  - Extreme negative price drops (Rt <= -5.0%) without immediate operational news
tools:
  - sectors_get_company_report
  - sectors_get_quarterly_financials
---

# Financial Health Stress Test Skill

## 1. Overview
The `financial-health-stress-test` skill verifies company solvency and liquidity through deterministic accounting calculations. It evaluates quarterly balance sheet metrics to validate or refute bankruptcy or debt default rumors circulating in social or retail channels.

## 2. Analytical Procedure (SOP)
1. **Retrieve Financial Statements**:
   - Query `/v2/quarterly-financials/{symbol}/` to obtain balance sheet and income statement items.
   - Query `/v2/company/report/{symbol}/?sections=valuation,financials` for fundamental ratios.
2. **Deterministic Compute Gate (NumPy / Python Ratios)**:
   - Calculate Liquidity:
     - $\text{Current Ratio} = \frac{\text{Current Assets}}{\text{Current Liabilities}}$
     - $\text{Quick Ratio} = \frac{\text{Cash \& Cash Equivalents}}{\text{Current Liabilities}}$
   - Calculate Solvency:
     - $\text{Debt to Equity (DER)} = \frac{\text{Total Debt}}{\text{Total Equity}}$
     - $\text{Interest Coverage Ratio (ICR)} = \frac{\text{EBIT}}{\text{Interest Expense}}$
3. **Rumor Refutation Analysis**:
   - If market claims allege insolvency or bond default, but Quick Ratio $> 1.0$ and ICR $\ge 2.0x$:
     - Tag claim as `CONTRADICTED`.
     - Cite exact cash balance and short-term liabilities from the official report.
4. **Taxonomy & Confidence Rubric**:
   - Official financial report facts: `SUPPORTED`, Confidence `1.00`.
   - Refuted insolvency rumors: `CONTRADICTED`, Confidence `1.00`.
   - Unverified financial rumors: `UNCERTAIN`, Confidence `0.65`.
