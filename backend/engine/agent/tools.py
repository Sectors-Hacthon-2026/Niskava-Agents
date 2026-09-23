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
from engine.osint.harvester import DualEngineOSINTHarvester, OSINTItem
from engine.quant.anomaly import AnomalyResult, detect_historical_anomalies
from engine.sectors.client import SectorsAPIClient
from engine.skills.registry import SkillsRegistry

# Ticker-ticker yang merepresentasikan indeks pasar, bukan emiten perusahaan individual.
# Jika digunakan di harvest_market_news, harus di-route ke general market news.
_INDEX_TICKERS: frozenset[str] = frozenset({"IHSG", "JCI", "IDX", "COMPOSITE"})

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
}


class NiskavaToolRegistry:
    """Provides structured, callable tools for the ReAct Agent."""

    def __init__(
        self,
        db_path: str,
        sectors_client: Optional[SectorsAPIClient] = None,
        osint_harvester: Optional[DualEngineOSINTHarvester] = None,
        mock_mode: Optional[bool] = None,
        skills_registry: Optional[SkillsRegistry] = None,
        memory: Optional[LocalGraphMemory] = None,
    ):
        self.db_path = os.path.expanduser(db_path)
        self.mock_mode = mock_mode
        self.sectors_client = sectors_client or SectorsAPIClient(
            db_path=self.db_path, mock_mode=self.mock_mode
        )
        self.osint_harvester = osint_harvester or DualEngineOSINTHarvester(
            mock_mode=self.mock_mode
        )
        self.skills_registry = skills_registry or SkillsRegistry()
        self.memory = memory or LocalGraphMemory(db_path=self.db_path)

    def execute_skill(self, skill_id: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """Execute a Layer 3 Domain Skill and return structured findings dict."""
        context = {
            "sectors_client": self.sectors_client,
            "osint_harvester": self.osint_harvester,
            "db_path": self.db_path,
            "mock_mode": self.mock_mode,
        }
        res = self.skills_registry.execute_skill(skill_id, arguments, context)
        return res.to_dict()

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
                    'subsector_peers', 'mining_detail', 'news').
            ticker: IDX 4-letter ticker (case-insensitive, auto-uppercased).
            params: Optional domain-specific extra parameters
                    (e.g. {'slug': '...'} for mining_detail).

        Returns:
            Raw API response as returned by the underlying SectorsAPIClient method.

        Raises:
            ValueError: If `domain` is not in _SECTORS_DOMAIN_MAP.
        """
        clean_ticker = ticker.upper() if ticker else ""

        # Intelligently default domain when omitted or empty to prevent ReAct loop crashes
        if not domain:
            if clean_ticker in _INDEX_TICKERS:
                domain = "news"
            else:
                domain = "candles"

        method_name = _SECTORS_DOMAIN_MAP.get(domain)
        if not method_name:
            supported = ", ".join(sorted(_SECTORS_DOMAIN_MAP.keys()))
            raise ValueError(
                f"Unknown domain: '{domain}'. Supported domains: {supported}"
            )

        client_method = getattr(self.sectors_client, method_name)

        # Domains with a non-ticker primary key
        if domain == "subsector_peers":
            slug = params.get("subsector", clean_ticker.lower())
            return client_method(slug)
        if domain == "mining_detail":
            slug = params.get("slug", clean_ticker.lower())
            return client_method(slug)

        return client_method(clean_ticker)

    def search_osint(self, ticker: str, query: str = "") -> List[Dict[str, Any]]:
        """Universal gateway to the Dual-Engine OSINT harvester.

        Fetches curated Sectors news and targeted Google News RSS results for
        the given ticker. Returns a list of OSINTItem dicts (sanitised, no raw HTML).

        Args:
            ticker: IDX 4-letter ticker. Pass empty string for general market news.
            query: Optional extra keyword to narrow Google News RSS dorking.

        Returns:
            List of dicts, each with keys: title, url, published_at, source, snippet.
        """
        clean_ticker = ticker.upper() if ticker else ""

        if not clean_ticker or clean_ticker in _INDEX_TICKERS:
            sectors_news = self.sectors_client.get_news(None)
            items: List[OSINTItem] = self.osint_harvester.harvest(
                ticker="IHSG",
                company_name="Pasar Modal Indonesia",
                sectors_news_items=sectors_news,
            )
            return [item.model_dump() for item in items]

        report = self.get_company_fundamentals(clean_ticker)
        company_name = report.get("company_name", clean_ticker)
        sectors_news = self.sectors_client.get_news(clean_ticker)
        items = self.osint_harvester.harvest(
            ticker=clean_ticker,
            company_name=company_name,
            sectors_news_items=sectors_news,
        )
        return [item.model_dump() for item in items]

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
        candles = self.get_daily_candles(ticker)
        anomalies = detect_historical_anomalies(
            daily_candles=candles,
            volume_z_threshold=volume_z_threshold,
            return_threshold_pct=return_threshold_pct,
        )
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
        """Harvest curated news and targeted IDX regulatory filings via Dual-Engine OSINT."""
        if not ticker:
            # General market headlines when no specific ticker is provided
            sectors_news = self.sectors_client.get_news(None)
            items: List[OSINTItem] = self.osint_harvester.harvest(
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
        items: List[OSINTItem] = self.osint_harvester.harvest(
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
                    "Jalankan Standard Operating Procedure (SOP) analis ekuitas domain. "
                    "Gunakan ini untuk investigasi mendalam terstruktur. "
                    "skill_id tersedia: "
                    "market_anomaly_recon (scan lonjakan volume MA20/Z-score & abnormal return via NumPy), "
                    "event_causality_audit (audit kausalitas berita vs lonjakan volume: LIKELY_CATALYST/PRECEDED_ANNOUNCEMENT), "
                    "insider_bandarmology_forensic (audit akumulasi top broker C3>=65% & transaksi direksi/komisaris), "
                    "financial_health_stress_test (audit likuiditas Current/Quick, solvabilitas DER, sanggahan rumor gagal bayar), "
                    "mining_commodity_divergence (uji korelasi emiten tambang vs harga spot komoditas: Nikel, Batubara), "
                    "peer_valuation_benchmark (benchmark valuasi PER/PBV vs median rekan subsektor IDX)."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "skill_id": {
                            "type": "string",
                            "description": (
                                "ID skill yang akan dijalankan. Pilih salah satu: "
                                "market_anomaly_recon, event_causality_audit, "
                                "insider_bandarmology_forensic, financial_health_stress_test, "
                                "mining_commodity_divergence, peer_valuation_benchmark."
                            ),
                        },
                        "arguments": {
                            "type": "object",
                            "description": (
                                "Parameter skill. Minimal wajib: {'ticker': 'ANTM'}. "
                                "Opsional: 'days' (int), 'subsector' (str untuk peer_valuation_benchmark)."
                            ),
                        },
                    },
                    "required": ["skill_id", "arguments"],
                },
            },
            {
                "name": "query_sectors",
                "description": (
                    "Router universal ke Sectors Financial API v2. "
                    "Gunakan `domain` untuk memilih jenis data: "
                    "candles (OHLCV harian), "
                    "fundamentals (profil & rasio keuangan emiten), "
                    "foreign_flow (akumulasi/distribusi dana asing), "
                    "suspensions (suspensi & pengumuman UMA resmi BEI), "
                    "filings (kepemilikan orang dalam/insider trading), "
                    "broker_summary (top buyer/seller broker), "
                    "corporate_actions (dividen, split, rights issue), "
                    "subsector_peers (komparasi rekan subsektor), "
                    "mining_detail (detail operasional tambang & smelter), "
                    "news (berita terkurasi Sectors API)."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "domain": {
                            "type": "string",
                            "description": (
                                "Jenis data Sectors API. Wajib diisi. Pilih: "
                                "candles | fundamentals | foreign_flow | suspensions | "
                                "filings | broker_summary | corporate_actions | "
                                "subsector_peers | mining_detail | news."
                            ),
                        },
                        "ticker": {
                            "type": "string",
                            "description": "Kode ticker IDX 4 huruf (contoh: ANTM, BBCA). Tidak case-sensitive.",
                        },
                        "params": {
                            "type": "object",
                            "description": (
                                "Parameter tambahan opsional, misalnya: "
                                "{'subsector': 'metals-mining'} untuk subsector_peers, "
                                "{'slug': 'antm'} untuk mining_detail."
                            ),
                        },
                    },
                    "required": ["domain", "ticker"],
                },
            },
            {
                "name": "search_osint",
                "description": (
                    "Router universal ke mesin Dual-Engine OSINT. "
                    "Memanen berita terkurasi dari Sectors v2 API dan melakukan "
                    "targeted boolean dorking ke Google News RSS untuk menemukan "
                    "keterbukaan informasi IDXnet, Kontan, Bisnis, dan CNBC Indonesia. "
                    "Kosongkan ticker untuk berita pasar modal umum terkini."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {
                            "type": "string",
                            "description": (
                                "Kode ticker IDX (contoh: ANTM). "
                                "Kosongkan ('') untuk mengambil berita pasar modal umum."
                            ),
                        },
                        "query": {
                            "type": "string",
                            "description": "Kata kunci tambahan untuk mempersempit pencarian berita (opsional).",
                        },
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "query_memory",
                "description": (
                    "Router universal ke mesin memori graf percakapan lokal (SQLite + NetworkX). "
                    "Ambil konteks riwayat, relasi entitas, dan observasi masa lalu "
                    "terkait suatu ticker atau entitas pasar modal lintas sesi percakapan. "
                    "Gunakan sebelum memulai investigasi baru untuk mengecek riwayat temuan."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "concept_or_ticker": {
                            "type": "string",
                            "description": "Entitas atau ticker yang dicari dalam memori lokal (contoh: ANTM, Hari Darmawan).",
                        },
                        "radius": {
                            "type": "integer",
                            "description": "Kedalaman hop ego-graph (default 2, maks 3).",
                            "default": 2,
                        },
                    },
                    "required": ["concept_or_ticker"],
                },
            },
        ]

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
            return self.execute_skill(
                skill_id=arguments.get("skill_id", ""),
                arguments=arguments.get("arguments", {}),
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
            "get_mining_detail": lambda args: self.get_mining_detail(
                slug=args.get("slug", ""),
            ),
            "sectors_get_mining_detail": lambda args: self.get_mining_detail(
                slug=args.get("slug", ""),
            ),
            "harvest_market_news": lambda args: self.harvest_market_news(
                ticker=ticker if ticker and ticker not in _INDEX_TICKERS else None,
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
            "search_osint": lambda args: self.search_osint(
                ticker=args.get("ticker", ""),
                query=args.get("query", ""),
            ),
            "query_memory": lambda args: self.query_memory(
                concept_or_ticker=args.get("concept_or_ticker", args.get("ticker", "")),
                radius=int(args.get("radius", 2)),
            ),
        }

        handler = handlers.get(tool_name)
        if not handler:
            raise ValueError(f"Tool '{tool_name}' is not registered in Niskava Tool Registry.")

        return handler(arguments)

