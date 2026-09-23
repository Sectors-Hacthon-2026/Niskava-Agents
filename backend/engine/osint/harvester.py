"""Resilient Dual-Engine OSINT Harvester for IDX Disclosures and Market News.

Implements Decision 07 (Resilient Dual-Engine OSINT Architecture):
- Engine 1: Sectors v2 Curated News API
- Engine 2: Google News RSS Search Engine (targeted secondary disclosure dorking)
- Sanitizer: Trafilatura HTML cleaner
- Injection Defense: XML-wrapped <evidence_context> per Hermes pattern
"""

import os
import urllib.parse
from typing import Any, Dict, List, Optional
import feedparser
from pydantic import BaseModel, Field
import requests
import trafilatura


class OSINTItem(BaseModel):
    """Structured representation of an OSINT news or disclosure item."""
    title: str
    source_name: str
    source_url: str
    publication_date: str
    snippet: str
    full_text: Optional[str] = None
    is_disclosure: bool = False
    source_type: str = "NEWS"  # 'NEWS', 'DISCLOSURE', 'COMMUNITY'


class DualEngineOSINTHarvester:
    """Orchestrates news harvesting across Sectors v2 News and Google News RSS."""

    GOOGLE_NEWS_RSS_BASE = "https://news.google.com/rss/search"

    def __init__(self, mock_mode: Optional[bool] = None):
        if mock_mode is not None:
            self.mock_mode = mock_mode
        else:
            self.mock_mode = (
                os.environ.get("MOCK_SECTORS", "0") in ("1", "true", "True")
                or os.environ.get("NISKAVA_OFFLINE", "0") in ("1", "true", "True")
            )

    def harvest(
        self,
        ticker: str,
        company_name: Optional[str] = None,
        sectors_news_items: Optional[List[Dict[str, Any]]] = None,
    ) -> List[OSINTItem]:
        """Harvest OSINT items from both engines and deduplicate."""
        if self.mock_mode:
            return self._generate_mock_osint(ticker)

        results: List[OSINTItem] = []
        seen_titles = set()

        # 1. Ingest Engine 1 items (Sectors v2 Curated News)
        raw_items: List[Any] = []
        if isinstance(sectors_news_items, dict):
            dict_items = sectors_news_items.get("results") or sectors_news_items.get("data") or []
            if isinstance(dict_items, list):
                raw_items = dict_items
        elif isinstance(sectors_news_items, list):
            raw_items = sectors_news_items

        for item in raw_items:
            if not isinstance(item, dict):
                continue
            title = str(item.get("title", "")).strip()
            if title and title not in seen_titles:
                seen_titles.add(title)
                results.append(
                    OSINTItem(
                        title=title,
                        source_name=str(item.get("source", "Sectors News")),
                        source_url=str(item.get("url", "")),
                        publication_date=str(item.get("publish_date", "")),
                        snippet=str(item.get("snippet", "")),
                        is_disclosure=self._is_disclosure_headline(title),
                        source_type="DISCLOSURE" if self._is_disclosure_headline(title) else "NEWS",
                    )
                )

        # 2. Ingest Engine 2 items (Google News RSS with targeted secondary disclosure dorking)
        rss_items = self._fetch_google_news_rss(ticker, company_name)
        for item in rss_items:
            if item.title not in seen_titles:
                seen_titles.add(item.title)
                results.append(item)

        return results

    def _fetch_google_news_rss(
        self, ticker: str, company_name: Optional[str] = None
    ) -> List[OSINTItem]:
        """Fetch targeted syndication news from Google News RSS in Indonesia."""
        query_terms = [f'"{ticker}"']
        if company_name:
            # Add short company name if distinct
            short_name = company_name.replace("PT", "").replace("Tbk", "").strip()
            if len(short_name) > 2:
                query_terms.append(f'"{short_name}"')

        disclosure_keywords = '(keterbukaan OR bursa OR laba OR dividen OR smelter OR akuisisi)'
        raw_query = f"{' OR '.join(query_terms)} {disclosure_keywords}"
        encoded_query = urllib.parse.quote(raw_query)
        rss_url = f"{self.GOOGLE_NEWS_RSS_BASE}?q={encoded_query}&hl=id&gl=ID&ceid=ID:id"

        items: List[OSINTItem] = []
        try:
            headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Niskava-Agent/1.0"}
            resp = requests.get(rss_url, headers=headers, timeout=10)
            if resp.status_code == 200:
                feed = feedparser.parse(resp.content)
                for entry in feed.entries[:8]:  # Limit top 8 freshest syndications
                    title = getattr(entry, "title", "").strip()
                    link = getattr(entry, "link", "")
                    published = getattr(entry, "published", "")
                    summary = getattr(entry, "summary", "")

                    source_name = "Google News"
                    if hasattr(entry, "source") and hasattr(entry.source, "title"):
                        source_name = entry.source.title

                    is_disclosure = self._is_disclosure_headline(title)
                    items.append(
                        OSINTItem(
                            title=title,
                            source_name=source_name,
                            source_url=link,
                            publication_date=published,
                            snippet=summary,
                            is_disclosure=is_disclosure,
                            source_type="DISCLOSURE" if is_disclosure else "NEWS",
                        )
                    )
        except Exception:
            pass

        return items

    def sanitize_article_text(self, html_or_url: str) -> Optional[str]:
        """Use Trafilatura to cleanly extract main article content without ads/boilerplate."""
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
        escaped = (
            text.replace("&", "&amp;")
            .replace("<", "&lt;")
            .replace(">", "&gt;")
            .replace('"', "&quot;")
            .replace("'", "&apos;")
        )
        return escaped

    def wrap_in_evidence_context(self, items: List[OSINTItem]) -> str:
        """Format harvested items into an isolated XML structure.

        Mitigates indirect prompt injection by separating data context
        from system instructions and escaping all embedded XML tags.
        """
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

    def _is_disclosure_headline(self, title: str) -> bool:
        low = title.lower()
        keywords = ["keterbukaan", "penjelasan bursa", "volatilitas", "suspensi", "dividen", "rups", "smelter"]
        return any(k in low for k in keywords)

    def _generate_mock_osint(self, ticker: str) -> List[OSINTItem]:
        """Generate static mock OSINT items for tests and offline usage."""
        return [
            OSINTItem(
                title=f"{ticker} Resmikan Uji Coba Smelter Feronikel Baru di Halmahera Timur",
                source_name="IDX Channel",
                source_url="https://idxchannel.com/market/antm-smelter-halmahera",
                publication_date="2026-09-12T07:30:00Z",
                snippet="PT Aneka Tambang Tbk (ANTM) mengumumkan penyelesaian proyek hilirisasi nikel berkapasitas 13.500 TNi per tahun.",
                is_disclosure=True,
                source_type="DISCLOSURE",
            ),
            OSINTItem(
                title=f"Harga Komoditas Nikel Menguat, Saham {ticker} Melesat 8%",
                source_name="CNBC Indonesia",
                source_url="https://cnbcindonesia.com/market/antm-nikel-rally",
                publication_date="2026-09-12T10:15:00Z",
                snippet="Sentimen reli komoditas nikel di London Metal Exchange memberikan angin segar bagi saham-saham tambang BUMN.",
                is_disclosure=False,
                source_type="NEWS",
            ),
        ]
