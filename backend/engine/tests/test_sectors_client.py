"""Unit and Integration Tests for SectorsAPIClient.

Verifies:
- All 9 Sectors API v2 endpoints return structured data
- Law 5: Credit Budget Discipline & Local SQLite Caching
- Permanent caching for historical daily candles (T < today)
- TTL-based caching for fundamental reports and news
"""

import os
import sqlite3
import pytest
from engine.sectors.client import SectorsAPIClient


@pytest.fixture
def client(tmp_path):
    db_file = str(tmp_path / "test_sectors.db")
    return SectorsAPIClient(db_path=db_file, mock_mode=True)


def test_get_daily_candles(client):
    candles = client.get_daily_candles("ANTM")
    assert isinstance(candles, list)
    assert len(candles) == 30
    assert "open" in candles[0]
    assert "close" in candles[0]
    assert "volume" in candles[0]
    assert "date" in candles[0]


def test_get_company_report(client):
    report = client.get_company_report("ANTM")
    assert isinstance(report, dict)
    assert report["symbol"] == "ANTM"
    assert "company_name" in report
    assert "pe_ratio" in report


def test_get_foreign_flow(client):
    flow = client.get_foreign_flow("ANTM")
    assert isinstance(flow, list)
    assert len(flow) >= 1
    assert "net_foreign_buy" in flow[0]


def test_get_news(client):
    news = client.get_news("ANTM")
    assert isinstance(news, list)
    assert len(news) >= 1
    assert "title" in news[0]
    assert "url" in news[0]


def test_get_suspensions(client):
    suspensions = client.get_suspensions("ANTM")
    assert isinstance(suspensions, list)
    assert len(suspensions) >= 1
    assert "pdf_url" in suspensions[0]
    assert "reason" in suspensions[0]


def test_get_corporate_actions(client):
    actions = client.get_corporate_actions("ANTM")
    assert isinstance(actions, list)
    assert len(actions) >= 1
    assert actions[0]["action_type"] == "DIVIDEND"
    assert "cum_date" in actions[0]


def test_get_filings(client):
    filings = client.get_filings("ANTM")
    assert isinstance(filings, list)
    assert len(filings) >= 1
    assert "insider_name" in filings[0]
    assert "shares" in filings[0]


def test_get_broker_summary(client):
    summary = client.get_broker_summary("ANTM")
    assert isinstance(summary, dict)
    assert "top_buyers" in summary
    assert "top_sellers" in summary
    assert len(summary["top_buyers"]) >= 1


def test_get_subsector_peers(client):
    peers = client.get_subsector_peers("metals-and-minerals-mining")
    assert isinstance(peers, dict)
    assert "peer_count" in peers
    assert "peers" in peers


def test_get_mining_detail(client):
    detail = client.get_mining_detail("aneka-tambang")
    assert isinstance(detail, dict)
    assert detail["commodity"] == "NICKEL"
    assert "smelter_count" in detail


def test_get_commodity_price(client):
    prices = client.get_commodity_price("nickel")
    assert isinstance(prices, list)
    assert len(prices) >= 1
    assert "price" in prices[0]
    assert "date" in prices[0]


def test_get_quarterly_financials(client):
    fin = client.get_quarterly_financials("ANTM")
    assert isinstance(fin, list)
    assert len(fin) >= 1
    assert "current_assets" in fin[0]
    assert "current_liabilities" in fin[0]


def test_get_broker_registry(client):
    reg = client.get_broker_registry()
    assert isinstance(reg, list)
    assert len(reg) >= 1
    assert "code" in reg[0]
    assert "cohort" in reg[0]


def test_get_subsectors(client):
    subs = client.get_subsectors()
    assert isinstance(subs, list)
    assert len(subs) >= 1
    assert "subsector" in subs[0]


def test_sqlite_caching_law_5(client, tmp_path):
    """Verify Law 5: Caching prevents redundant HTTP calls."""
    # First call: populates cache
    client.get_daily_candles("ANTM")
    client.get_company_report("ANTM")

    # Inspect SQLite database directly
    with sqlite3.connect(client.db_path) as conn:
        cursor = conn.cursor()
        # Candlestick data must have expires_at = NULL (permanent)
        cursor.execute("SELECT expires_at FROM sectors_cache WHERE endpoint LIKE '%/daily/%'")
        candle_rows = cursor.fetchall()
        assert len(candle_rows) >= 1
        assert candle_rows[0][0] is None, "Historical candlestick must have permanent cache (expires_at is NULL)"

        # Company report must have TTL (expires_at is NOT NULL)
        cursor.execute("SELECT expires_at FROM sectors_cache WHERE endpoint LIKE '%/company/report/%'")
        report_rows = cursor.fetchall()
        assert len(report_rows) >= 1
        assert report_rows[0][0] is not None, "Company report must have TTL set"

    # Second call: served directly from cache
    cached_candles = client.get_daily_candles("ANTM")
    assert len(cached_candles) == 30


def test_sectors_api_error_raised_in_online_mode(tmp_path):
    """Verify that in online mode (mock_mode=False), HTTP errors raise SectorsAPIError instead of returning mock data."""
    db_file = str(tmp_path / "test_sectors_live.db")
    live_client = SectorsAPIClient(
        db_path=db_file,
        api_key="invalid-test-key",
        mock_mode=False,
        base_url="https://127.0.0.1:9999",  # unreachable endpoint
    )

    with pytest.raises(Exception) as exc_info:
        live_client.get_company_report("ANTM")

    # In online mode, it must raise SectorsAPIError and NOT return mock data silently
    assert "SectorsAPIError" in type(exc_info.value).__name__


def test_force_refresh_bypasses_cache(client, monkeypatch):
    """Verify that force_refresh=True forces a fresh request and updates cache."""
    # First call: populates cache
    initial = client.get_daily_candles("ANTM")
    assert len(initial) == 30

    # Spy or count _request calls
    call_count = 0
    original_generate = client._generate_mock_data

    def counting_generate(endpoint, params=None):
        nonlocal call_count
        call_count += 1
        return original_generate(endpoint, params)

    monkeypatch.setattr(client, "_generate_mock_data", counting_generate)

    # Calling without force_refresh uses cache (no new mock_data generation)
    cached = client.get_daily_candles("ANTM", force_refresh=False)
    assert len(cached) == 30
    assert call_count == 0

    # Calling with force_refresh=True bypasses cache
    refreshed = client.get_daily_candles("ANTM", force_refresh=True)
    assert len(refreshed) == 30
    assert call_count == 1


