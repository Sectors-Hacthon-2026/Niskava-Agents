"""Tests for dynamic multi-sector mock data generator in SectorsAPIClient."""

from datetime import datetime
from engine.sectors.client import SectorsAPIClient


def test_mock_news_generates_today_date_and_multi_sector(tmp_path):
    client = SectorsAPIClient(db_path=str(tmp_path / "test.db"), mock_mode=True)
    news = client.get_news(None)
    
    assert len(news) >= 3
    today_str = datetime.now().strftime("%Y-%m-%d")
    # At least one article must have today's publication date
    assert any(today_str in item.get("publish_date", "") for item in news)
    
    # Must cover multiple sectors (e.g. Banking, Mining, Energy/Telco)
    titles = " ".join(item.get("title", "") for item in news)
    assert any(term in titles for term in ("IHSG", "Perbankan", "Tambang", "Smelter", "Energi"))


def test_mock_subsectors_endpoint(tmp_path):
    client = SectorsAPIClient(db_path=str(tmp_path / "test.db"), mock_mode=True)
    subsectors = client.get_subsectors()
    assert isinstance(subsectors, list)
    assert len(subsectors) > 0
    assert any("metals" in str(s).lower() or "bank" in str(s).lower() for s in subsectors)
