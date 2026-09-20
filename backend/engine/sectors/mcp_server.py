"""Sectors Financial API v2 Model Context Protocol (MCP) Server (Legacy Wrapper).

Maintains backward compatibility while delegating to modular tool implementations.
For the unified full-stack MCP server, use engine.mcp.UnifiedMCPServer.
"""

import json
import os
import sys
from typing import Any, Dict, List, Optional

from engine.mcp.tools.sectors import (
    execute_sectors_tool,
    get_sectors_tool_definitions,
)
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
        """Return MCP standardized tool schemas for Sectors API."""
        return get_sectors_tool_definitions()

    def execute_tool(self, name: str, arguments: Dict[str, Any]) -> Any:
        """Dispatch tool call to SectorsAPIClient."""
        return execute_sectors_tool(self.client, name, arguments)

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
