"""Deterministic bandarmology and broker flow calculations using pure NumPy.

Abides by Law 1 (Deterministic Before Generative):
Calculates Top-3 Buyer Concentration (C3) and cohort classifications
deterministically before any LLM evaluation.
"""

from typing import Any, Dict, List, Optional
import numpy as np


def compute_top3_buyer_concentration(
    top_buyers: List[Dict[str, Any]],
    total_market_volume: float,
) -> float:
    """Compute Top-3 Buyer Concentration Ratio (C3).

    Formula:
        C3 = sum(Volume Buyer_j for j in 1..3) / Total Market Volume

    Args:
        top_buyers: List of buyer dicts with 'net_buy_shares' or 'volume'.
        total_market_volume: Total traded shares during observation period.

    Returns:
        Concentration ratio as a float between 0.0 and 1.0.
    """
    if total_market_volume <= 0 or not top_buyers:
        return 0.0

    volumes = []
    for b in top_buyers[:3]:
        vol = float(b.get("net_buy_shares", b.get("volume", 0.0)))
        volumes.append(max(0.0, vol))

    top3_sum = float(np.sum(volumes))
    ratio = top3_sum / float(total_market_volume)
    return float(np.clip(ratio, 0.0, 1.0))


def classify_broker_cohort(
    broker_code: str,
    registry: Optional[List[Dict[str, Any]]] = None,
) -> Dict[str, str]:
    """Classify a broker by domicile (Foreign/Domestic) and cohort (Institutional/Retail).

    Args:
        broker_code: Official 2-letter IDX broker code (e.g. 'CS', 'YP').
        registry: List of broker registry items from /v2/broker-registry/.

    Returns:
        Dict with 'code', 'name', 'domicile', and 'cohort'.
    """
    code = broker_code.upper().strip()
    if registry:
        for item in registry:
            if item.get("code", "").upper() == code:
                return {
                    "code": code,
                    "name": item.get("name", f"Broker {code}"),
                    "domicile": item.get("domicile", "DOMESTIC"),
                    "cohort": item.get("cohort", "RETAIL"),
                }

    # Standard IDX heuristics if registry not available
    foreign_institutional = {"CS", "ZP", "AK", "BK", "CG", "KZ", "MS", "RX", "YU"}
    domestic_institutional = {"CC", "DX", "LG", "NI", "OD"}
    domestic_retail = {"YP", "PD", "XC", "XL", "SQ", "KK"}

    if code in foreign_institutional:
        return {"code": code, "name": f"Broker {code}", "domicile": "FOREIGN", "cohort": "INSTITUTION"}
    if code in domestic_institutional:
        return {"code": code, "name": f"Broker {code}", "domicile": "DOMESTIC", "cohort": "INSTITUTION"}
    if code in domestic_retail:
        return {"code": code, "name": f"Broker {code}", "domicile": "DOMESTIC", "cohort": "RETAIL"}

    return {"code": code, "name": f"Broker {code}", "domicile": "DOMESTIC", "cohort": "UNKNOWN"}
