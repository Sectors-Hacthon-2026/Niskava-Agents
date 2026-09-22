"""Unit Tests for MCP Tool Execution Timeout.

Verifies:
- MCP_TOOL_TIMEOUT_SECONDS constant definition and default value
- Timeout enforcement terminates slow-running tools and raises TimeoutError
- Normal tool execution completes successfully within timeout window
"""

import sqlite3
import time
from unittest.mock import patch
import pytest

from engine.mcp.server import MCP_TOOL_TIMEOUT_SECONDS, UnifiedMCPServer


@pytest.fixture
def mcp_server(tmp_path):
    """Fixture providing a mock UnifiedMCPServer instance."""
    db_file = str(tmp_path / "test_mcp_timeout.db")
    conn = sqlite3.connect(db_file)
    conn.execute("""
        CREATE TABLE IF NOT EXISTS sectors_cache (
            cache_key TEXT PRIMARY KEY,
            endpoint TEXT NOT NULL,
            payload_json TEXT NOT NULL,
            created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
            expires_at TEXT
        );
    """)
    conn.commit()
    conn.close()

    return UnifiedMCPServer(db_path=db_file, mock_mode=True)


def test_mcp_tool_timeout_constant_exists():
    """Ensure MCP_TOOL_TIMEOUT_SECONDS exists, is an integer, and has expected default."""
    assert isinstance(MCP_TOOL_TIMEOUT_SECONDS, int)
    assert MCP_TOOL_TIMEOUT_SECONDS > 0
    assert MCP_TOOL_TIMEOUT_SECONDS == 30


def test_mcp_tool_timeout_terminates_slow_tool(mcp_server):
    """Mock execute_sectors_tool with slow sleep and patched timeout of 1 second."""
    def slow_tool(*args, **kwargs):
        time.sleep(2.0)
        return [{"symbol": "ANTM", "close": 1500}]

    with patch("engine.mcp.server.MCP_TOOL_TIMEOUT_SECONDS", 1), \
         patch("engine.mcp.server.execute_sectors_tool", side_effect=slow_tool):
        # 1. Direct execute_tool raises TimeoutError
        with pytest.raises(TimeoutError) as exc_info:
            mcp_server.execute_tool(
                "sectors_get_daily_candles",
                {"ticker": "ANTM", "days": 5},
            )
        assert "melebihi batas waktu" in str(exc_info.value)
        assert "1s" in str(exc_info.value)

        # 2. handle_request returns isError: True with informative error message
        resp = mcp_server.handle_request({
            "jsonrpc": "2.0",
            "id": 101,
            "method": "tools/call",
            "params": {
                "name": "sectors_get_daily_candles",
                "arguments": {"ticker": "ANTM", "days": 5},
            },
        })
        assert resp["result"]["isError"] is True
        error_msg = resp["result"]["content"][0]["text"]
        assert "melebihi batas waktu" in error_msg
        assert "1s" in error_msg


def test_mcp_tool_normal_execution_completes(mcp_server):
    """Verify normal tool execution completes successfully within timeout window."""
    result = mcp_server.execute_tool(
        "sectors_get_daily_candles",
        {"ticker": "ANTM", "days": 5},
    )
    assert isinstance(result, list)
    assert len(result) == 5

    # Also test via handle_request
    resp = mcp_server.handle_request({
        "jsonrpc": "2.0",
        "id": 102,
        "method": "tools/call",
        "params": {
            "name": "sectors_get_daily_candles",
            "arguments": {"ticker": "ANTM", "days": 5},
        },
    })
    assert resp["id"] == 102
    assert resp["result"]["isError"] is False
    assert len(resp["result"]["content"]) > 0
