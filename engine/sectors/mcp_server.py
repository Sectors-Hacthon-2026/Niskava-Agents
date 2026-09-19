"""Sectors Financial API v2 Model Context Protocol (MCP) Server.

Implements standard MCP JSON-RPC 2.0 (stdio) protocol for exposing
IDX ground truth market data to AI agents (Claude Desktop, Cursor, Hermes, Niskava).

Complies with:
- Law 1: Deterministic Before Generative (Structured factual data)
- Law 4: Local-First Data Sovereignty (Transparent SQLite caching)
- Law 5: Credit Budget Discipline (Local sectors_cache)
"""

import json
import os
import sys
from typing import Any, Dict, List, Optional

from engine.sectors.client import SectorsAPIClient


class SectorsMCPServer:
    """Model Context Protocol (MCP) Server exposing Sectors Financial API v2."""

    SERVER_NAME = "sectors-mcp-server"
    SERVER_VERSION = "1.0.0"

    def __init__(
        self,
        db_path: str = "~/.niskava/niskava.db",
        api_key: Optional[str] = None,
        mock_mode: Optional[bool] = None,
        client: Optional[SectorsAPIClient] = None,
    ):
        self.db_path = os.path.expanduser(db_path)
        self.client = client or SectorsAPIClient(
            db_path=self.db_path,
            api_key=api_key,
            mock_mode=mock_mode,
        )

    def get_tool_definitions(self) -> List[Dict[str, Any]]:
        """Return MCP standardized tool schemas."""
        return [
            {
                "name": "sectors_get_daily_candles",
                "description": "Ambil deret waktu harga dan volume perdagangan harian (OHLCV) saham IDX dari Sectors API v2.",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "ticker": {"type": "string", "description": "Kode ticker 4 huruf IDX (contoh: ANTM, BBRI)"},
                        "days": {"type": "integer", "description": "Jendela observasi harian (default: 30)", "default": 30},
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
                        "ticker": {"type": "string", "description": "Kode ticker 4 huruf IDX (contoh: ANTM)"},
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
                        "ticker": {"type": "string", "description": "Kode ticker 4 huruf IDX"},
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
                        "ticker": {"type": "string", "description": "Kode ticker 4 huruf IDX"},
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
                        "ticker": {"type": "string", "description": "Kode ticker 4 huruf IDX"},
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
                        "ticker": {"type": "string", "description": "Kode ticker 4 huruf IDX"},
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
                        "ticker": {"type": "string", "description": "Kode ticker 4 huruf IDX"},
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

    def execute_tool(self, name: str, arguments: Dict[str, Any]) -> Any:
        """Dispatch tool call to SectorsAPIClient."""
        ticker = arguments.get("ticker", "").upper()
        if name == "sectors_get_daily_candles":
            days = arguments.get("days", 30)
            candles = self.client.get_daily_candles(ticker)
            if days and len(candles) > days:
                return candles[-days:]
            return candles

        if name == "sectors_get_company_report":
            sections = arguments.get("sections", "valuation,financials,peers")
            return self.client.get_company_report(ticker, sections=sections)

        if name == "sectors_get_foreign_flow":
            return self.client.get_foreign_flow(ticker)

        if name == "sectors_get_suspensions":
            return self.client.get_suspensions(ticker)

        if name == "sectors_get_corporate_actions":
            return self.client.get_corporate_actions(ticker)

        if name == "sectors_get_filings":
            return self.client.get_filings(ticker)

        if name == "sectors_get_broker_summary":
            return self.client.get_broker_summary(ticker)

        if name == "sectors_get_subsector_peers":
            subsector = arguments.get("subsector", "")
            return self.client.get_subsector_peers(subsector)

        if name == "sectors_get_mining_detail":
            slug = arguments.get("slug", "")
            return self.client.get_mining_detail(slug)

        raise ValueError(f"Unknown MCP tool: {name}")

    def handle_request(self, request: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Handle a single JSON-RPC 2.0 request dictionary."""
        req_id = request.get("id")
        method = request.get("method")
        params = request.get("params", {})

        if method == "initialize":
            return {
                "jsonrpc": "2.0",
                "id": req_id,
                "result": {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {
                        "tools": {"listChanged": False},
                    },
                    "serverInfo": {
                        "name": self.SERVER_NAME,
                        "version": self.SERVER_VERSION,
                    },
                },
            }

        if method == "notifications/initialized":
            return None

        if method == "tools/list":
            return {
                "jsonrpc": "2.0",
                "id": req_id,
                "result": {
                    "tools": self.get_tool_definitions(),
                },
            }

        if method == "tools/call":
            tool_name = params.get("name", "")
            tool_args = params.get("arguments", {})
            try:
                data = self.execute_tool(tool_name, tool_args)
                return {
                    "jsonrpc": "2.0",
                    "id": req_id,
                    "result": {
                        "content": [
                            {
                                "type": "text",
                                "text": json.dumps(data, indent=2, ensure_ascii=False),
                            }
                        ],
                        "isError": False,
                    },
                }
            except Exception as exc:
                return {
                    "jsonrpc": "2.0",
                    "id": req_id,
                    "result": {
                        "content": [
                            {
                                "type": "text",
                                "text": f"Error executing tool '{tool_name}': {str(exc)}",
                            }
                        ],
                        "isError": True,
                    },
                }

        # Unknown method error
        if req_id is not None:
            return {
                "jsonrpc": "2.0",
                "id": req_id,
                "error": {
                    "code": -32601,
                    "message": f"Method not found: {method}",
                },
            }
        return None

    def run_stdio(self) -> None:
        """Read JSON-RPC from STDIN line-by-line and write responses to STDOUT."""
        for line in sys.stdin:
            line = line.strip()
            if not line:
                continue
            try:
                request = json.loads(line)
                response = self.handle_request(request)
                if response is not None:
                    sys.stdout.write(json.dumps(response) + "\n")
                    sys.stdout.flush()
            except Exception as exc:
                err_resp = {
                    "jsonrpc": "2.0",
                    "id": None,
                    "error": {"code": -32700, "message": f"Parse error: {str(exc)}"},
                }
                sys.stdout.write(json.dumps(err_resp) + "\n")
                sys.stdout.flush()


def main() -> None:
    server = SectorsMCPServer()
    server.run_stdio()


if __name__ == "__main__":
    main()
