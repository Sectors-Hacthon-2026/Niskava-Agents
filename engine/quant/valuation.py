"""Deterministic subsector relative valuation and peer benchmarking using pure NumPy.

Abides by Law 1 (Deterministic Before Generative) and Law 2 (Non-Advisory Boundary):
Evaluates valuation percentiles (Median, IQR, Z-val) without ever generating
price targets or buy/sell recommendations.
"""

from typing import Any, Dict, List, Tuple
import numpy as np


def compute_subsector_percentile(
    target_val: float,
    peer_values: List[float],
) -> Tuple[float, float, float, str]:
    """Compute peer distribution median, IQR, and target valuation z-score.

    Formula:
        z_val = (target_val - median) / IQR

    Args:
        target_val: Metric for target ticker (e.g. P/E of 12.4).
        peer_values: Metric values across all subsector peers.

    Returns:
        tuple[median, iqr, z_val, valuation_posture]
    """
    valid_peers = [float(v) for v in peer_values if v is not None and not np.isnan(float(v)) and float(v) > 0]
    if not valid_peers:
        return target_val, 0.0, 0.0, "INSUFFICIENT_PEER_DATA"

    arr = np.array(valid_peers, dtype=np.float64)
    median = float(np.median(arr))
    q75, q25 = np.percentile(arr, [75, 25])
    iqr = float(q75 - q25)

    if iqr > 0:
        z_val = round((target_val - median) / iqr, 2)
    else:
        std = float(np.std(arr))
        z_val = round((target_val - median) / std, 2) if std > 0 else 0.0

    if z_val <= -0.5:
        posture = "TRADING_AT_DISCOUNT"
    elif z_val >= 0.5:
        posture = "TRADING_AT_PREMIUM"
    else:
        posture = "FAIRLY_ALIGNED_WITH_PEERS"

    return round(median, 2), round(iqr, 2), z_val, posture
