"""Unit Tests for Unified MCP Server (Niskava MCP Engine).

Verifies:
- Standard JSON-RPC 2.0 protocol compliance (version 2024-11-05)
- tools/list schema validation across all 14 tools (Sectors, OSINT, Quant, Memory)
- tools/call execution across all domains
- resources/list and resources/read functionality
- prompts/list and prompts/get functionality
- Error handling for invalid/unregistered tools, methods, resources, and prompts
"""

import json
import sqlite3
import pytest
from engine.mcp.server import UnifiedMCPServer


@pytest.fixture
def mcp_server(tmp_path):
    db_file = str(tmp_path / "test_unified_mcp.db")
    # Initialize basic sqlite schema needed for memory tests
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
    conn.execute("""
        CREATE TABLE IF NOT EXISTS memory_nodes (
            id TEXT PRIMARY KEY,
            label TEXT NOT NULL,
            node_type TEXT NOT NULL,
            metadata_json TEXT,
            last_observed_at TEXT NOT NULL,
            created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
    """)
    conn.execute("""
        CREATE TABLE IF NOT EXISTS memory_edges (
            source_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
            target_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
            relation TEXT NOT NULL,
            context_snippet TEXT,
            session_id TEXT,
            weight REAL NOT NULL DEFAULT 1.0,
            last_observed_at TEXT NOT NULL,
            created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (source_id, target_id, relation)
        );
    """)
    conn.commit()
    conn.close()

    return UnifiedMCPServer(db_path=db_file, mock_mode=True)


def test_unified_mcp_initialize(mcp_server):
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
    assert "tools" in resp["result"]["capabilities"]
    assert "resources" in resp["result"]["capabilities"]
    assert "prompts" in resp["result"]["capabilities"]
    assert resp["result"]["serverInfo"]["name"] == "niskava-mcp-engine"
    assert resp["result"]["protocolVersion"] == "2024-11-05"


def test_unified_mcp_ping(mcp_server):
    req = {"jsonrpc": "2.0", "id": 2, "method": "ping"}
    resp = mcp_server.handle_request(req)
    assert resp["jsonrpc"] == "2.0"
    assert resp["id"] == 2
    assert resp["result"] == {}


def test_unified_mcp_initialized_notification(mcp_server):
    req = {"jsonrpc": "2.0", "method": "notifications/initialized"}
    resp = mcp_server.handle_request(req)
    assert resp is None


def test_unified_mcp_tools_list(mcp_server):
    req = {"jsonrpc": "2.0", "id": 3, "method": "tools/list"}
    resp = mcp_server.handle_request(req)
    assert resp["id"] == 3
    tools = resp["result"]["tools"]
    assert len(tools) == 14

    tool_names = [t["name"] for t in tools]
    expected_tools = [
        # Sectors tools (9)
        "sectors_get_daily_candles",
        "sectors_get_company_report",
        "sectors_get_foreign_flow",
        "sectors_get_suspensions",
        "sectors_get_corporate_actions",
        "sectors_get_filings",
        "sectors_get_broker_summary",
        "sectors_get_subsector_peers",
        "sectors_get_mining_detail",
        # OSINT tools (2)
        "osint_harvest_market_news",
        "osint_extract_article_content",
        # Quant tools (1)
        "quant_compute_anomalies",
        # Memory tools (2)
        "memory_recall_context",
        "memory_store_observation",
    ]
    for expected in expected_tools:
        assert expected in tool_names, f"Missing tool: {expected}"

    for t in tools:
        assert "description" in t
        assert "inputSchema" in t
        assert t["inputSchema"]["type"] == "object"


def test_unified_mcp_sectors_tool_call(mcp_server):
    req = {
        "jsonrpc": "2.0",
        "id": 4,
        "method": "tools/call",
        "params": {
            "name": "sectors_get_daily_candles",
            "arguments": {"ticker": "ANTM", "days": 10},
        },
    }
    resp = mcp_server.handle_request(req)
    assert resp["id"] == 4
    assert resp["result"]["isError"] is False
    content = resp["result"]["content"][0]["text"]
    data = json.loads(content)
    assert isinstance(data, list)
    assert len(data) == 10


def test_unified_mcp_osint_tool_calls(mcp_server):
    # 1. Test osint_harvest_market_news
    req1 = {
        "jsonrpc": "2.0",
        "id": 5,
        "method": "tools/call",
        "params": {
            "name": "osint_harvest_market_news",
            "arguments": {"ticker": "ANTM", "company_name": "Aneka Tambang"},
        },
    }
    resp1 = mcp_server.handle_request(req1)
    assert resp1["id"] == 5
    assert resp1["result"]["isError"] is False
    items = json.loads(resp1["result"]["content"][0]["text"])
    assert isinstance(items, list)
    assert len(items) > 0

    # 2. Test osint_extract_article_content
    sample_html = "<html><body><h1>Judul Berita</h1><p>Ini adalah isi berita saham yang bersih.</p></body></html>"
    req2 = {
        "jsonrpc": "2.0",
        "id": 6,
        "method": "tools/call",
        "params": {
            "name": "osint_extract_article_content",
            "arguments": {"url_or_html": sample_html},
        },
    }
    resp2 = mcp_server.handle_request(req2)
    assert resp2["id"] == 6
    assert resp2["result"]["isError"] is False
    res2_data = json.loads(resp2["result"]["content"][0]["text"])
    assert "clean_text" in res2_data
    assert "evidence_context" in res2_data
    assert "<evidence_context>" in res2_data["evidence_context"]


def test_unified_mcp_quant_tool_call(mcp_server):
    req = {
        "jsonrpc": "2.0",
        "id": 7,
        "method": "tools/call",
        "params": {
            "name": "quant_compute_anomalies",
            "arguments": {
                "ticker": "ANTM",
                "volume_z_threshold": 1.5,
                "return_threshold_pct": 3.0,
            },
        },
    }
    resp = mcp_server.handle_request(req)
    assert resp["id"] == 7
    assert resp["result"]["isError"] is False
    anomalies = json.loads(resp["result"]["content"][0]["text"])
    assert isinstance(anomalies, list)


def test_unified_mcp_memory_tool_calls(mcp_server):
    # 1. Store observation
    store_req = {
        "jsonrpc": "2.0",
        "id": 8,
        "method": "tools/call",
        "params": {
            "name": "memory_store_observation",
            "arguments": {
                "source_label": "ANTM",
                "source_type": "TICKER",
                "relation": "OPERATES",
                "target_label": "Smelter Haltim",
                "target_type": "FACILITY",
                "context_snippet": "Uji coba fasilitas feronikel Halmahera Timur",
            },
        },
    }
    store_resp = mcp_server.handle_request(store_req)
    assert store_resp["id"] == 8
    assert store_resp["result"]["isError"] is False
    store_data = json.loads(store_resp["result"]["content"][0]["text"])
    assert store_data["status"] == "STORED"

    # 2. Recall context
    recall_req = {
        "jsonrpc": "2.0",
        "id": 9,
        "method": "tools/call",
        "params": {
            "name": "memory_recall_context",
            "arguments": {"query_entity": "ANTM"},
        },
    }
    recall_resp = mcp_server.handle_request(recall_req)
    assert recall_resp["id"] == 9
    assert recall_resp["result"]["isError"] is False
    recall_data = json.loads(recall_resp["result"]["content"][0]["text"])
    assert recall_data["nodes_found"] >= 1
    assert len(recall_data["edges"]) >= 1
    assert recall_data["edges"][0]["relation"] == "OPERATES"


def test_unified_mcp_resources(mcp_server):
    # List resources
    list_req = {"jsonrpc": "2.0", "id": 10, "method": "resources/list"}
    list_resp = mcp_server.handle_request(list_req)
    assert list_resp["id"] == 10
    resources = list_resp["result"]["resources"]
    assert len(resources) == 2
    uris = [r["uri"] for r in resources]
    assert "niskava://status" in uris
    assert "niskava://cache-stats" in uris

    # Read status resource
    read_req1 = {
        "jsonrpc": "2.0",
        "id": 11,
        "method": "resources/read",
        "params": {"uri": "niskava://status"},
    }
    read_resp1 = mcp_server.handle_request(read_req1)
    assert read_resp1["id"] == 11
    content1 = read_resp1["result"]["contents"][0]
    assert content1["uri"] == "niskava://status"
    status_data = json.loads(content1["text"])
    assert status_data["status"] == "HEALTHY"

    # Read cache-stats resource
    read_req2 = {
        "jsonrpc": "2.0",
        "id": 12,
        "method": "resources/read",
        "params": {"uri": "niskava://cache-stats"},
    }
    read_resp2 = mcp_server.handle_request(read_req2)
    assert read_resp2["id"] == 12
    content2 = read_resp2["result"]["contents"][0]
    assert content2["uri"] == "niskava://cache-stats"
    cache_data = json.loads(content2["text"])
    assert "sectors_cache_total" in cache_data


def test_unified_mcp_prompts(mcp_server):
    # List prompts
    list_req = {"jsonrpc": "2.0", "id": 13, "method": "prompts/list"}
    list_resp = mcp_server.handle_request(list_req)
    assert list_resp["id"] == 13
    prompts = list_resp["result"]["prompts"]
    assert len(prompts) == 2
    prompt_names = [p["name"] for p in prompts]
    assert "investigate_ticker_anomaly" in prompt_names
    assert "bandarmology_insider_audit" in prompt_names

    # Get prompt
    get_req = {
        "jsonrpc": "2.0",
        "id": 14,
        "method": "prompts/get",
        "params": {
            "name": "investigate_ticker_anomaly",
            "arguments": {"ticker": "ANTM", "days": "30"},
        },
    }
    get_resp = mcp_server.handle_request(get_req)
    assert get_resp["id"] == 14
    prompt_res = get_resp["result"]
    assert "messages" in prompt_res
    msg_text = prompt_res["messages"][0]["content"]["text"]
    assert "ANTM" in msg_text
    assert "sectors_get_daily_candles" in msg_text


def test_unified_mcp_errors(mcp_server):
    # Unknown method
    resp1 = mcp_server.handle_request({"jsonrpc": "2.0", "id": 90, "method": "invalid/method"})
    assert resp1["error"]["code"] == -32601

    # Unknown tool
    resp2 = mcp_server.handle_request({
        "jsonrpc": "2.0",
        "id": 91,
        "method": "tools/call",
        "params": {"name": "invalid_tool", "arguments": {}},
    })
    assert resp2["result"]["isError"] is True
    assert "Unknown MCP tool" in resp2["result"]["content"][0]["text"]

    # Unknown resource
    resp3 = mcp_server.handle_request({
        "jsonrpc": "2.0",
        "id": 92,
        "method": "resources/read",
        "params": {"uri": "niskava://unknown"},
    })
    assert resp3["error"]["code"] == -32602

    # Unknown prompt
    resp4 = mcp_server.handle_request({
        "jsonrpc": "2.0",
        "id": 93,
        "method": "prompts/get",
        "params": {"name": "unknown_prompt", "arguments": {}},
    })
    assert resp4["error"]["code"] == -32602


def test_ticker_sanitization(mcp_server):
    """Test that dirty tickers like 'ANTM.JK', 'IDX:BBRI', ' bbca ' are sanitized properly."""
    # 1. Ticker with .JK suffix
    req1 = {
        "jsonrpc": "2.0",
        "id": 201,
        "method": "tools/call",
        "params": {
            "name": "sectors_get_company_report",
            "arguments": {"ticker": "ANTM.JK"},
        },
    }
    resp1 = mcp_server.handle_request(req1)
    assert resp1["id"] == 201
    assert resp1["result"]["isError"] is False

    # 2. Ticker with IDX: prefix
    req2 = {
        "jsonrpc": "2.0",
        "id": 202,
        "method": "tools/call",
        "params": {
            "name": "quant_compute_anomalies",
            "arguments": {"ticker": "IDX:ANTM"},
        },
    }
    resp2 = mcp_server.handle_request(req2)
    assert resp2["id"] == 202
    assert resp2["result"]["isError"] is False

    # 3. Invalid ticker (e.g. numbers or too long)
    req3 = {
        "jsonrpc": "2.0",
        "id": 203,
        "method": "tools/call",
        "params": {
            "name": "sectors_get_daily_candles",
            "arguments": {"ticker": "INVALID_123"},
        },
    }
    resp3 = mcp_server.handle_request(req3)
    assert resp3["id"] == 203
    assert resp3["result"]["isError"] is True
    assert "Kode ticker IDX tidak valid" in resp3["result"]["content"][0]["text"]


def test_quant_insufficient_candles_guard(tmp_path):
    """Test that quant anomaly calculation handles < 20 candles gracefully."""
    db_file = str(tmp_path / "test_quant_guard.db")
    server = UnifiedMCPServer(db_path=db_file, mock_mode=True)

    # Monkeypatch client.get_daily_candles to return only 5 candles
    server.client.get_daily_candles = lambda ticker: [
        {"close": 1000 + i, "volume": 1000000, "date": f"2026-09-0{i+1}"}
        for i in range(5)
    ]

    req = {
        "jsonrpc": "2.0",
        "id": 301,
        "method": "tools/call",
        "params": {
            "name": "quant_compute_anomalies",
            "arguments": {"ticker": "ANTM"},
        },
    }
    resp = server.handle_request(req)
    assert resp["id"] == 301
    assert resp["result"]["isError"] is False
    data = json.loads(resp["result"]["content"][0]["text"])
    assert data["status"] == "INSUFFICIENT_DATA"
    assert "Minimal 20 hari bursa diperlukan" in data["message"]
    assert data["anomalies"] == []


def test_memory_auto_ddl_and_recency_decay(tmp_path):
    """Test that memory tool self-heals by creating tables in a clean DB and calculates recency decay."""
    # Brand new DB with no tables
    fresh_db = str(tmp_path / "fresh_unmigrated.db")
    server = UnifiedMCPServer(db_path=fresh_db, mock_mode=True)

    # 1. Store observation in fresh DB (should auto-create tables)
    store_req = {
        "jsonrpc": "2.0",
        "id": 401,
        "method": "tools/call",
        "params": {
            "name": "memory_store_observation",
            "arguments": {
                "source_label": "BBCA",
                "source_type": "TICKER",
                "relation": "PARTNERS_WITH",
                "target_label": "GOTO",
                "target_type": "TICKER",
                "context_snippet": "Kolaborasi pembayaran digital",
            },
        },
    }
    store_resp = server.handle_request(store_req)
    assert store_resp["id"] == 401
    assert store_resp["result"]["isError"] is False

    # 2. Recall context and verify effective_weight and decay_factor
    recall_req = {
        "jsonrpc": "2.0",
        "id": 402,
        "method": "tools/call",
        "params": {
            "name": "memory_recall_context",
            "arguments": {"query_entity": "BBCA"},
        },
    }
    recall_resp = server.handle_request(recall_req)
    assert recall_resp["id"] == 402
    assert recall_resp["result"]["isError"] is False
    recall_data = json.loads(recall_resp["result"]["content"][0]["text"])
    assert len(recall_data["edges"]) >= 1
    edge = recall_data["edges"][0]
    assert "effective_weight" in edge
    assert "decay_factor" in edge
    assert "days_ago" in edge
    assert edge["effective_weight"] <= edge["weight"]

