"""Unit tests for Niskava Tool Registry & MCP Tool Dispatching."""

from engine.agent.tools import NiskavaToolRegistry


def test_tool_registry_definitions(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    defs = registry.get_tool_definitions()

    tool_names = [d["name"] for d in defs]
    assert "get_daily_candles" in tool_names
    assert "compute_quant_anomalies" in tool_names
    assert "get_company_fundamentals" in tool_names
    assert "get_foreign_flow" in tool_names
    assert "get_suspensions" in tool_names
    assert "get_corporate_actions" in tool_names
    assert "get_filings" in tool_names
    assert "get_broker_summary" in tool_names
    assert "get_subsector_peers" in tool_names
    assert "get_mining_detail" in tool_names
    assert "harvest_market_news" in tool_names


def test_tool_execution_direct_and_mcp_aliases(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)

    # Test direct names
    candles = registry.execute_tool("get_daily_candles", {"ticker": "ANTM", "days": 30})
    assert len(candles) == 30

    anomalies = registry.execute_tool("compute_quant_anomalies", {"ticker": "ANTM"})
    assert len(anomalies) >= 1
    assert anomalies[0]["z_score"] > 2.5

    # Test MCP aliases
    candles_mcp = registry.execute_tool("sectors_get_daily_candles", {"ticker": "ANTM", "days": 30})
    assert len(candles_mcp) == 30

    report = registry.execute_tool("sectors_get_company_report", {"ticker": "ANTM"})
    assert report["symbol"] == "ANTM"

    flow = registry.execute_tool("sectors_get_foreign_flow", {"ticker": "ANTM"})
    assert len(flow) >= 1

    suspensions = registry.execute_tool("sectors_get_suspensions", {"ticker": "ANTM"})
    assert len(suspensions) >= 1

    corp_actions = registry.execute_tool("sectors_get_corporate_actions", {"ticker": "ANTM"})
    assert len(corp_actions) >= 1

    filings = registry.execute_tool("sectors_get_filings", {"ticker": "ANTM"})
    assert len(filings) >= 1

    brokers = registry.execute_tool("sectors_get_broker_summary", {"ticker": "ANTM"})
    assert "top_buyers" in brokers

    peers = registry.execute_tool("sectors_get_subsector_peers", {"subsector": "metals-and-minerals-mining"})
    assert "peer_count" in peers

    mining = registry.execute_tool("sectors_get_mining_detail", {"slug": "aneka-tambang"})
    assert mining["commodity"] == "NICKEL"
