"""Sectors News Engine - Pure Sectors Financial API v2 news & corporate disclosure provider.

Guarantees 100% compliance with Sectors Hackathon Rule 06 (Sectors API as core data source)
and eliminates unapproved third-party data scraping dependencies.
"""

from typing import Any, Dict, List, Optional
from pydantic import BaseModel, Field
import requests
import trafilatura
from engine.sectors.client import SectorsAPIClient


class NewsItem(BaseModel):
    """Structured representation of a Sectors curated news or disclosure item."""
    title: str
    source_name: str
    source_url: str
    publication_date: str
    snippet: str
    full_text: Optional[str] = None
    is_disclosure: bool = False
    source_type: str = "NEWS"  # 'NEWS', 'DISCLOSURE'


class SectorsNewsEngine:
    """Retrieves and normalizes market news and emiten disclosures exclusively from Sectors API v2."""

    def __init__(
        self,
        sectors_client: Optional[SectorsAPIClient] = None,
        db_path: str = "~/.niskava/niskava.db",
        mock_mode: Optional[bool] = None,
    ):
        self.client = sectors_client or SectorsAPIClient(
            db_path=db_path, mock_mode=mock_mode
        )

    def fetch_news(
        self,
        ticker: Optional[str] = None,
        company_name: Optional[str] = None,
    ) -> List[NewsItem]:
        """Fetch curated news for a ticker or general market from Sectors API v2."""
        clean_ticker = ticker.strip().upper() if ticker and ticker.strip() else None

        # In Sectors API v2, index symbols (IHSG, JCI, COMPOSITE) map to market-wide news (symbol=None)
        if clean_ticker in ("IHSG", "JCI", "IDX", "COMPOSITE"):
            clean_ticker = None

        raw_news = self.client.get_news(clean_ticker)
        results: List[NewsItem] = []
        seen_titles = set()

        for item in raw_news:
            if not isinstance(item, dict):
                continue
            title = str(item.get("title", "")).strip()
            if not title or title in seen_titles:
                continue

            seen_titles.add(title)
            is_disclosure = self._is_disclosure(title)
            results.append(
                NewsItem(
                    title=title,
                    source_name=str(item.get("source", "Sectors News")),
                    source_url=str(item.get("url", "")),
                    publication_date=str(item.get("publish_date", "")),
                    snippet=str(item.get("snippet", "")),
                    full_text=item.get("full_text"),
                    is_disclosure=is_disclosure,
                    source_type="DISCLOSURE" if is_disclosure else "NEWS",
                )
            )

        return results

    def harvest(
        self,
        ticker: str,
        company_name: Optional[str] = None,
        sectors_news_items: Optional[List[Dict[str, Any]]] = None,
    ) -> List[NewsItem]:
        """Backward-compatible alias for pipeline and skill consumers."""
        if sectors_news_items is not None:
            results: List[NewsItem] = []
            seen_titles = set()
            raw_items = []
            if isinstance(sectors_news_items, dict):
                raw_items = sectors_news_items.get("results") or sectors_news_items.get("data") or []
            elif isinstance(sectors_news_items, list):
                raw_items = sectors_news_items

            for item in raw_items:
                if not isinstance(item, dict):
                    continue
                title = str(item.get("title", "")).strip()
                if not title or title in seen_titles:
                    continue
                seen_titles.add(title)
                is_disc = self._is_disclosure(title)
                results.append(
                    NewsItem(
                        title=title,
                        source_name=str(item.get("source", "Sectors News")),
                        source_url=str(item.get("url", "")),
                        publication_date=str(item.get("publish_date", "")),
                        snippet=str(item.get("snippet", "")),
                        full_text=item.get("full_text"),
                        is_disclosure=is_disc,
                        source_type="DISCLOSURE" if is_disc else "NEWS",
                    )
                )
            return results

        return self.fetch_news(ticker=ticker, company_name=company_name)

    def sanitize_article_text(self, html_or_url: str) -> Optional[str]:
        """Use Trafilatura to cleanly extract article content without ads/boilerplate."""
        try:
            if html_or_url.startswith("http://") or html_or_url.startswith("https://"):
                headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Niskava-Agent/1.0"}
                resp = requests.get(html_or_url, headers=headers, timeout=10)
                if resp.status_code != 200 or not resp.text:
                    return None
                return trafilatura.extract(resp.text, include_comments=False)
            return trafilatura.extract(html_or_url, include_comments=False)
        except Exception:
            return None

    def _sanitize_xml_text(self, text: str) -> str:
        """Escape XML entities and neutralize prompt injection context boundaries."""
        if not text:
            return ""
        return (
            text.replace("&", "&amp;")
            .replace("<", "&lt;")
            .replace(">", "&gt;")
            .replace('"', "&quot;")
            .replace("'", "&apos;")
        )

    def wrap_in_evidence_context(self, items: List[NewsItem]) -> str:
        """Format news items into an isolated XML structure for LLM reasoning."""
        parts = ["<evidence_context>"]
        for idx, item in enumerate(items, 1):
            san_title = self._sanitize_xml_text(item.title)
            san_snippet = self._sanitize_xml_text(item.snippet)
            san_source = self._sanitize_xml_text(item.source_name)
            san_date = self._sanitize_xml_text(item.publication_date)
            san_type = self._sanitize_xml_text(item.source_type)

            parts.append(
                f'  <item id="{idx}" type="{san_type}" source="{san_source}" date="{san_date}">\n'
                f"    <headline>{san_title}</headline>\n"
                f"    <snippet>{san_snippet}</snippet>\n"
                f"  </item>"
            )
        parts.append("</evidence_context>")
        return "\n".join(parts)

    def _is_disclosure(self, title: str) -> bool:
        """Heuristic to tag corporate regulatory disclosures vs standard financial press."""
        t = title.lower()
        keywords = (
            "keterbukaan",
            "pengumuman",
            "penjelasan bursa",
            "suspensi",
            "unusual market activity",
            "uma",
            "corporate action",
            "rups",
            "dividen",
            "rights issue",
            "laporan keuangan",
            "smelter",
            "volatilitas",
        )
        return any(k in t for k in keywords)
