"""Test guard untuk harvest_market_news dengan ticker IHSG atau kosong.

Memverifikasi Bug #2 fix: execute_tool harus mengembalikan general
market news (bukan crash) untuk ticker IHSG, '', atau None.
"""
import pytest
from engine.agent.tools import NiskavaToolRegistry

NON_COMPANY_TICKERS = ["IHSG", "JCI", "IDX", ""]


def _make_registry(tmp_path) -> NiskavaToolRegistry:
    return NiskavaToolRegistry(
        db_path=str(tmp_path / "guard_test.db"), mock_mode=True
    )


@pytest.mark.parametrize("bad_ticker", NON_COMPANY_TICKERS)
def test_harvest_market_news_does_not_crash_for_index_ticker(tmp_path, bad_ticker):
    """harvest_market_news via execute_tool tidak boleh crash untuk ticker index."""
    registry = _make_registry(tmp_path)
    result = registry.execute_tool(
        "harvest_market_news", {"ticker": bad_ticker}
    )
    assert isinstance(result, list), (
        f"Expected list, got {type(result)} for ticker='{bad_ticker}'"
    )


def test_harvest_market_news_direct_none_returns_list(tmp_path):
    """harvest_market_news(None) harus mengembalikan list (general news)."""
    registry = _make_registry(tmp_path)
    result = registry.harvest_market_news(ticker=None)
    assert isinstance(result, list)


def test_harvest_market_news_direct_ihsg_string_returns_list(tmp_path):
    """harvest_market_news('IHSG') harus mengembalikan list tanpa crash."""
    registry = _make_registry(tmp_path)
    result = registry.harvest_market_news(ticker="IHSG")
    assert isinstance(result, list)


def test_execute_tool_harvest_no_ticker_returns_list(tmp_path):
    """execute_tool('harvest_market_news', {}) tanpa ticker harus OK."""
    registry = _make_registry(tmp_path)
    result = registry.execute_tool("harvest_market_news", {})
    assert isinstance(result, list)
