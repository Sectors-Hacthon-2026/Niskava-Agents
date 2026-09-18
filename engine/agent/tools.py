"""Deterministic Tool Registry for Niskava Autonomous ReAct Agent.

Complies strictly with:
- Law 1: Deterministic Before Generative (NumPy calculates Z-scores, never LLM)
- Law 2: Strict Financial Non-Advisory Boundary (3-Tier Taxonomy)
- Law 5: Credit Budget Discipline (SQLite sectors_cache)
"""

import os
from typing import Any, Callable, Dict, List, Optional

from engine.osint.harvester import DualEngineOSINTHarvester, OSINTItem
from engine.quant.anomaly import AnomalyResult, detect_historical_anomalies
from engine.sectors.client import SectorsAPIClient


class NiskavaToolRegistry:
    """Provides structured, callable tools for the ReAct Agent."""

    def __init__(
        self,
        db_path: str,
        sectors_client: Optional[SectorsAPIClient] = None,
        osint_harvester: Optional[DualEngineOSINTHarvester] = None,
        mock_mode: Optional[bool] = None,
    ):
        self.db_path = os.path.expanduser(db_path)
        self.mock_mode = mock_mode
        self.sectors_client = sectors_client or SectorsAPIClient(
            db_path=self.db_path, mock_mode=self.mock_mode
        )
        self.osint_harvester = osint_harvester or DualEngineOSINTHarvester(
            mock_mode=self.mock_mode
        )

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

    def harvest_market_news(
        self,
        ticker: str,
        company_name: Optional[str] = None,
    ) -> List[Dict[str, Any]]:
        """Harvest curated news and targeted IDX regulatory filings via Dual-Engine OSINT."""
        if not company_name:
            report = self.get_company_fundamentals(ticker)
            company_name = report.get("company_name", ticker)

        sectors_news = self.sectors_client.get_news(ticker)
        items: List[OSINTItem] = self.osint_harvester.harvest(
            ticker=ticker.upper(),
            company_name=company_name,
            sectors_news_items=sectors_news,
        )
        return [item.model_dump() for item in items]

    def get_tool_definitions(self) -> List[Dict[str, Any]]:
        """Return JSON-schema compatible tool definitions for LLM function calling."""
        return [
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
                "name": "harvest_market_news",
                "description": "Panen berita pasar modal terkurasi dan keterbukaan informasi bursa resmi menggunakan arsitektur Dual-Engine OSINT.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker IDX"},
                        "company_name": {"type": "string", "description": "Nama resmi perseroan (opsional)"},
                    },
                    "required": ["ticker"],
                },
            },
        ]

    def execute_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Any:
        """Dynamically dispatch and execute a registered tool."""
        handlers: Dict[str, Callable[..., Any]] = {
            "get_daily_candles": lambda args: self.get_daily_candles(
                ticker=args.get("ticker", ""),
                days=args.get("days", 30),
            ),
            "compute_quant_anomalies": lambda args: self.compute_quant_anomalies(
                ticker=args.get("ticker", ""),
                volume_z_threshold=float(args.get("volume_z_threshold", 2.5)),
            ),
            "get_company_fundamentals": lambda args: self.get_company_fundamentals(
                ticker=args.get("ticker", ""),
            ),
            "get_foreign_flow": lambda args: self.get_foreign_flow(
                ticker=args.get("ticker", ""),
            ),
            "harvest_market_news": lambda args: self.harvest_market_news(
                ticker=args.get("ticker", ""),
                company_name=args.get("company_name"),
            ),
        }

        handler = handlers.get(tool_name)
        if not handler:
            raise ValueError(f"Tool '{tool_name}' tidak terdaftar di Niskava Tool Registry.")

        return handler(arguments)
