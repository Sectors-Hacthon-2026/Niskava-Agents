"""Test bahwa query_sectors dengan ticker indeks dan domain kosong
mengembalikan respons yang diperkaya (news + context) bukan hanya list news mentah.

Memverifikasi fix: model tidak lagi mendapat data yang membingungkan
saat user bertanya tentang 'market hari ini'.
"""
import pytest

from engine.agent.tools import NiskavaToolRegistry, _INDEX_TICKERS


def _make_registry(tmp_path) -> NiskavaToolRegistry:
    return NiskavaToolRegistry(
        db_path=str(tmp_path / "enrich_test.db"), mock_mode=True
    )


@pytest.mark.parametrize("idx_ticker", sorted(_INDEX_TICKERS))
def test_index_ticker_empty_domain_returns_enriched_dict(tmp_path, idx_ticker):
    """query_sectors(domain='', ticker=idx_ticker) harus mengembalikan dict
    berisi key 'type', 'news', dan 'market_context'."""
    registry = _make_registry(tmp_path)
    result = registry.query_sectors(domain="", ticker=idx_ticker)

    assert isinstance(result, dict), (
        f"Expected dict for index ticker '{idx_ticker}', got {type(result)}"
    )
    assert result.get("type") == "market_overview", (
        f"Expected type='market_overview', got '{result.get('type')}'"
    )
    assert "news" in result, "Result must contain 'news' key"
    assert isinstance(result["news"], list), "'news' must be a list"
    assert "market_context" in result, "Result must contain 'market_context' key"


def test_explicit_news_domain_index_ticker_returns_raw_list(tmp_path):
    """query_sectors(domain='news', ticker='IHSG') harus tetap mengembalikan
    list mentah (backward compatibility)."""
    registry = _make_registry(tmp_path)
    result = registry.query_sectors(domain="news", ticker="IHSG")

    # Explicit domain=news should return raw list as before
    assert isinstance(result, list), (
        f"Explicit domain='news' should return list, got {type(result)}"
    )


def test_normal_ticker_empty_domain_returns_candles_list(tmp_path):
    """query_sectors(domain='', ticker='ANTM') harus mengembalikan list
    candles (bukan enriched dict) — non-index ticker tidak terpengaruh."""
    registry = _make_registry(tmp_path)
    result = registry.query_sectors(domain="", ticker="ANTM")

    assert isinstance(result, list), (
        f"Non-index ticker should return list candles, got {type(result)}"
    )
