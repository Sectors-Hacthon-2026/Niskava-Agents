"""Unified Model Context Protocol (MCP) Engine for Niskava Agent.

Exposes IDX market data, OSINT intelligence, deterministic quant math,
and graph memory to AI agents (Claude Desktop, Cursor, Antigravity, Niskava).
"""

from typing import Any

__all__ = ["UnifiedMCPServer"]


def __getattr__(name: str) -> Any:
    if name == "UnifiedMCPServer":
        from engine.mcp.server import UnifiedMCPServer
        return UnifiedMCPServer
    raise AttributeError(f"module {__name__!r} has no attribute {name!r}")
