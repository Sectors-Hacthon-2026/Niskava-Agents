"""Sectors Financial API v2 MCP Tool Definitions and Dispatcher.

Complies strictly with:
- Law 1: Deterministic Before Generative (Structured factual data)
- Law 4: Local-First Data Sovereignty (Transparent SQLite caching)
- Law 5: Credit Budget Discipline (Local sectors_cache)
"""

import re
from typing import Any, Dict, List
from engine.sectors.client import SectorsAPIClient


def sanitize_ticker(raw: str) -> str:
    """Normalize raw ticker string to standard 4-5 letter IDX symbol.

    Strips leading/trailing whitespace, common prefixes (IDX:, BEI:),
    and regional suffixes (.JK, .IJ).
    """
    if not raw:
        raise ValueError("Kode ticker saham tidak boleh kosong.")
    cleaned = str(raw).strip().upper()
    cleaned = re.sub(r"^(IDX|BEI):", "", cleaned)
    cleaned = re.sub(r"\.(JK|IJ)$", "", cleaned)
    cleaned = cleaned.strip()
    if not re.match(r"^[A-Z]{4,5}$", cleaned):
        raise ValueError(
            f"Kode ticker IDX tidak valid: '{raw}'. Format yang valid adalah 4-5 huruf alfabet (contoh: ANTM, BBRI, BRIS)."
        )
    return cleaned


def get_sectors_tool_definitions() -> List[Dict[str, Any]]:
    """Return standardized MCP schemas for Sectors Financial API tools."""
    return [
        {
            "name": "sectors_get_daily_candles",
            "description": "Ambil deret waktu harga dan volume perdagangan harian (OHLCV) saham IDX dari Sectors API v2.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "Kode ticker 4 huruf IDX (contoh: ANTM, BBRI)",
                    },
                    "days": {
                        "type": "integer",
                        "description": "Jendela observasi harian (default: 30)",
                        "default": 30,
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_company_report",
            "description": "Ambil profil fundamental, valuasi (PE, PBV, ROE), dan gambaran umum emiten dari Sectors API v2.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "Kode ticker 4 huruf IDX (contoh: ANTM)",
                    },
                    "sections": {
                        "type": "string",
                        "description": "Bagian laporan yang diambil (default: 'valuation,financials,peers')",
                        "default": "valuation,financials,peers",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_foreign_flow",
            "description": "Ambil data akumulasi dan distribusi modal investor asing (Net Foreign Flow) per saham.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "Kode ticker 4 huruf IDX",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_suspensions",
            "description": "Ambil catatan suspensi resmi bursa, pengumuman UMA, dan tautan surat pengumuman PDF resmi BEI.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "Kode ticker 4 huruf IDX",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_corporate_actions",
            "description": "Ambil jadwal aksi korporasi emiten (dividen, stock split, rights issue).",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "Kode ticker 4 huruf IDX",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_filings",
            "description": "Ambil laporan transaksi kepemilikan orang dalam (insider trading) direksi dan komisaris.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "Kode ticker 4 huruf IDX",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_broker_summary",
            "description": "Ambil ringkasan broker pembeli bersih (top buyers) dan penjual bersih (top sellers) teratas.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "Kode ticker 4 huruf IDX",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_subsector_peers",
            "description": "Ambil data komparasi emiten dan rata-rata industri subsektor untuk analisis divergensi.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "subsector": {
                        "type": "string",
                        "description": "Slug subsektor industri (contoh: 'metals-and-minerals-mining')",
                    },
                },
                "required": ["subsector"],
            },
        },
        {
            "name": "sectors_get_mining_detail",
            "description": "Ambil detail operasional tambang, izin konsesi IUP, dan lokasi smelter emiten pertambangan.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "slug": {
                        "type": "string",
                        "description": "Slug emiten tambang (contoh: 'aneka-tambang')",
                    },
                },
                "required": ["slug"],
            },
        },
    ]


KNOWN_SECTORS_TOOLS = {
    "sectors_get_daily_candles",
    "sectors_get_company_report",
    "sectors_get_foreign_flow",
    "sectors_get_suspensions",
    "sectors_get_corporate_actions",
    "sectors_get_filings",
    "sectors_get_broker_summary",
    "sectors_get_subsector_peers",
    "sectors_get_mining_detail",
}


def execute_sectors_tool(
    client: SectorsAPIClient, name: str, arguments: Dict[str, Any]
) -> Any:
    """Execute a Sectors API tool call against SectorsAPIClient."""
    if name not in KNOWN_SECTORS_TOOLS:
        raise ValueError(f"Unknown MCP tool: {name}")

    if name == "sectors_get_subsector_peers":
        subsector = arguments.get("subsector", "")
        return client.get_subsector_peers(subsector)

    if name == "sectors_get_mining_detail":
        slug = arguments.get("slug", "")
        return client.get_mining_detail(slug)

    ticker = sanitize_ticker(arguments.get("ticker", ""))

    if name == "sectors_get_daily_candles":
        days = arguments.get("days", 30)
        candles = client.get_daily_candles(ticker)
        if days and len(candles) > days:
            return candles[-days:]
        return candles

    if name == "sectors_get_company_report":
        sections = arguments.get("sections", "valuation,financials,peers")
        return client.get_company_report(ticker, sections=sections)

    if name == "sectors_get_foreign_flow":
        return client.get_foreign_flow(ticker)

    if name == "sectors_get_suspensions":
        return client.get_suspensions(ticker)

    if name == "sectors_get_corporate_actions":
        return client.get_corporate_actions(ticker)

    if name == "sectors_get_filings":
        return client.get_filings(ticker)

    if name == "sectors_get_broker_summary":
        return client.get_broker_summary(ticker)

    if name == "sectors_get_subsector_peers":
        subsector = arguments.get("subsector", "")
        return client.get_subsector_peers(subsector)

    if name == "sectors_get_mining_detail":
        slug = arguments.get("slug", "")
        return client.get_mining_detail(slug)

    raise ValueError(f"Unknown MCP tool: {name}")
