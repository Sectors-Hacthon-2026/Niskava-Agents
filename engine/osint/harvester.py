"""Resilient Dual-Engine OSINT Harvester for IDX Disclosures and Market News.

Implements Decision 07 (Resilient Dual-Engine OSINT Architecture) & Multi-Level Fallback:
- Engine 1: Sectors v2 Curated News API
- Engine 2: Google News RSS Search Engine (targeted secondary disclosure dorking)
- Engine 3 (Fallback): DuckDuckGo HTML Web Dorking
- Sanitizer: Trafilatura HTML cleaner + Indonesian Media Regex Boilerplate Cleaner
- Injection Defense: XML-wrapped <evidence_context> per Hermes pattern
"""

import os
import re
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
    """Orchestrates news harvesting across Sectors v2 News, Google News RSS, and DuckDuckGo."""

    GOOGLE_NEWS_RSS_BASE = "https://news.google.com/rss/search"
    DUCKDUCKGO_HTML_BASE = "https://html.duckduckgo.com/html/"

    NOISE_PATTERNS = [
        r"Baca Juga:.*$",
        r"Simak Video:.*$",
        r"Pilihan Redaksi:.*$",
        r"ADVERTISEMENT.*$",
        r"SCROLL TO CONTINUE WITH CONTENT.*$",
        r"JAKARTA, \w+ --",
    ]

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
        """Harvest OSINT items using multi-level fallback engine."""
        if self.mock_mode:
            return self._generate_mock_osint(ticker)

        results: List[OSINTItem] = []
        seen_titles = set()

        # 1. Ingest Engine 1 items (Sectors v2 Curated News)
        if sectors_news_items:
            for item in sectors_news_items:
                title = item.get("title", "").strip()
                if title and title not in seen_titles:
                    seen_titles.add(title)
                    results.append(
                        OSINTItem(
                            title=title,
                            source_name=item.get("source", "Sectors News"),
                            source_url=item.get("url", ""),
                            publication_date=item.get("publish_date", ""),
                            snippet=self._clean_snippet(item.get("snippet", "") or title),
                            is_disclosure=self._is_disclosure_headline(title),
                            source_type="DISCLOSURE" if self._is_disclosure_headline(title) else "NEWS",
                        )
                    )

        # 2. Ingest Engine 2 items (Google News RSS)
        rss_items = self._fetch_google_news_rss(ticker, company_name)
        for item in rss_items:
            if item.title not in seen_titles:
                seen_titles.add(item.title)
                results.append(item)

        # 3. Fallback Engine 3: If results < 3, trigger DuckDuckGo HTML Fallback
        if len(results) < 3:
            ddg_items = self._fetch_duckduckgo_fallback(ticker, company_name)
            for item in ddg_items:
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
            short_name = company_name.replace("PT", "").replace("Tbk", "").strip()
            if len(short_name) > 2:
                query_terms.append(f'"{short_name}"')

        disclosure_keywords = '(keterbukaan OR bursa OR laba OR dividen OR smelter OR akuisisi)'
        raw_query = f"{' OR '.join(query_terms)} {disclosure_keywords}"
        encoded_query = urllib.parse.quote(raw_query)
        rss_url = f"{self.GOOGLE_NEWS_RSS_BASE}?q={encoded_query}&hl=id&gl=ID&ceid=ID:id"

        items: List[OSINTItem] = []
        try:
            feed = feedparser.parse(rss_url)
            for entry in feed.entries[:8]:
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
                        snippet=self._clean_snippet(summary or title),
                        is_disclosure=is_disclosure,
                        source_type="DISCLOSURE" if is_disclosure else "NEWS",
                    )
                )
        except Exception:
            pass

        return items

    def _fetch_duckduckgo_fallback(
        self, ticker: str, company_name: Optional[str] = None
    ) -> List[OSINTItem]:
        """Fallback web dorking via DuckDuckGo HTML without API keys."""
        items: List[OSINTItem] = []
        query = f'"{ticker}" saham keterbukaan informasi bursa'
        try:
            headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}
            resp = requests.post(self.DUCKDUCKGO_HTML_BASE, data={"q": query}, headers=headers, timeout=5)
            if resp.status_code == 200:
                # Basic parsing of DDG HTML result links
                matches = re.findall(r'<a class="result__url" href="([^"]+)".*?>.*?</a>.*?<a class="result__snippet[^"]*"[^>]*>(.*?)</a>', resp.text, re.DOTALL)
                for link, snip in matches[:4]:
                    clean_snip = re.sub(r"<[^>]+>", "", snip).strip()
                    title = clean_snip[:70] + "..." if len(clean_snip) > 70 else clean_snip
                    is_disc = self._is_disclosure_headline(title)
                    items.append(
                        OSINTItem(
                            title=title,
                            source_name="DuckDuckGo Web Search",
                            source_url=link.strip(),
                            publication_date="",
                            snippet=self._clean_snippet(clean_snip),
                            is_disclosure=is_disc,
                            source_type="DISCLOSURE" if is_disc else "NEWS",
                        )
                    )
        except Exception:
            pass
        return items

    def sanitize_article_text(self, html_or_url: str, fallback_snippet: Optional[str] = None) -> Optional[str]:
        """Multi-Level Trafilatura Sanitizer with Media Noise Cleaner."""
        extracted: Optional[str] = None
        try:
            if html_or_url.startswith("http://") or html_or_url.startswith("https://"):
                downloaded = trafilatura.fetch_url(html_or_url)
                if downloaded:
                    extracted = trafilatura.extract(downloaded, include_comments=False)
            else:
                extracted = trafilatura.extract(html_or_url, include_comments=False)
        except Exception:
            extracted = None

        # Level 2 Fallback: If trafilatura failed or returned empty text
        if not extracted or len(extracted.strip()) < 50:
            extracted = fallback_snippet or None

        if not extracted:
            return None

        # Apply Indonesian media noise regex filters
        cleaned = extracted
        for pat in self.NOISE_PATTERNS:
            cleaned = re.sub(pat, "", cleaned, flags=re.IGNORECASE | re.MULTILINE)

        return cleaned.strip()

    def wrap_in_evidence_context(self, items: List[OSINTItem]) -> str:
        """Format harvested items into an isolated XML structure for LLM context."""
        parts = ["<evidence_context>"]
        for idx, item in enumerate(items, 1):
            parts.append(
                f'  <item id="{idx}" type="{item.source_type}" source="{item.source_name}" date="{item.publication_date}">\n'
                f"    <headline>{item.title}</headline>\n"
                f"    <snippet>{item.snippet}</snippet>\n"
                f"  </item>"
            )
        parts.append("</evidence_context>")
        return "\n".join(parts)

    def _clean_snippet(self, raw_snippet: str) -> str:
        """Strip HTML tags and trim snippet length."""
        clean = re.sub(r"<[^>]+>", "", raw_snippet).strip()
        return clean[:300]

    def _is_disclosure_headline(self, title: str) -> bool:
        low = title.lower()
        keywords = ["keterbukaan", "penjelasan bursa", "volatilitas", "suspensi", "dividen", "rups", "smelter", "laporan keuangan"]
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
