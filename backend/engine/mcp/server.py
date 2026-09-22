"""Unified Model Context Protocol (MCP) Server for Niskava Agent.

Implements standard MCP JSON-RPC 2.0 (stdio) protocol version 2024-11-05
exposing:
- Sectors Financial API v2 Market Data Primitives (9 tools)
- Dual-Engine OSINT Harvester & Content Sanitizer (2 tools)
- Deterministic Quantitative Math Calculations (1 tool)
- Associative Local Graph Memory & Recall (2 tools)
- MCP Resources (System status & Cache metrics)
- MCP Prompts (Standardized investigative workflows)

Complies with:
- Law 1: Deterministic Before Generative
- Law 2: Strict Financial Non-Advisory Boundary
- Law 4: Local-First Data Sovereignty (SQLite)
- Law 5: Credit Budget Discipline
- Law 6: Local Conversational Graph Memory Engine
"""

import concurrent.futures
import json
import os
import sys
from typing import Any, Dict, List, Optional

from engine.mcp.prompts import get_prompt_definitions, get_prompt_messages
from engine.mcp.resources import get_resource_definitions, read_resource
from engine.mcp.tools.memory import execute_memory_tool, get_memory_tool_definitions
from engine.mcp.tools.osint import execute_osint_tool, get_osint_tool_definitions
from engine.mcp.tools.quant import execute_quant_tool, get_quant_tool_definitions
from engine.mcp.tools.sectors import (
    execute_sectors_tool,
    get_sectors_tool_definitions,
)
from engine.osint.harvester import DualEngineOSINTHarvester
from engine.sectors.client import SectorsAPIClient

MCP_TOOL_TIMEOUT_SECONDS: int = int(os.environ.get("MCP_TOOL_TIMEOUT_SECONDS", "30"))


class UnifiedMCPServer:
    """Unified Model Context Protocol (MCP) Server for Niskava."""

    SERVER_NAME = "niskava-mcp-engine"
    SERVER_VERSION = "1.0.0"
    PROTOCOL_VERSION = "2024-11-05"

    def __init__(
        self,
        db_path: Optional[str] = None,
        api_key: Optional[str] = None,
        mock_mode: Optional[bool] = None,
        client: Optional[SectorsAPIClient] = None,
        harvester: Optional[DualEngineOSINTHarvester] = None,
    ):
        resolved_db = db_path or os.environ.get("NISKAVA_DB_PATH", "~/.niskava/niskava.db")
        self.db_path = os.path.expanduser(resolved_db)
        self.mock_mode = mock_mode
        self.client = client or SectorsAPIClient(
            db_path=self.db_path,
            api_key=api_key,
            mock_mode=mock_mode,
        )
        self.harvester = harvester or DualEngineOSINTHarvester(
            mock_mode=mock_mode
        )

    def get_tool_definitions(self) -> List[Dict[str, Any]]:
        """Return all unified MCP tool schemas (Sectors, OSINT, Quant, Memory)."""
        tools: List[Dict[str, Any]] = []
        tools.extend(get_sectors_tool_definitions())
        tools.extend(get_osint_tool_definitions())
        tools.extend(get_quant_tool_definitions())
        tools.extend(get_memory_tool_definitions())
        return tools

    def execute_tool(self, name: str, arguments: Dict[str, Any]) -> Any:
        """Dispatch tool execution to the appropriate domain module."""
        def _dispatch() -> Any:
            if name.startswith("sectors_"):
                return execute_sectors_tool(self.client, name, arguments)
            if name.startswith("osint_"):
                return execute_osint_tool(self.harvester, self.client, name, arguments)
            if name.startswith("quant_"):
                return execute_quant_tool(self.client, name, arguments)
            if name.startswith("memory_"):
                return execute_memory_tool(self.db_path, name, arguments)
            raise ValueError(f"Unknown MCP tool: {name}")

        with concurrent.futures.ThreadPoolExecutor(max_workers=1) as executor:
            future = executor.submit(_dispatch)
            try:
                return future.result(timeout=MCP_TOOL_TIMEOUT_SECONDS)
            except concurrent.futures.TimeoutError:
                raise TimeoutError(
                    f"MCP tool '{name}' melebihi batas waktu {MCP_TOOL_TIMEOUT_SECONDS}s. "
                    "Periksa koneksi jaringan atau naikkan MCP_TOOL_TIMEOUT_SECONDS."
                )

    def get_resource_definitions(self) -> List[Dict[str, Any]]:
        """Return all MCP resource schemas."""
        return get_resource_definitions()

    def read_resource_content(self, uri: str) -> Dict[str, Any]:
        """Read MCP resource by URI."""
        return read_resource(uri, self.db_path, self.client)

    def get_prompt_definitions(self) -> List[Dict[str, Any]]:
        """Return all MCP prompt templates."""
        return get_prompt_definitions()

    def get_prompt(self, name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """Render prompt template."""
        return get_prompt_messages(name, arguments)

    def handle_request(self, request: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Process a single JSON-RPC 2.0 request."""
        req_id = request.get("id")
        method = request.get("method")
        params = request.get("params", {})

        if method == "initialize":
            return {
                "jsonrpc": "2.0",
                "id": req_id,
                "result": {
                    "protocolVersion": self.PROTOCOL_VERSION,
                    "capabilities": {
                        "tools": {"listChanged": False},
                        "resources": {"subscribe": False, "listChanged": False},
                        "prompts": {"listChanged": False},
                    },
                    "serverInfo": {
                        "name": self.SERVER_NAME,
                        "version": self.SERVER_VERSION,
                    },
                },
            }

        if method == "notifications/initialized":
            return None

        if method == "ping":
            return {
                "jsonrpc": "2.0",
                "id": req_id,
                "result": {},
            }

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

        if method == "resources/list":
            return {
                "jsonrpc": "2.0",
                "id": req_id,
                "result": {
                    "resources": self.get_resource_definitions(),
                },
            }

        if method == "resources/read":
            uri = params.get("uri", "")
            try:
                res_data = self.read_resource_content(uri)
                return {
                    "jsonrpc": "2.0",
                    "id": req_id,
                    "result": {
                        "contents": [
                            {
                                "uri": uri,
                                "mimeType": "application/json",
                                "text": json.dumps(res_data, indent=2, ensure_ascii=False),
                            }
                        ]
                    },
                }
            except Exception as exc:
                return {
                    "jsonrpc": "2.0",
                    "id": req_id,
                    "error": {
                        "code": -32602,
                        "message": f"Resource error: {str(exc)}",
                    },
                }

        if method == "prompts/list":
            return {
                "jsonrpc": "2.0",
                "id": req_id,
                "result": {
                    "prompts": self.get_prompt_definitions(),
                },
            }

        if method == "prompts/get":
            prompt_name = params.get("name", "")
            prompt_args = params.get("arguments", {})
            try:
                prompt_result = self.get_prompt(prompt_name, prompt_args)
                return {
                    "jsonrpc": "2.0",
                    "id": req_id,
                    "result": prompt_result,
                }
            except Exception as exc:
                return {
                    "jsonrpc": "2.0",
                    "id": req_id,
                    "error": {
                        "code": -32602,
                        "message": f"Prompt error: {str(exc)}",
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
    server = UnifiedMCPServer()
    server.run_stdio()


if __name__ == "__main__":
    main()
