"""Deterministic Tool Registry for Niskava Autonomous ReAct Agent.

Complies strictly with:
- Law 1: Deterministic Before Generative (NumPy calculates Z-scores, never LLM)
- Law 2: Strict Financial Non-Advisory Boundary (3-Tier Taxonomy)
- Law 5: Credit Budget Discipline (SQLite sectors_cache)
"""

import json
import os
from typing import Any, Callable, Dict, List, Optional

from engine.memory.graph_memory import LocalGraphMemory
from engine.quant.anomaly import AnomalyResult, detect_historical_anomalies
from engine.sectors.client import SectorsAPIClient, SectorsAPIError
from engine.sectors.news_engine import NewsItem, SectorsNewsEngine
from engine.skills.registry import SkillsRegistry


# Backward compatibility alias
OSINTItem = NewsItem

_INDEX_TICKERS: frozenset[str] = frozenset(
    {"IHSG", "JCI", "IDX", "COMPOSITE", "JKSE", "^JKSE", "^IDX"}
)


def _is_index_ticker(ticker: Optional[str]) -> bool:
    """Return True if ticker represents a composite market index rather than an individual stock."""
    if not ticker:
        return False
    clean = str(ticker).upper().strip()
    return clean in _INDEX_TICKERS or clean.lstrip("^") in _INDEX_TICKERS


# Domain key → SectorsAPIClient method name mapping for query_sectors gateway.
_SECTORS_DOMAIN_MAP: dict[str, str] = {
    "candles": "get_daily_candles",
    "fundamentals": "get_company_report",
    "foreign_flow": "get_foreign_flow",
    "suspensions": "get_suspensions",
    "filings": "get_filings",
    "broker_summary": "get_broker_summary",
    "corporate_actions": "get_corporate_actions",
    "subsector_peers": "get_subsector_peers",
    "mining_detail": "get_mining_detail",
    "news": "get_news",
    "subsectors": "get_subsectors",
    "top_changes": "get_top_changes",
    "most_traded": "get_most_traded",
}


class NiskavaToolRegistry:
    """Provides structured, callable tools for the ReAct Agent."""

    def __init__(
        self,
        db_path: str,
        sectors_client: Optional[SectorsAPIClient] = None,
        osint_harvester: Optional[Any] = None,
        news_harvester: Optional[Any] = None,
        mock_mode: Optional[bool] = None,
        skills_registry: Optional[SkillsRegistry] = None,
        memory: Optional[LocalGraphMemory] = None,
    ):
        self.db_path = os.path.expanduser(db_path)
        self.mock_mode = mock_mode
        self.sectors_client = sectors_client or SectorsAPIClient(
            db_path=self.db_path, mock_mode=self.mock_mode
        )
        self.news_engine = SectorsNewsEngine(
            sectors_client=self.sectors_client,
            db_path=self.db_path,
            mock_mode=self.mock_mode,
        )
        self.news_harvester = news_harvester or osint_harvester or self.news_engine
        # Backward-compatible reference
        self.osint_harvester = self.news_harvester
        self.skills_registry = skills_registry or SkillsRegistry()
        self.memory = memory or LocalGraphMemory(db_path=self.db_path)

    def execute_skill(self, skill_id: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """Execute a Layer 3 Domain Skill and return structured findings dict."""
        context = {
            "sectors_client": self.sectors_client,
            "news_harvester": self.news_harvester,
            "osint_harvester": self.osint_harvester,
            "db_path": self.db_path,
            "mock_mode": self.mock_mode,
        }
        if hasattr(self, "emitter") and callable(self.emitter):
            context["emitter"] = self.emitter
        try:
            res = self.skills_registry.execute_skill(skill_id, arguments, context)
            return res.to_dict()
        except Exception as e:
            ticker = arguments.get("ticker", "")
            return {
                "skill_id": skill_id,
                "verification_status": "UNCERTAIN",
                "confidence_score": 0.65,
                "metrics": {"error": str(e), "ticker": ticker},
                "evidence": [],
                "summary": f"Skill '{skill_id}' menghadapi kendala data eksternal ({str(e)}). Melanjutkan investigasi dengan data fundamental dan historis yang tersedia.",
            }

    # ---------------------------------------------------------------------------
    # Gateway Primitive Methods — Progressive Skill Disclosure (ADR-11)
    # ---------------------------------------------------------------------------

    def query_sectors(self, domain: str, ticker: str, params: Dict[str, Any] = {}) -> Any:
        """Universal gateway to Sectors Financial API v2.

        Routes to the appropriate SectorsAPIClient method based on `domain`.
        Always checks SQLite sectors_cache first via the client (Law 5).

        Args:
            domain: One of the keys in _SECTORS_DOMAIN_MAP
                    ('candles', 'fundamentals', 'foreign_flow', 'suspensions',
                    'filings', 'broker_summary', 'corporate_actions',
                    'subsector_peers', 'mining_detail', 'news', 'subsectors').
            ticker: IDX 4-letter ticker (case-insensitive, auto-uppercased).
            params: Optional domain-specific extra parameters
                    (e.g. {'slug': '...'} for mining_detail).

        Returns:
            Raw API response as returned by the underlying SectorsAPIClient method.

        Raises:
            ValueError: If `domain` is not in _SECTORS_DOMAIN_MAP.
        """
        clean_ticker = ticker.upper().strip() if ticker else ""
        is_index = _is_index_ticker(clean_ticker)

        if is_index and domain in ("", None):
            # Return enriched market overview instead of raw news list
            # to prevent model confusion and duplicate tool call loops
            news = self.sectors_client.get_news(None)
            context_parts = [
                f"Data pasar umum IDX per hari ini.",
                f"Jumlah artikel berita terkini: {len(news)}.",
            ]
            if news:
                top_headlines = [n.get("title", "") for n in news[:3] if isinstance(n, dict)]
                if top_headlines:
                    context_parts.append(
                        "Headline: " + "; ".join(top_headlines)
                    )
            return {
                "type": "market_overview",
                "ticker": clean_ticker,
                "status": "active",
                "news": news,
                "market_context": " ".join(context_parts),
            }

        if is_index and domain == "candles":
            return {
                "error": f"Candlestick OHLCV data is unavailable for composite market index '{clean_ticker}'.",
                "ticker": clean_ticker,
                "type": "index_candle_limitation",
                "message": (
                    f"Sectors Financial API v2 provides daily OHLCV candles exclusively for individual IDX company stocks (e.g. BBCA, BBRI, ANTM), "
                    f"not for composite market indices ({clean_ticker})."
                ),
                "recommended_alternatives": [
                    "Use domain='top_changes' (classification='top_gainers' or 'top_losers') or domain='most_traded' to assess market breadth.",
                    "Use domain='subsectors' to evaluate sector performance.",
                    "Analyze key index-mover heavyweights (BBCA, BBRI, BMRI, TLKM, ASII) via domain='candles' or skill 'market_anomaly_recon' as proxies for market movement.",
                ],
            }

        # Intelligently default domain when omitted or empty to prevent ReAct loop crashes
        if not domain:
            domain = "candles"

        method_name = _SECTORS_DOMAIN_MAP.get(domain)
        if not method_name:
            supported = ", ".join(sorted(_SECTORS_DOMAIN_MAP.keys()))
            raise ValueError(
                f"Unknown domain: '{domain}'. Supported domains: {supported}"
            )

        client_method = getattr(self.sectors_client, method_name)
        # Flatten nested params if model passes {'params': {'classification': ...}}
        merged_params: Dict[str, Any] = {}
        if isinstance(params, dict):
            merged_params.update(params)
            if "params" in params and isinstance(params["params"], dict):
                merged_params.update(params["params"])

        force_refresh = bool(merged_params.get("force_refresh", False))

        try:
            kwargs = {}
            if force_refresh:
                kwargs["force_refresh"] = True

            if domain == "subsectors":
                return client_method(**kwargs)
            if domain == "top_changes":
                cls_val = (merged_params.get("classification") or merged_params.get("classifications") or "top_gainers")
                period_val = (merged_params.get("period") or merged_params.get("periods") or "1d")
                n_stock_val = int(merged_params.get("n_stock", 5)) if "n_stock" in merged_params else 5
                return client_method(classification=cls_val, period=period_val, n_stock=n_stock_val, **kwargs)
            if domain == "most_traded":
                n_stock = int(merged_params.get("n_stock", 5)) if "n_stock" in merged_params else 5
                start = merged_params.get("start")
                end = merged_params.get("end")
                return client_method(start=start, end=end, n_stock=n_stock, **kwargs)

            # Domains with a non-ticker primary key
            if domain == "subsector_peers":
                raw_slug = None
                if isinstance(params, dict):
                    raw_slug = params.get("subsector") or params.get("sub_sector") or params.get("slug")
                if not raw_slug and clean_ticker:
                    try:
                        rep = self.sectors_client.get_company_report(clean_ticker)
                        ov = rep.get("overview", {}) if isinstance(rep.get("overview"), dict) else {}
                        raw_sub = rep.get("sub_sector") or rep.get("subsector") or ov.get("sub_sector") or ov.get("subsector")
                        if raw_sub:
                            raw_slug = re.sub(r'[^a-z0-9]+', '-', raw_sub.lower()).strip('-')
                    except Exception:
                        pass
                slug = raw_slug or clean_ticker.lower()
                return client_method(slug, **kwargs)
            if domain == "mining_detail":
                slug = params.get("slug", clean_ticker.lower()) if isinstance(params, dict) else clean_ticker.lower()
                return client_method(slug, **kwargs)

            return client_method(clean_ticker, **kwargs)

        except SectorsAPIError as err:
            return {
                "error": True,
                "error_type": "SECTORS_API_ERROR",
                "status_code": err.status_code,
                "message": f"Koneksi Sectors Financial API gagal: {str(err)}. Periksa koneksi internet atau SECTORS_API_KEY di Pengaturan.",
            }


    def search_news(self, ticker: str, query: str = "", force_refresh: bool = False) -> List[Dict[str, Any]]:
        """Universal gateway to the Sectors News and Corporate Disclosure engine.

        Fetches curated news and corporate disclosures directly from Sectors
        Financial API v2 for the given ticker. Returns a list of sanitized news dicts.

        Args:
            ticker: IDX 4-letter ticker. Pass empty string for general market news.
            query: Optional search keyword or context filter.
            force_refresh: Whether to bypass cache and fetch latest news.

        Returns:
            List of dicts, each with keys: title, source_name, source_url, publication_date, snippet.
        """
        clean_ticker = ticker.upper() if ticker else ""

        try:
            if not clean_ticker or _is_index_ticker(clean_ticker):
                sectors_news = self.sectors_client.get_news(None, force_refresh=force_refresh)
                items: List[NewsItem] = self.news_harvester.harvest(
                    ticker="IHSG",
                    company_name="Pasar Modal Indonesia",
                    sectors_news_items=sectors_news,
                    query=query,
                )
                return [item.model_dump() for item in items]

            try:
                report = self.get_company_fundamentals(clean_ticker)
                company_name = report.get("company_name", clean_ticker) if isinstance(report, dict) else clean_ticker
            except Exception:
                company_name = clean_ticker

            try:
                sectors_news = self.sectors_client.get_news(clean_ticker, force_refresh=force_refresh)
            except Exception:
                sectors_news = []

            items = self.news_harvester.harvest(
                ticker=clean_ticker,
                company_name=company_name,
                sectors_news_items=sectors_news,
                query=query,
            )
            return [item.model_dump() for item in items]
        except Exception as err:
            return [{
                "title": f"Gagal mengambil berita terkini: {str(err)}",
                "source_name": "Sectors API",
                "source_url": "",
                "publication_date": "",
                "snippet": "Terjadi kendala saat menghubungi Sectors Financial API. Silakan periksa kunci API Anda di Settings.",
            }]


    # Backward-compatible alias
    search_osint = search_news

    def query_memory(self, concept_or_ticker: str, radius: int = 2) -> Dict[str, Any]:
        """Universal gateway to the local conversational graph memory engine.

        Performs ego-graph traversal (≤ radius hops) with exponential recency
        decay from SQLite memory_nodes/memory_edges via NetworkX.

        Args:
            concept_or_ticker: Entity label to query (e.g. 'ANTM', 'Hari Darmawan').
            radius: Ego-graph hop radius (default 2, max recommended 3).

        Returns:
            Dict with keys: query, nodes_found (int), nodes (list), edges (list).
        """
        res = self.memory.retrieve_ego_subgraph(
            entity_query=concept_or_ticker, radius=radius
        )
        return {
            "query": res["query"],
            "nodes_found": len(res["nodes"]),
            "nodes": res["nodes"],
            "edges": res["edges"],
        }

    def get_daily_candles(self, ticker: str, days: int = 30) -> List[Dict[str, Any]]:
        """Retrieve daily OHLCV candlesticks for the specified ticker."""
        return self.sectors_client.get_daily_candles(ticker.upper())

    def compute_quant_anomalies(
        self,
        ticker: str,
        volume_z_threshold: float = 2.5,
        return_threshold_pct: float = 5.0,
    ) -> List[Dict[str, Any]]:
        """Compute rolling volume Z-scores (MA20) and price return anomalies deterministically via NumPy."""
        clean_ticker = ticker.upper().strip()
        candles = self.get_daily_candles(clean_ticker)
        anomalies = detect_historical_anomalies(
            daily_candles=candles,
            volume_z_threshold=volume_z_threshold,
            return_threshold_pct=return_threshold_pct,
        )
        if hasattr(self, "emitter") and callable(self.emitter):
            for a in anomalies:
                self.emitter({
                    "event": "anomaly_detected",
                    "ticker": clean_ticker,
                    "anomaly_date": a.date,
                    "metric_type": a.classification,
                    "metric_value": a.metric_value,
                    "baseline_value": a.baseline_value,
                    "z_score": round(a.z_score, 2),
                    "price_change_pct": round(a.price_change_pct, 2),
                    "description": a.description,
                })
        return [a.to_dict() for a in anomalies]

    def get_company_fundamentals(self, ticker: str) -> Dict[str, Any]:
        """Fetch fundamental company report and industrial classification."""
        return self.sectors_client.get_company_report(ticker.upper())

    def get_foreign_flow(self, ticker: str) -> List[Dict[str, Any]]:
        """Fetch net foreign inflow / outflow data."""
        return self.sectors_client.get_foreign_flow(ticker.upper())

    def get_suspensions(self, ticker: str) -> List[Dict[str, Any]]:
        """Fetch exchange suspension notices and official IDX PDF announcements."""
        return self.sectors_client.get_suspensions(ticker.upper())

    def get_corporate_actions(self, ticker: str) -> List[Dict[str, Any]]:
        """Fetch scheduled corporate actions (dividends, splits, rights issue)."""
        return self.sectors_client.get_corporate_actions(ticker.upper())

    def get_filings(self, ticker: str) -> List[Dict[str, Any]]:
        """Fetch insider trading and substantial shareholder filings."""
        return self.sectors_client.get_filings(ticker.upper())

    def get_broker_summary(self, ticker: str) -> Dict[str, Any]:
        """Fetch top broker accumulation and distribution summary."""
        return self.sectors_client.get_broker_summary(ticker.upper())

    def get_subsector_peers(self, subsector: str) -> Dict[str, Any]:
        """Fetch industrial subsector peers and valuation benchmarks."""
        return self.sectors_client.get_subsector_peers(subsector.lower())

    def get_mining_detail(self, slug: str) -> Dict[str, Any]:
        """Fetch operational mining concession and smelter details."""
        return self.sectors_client.get_mining_detail(slug.lower())

    def harvest_market_news(
        self,
        ticker: Optional[str] = None,
        company_name: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        """Harvest curated news and corporate disclosures exclusively from Sectors Financial API v2."""
        if not ticker:
            # General market headlines when no specific ticker is provided
            sectors_news = self.sectors_client.get_news(None)
            items: List[NewsItem] = self.news_harvester.harvest(
                ticker="IHSG",
                company_name="Pasar Modal Indonesia",
                sectors_news_items=sectors_news,
            )
            return [item.model_dump() for item in items]

        clean_ticker = ticker.upper()
        if not company_name:
            report = self.get_company_fundamentals(clean_ticker)
            company_name = report.get("company_name", clean_ticker)

        sectors_news = self.sectors_client.get_news(clean_ticker)
        items: List[NewsItem] = self.news_harvester.harvest(
            ticker=clean_ticker,
            company_name=company_name,
            sectors_news_items=sectors_news,
        )
        return [item.model_dump() for item in items]

    def memory_recall_context(
        self,
        query_entity: str,
        radius: int = 2,
    ) -> Dict[str, Any]:
        """Recall associative past memories and observations around an entity from local graph."""
        res = self.memory.retrieve_ego_subgraph(entity_query=query_entity, radius=radius)
        return {
            "query": res["query"],
            "nodes_found": len(res["nodes"]),
            "nodes": res["nodes"],
            "edges": res["edges"],
        }

    def memory_store_observation(
        self,
        source_label: str,
        relation: str,
        target_label: str,
        source_type: str = "TICKER",
        target_type: str = "ENTITY",
        context_snippet: str = "",
        session_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Store a new observed relation between two entities into local graph memory."""
        return self.memory.store_observation(
            source_label=source_label,
            relation=relation,
            target_label=target_label,
            source_type=source_type,
            target_type=target_type,
            context_snippet=context_snippet,
            session_id=session_id,
        )

    def memory_find_connection(
        self,
        source_entity: str,
        target_entity: str,
    ) -> Dict[str, Any]:
        """Find the shortest connection path between two entities in the knowledge graph."""
        path = self.memory.find_shortest_path(source_entity, target_entity)
        return {
            "source": source_entity,
            "target": target_entity,
            "path_found": path is not None,
            "hops": len(path) - 1 if path else 0,
            "path": path or [],
        }

    def memory_get_graph_stats(self) -> Dict[str, Any]:
        """Get summary graph topological statistics and top central entities."""
        return self.memory.get_graph_stats()

    def inspect_document(
        self,
        doc_path: str,
        query: Optional[str] = None,
        page: Optional[int] = None,
    ) -> Dict[str, Any]:
        """Inspect, search, or read specific pages of an uploaded document locally.

        Complies with Law 1 (Deterministic Before Generative) and Law 4 (Local-First).

        Args:
            doc_path: Local file path of the document to inspect.
            query: Optional keyword or phrase to search within document text.
            page: Optional 1-indexed page number to extract.

        Returns:
            Dict containing document metadata, page content, or search matches.
        """
        if not doc_path:
            return {"error": True, "message": "Missing doc_path parameter."}

        clean_path = os.path.expanduser(str(doc_path).strip())
        if not os.path.exists(clean_path):
            # Fallback: check ~/.niskava/uploads/ recursively for matching filename
            base_filename = os.path.basename(clean_path)
            uploads_dir = os.path.expanduser("~/.niskava/uploads")
            matched_path = None
            if os.path.exists(uploads_dir):
                for root, _, files in os.walk(uploads_dir):
                    for f in files:
                        if f == base_filename or f.endswith("_" + base_filename) or base_filename in f:
                            matched_path = os.path.join(root, f)
                            break
                    if matched_path:
                        break
            if matched_path and os.path.exists(matched_path):
                clean_path = matched_path
            else:
                return {"error": True, "message": f"Document not found at path: {clean_path}"}

        from engine.skills.document_audit.parser import (
            get_document_page,
            parse_document,
            search_document,
        )

        try:
            parsed = parse_document(clean_path)
        except Exception as exc:
            return {
                "error": True,
                "message": f"Failed to parse document: {str(exc)}",
            }

        filename = parsed.get("filename", os.path.basename(clean_path))
        total_pages = parsed.get("page_count", 0)

        if page is not None:
            try:
                page_int = int(page)
            except (ValueError, TypeError):
                return {
                    "error": True,
                    "message": f"Invalid page number: '{page}'. Page must be an integer.",
                }

            page_text = get_document_page(parsed, page_int)
            if page_text is None:
                return {
                    "error": True,
                    "message": f"Page {page_int} out of range (document has {total_pages} pages).",
                    "filename": filename,
                    "total_pages": total_pages,
                }

            return {
                "filename": filename,
                "doc_path": clean_path,
                "page": page_int,
                "total_pages": total_pages,
                "content": page_text,
            }

        if query:
            matches = search_document(parsed, str(query).strip(), max_matches=5)
            return {
                "filename": filename,
                "doc_path": clean_path,
                "query": query,
                "total_matches": len(matches),
                "matches": matches,
            }

        # Default overview: metadata and first 2-3 pages preview
        preview_pages = [
            {"page": p["page_number"], "preview": p["text"][:500]}
            for p in parsed.get("pages", [])[:3]
        ]
        return {
            "filename": filename,
            "doc_path": clean_path,
            "format": parsed.get("format", ""),
            "file_size": parsed.get("file_size", 0),
            "total_pages": total_pages,
            "preview": preview_pages,
        }

    def get_tool_definitions(self) -> List[Dict[str, Any]]:
        """Return the 4 lean gateway tool definitions for LLM function calling.

        Progressive Skill Disclosure Architecture (ADR-11): the LLM sees only
        4 universal gateway primitives. Each gateway routes internally to the full
        set of atomic Sectors API methods and domain skills via execute_tool() —
        preserving full backward compatibility.
        """
        return [
            {
                "name": "execute_skill",
                "description": (
                    "Execute a specialized domain equity research Standard Operating Procedure (SOP). "
                    "Use this for structured deep-dive market investigations. "
                    "Available skill_ids: "
                    "market_anomaly_recon (scan MA20 volume spikes/Z-scores and abnormal returns via NumPy), "
                    "event_causality_audit (audit news causality vs volume surge: LIKELY_CATALYST/PRECEDED_ANNOUNCEMENT), "
                    "insider_bandarmology_forensic (audit top broker accumulation C3>=65% and insider trading filings), "
                    "financial_health_stress_test (stress-test liquidity/solvency ratios and evaluate default rumors), "
                    "mining_commodity_divergence (test mining company correlation against global spot commodity benchmarks), "
                    "peer_valuation_benchmark (benchmark PER/PBV multiples against IDX subsector median), "
                    "document_audit (audit and inspect uploaded local financial documents, prospectuses, or disclosures), "
                    "investigation_report_pdf (generate institutional PDF audit trail report; ONLY when user asks to export/save/print PDF)."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "skill_id": {
                            "type": "string",
                            "description": (
                                "Target skill ID to execute. Choose one: "
                                "market_anomaly_recon, event_causality_audit, "
                                "insider_bandarmology_forensic, financial_health_stress_test, "
                                "mining_commodity_divergence, peer_valuation_benchmark, "
                                "document_audit, investigation_report_pdf."
                            ),
                        },
                        "arguments": {
                            "type": "object",
                            "description": (
                                "Skill parameters. For stock analysis skills: {'ticker': 'ANTM'} is required. "
                                "For investigation_report_pdf: requires either 'summary' or 'blocks' (array of callout/markdown/table/key_value). Optional: 'ticker' (defaults to 'MARKET'), 'title', 'custom_tables', 'metrics', 'evidence'."
                            ),
                        },
                    },
                    "required": ["skill_id", "arguments"],
                },
            },
            {
                "name": "query_sectors",
                "description": (
                    "Universal router to Sectors Financial API v2. "
                    "Use `domain` to select the dataset: "
                    "candles (daily OHLCV time series), "
                    "fundamentals (company profile, financial statements, and valuation ratios), "
                    "foreign_flow (net foreign institutional accumulation/distribution), "
                    "suspensions (official IDX suspension notices and Unusual Market Activity / UMA), "
                    "filings (insider trading transactions by directors and commissioners), "
                    "broker_summary (top buying and selling brokerage participants), "
                    "corporate_actions (cash dividends, stock splits, rights issues), "
                    "subsector_peers (subsector peer comparison and valuation multiples), "
                    "mining_detail (operational mining concessions, IUP permits, and smelter assets), "
                    "news (curated financial news from Sectors API), "
                    "subsectors (official list of IDX sectors and subsectors), "
                    "top_changes (top gainers or top losers on IDX), "
                    "most_traded (most active/traded stocks by volume or turnover)."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "domain": {
                            "type": "string",
                            "description": (
                                "Sectors API dataset domain. Required. Choose one: "
                                "candles | fundamentals | foreign_flow | suspensions | "
                                "filings | broker_summary | corporate_actions | "
                                "subsector_peers | mining_detail | news | subsectors | "
                                "top_changes | most_traded."
                            ),
                        },
                        "ticker": {
                            "type": "string",
                            "description": "4-letter IDX stock ticker symbol (e.g. ANTM, BBCA). Case-insensitive. Optional or empty string for subsectors, top_changes, most_traded, or macro index overview.",
                        },
                        "params": {
                            "type": "object",
                            "description": (
                                "Optional extra query parameters, e.g.: "
                                "{'classification': 'top_gainers'|'top_losers', 'period': '1d'} for top_changes, "
                                "{'n_stock': 10} for most_traded, "
                                "{'subsector': 'metals-mining'} for subsector_peers, "
                                "{'slug': 'antm'} for mining_detail."
                            ),
                        },
                    },
                    "required": ["domain", "ticker"],
                },
            },
            {
                "name": "search_news",
                "description": (
                    "Universal router to Sectors Curated News & Corporate Disclosures Engine. "
                    "Fetches verified financial news and official company announcements directly "
                    "from Sectors Financial API v2 (/v2/news/). "
                    "Pass empty string for ticker to retrieve macro IDX market news."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {
                            "type": "string",
                            "description": "4-letter IDX stock ticker. Pass empty string for general market news.",
                        },
                        "query": {
                            "type": "string",
                            "description": "Optional search keyword or sector filtering query.",
                        },
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "query_memory",
                "description": (
                    "Universal router to local associative graph memory (NetworkX + SQLite WAL). "
                    "Traverses ego-graphs up to 2 hops with temporal recency decay to recall "
                    "past findings, entities, and discussion context across sessions."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "concept_or_ticker": {
                            "type": "string",
                            "description": "Stock ticker or market concept to recall (e.g. 'ANTM', 'smelter').",
                        },
                        "radius": {
                            "type": "integer",
                            "description": "Ego-graph traversal radius (1 or 2, default: 2).",
                            "default": 2,
                        },
                    },
                    "required": ["concept_or_ticker"],
                },
            },
        ]

    def get_all_tool_definitions(self) -> List[Dict[str, Any]]:
        """Return full list of tool definitions including gateways and inspect_document."""
        defs = list(self.get_tool_definitions())
        defs.append({
            "name": "inspect_document",
            "description": (
                "Inspect, search, or read specific pages of an uploaded document "
                "(PDF, TXT, CSV, financial reports) locally without context bloat."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "doc_path": {
                        "type": "string",
                        "description": "Local file path of the document to inspect.",
                    },
                    "query": {
                        "type": "string",
                        "description": "Optional keyword or phrase to search within the document.",
                    },
                    "page": {
                        "type": "integer",
                        "description": "Optional 1-indexed page number to read directly.",
                    },
                },
                "required": ["doc_path"],
            },
        })
        return defs

    def execute_tool(self, tool_name: str, arguments: Any) -> Any:
        """Dynamically dispatch and execute a registered tool (supporting direct & MCP names)."""
        if arguments is None:
            arguments = {}
        elif isinstance(arguments, str):
            trimmed = arguments.strip()
            if trimmed.startswith("{") and trimmed.endswith("}"):
                try:
                    parsed = json.loads(trimmed)
                    arguments = parsed if isinstance(parsed, dict) else {"ticker": trimmed}
                except Exception:
                    arguments = {"ticker": trimmed}
            else:
                arguments = {"ticker": trimmed} if trimmed else {}
        elif not isinstance(arguments, dict):
            arguments = {}

        ticker = str(arguments.get("ticker") or arguments.get("symbol") or "").upper()
        raw_days = arguments.get("days")
        if raw_days is None:
            raw_days = arguments.get("lookback_days", 30)
        try:
            days = int(raw_days)
            if days <= 0:
                days = 30
        except (ValueError, TypeError):
            days = 30

        # Check for Layer 3 Domain Skill execution
        if tool_name == "execute_skill":
            sub_args = arguments.get("arguments", {})
            if not isinstance(sub_args, dict) or not sub_args:
                sub_args = {k: v for k, v in arguments.items() if k != "skill_id"}
            return self.execute_skill(
                skill_id=arguments.get("skill_id", ""),
                arguments=sub_args,
            )
        if tool_name.startswith("skill_"):
            skill_id = tool_name[6:].replace("_", "-")
            return self.execute_skill(skill_id=skill_id, arguments=arguments)
        if self.skills_registry.get_skill(tool_name):
            return self.execute_skill(skill_id=tool_name, arguments=arguments)

        handlers: Dict[str, Callable[..., Any]] = {
            "get_daily_candles": lambda args: self.get_daily_candles(
                ticker=ticker,
                days=days,
            ),
            "sectors_get_daily_candles": lambda args: self.get_daily_candles(
                ticker=ticker,
                days=days,
            ),
            "compute_quant_anomalies": lambda args: self.compute_quant_anomalies(
                ticker=ticker,
                volume_z_threshold=float(args.get("volume_z_threshold", 2.5)),
            ),
            "get_company_fundamentals": lambda args: self.get_company_fundamentals(
                ticker=ticker,
            ),
            "sectors_get_company_report": lambda args: self.get_company_fundamentals(
                ticker=ticker,
            ),
            "query_company_profile": lambda args: self.get_company_fundamentals(
                ticker=ticker,
            ),
            "get_foreign_flow": lambda args: self.get_foreign_flow(
                ticker=ticker,
            ),
            "sectors_get_foreign_flow": lambda args: self.get_foreign_flow(
                ticker=ticker,
            ),
            "get_suspensions": lambda args: self.get_suspensions(
                ticker=ticker,
            ),
            "sectors_get_suspensions": lambda args: self.get_suspensions(
                ticker=ticker,
            ),
            "get_corporate_actions": lambda args: self.get_corporate_actions(
                ticker=ticker,
            ),
            "sectors_get_corporate_actions": lambda args: self.get_corporate_actions(
                ticker=ticker,
            ),
            "get_filings": lambda args: self.get_filings(
                ticker=ticker,
            ),
            "sectors_get_filings": lambda args: self.get_filings(
                ticker=ticker,
            ),
            "get_broker_summary": lambda args: self.get_broker_summary(
                ticker=ticker,
            ),
            "sectors_get_broker_summary": lambda args: self.get_broker_summary(
                ticker=ticker,
            ),
            "get_subsector_peers": lambda args: self.get_subsector_peers(
                subsector=args.get("subsector", ""),
            ),
            "sectors_get_subsector_peers": lambda args: self.get_subsector_peers(
                subsector=args.get("subsector", ""),
            ),
            "get_subsectors": lambda args: self.sectors_client.get_subsectors(),
            "sectors_get_subsectors": lambda args: self.sectors_client.get_subsectors(),
            "get_mining_detail": lambda args: self.get_mining_detail(
                slug=args.get("slug", ""),
            ),
            "sectors_get_mining_detail": lambda args: self.get_mining_detail(
                slug=args.get("slug", ""),
            ),
            "harvest_market_news": lambda args: self.harvest_market_news(
                ticker=ticker if ticker and not _is_index_ticker(ticker) else None,
                company_name=args.get("company_name"),
            ),
            "memory_recall_context": lambda args: self.memory_recall_context(
                query_entity=args.get("query_entity", ticker),
                radius=int(args.get("radius", 2)),
            ),
            "memory_store_observation": lambda args: self.memory_store_observation(
                source_label=args.get("source_label", ""),
                relation=args.get("relation", ""),
                target_label=args.get("target_label", ""),
                source_type=args.get("source_type", "TICKER"),
                target_type=args.get("target_type", "ENTITY"),
                context_snippet=args.get("context_snippet", ""),
                session_id=args.get("session_id"),
            ),
            "memory_find_connection": lambda args: self.memory_find_connection(
                source_entity=args.get("source_entity", ""),
                target_entity=args.get("target_entity", ""),
            ),
            "memory_get_graph_stats": lambda args: self.memory_get_graph_stats(),
            "query_sectors": lambda args: self.query_sectors(
                domain=args.get("domain", ""),
                ticker=args.get("ticker", ""),
                params={k: v for k, v in args.items() if k not in ("domain", "ticker")},
            ),
            "search_news": lambda args: self.search_news(
                ticker=args.get("ticker", ""),
                query=args.get("query", ""),
            ),
            "search_osint": lambda args: self.search_news(
                ticker=args.get("ticker", ""),
                query=args.get("query", ""),
            ),
            "sectors_search_news": lambda args: self.search_news(
                ticker=args.get("ticker", ""),
                query=args.get("query", ""),
            ),
            "query_memory": lambda args: self.query_memory(
                concept_or_ticker=args.get("concept_or_ticker", args.get("ticker", "")),
                radius=int(args.get("radius", 2)),
            ),
            "inspect_document": lambda args: self.inspect_document(
                doc_path=args.get("doc_path") or args.get("file_path", ""),
                query=args.get("query"),
                page=args.get("page"),
            ),
        }

        handler = handlers.get(tool_name)
        if not handler:
            raise ValueError(f"Tool '{tool_name}' is not registered in Niskava Tool Registry.")

        return handler(arguments)

