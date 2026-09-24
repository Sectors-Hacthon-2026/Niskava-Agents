"""Tests for SectorsNewsEngine ensuring 100% Sectors API data purity."""

from unittest.mock import MagicMock
import pytest
from engine.sectors.client import SectorsAPIClient
from engine.sectors.news_engine import SectorsNewsEngine, NewsItem


def test_news_item_model():
    item = NewsItem(
        title="ANTM Catat Kenaikan Laba Bersih",
        source_name="Sectors News",
        source_url="https://sectors.app/news/123",
        publication_date="2026-09-20",
        snippet="PT Aneka Tambang Tbk membukukan lonjakan kinerja...",
        is_disclosure=False,
    )
    assert item.title == "ANTM Catat Kenaikan Laba Bersih"
    assert item.source_type == "NEWS"
    assert item.is_disclosure is False


def test_fetch_news_for_ticker(tmp_path):
    client = SectorsAPIClient(db_path=str(tmp_path / "cache.db"), api_key="test-key")
    client.get_news = MagicMock(return_value=[
        {
            "title": "ANTM Umumkan Smelter Baru",
            "source": "Sectors Media",
            "url": "https://sectors.app/news/456",
            "publish_date": "2026-09-22",
            "snippet": "Pembangunan smelter nikel rampung...",
        }
    ])

    engine = SectorsNewsEngine(sectors_client=client)
    items = engine.fetch_news(ticker="ANTM")

    assert len(items) == 1
    assert items[0].title == "ANTM Umumkan Smelter Baru"
    assert items[0].source_name == "Sectors Media"
    assert items[0].source_url == "https://sectors.app/news/456"
    assert items[0].publication_date == "2026-09-22"
    client.get_news.assert_called_once_with("ANTM")


def test_fetch_news_market_overview(tmp_path):
    client = SectorsAPIClient(db_path=str(tmp_path / "cache.db"), api_key="test-key")
    client.get_news = MagicMock(return_value=[
        {
            "title": "IHSG Menghijau Ditopang Sektor Tambang",
            "source": "IDX Curated",
            "url": "https://sectors.app/news/789",
            "publish_date": "2026-09-23",
            "snippet": "Indeks Harga Saham Gabungan menguat 0.8%...",
        }
    ])

    engine = SectorsNewsEngine(sectors_client=client)
    items = engine.fetch_news(ticker=None)

    assert len(items) == 1
    assert items[0].title == "IHSG Menghijau Ditopang Sektor Tambang"
    client.get_news.assert_called_once_with(None)


def test_fetch_news_handles_empty_and_malformed(tmp_path):
    client = SectorsAPIClient(db_path=str(tmp_path / "cache.db"), api_key="test-key")
    client.get_news = MagicMock(return_value=[
        "malformed_string_entry",
        None,
        {"title": "", "source": ""},
        {"title": "Valid News", "source": "Sectors", "url": "https://sectors.app", "publish_date": "2026-09-24"},
    ])

    engine = SectorsNewsEngine(sectors_client=client)
    items = engine.fetch_news(ticker="ANTM")

    assert len(items) == 1
    assert items[0].title == "Valid News"
