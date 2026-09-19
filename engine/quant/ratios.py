"""Deterministic financial statement and liquidity/solvency ratio calculations.

Abides by Law 1 (Deterministic Before Generative):
Never let LLM calculate accounting ratios, coverage figures, or percentages.
All financial metrics are computed strictly via pure Python arithmetic.
"""

from typing import Dict, Any


def compute_current_ratio(current_assets: float, current_liabilities: float) -> float:
    """Compute Current Ratio: Current Assets / Current Liabilities."""
    if current_liabilities <= 0:
        return 0.0
    return round(float(current_assets) / float(current_liabilities), 2)


def compute_quick_ratio(cash_and_equivalents: float, current_liabilities: float) -> float:
    """Compute Quick Ratio: Cash & Equivalents / Current Liabilities."""
    if current_liabilities <= 0:
        return 0.0
    return round(float(cash_and_equivalents) / float(current_liabilities), 2)


def compute_debt_to_equity(total_debt: float, total_equity: float) -> float:
    """Compute Debt to Equity Ratio (DER): Total Debt / Total Equity."""
    if total_equity <= 0:
        return 0.0
    return round(float(total_debt) / float(total_equity), 2)


def compute_interest_coverage(ebit: float, interest_expense: float) -> float:
    """Compute Interest Coverage Ratio (ICR): EBIT / Interest Expense."""
    if interest_expense <= 0:
        return 999.0 if ebit > 0 else 0.0
    return round(float(ebit) / float(interest_expense), 2)


def evaluate_financial_health(
    current_assets: float,
    cash_and_equivalents: float,
    current_liabilities: float,
    total_debt: float,
    total_equity: float,
    ebit: float,
    interest_expense: float,
) -> Dict[str, Any]:
    """Calculate and classify complete balance sheet liquidity and solvency health."""
    current_ratio = compute_current_ratio(current_assets, current_liabilities)
    quick_ratio = compute_quick_ratio(cash_and_equivalents, current_liabilities)
    der = compute_debt_to_equity(total_debt, total_equity)
    icr = compute_interest_coverage(ebit, interest_expense)

    if quick_ratio >= 1.5 and current_ratio >= 2.0:
        liquidity_grade = "PRISTINE"
    elif quick_ratio >= 1.0 and current_ratio >= 1.2:
        liquidity_grade = "HEALTHY"
    elif current_ratio >= 1.0:
        liquidity_grade = "ADEQUATE"
    else:
        liquidity_grade = "DISTRESSED"

    if der <= 1.0 and icr >= 3.0:
        solvency_grade = "STRONG"
    elif der <= 2.0 and icr >= 1.5:
        solvency_grade = "MODERATE"
    else:
        solvency_grade = "OVERLEVERAGED"

    return {
        "current_ratio": current_ratio,
        "quick_ratio": quick_ratio,
        "liquidity_grade": liquidity_grade,
        "der": der,
        "interest_coverage": icr,
        "solvency_grade": solvency_grade,
    }
