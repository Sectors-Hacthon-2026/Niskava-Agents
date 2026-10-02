"""Tests for SectorsAPIClient market screening endpoints (top-changes & most-traded)."""
import os
import tempfile
import pytest
from engine.sectors.client import SectorsAPIClient


@pytest.fixture
def mock_client():
    with tempfile.NamedTemporaryFile(suffix=".db") as tmp:
        client = SectorsAPIClient(db_path=tmp.name, mock_mode=True)
        yield client


def test_get_top_changes_mock_gainers(mock_client):
    data = mock_client.get_top_changes(classification="top_gainers", period="1d")
    assert isinstance(data, list)
    assert len(data) > 0
    first = data[0]
    assert "symbol" in first
    assert "change" in first or "price_change" in first or "return" in first


def test_get_top_changes_mock_losers(mock_client):
    data = mock_client.get_top_changes(classification="top_losers", period="1d")
    assert isinstance(data, list)
    assert len(data) > 0


def test_get_most_traded_mock(mock_client):
    data = mock_client.get_most_traded(n_stock=5)
    assert isinstance(data, list)
    assert len(data) > 0
    first = data[0]
    assert "symbol" in first
    assert "volume" in first or "total_volume" in first or "turnover" in first
