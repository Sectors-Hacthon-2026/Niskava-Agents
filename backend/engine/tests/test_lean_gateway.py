"""Tests for Task 1: Gateway Primitive Methods in NiskavaToolRegistry.

Verifies that query_sectors, search_osint, and query_memory correctly
route to the underlying client/harvester/memory layer using mocks.
"""

from __future__ import annotations

from typing import Any
from unittest.mock import MagicMock, patch

import pytest

from engine.agent.tools import NiskavaToolRegistry


@pytest.fixture()
def registry(tmp_path) -> NiskavaToolRegistry:
    """Return a fully-mocked NiskavaToolRegistry bound to a tmp db path."""
    db = str(tmp_path / "test.db")
    reg = NiskavaToolRegistry(
        db_path=db,
        sectors_client=MagicMock(),
        osint_harvester=MagicMock(),
        skills_registry=MagicMock(),
        memory=MagicMock(),
    )
    return reg


# ---------------------------------------------------------------------------
# query_sectors routing tests
# ---------------------------------------------------------------------------

class TestQuerySectors:
    def test_candles_domain_routes_to_get_daily_candles(self, registry):
        registry.sectors_client.get_daily_candles.return_value = [{"date": "2026-09-01", "close": 1500}]
        result = registry.query_sectors(domain="candles", ticker="ANTM")
        registry.sectors_client.get_daily_candles.assert_called_once_with("ANTM")
        assert result == [{"date": "2026-09-01", "close": 1500}]

    def test_fundamentals_domain_routes_to_get_company_report(self, registry):
        registry.sectors_client.get_company_report.return_value = {"company_name": "Aneka Tambang"}
        result = registry.query_sectors(domain="fundamentals", ticker="ANTM")
        registry.sectors_client.get_company_report.assert_called_once_with("ANTM")
        assert result["company_name"] == "Aneka Tambang"

    def test_foreign_flow_domain_routes_correctly(self, registry):
        registry.sectors_client.get_foreign_flow.return_value = [{"date": "2026-09-01", "net": 1_000_000}]
        result = registry.query_sectors(domain="foreign_flow", ticker="ANTM")
        registry.sectors_client.get_foreign_flow.assert_called_once_with("ANTM")

    def test_suspensions_domain_routes_correctly(self, registry):
        registry.sectors_client.get_suspensions.return_value = []
        result = registry.query_sectors(domain="suspensions", ticker="ANTM")
        registry.sectors_client.get_suspensions.assert_called_once_with("ANTM")

    def test_filings_domain_routes_correctly(self, registry):
        registry.sectors_client.get_filings.return_value = []
        result = registry.query_sectors(domain="filings", ticker="ANTM")
        registry.sectors_client.get_filings.assert_called_once_with("ANTM")

    def test_broker_summary_domain_routes_correctly(self, registry):
        registry.sectors_client.get_broker_summary.return_value = {"top_buyers": []}
        result = registry.query_sectors(domain="broker_summary", ticker="ANTM")
        registry.sectors_client.get_broker_summary.assert_called_once_with("ANTM")

    def test_corporate_actions_domain_routes_correctly(self, registry):
        registry.sectors_client.get_corporate_actions.return_value = []
        result = registry.query_sectors(domain="corporate_actions", ticker="ANTM")
        registry.sectors_client.get_corporate_actions.assert_called_once_with("ANTM")

    def test_ticker_is_uppercased(self, registry):
        registry.sectors_client.get_daily_candles.return_value = []
        registry.query_sectors(domain="candles", ticker="antm")
        registry.sectors_client.get_daily_candles.assert_called_once_with("ANTM")

    def test_unknown_domain_raises_value_error(self, registry):
        with pytest.raises(ValueError, match="Domain tidak dikenal"):
            registry.query_sectors(domain="nonexistent_domain", ticker="ANTM")


# ---------------------------------------------------------------------------
# search_osint tests
# ---------------------------------------------------------------------------

class TestSearchOsint:
    def test_harvest_is_called_with_ticker(self, registry):
        registry.sectors_client.get_news.return_value = []
        registry.sectors_client.get_company_report.return_value = {"company_name": "Aneka Tambang"}
        mock_item = MagicMock()
        mock_item.model_dump.return_value = {"title": "ANTM news", "url": "http://example.com"}
        registry.osint_harvester.harvest.return_value = [mock_item]

        result = registry.search_osint(ticker="ANTM")

        registry.osint_harvester.harvest.assert_called_once()
        assert len(result) == 1
        assert result[0]["title"] == "ANTM news"

    def test_empty_ticker_harvests_general_market_news(self, registry):
        registry.sectors_client.get_news.return_value = []
        mock_item = MagicMock()
        mock_item.model_dump.return_value = {"title": "Market headlines"}
        registry.osint_harvester.harvest.return_value = [mock_item]

        result = registry.search_osint(ticker="")

        registry.osint_harvester.harvest.assert_called_once()
        call_kwargs = registry.osint_harvester.harvest.call_args
        # General market harvest uses IHSG as ticker
        assert (
            call_kwargs.kwargs.get("ticker") == "IHSG"
            or (call_kwargs.args and call_kwargs.args[0] == "IHSG")
        )


# ---------------------------------------------------------------------------
# query_memory tests
# ---------------------------------------------------------------------------

class TestQueryMemory:
    def test_delegates_to_memory_retrieve_ego_subgraph(self, registry):
        registry.memory.retrieve_ego_subgraph.return_value = {
            "query": "ANTM",
            "nodes": [{"label": "ANTM"}],
            "edges": [],
        }
        result = registry.query_memory(concept_or_ticker="ANTM")
        registry.memory.retrieve_ego_subgraph.assert_called_once_with(
            entity_query="ANTM", radius=2
        )
        assert result["nodes_found"] == 1
        assert result["query"] == "ANTM"

    def test_custom_radius_is_passed_through(self, registry):
        registry.memory.retrieve_ego_subgraph.return_value = {
            "query": "ANTM", "nodes": [], "edges": []
        }
        registry.query_memory(concept_or_ticker="ANTM", radius=1)
        registry.memory.retrieve_ego_subgraph.assert_called_once_with(
            entity_query="ANTM", radius=1
        )
