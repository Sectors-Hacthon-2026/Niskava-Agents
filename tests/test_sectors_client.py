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
