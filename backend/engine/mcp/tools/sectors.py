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
            "description": "Fetch daily OHLCV candlestick price and volume series from Sectors API v2.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "4-letter IDX stock ticker symbol (e.g. ANTM, BBRI)",
                    },
                    "days": {
                        "type": "integer",
                        "description": "Daily observation lookback window (default: 30)",
                        "default": 30,
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_company_report",
            "description": "Fetch company fundamental profile, valuation metrics (PE, PBV, ROE), and overview from Sectors API v2.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "4-letter IDX stock ticker symbol (e.g. ANTM)",
                    },
                    "sections": {
                        "type": "string",
                        "description": "Report sections to retrieve (default: 'valuation,financials,peers')",
                        "default": "valuation,financials,peers",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_foreign_flow",
            "description": "Fetch foreign investor net capital accumulation and distribution flow (Net Foreign Flow).",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "4-letter IDX stock ticker symbol",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_suspensions",
            "description": "Fetch official IDX trading suspensions, Unusual Market Activity (UMA) notices, and regulatory announcement URLs.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "4-letter IDX stock ticker symbol",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_corporate_actions",
            "description": "Fetch corporate action schedules (cash dividends, stock splits, rights issues).",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "4-letter IDX stock ticker symbol",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_filings",
            "description": "Fetch insider trading disclosures and ownership filings by directors and commissioners.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "4-letter IDX stock ticker symbol",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_broker_summary",
            "description": "Fetch top net buying and selling brokerage participants.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "4-letter IDX stock ticker symbol",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "sectors_get_subsector_peers",
            "description": "Fetch subsector industry peers and valuation multiples for divergence analysis.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "subsector": {
                        "type": "string",
                        "description": "Subsector industry slug (e.g. 'metals-and-minerals-mining')",
                    },
                },
                "required": ["subsector"],
            },
        },
        {
            "name": "sectors_get_mining_detail",
            "description": "Fetch mining operational details, IUP concession permits, and smelter asset data.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "slug": {
                        "type": "string",
                        "description": "Mining company slug (e.g. 'aneka-tambang')",
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
