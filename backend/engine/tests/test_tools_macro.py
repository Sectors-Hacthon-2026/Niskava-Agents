"""Tests for macro sector and query routing in NiskavaToolRegistry."""

from engine.agent.tools import NiskavaToolRegistry


def test_search_news_forwards_query(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "cache.db"), mock_mode=True)
    results = registry.search_news(ticker="", query="perbankan")
    assert isinstance(results, list)
    assert len(results) > 0
    assert any("Perbankan" in r.get("title", "") or "IHSG" in r.get("title", "") for r in results)


def test_query_sectors_subsectors_domain(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "cache.db"), mock_mode=True)
    res = registry.query_sectors(domain="subsectors", ticker="")
    assert isinstance(res, list)
    assert len(res) > 0
    assert any(s.get("subsector") == "banks" or "metals" in str(s) for s in res)


def test_query_sectors_market_overview_on_index_ticker(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "cache.db"), mock_mode=True)
    res = registry.query_sectors(domain="", ticker="IDX")
    assert isinstance(res, dict)
    assert res.get("type") == "market_overview"
    assert len(res.get("news", [])) > 0
