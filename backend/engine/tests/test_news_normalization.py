"""Targeted tests for Sectors News normalization and DualEngineOSINTHarvester defense."""

from unittest.mock import MagicMock
import pytest
from engine.sectors.client import SectorsAPIClient
from engine.osint.harvester import DualEngineOSINTHarvester


def test_sectors_client_get_news_dict_with_results(tmp_path):
    """Verify get_news() extracts list from {'results': [...]}."""
    client = SectorsAPIClient(db_path=str(tmp_path / "cache.db"), api_key="test-key")
    client._request = MagicMock(return_value={
        "results": [
            {"title": "Kinerja ANTM Meningkat", "source": "CNBC", "url": "https://cnbc.com", "publish_date": "2026-09-20"},
            "invalid_non_dict_element",
        ]
    })

    news = client.get_news("ANTM")
    assert isinstance(news, list)
    assert len(news) == 1
    assert news[0]["title"] == "Kinerja ANTM Meningkat"


def test_sectors_client_get_news_dict_with_data(tmp_path):
    """Verify get_news() extracts list from {'data': [...]}."""
    client = SectorsAPIClient(db_path=str(tmp_path / "cache.db"), api_key="test-key")
    client._request = MagicMock(return_value={
        "data": [
            {"title": "BBCA Tebar Dividen", "source": "Bisnis", "url": "https://bisnis.com", "publish_date": "2026-09-21"}
        ]
    })

    news = client.get_news("BBCA")
    assert isinstance(news, list)
    assert len(news) == 1
    assert news[0]["title"] == "BBCA Tebar Dividen"


def test_sectors_client_get_news_raw_list(tmp_path):
    """Verify get_news() returns filtered list when raw response is list."""
    client = SectorsAPIClient(db_path=str(tmp_path / "cache.db"), api_key="test-key")
    client._request = MagicMock(return_value=[
        {"title": "IHSG Ditutup Menguat", "source": "Kontan"},
        "bad_entry",
    ])

    news = client.get_news(None)
    assert isinstance(news, list)
    assert len(news) == 1
    assert news[0]["title"] == "IHSG Ditutup Menguat"


def test_sectors_client_get_news_malformed(tmp_path):
    """Verify get_news() handles unexpected non-dict/non-list safely."""
    client = SectorsAPIClient(db_path=str(tmp_path / "cache.db"), api_key="test-key")
    client._request = MagicMock(return_value="Unexpected String Error")
    assert client.get_news("ANTM") == []

    client._request = MagicMock(return_value={"error": "Not Found"})
    assert client.get_news("ANTM") == []


def test_harvester_with_dict_payload():
    """Verify DualEngineOSINTHarvester handles sectors_news_items passed as dict without 'str' object error."""
    harvester = DualEngineOSINTHarvester(mock_mode=False)
    harvester._fetch_google_news_rss = MagicMock(return_value=[])

    dict_payload = {
        "results": [
            {
                "title": "BREN Catat Rekor Tertinggi",
                "source": "IDX News",
                "url": "https://idx.co.id",
                "publish_date": "2026-09-22",
                "snippet": "Saham BREN melesat...",
            }
        ]
    }

    # Should not raise AttributeError: 'str' object has no attribute 'get'
    items = harvester.harvest(ticker="BREN", sectors_news_items=dict_payload)
    assert len(items) == 1
    assert items[0].title == "BREN Catat Rekor Tertinggi"
    assert items[0].source_name == "IDX News"


def test_harvester_with_malformed_items():
    """Verify DualEngineOSINTHarvester skips strings and non-dict items inside sectors_news_items."""
    harvester = DualEngineOSINTHarvester(mock_mode=False)
    harvester._fetch_google_news_rss = MagicMock(return_value=[])

    corrupted_list = [
        "just_a_string_headline",
        None,
        12345,
        {"title": "Valid News Item", "source": "Kontan", "url": "https://kontan.co.id"},
    ]

    items = harvester.harvest(ticker="ANTM", sectors_news_items=corrupted_list)
    assert len(items) == 1
    assert items[0].title == "Valid News Item"
