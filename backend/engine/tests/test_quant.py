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


def test_foreign_flow_z_score():
    import numpy as np
    from engine.quant.anomaly import compute_foreign_flow_z_score

    flows = np.array([10.0] * 10 + [20.0] * 10)
    mu, sigma, z = compute_foreign_flow_z_score(flows, 30.0)
    assert mu == 15.0
    assert sigma == 5.0
    assert z == 3.0


def test_bandarmology_quant():
    from engine.quant.bandarmology import compute_top3_buyer_concentration, classify_broker_cohort

    top_buyers = [
        {"broker_code": "CS", "net_buy_shares": 50_000_000},
        {"broker_code": "ZP", "net_buy_shares": 30_000_000},
        {"broker_code": "AK", "net_buy_shares": 20_000_000},
        {"broker_code": "YP", "net_buy_shares": 5_000_000},
    ]
    c3 = compute_top3_buyer_concentration(top_buyers, total_market_volume=140_000_000)
    assert round(c3, 3) == round(100_000_000 / 140_000_000, 3)

    cohort = classify_broker_cohort("CS")
    assert cohort["domicile"] == "FOREIGN"
    assert cohort["cohort"] == "INSTITUTION"


def test_financial_ratios_quant():
    from engine.quant.ratios import evaluate_financial_health

    health = evaluate_financial_health(
        current_assets=14_000_000,
        cash_and_equivalents=9_400_000,
        current_liabilities=5_000_000,
        total_debt=4_600_000,
        total_equity=24_000_000,
        ebit=3_200_000,
        interest_expense=380_000,
    )
    assert health["current_ratio"] == 2.8
    assert health["quick_ratio"] == 1.88
    assert health["liquidity_grade"] == "PRISTINE"
    assert health["der"] == 0.19
    assert health["solvency_grade"] == "STRONG"


def test_pearson_correlation_quant():
    from engine.quant.correlation import compute_pearson_correlation, align_and_correlate_series

    # Perfectly correlated
    x = [1.0, 2.0, 3.0, 4.0, 5.0]
    y = [2.0, 4.0, 6.0, 8.0, 10.0]
    r = compute_pearson_correlation(x, y)
    assert r == 1.0

    candles = [
        {"date": "2026-09-01", "close": 1000},
        {"date": "2026-09-02", "close": 1050},
        {"date": "2026-09-03", "close": 1100},
        {"date": "2026-09-04", "close": 1150},
    ]
    comm = [
        {"date": "2026-09-01", "price": 100},
        {"date": "2026-09-02", "price": 105},
        {"date": "2026-09-03", "price": 110},
        {"date": "2026-09-04", "price": 115},
    ]
    r_val, s_ret, c_ret, div_cls = align_and_correlate_series(candles, comm)
    assert r_val > 0.9
    assert div_cls == "COMMODITY_DRIVEN"


def test_subsector_percentile_valuation():
    from engine.quant.valuation import compute_subsector_percentile

    peers = [10.0, 12.0, 14.0, 16.0, 18.0, 20.0, 22.0]
    median, iqr, z_val, posture = compute_subsector_percentile(target_val=10.0, peer_values=peers)
    assert median == 16.0
    assert posture == "TRADING_AT_DISCOUNT"

