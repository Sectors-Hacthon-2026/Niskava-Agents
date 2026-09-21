"""Deterministic Tool Registry for Niskava Autonomous ReAct Agent.

Complies strictly with:
- Law 1: Deterministic Before Generative (NumPy calculates Z-scores, never LLM)
- Law 2: Strict Financial Non-Advisory Boundary (3-Tier Taxonomy)
- Law 5: Credit Budget Discipline (SQLite sectors_cache)
"""

import os
from typing import Any, Callable, Dict, List, Optional

from engine.memory.graph_memory import LocalGraphMemory
from engine.osint.harvester import DualEngineOSINTHarvester, OSINTItem
from engine.quant.anomaly import AnomalyResult, detect_historical_anomalies
from engine.sectors.client import SectorsAPIClient
from engine.skills.registry import SkillsRegistry


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
        """Return JSON-schema compatible tool definitions for LLM function calling."""
        definitions: List[Dict[str, Any]] = [
            {
                "name": "get_daily_candles",
                "description": "Ambil data deret waktu harga dan volume perdagangan harian saham IDX dari Sectors API v2.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker 4 huruf IDX (contoh: ANTM, BBCA)"},
                        "days": {"type": "integer", "description": "Jendela waktu observasi (default 30 hari)", "default": 30},
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "compute_quant_anomalies",
                "description": "Hitung anomali statistik volume (MA20 Z-Score) dan lonjakan harga abnormal secara deterministik menggunakan NumPy (Law 1).",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker IDX (contoh: ANTM)"},
                        "volume_z_threshold": {"type": "number", "description": "Ambang batas Z-Score lonjakan volume (default: 2.5)", "default": 2.5},
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "get_company_fundamentals",
                "description": "Ambil profil fundamental, nama resmi perseroan, kapitalisasi pasar, dan sektor industri emiten.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker IDX"},
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "get_foreign_flow",
                "description": "Ambil data akumulasi atau distribusi dana investor asing (Net Foreign Flow).",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker IDX"},
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "get_suspensions",
                "description": "Ambil riwayat suspensi bursa, pengumuman UMA, dan tautan surat pengumuman PDF resmi BEI.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker IDX"},
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "get_corporate_actions",
                "description": "Ambil jadwal aksi korporasi emiten (dividen, stock split, rights issue).",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker IDX"},
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "get_filings",
                "description": "Ambil pelaporan transaksi kepemilikan orang dalam (insider trading) direksi dan komisaris.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker IDX"},
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "get_broker_summary",
                "description": "Ambil daftar broker pembeli bersih (top buyers) dan penjual bersih (top sellers) teratas.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker IDX"},
                    },
                    "required": ["ticker"],
                },
            },
            {
                "name": "get_subsector_peers",
                "description": "Ambil data komparasi emiten dan rata-rata industri subsektor untuk analisis divergensi.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "subsector": {"type": "string", "description": "Slug subsektor industri"},
                    },
                    "required": ["subsector"],
                },
            },
            {
                "name": "get_mining_detail",
                "description": "Ambil detail operasional konsesi tambang dan fasilitas smelter emiten.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "slug": {"type": "string", "description": "Slug emiten tambang"},
                    },
                    "required": ["slug"],
                },
            },
            {
                "name": "harvest_market_news",
                "description": "Panen berita pasar modal terkurasi dan keterbukaan informasi bursa resmi menggunakan arsitektur Dual-Engine OSINT. Jika ticker tidak diisi, mengambil berita pasar modal umum terkini.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker IDX (opsional, kosongkan jika mencari berita pasar umum)"},
                        "company_name": {"type": "string", "description": "Nama resmi perseroan (opsional)"},
                    },
                    "required": [],
                },
            },
            {
                "name": "memory_recall_context",
                "description": "Ambil konteks masa lalu dan relasi graf memori lokal untuk suatu entitas/ticker pasar modal.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "query_entity": {"type": "string", "description": "Nama entitas atau ticker saham yang dicari (contoh: ANTM)"},
                        "radius": {"type": "integer", "description": "Kedalaman hop penelusuran Ego-Graph (default 2)", "default": 2},
                    },
                    "required": ["query_entity"],
                },
            },
            {
                "name": "memory_store_observation",
                "description": "Simpan observasi relasi baru ke dalam basis data graf memori lokal.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "source_label": {"type": "string", "description": "Label entitas asal"},
                        "relation": {"type": "string", "description": "Relasi/predikat penghubung (misal: HOLDS_AT, OPERATES)"},
                        "target_label": {"type": "string", "description": "Label entitas tujuan"},
                        "context_snippet": {"type": "string", "description": "Kutipan atau konteks bukti"},
                    },
                    "required": ["source_label", "relation", "target_label"],
                },
            },
            {
                "name": "memory_find_connection",
                "description": "Lacak rute hubungan terpendek (shortest path) antara dua entitas pasar untuk menemukan keterkaitan tersembunyi.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "source_entity": {"type": "string", "description": "Entitas pertama (contoh: ANTM)"},
                        "target_entity": {"type": "string", "description": "Entitas kedua (contoh: BBCA)"},
                    },
                    "required": ["source_entity", "target_entity"],
                },
            },
        ]
        # Append Layer 3 Domain Skills tool definitions
        definitions.extend(self.skills_registry.get_all_tool_definitions())
        return definitions

    def execute_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Any:
        """Dynamically dispatch and execute a registered tool (supporting direct & MCP names)."""
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
                ticker=ticker,
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
        }

        handler = handlers.get(tool_name)
        if not handler:
            raise ValueError(f"Tool '{tool_name}' tidak terdaftar di Niskava Tool Registry.")

        return handler(arguments)

