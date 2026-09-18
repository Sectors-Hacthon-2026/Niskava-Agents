"""Unit tests for deterministic quant anomaly detection."""

import pytest
from engine.quant.anomaly import compute_volume_z_score, detect_historical_anomalies


def test_volume_z_score_calculation():
    import numpy as np

    # Baseline 20 days with uniform volume 1000
    volumes = np.full(20, 1000.0)
    mu, sigma, z = compute_volume_z_score(volumes, 1000.0)
    assert mu == 1000.0
    assert sigma == 0.0
    assert z == 0.0

    # With variance: 10 days of 800, 10 days of 1200 -> mean 1000, std 200
    volumes_var = np.array([800.0] * 10 + [1200.0] * 10)
    mu2, sigma2, z2 = compute_volume_z_score(volumes_var, 1500.0)
    assert mu2 == 1000.0
    assert sigma2 == 200.0
    assert z2 == 2.5  # Exactly 2.5 sigma


def test_detect_historical_anomalies_volume_spike():
    # 20 normal baseline days + 1 surge day
    daily_candles = []
    for i in range(20):
        daily_candles.append({
            "date": f"2026-08-{i+1:02d}",
            "close": 1500.0,
            "volume": 1000.0 if i % 2 == 0 else 1200.0,
        })
    
    # 21st day: massive volume spike (5000)
    daily_candles.append({
        "date": "2026-08-21",
        "close": 1620.0,  # +8.0% price jump
        "volume": 5000.0,
    })

    anomalies = detect_historical_anomalies(
        daily_candles,
        volume_z_threshold=2.5,
        return_threshold_pct=5.0,
    )

    assert len(anomalies) == 1
    assert anomalies[0].date == "2026-08-21"
    assert anomalies[0].z_score > 2.5
    assert anomalies[0].price_change_pct == 8.0
