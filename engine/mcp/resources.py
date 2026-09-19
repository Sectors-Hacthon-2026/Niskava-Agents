"""MCP Resource Providers for Niskava Agent.

Exposes system state, database metrics, and cache statistics via standard
MCP resources/list and resources/read endpoints.
"""

import json
import os
import sqlite3
from typing import Any, Dict, List, Optional
from engine.sectors.client import SectorsAPIClient


def get_resource_definitions() -> List[Dict[str, Any]]:
    """Return list of standard MCP resources available on Niskava."""
    return [
        {
            "uri": "niskava://status",
            "name": "Niskava Agent System Status",
            "description": "Informasi status daemon, basis data SQLite, dan ketersediaan API key.",
            "mimeType": "application/json",
        },
        {
            "uri": "niskava://cache-stats",
            "name": "Sectors & OSINT Cache Statistics",
            "description": "Statistik jumlah entri cache lokal di sectors_cache untuk audit disiplin kredit.",
            "mimeType": "application/json",
        },
        {
            "uri": "niskava://graph-stats",
            "name": "Local Conversational Graph Memory Statistics",
            "description": "Statistik topologi graf memori percakapan, jumlah simpul entitas, relasi, dan entitas sentral.",
            "mimeType": "application/json",
        },
    ]


def read_resource(
    uri: str,
    db_path: str,
    sectors_client: Optional[SectorsAPIClient] = None,
) -> Dict[str, Any]:
    """Read contents of an MCP resource by URI."""
    if uri == "niskava://status":
        db_exists = os.path.exists(db_path)
        db_size_bytes = os.path.getsize(db_path) if db_exists else 0
        return {
            "status": "HEALTHY",
            "version": "1.0.0",
            "database_path": db_path,
            "database_exists": db_exists,
            "database_size_bytes": db_size_bytes,
            "mock_mode": getattr(sectors_client, "mock_mode", False),
            "sectors_api_configured": bool(
                getattr(sectors_client, "api_key", None)
            ),
        }

    if uri == "niskava://cache-stats":
        if not os.path.exists(db_path):
            return {"sectors_cache_count": 0, "status": "NO_DATABASE"}

        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        try:
            cursor.execute("SELECT COUNT(*) FROM sectors_cache")
            count = cursor.fetchone()[0]
            cursor.execute(
                "SELECT COUNT(*) FROM sectors_cache WHERE expires_at IS NULL"
            )
            permanent = cursor.fetchone()[0]
            return {
                "sectors_cache_total": count,
                "sectors_cache_permanent_historical": permanent,
                "credit_savings_policy": "Permanent cache on historical candles (Law 5)",
            }
        except Exception as exc:
            return {"error": str(exc)}
        finally:
            conn.close()

    if uri == "niskava://graph-stats":
        from engine.memory.graph_memory import LocalGraphMemory
        mem = LocalGraphMemory(db_path=db_path)
        stats = mem.get_graph_stats()
        return {
            "status": "OK",
            "database_path": db_path,
            **stats,
        }

    raise ValueError(f"Resource not found: {uri}")
