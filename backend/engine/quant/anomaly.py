"""Deterministic quantitative anomaly calculations using pure NumPy.

Abides by Law 1 (Deterministic Before Generative):
Never allow an LLM to calculate statistics, volume moving averages,
Z-scores, abnormal returns, or sector divergence.
"""

from dataclasses import asdict, dataclass
from typing import Any, Dict, List, Optional
import numpy as np


@dataclass(frozen=True)
class AnomalyResult:
    """Represents a single detected quantitative anomaly."""
    date: str
    metric_type: str  # 'VOLUME_SPIKE', 'PRICE_BREAKOUT', 'IDIOSYNCRATIC_CATALYST', etc.
    metric_value: float
    baseline_value: float
    z_score: float
    price_change_pct: float
    divergence_pct: float
    classification: str
    description: str

    def to_dict(self) -> Dict[str, Any]:
        return asdict(self)


def compute_volume_z_score(
    historical_volumes: np.ndarray,
    current_volume: float
) -> tuple[float, float, float]:
    """Compute rolling MA20 volume baseline and Volume Z-Score.

    Returns:
        tuple[mu_20, sigma_20, z_score]
    """
    if len(historical_volumes) == 0:
        return 0.0, 0.0, 0.0

    clean_volumes = np.nan_to_num(historical_volumes, nan=0.0, posinf=0.0, neginf=0.0)
    mu = float(np.mean(clean_volumes))
    sigma = float(np.std(clean_volumes))

    curr_v = 0.0 if (current_volume is None or np.isnan(current_volume) or np.isinf(current_volume)) else float(current_volume)

    if sigma > 0 and not np.isnan(sigma) and not np.isinf(sigma):
        z = (curr_v - mu) / sigma
    else:
        z = 0.0

    if np.isnan(z) or np.isinf(z):
        z = 0.0

    return (
        0.0 if np.isnan(mu) or np.isinf(mu) else mu,
        0.0 if np.isnan(sigma) or np.isinf(sigma) else sigma,
        z,
    )


def compute_foreign_flow_z_score(
    historical_flows: np.ndarray,
    current_flow: float
) -> tuple[float, float, float]:
    """Compute rolling MA20 foreign inflow baseline and Foreign Flow Z-Score (Fz).

    Returns:
        tuple[mu_20, sigma_20, f_z]
    """
    if len(historical_flows) == 0:
        return 0.0, 0.0, 0.0

    clean_flows = np.nan_to_num(historical_flows, nan=0.0, posinf=0.0, neginf=0.0)
    mu = float(np.mean(clean_flows))
    sigma = float(np.std(clean_flows))

    curr_f = 0.0 if (current_flow is None or np.isnan(current_flow) or np.isinf(current_flow)) else float(current_flow)

    if sigma > 0 and not np.isnan(sigma) and not np.isinf(sigma):
        z = (curr_f - mu) / sigma
    else:
        z = 0.0

    if np.isnan(z) or np.isinf(z):
        z = 0.0

    return (
        0.0 if np.isnan(mu) or np.isinf(mu) else mu,
        0.0 if np.isnan(sigma) or np.isinf(sigma) else sigma,
        z,
    )


def detect_historical_anomalies(
    daily_candles: List[Dict[str, Any]],
    sector_return_map: Optional[Dict[str, float]] = None,
    volume_z_threshold: float = 2.5,
    return_threshold_pct: float = 5.0,
    divergence_threshold_pct: float = 4.0,
    rolling_window: int = 20,
) -> List[AnomalyResult]:
    """Scan historical daily candlestick data for volume spikes and abnormal returns.

    Args:
        daily_candles: List of dicts ordered chronologically ascending, each having
                       'date', 'close', and 'volume'.
        sector_return_map: Optional mapping of date -> sector percentage return.
        volume_z_threshold: Z-Score threshold to trigger volume anomaly (default: 2.5).
        return_threshold_pct: Daily price return absolute percentage threshold (default: 5.0%).
        divergence_threshold_pct: Sector divergence threshold (default: 4.0%).
        rolling_window: Number of trading days for baseline volume (default: 20).

    Returns:
        List of AnomalyResult objects for dates where anomaly conditions were met.
    """
    if len(daily_candles) <= rolling_window:
        return []

    anomalies: List[AnomalyResult] = []
    sector_map = sector_return_map or {}

    def _safe_float(val: Any, default: float = 0.0) -> float:
        try:
            res = float(val)
            return default if (np.isnan(res) or np.isinf(res)) else res
        except (ValueError, TypeError):
            return default

    for i in range(rolling_window, len(daily_candles)):
        window = daily_candles[i - rolling_window : i]
        current_day = daily_candles[i]
        prev_day = daily_candles[i - 1]

        volumes = np.array([_safe_float(d.get("volume", 0.0)) for d in window], dtype=np.float64)
        v_t = _safe_float(current_day.get("volume", 0.0))
        mu_20, sigma_20, v_z = compute_volume_z_score(volumes, v_t)

        curr_close = _safe_float(current_day.get("close", 0.0))
        prev_close = _safe_float(prev_day.get("close", 0.0))

        if prev_close > 0:
            r_t = ((curr_close - prev_close) / prev_close) * 100.0
            if np.isnan(r_t) or np.isinf(r_t):
                r_t = 0.0
        else:
            r_t = 0.0

        target_date = str(current_day.get("date", ""))
        sec_ret = _safe_float(sector_map.get(target_date, 0.0))
        divergence = r_t - sec_ret

        is_volume_anomaly = v_z >= volume_z_threshold
        is_price_anomaly = abs(r_t) >= return_threshold_pct

        if is_volume_anomaly or is_price_anomaly:
            # Determine classification per documentation matrix
            if is_volume_anomaly and is_price_anomaly and abs(divergence) >= divergence_threshold_pct:
                classification = "IDIOSYNCRATIC_CATALYST"
                metric_type = "VOLUME_AND_PRICE_SURGE"
                desc = f"Unusual volume surge ({v_z:.2f}σ) with price breakout ({r_t:+.2f}%) divergent from sector ({divergence:+.2f}%)."
            elif is_volume_anomaly and not is_price_anomaly:
                classification = "VOLUME_ACCUMULATION"
                metric_type = "VOLUME_SPIKE"
                desc = f"Heavy volume spike ({v_z:.2f}σ) without matching price breakout ({r_t:+.2f}%)."
            elif not is_volume_anomaly and is_price_anomaly and abs(divergence) < divergence_threshold_pct:
                classification = "SECTOR_BETA_RALLY"
                metric_type = "PRICE_BREAKOUT"
                desc = f"Price move ({r_t:+.2f}%) driven largely by broader sector movement ({sec_ret:+.2f}%)."
            else:
                classification = "IDIOSYNCRATIC_CATALYST"
                metric_type = "PRICE_BREAKOUT" if is_price_anomaly else "VOLUME_SPIKE"
                desc = f"Abnormal move detected: Volume {v_z:.2f}σ, Price {r_t:+.2f}%."

            anomaly = AnomalyResult(
                date=target_date,
                metric_type=metric_type,
                metric_value=v_t if is_volume_anomaly else r_t,
                baseline_value=mu_20 if is_volume_anomaly else 0.0,
                z_score=round(float(v_z), 2),
                price_change_pct=round(float(r_t), 2),
                divergence_pct=round(float(divergence), 2),
                classification=classification,
                description=desc,
            )
            anomalies.append(anomaly)

    return anomalies
