"""Unit tests for Niskava Tool Registry."""

from engine.agent.tools import NiskavaToolRegistry


def test_tool_registry_definitions(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    defs = registry.get_tool_definitions()

    tool_names = [d["name"] for d in defs]
    assert "get_daily_candles" in tool_names
    assert "compute_quant_anomalies" in tool_names
    assert "get_company_fundamentals" in tool_names
    assert "get_foreign_flow" in tool_names
    assert "harvest_market_news" in tool_names


def test_tool_execution(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)

    candles = registry.execute_tool("get_daily_candles", {"ticker": "ANTM", "days": 30})
    assert len(candles) == 30

    anomalies = registry.execute_tool("compute_quant_anomalies", {"ticker": "ANTM"})
    assert len(anomalies) >= 1
    assert anomalies[0]["z_score"] > 2.5
