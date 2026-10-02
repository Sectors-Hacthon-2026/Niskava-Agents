"""Sectors Financial API v2 package."""

from engine.sectors.client import SectorsAPIClient, SectorsAPIError
from engine.sectors.news_engine import SectorsNewsEngine, NewsItem

__all__ = ["SectorsAPIClient", "SectorsAPIError", "SectorsNewsEngine", "NewsItem"]

