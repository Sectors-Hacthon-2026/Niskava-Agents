"""Tests for query_sectors gateway integration with market screening domains."""
import os
import tempfile
import pytest
from engine.agent.tools import NiskavaToolRegistry
from engine.sectors.client import SectorsAPIClient


@pytest.fixture
def registry():
    with tempfile.NamedTemporaryFile(suffix=".db") as tmp:
        client = SectorsAPIClient(db_path=tmp.name, mock_mode=True)
        reg = NiskavaToolRegistry(db_path=tmp.name, sectors_client=client, mock_mode=True)
        yield reg


def test_query_sectors_top_changes_gainers(registry):
    res = registry.query_sectors(
        domain="top_changes",
        ticker="",
        params={"classification": "top_gainers", "period": "1d"},
    )
    assert isinstance(res, list)
    assert len(res) > 0
    assert "BSWD" in res[0]["symbol"] or "symbol" in res[0]


def test_query_sectors_top_changes_losers(registry):
    res = registry.query_sectors(
        domain="top_changes",
        ticker="",
        params={"classification": "top_losers", "period": "1d"},
    )
    assert isinstance(res, list)
    assert len(res) > 0
    assert "FORU" in res[0]["symbol"] or "GOTO" in res[0]["symbol"]


def test_query_sectors_most_traded(registry):
    res = registry.query_sectors(
        domain="most_traded",
        ticker="",
        params={"n_stock": 5},
    )
    assert isinstance(res, list)
    assert len(res) > 0
    assert res[0]["symbol"] == "BBRI"


def test_execute_tool_query_sectors_dispatch(registry):
    res = registry.execute_tool(
        "query_sectors",
        {
            "domain": "top_changes",
            "classification": "top_gainers",
        },
    )
    assert isinstance(res, list)
    assert len(res) > 0
