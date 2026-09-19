"""Unit Tests for Sectors MCP Server.

Verifies:
- Standard JSON-RPC 2.0 protocol compliance
- tools/list schema validation
- tools/call execution across all 9 tools
- Error handling for invalid/unregistered tools
"""

import json
import pytest
from engine.sectors.mcp_server import SectorsMCPServer


@pytest.fixture
def mcp_server(tmp_path):
    db_file = str(tmp_path / "test_mcp.db")
    return SectorsMCPServer(db_path=db_file, mock_mode=True)


def test_mcp_initialize(mcp_server):
    req = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {"protocolVersion": "2024-11-05"},
    }
    resp = mcp_server.handle_request(req)
    assert resp["jsonrpc"] == "2.0"
    assert resp["id"] == 1
    assert "capabilities" in resp["result"]
    assert resp["result"]["serverInfo"]["name"] == "sectors-mcp-server"


def test_mcp_tools_list(mcp_server):
    req = {
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list",
    }
    resp = mcp_server.handle_request(req)
    assert resp["jsonrpc"] == "2.0"
    assert resp["id"] == 2
    tools = resp["result"]["tools"]
    assert len(tools) == 9

    tool_names = [t["name"] for t in tools]
    expected_tools = [
        "sectors_get_daily_candles",
        "sectors_get_company_report",
        "sectors_get_foreign_flow",
        "sectors_get_suspensions",
        "sectors_get_corporate_actions",
        "sectors_get_filings",
        "sectors_get_broker_summary",
        "sectors_get_subsector_peers",
        "sectors_get_mining_detail",
    ]
    for expected in expected_tools:
        assert expected in tool_names

    # Check schema completeness
    for t in tools:
        assert "description" in t
        assert "inputSchema" in t
        assert t["inputSchema"]["type"] == "object"


def test_mcp_tool_call_candles(mcp_server):
    req = {
        "jsonrpc": "2.0",
        "id": 3,
        "method": "tools/call",
        "params": {
            "name": "sectors_get_daily_candles",
            "arguments": {"ticker": "ANTM", "days": 30},
        },
    }
    resp = mcp_server.handle_request(req)
    assert resp["jsonrpc"] == "2.0"
    assert resp["id"] == 3
    assert resp["result"]["isError"] is False
    content = resp["result"]["content"]
    assert len(content) == 1
    assert content[0]["type"] == "text"

    data = json.loads(content[0]["text"])
    assert isinstance(data, list)
    assert len(data) == 30


def test_mcp_tool_call_all_endpoints(mcp_server):
    """Test calling each of the remaining 8 tools."""
    tool_calls = [
        ("sectors_get_company_report", {"ticker": "ANTM"}),
        ("sectors_get_foreign_flow", {"ticker": "ANTM"}),
        ("sectors_get_suspensions", {"ticker": "ANTM"}),
        ("sectors_get_corporate_actions", {"ticker": "ANTM"}),
        ("sectors_get_filings", {"ticker": "ANTM"}),
        ("sectors_get_broker_summary", {"ticker": "ANTM"}),
        ("sectors_get_subsector_peers", {"subsector": "metals-and-minerals-mining"}),
        ("sectors_get_mining_detail", {"slug": "aneka-tambang"}),
    ]

    for idx, (name, args) in enumerate(tool_calls, start=10):
        req = {
            "jsonrpc": "2.0",
            "id": idx,
            "method": "tools/call",
            "params": {"name": name, "arguments": args},
        }
        resp = mcp_server.handle_request(req)
        assert resp["id"] == idx
        assert resp["result"]["isError"] is False
        text_data = resp["result"]["content"][0]["text"]
        parsed = json.loads(text_data)
        assert parsed is not None


def test_mcp_unknown_tool_error(mcp_server):
    req = {
        "jsonrpc": "2.0",
        "id": 99,
        "method": "tools/call",
        "params": {"name": "non_existent_tool", "arguments": {}},
    }
    resp = mcp_server.handle_request(req)
    assert resp["id"] == 99
    assert resp["result"]["isError"] is True
    assert "Unknown MCP tool" in resp["result"]["content"][0]["text"]


def test_mcp_unknown_method(mcp_server):
    req = {
        "jsonrpc": "2.0",
        "id": 100,
        "method": "unknown/method",
    }
    resp = mcp_server.handle_request(req)
    assert resp["id"] == 100
    assert "error" in resp
    assert resp["error"]["code"] == -32601
