"""Unit tests for Niskava ReAct Agent."""

from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


def test_react_agent_investigation_cycle(tmp_path):
    events = []

    def record_event(ev):
        events.append(ev)

    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    agent = NiskavaReActAgent(
        tool_registry=registry,
        emitter=record_event,
        mock_mode=True,
    )

    result = agent.investigate(ticker="ANTM", days=30)

    assert result["ticker"] == "ANTM"
    assert len(result["anomalies"]) >= 1
    assert len(result["findings"]) >= 1

    event_types = [e["event"] for e in events]
    assert "session_start" in event_types
    assert "agent_thought" in event_types
    assert "agent_tool_call" in event_types
    assert "agent_observation" in event_types
    assert "finding_emitted" in event_types
    assert "session_complete" in event_types


def test_system_prompt_includes_comprehensive_tool_catalog():
    from engine.agent.react_agent import get_system_prompt
    tools_summary = [
        {"name": "get_broker_summary", "description": "Fetch top broker accumulation"},
        {"name": "get_filings", "description": "Fetch insider filings"},
        {"name": "skill_market_anomaly_recon", "description": "Recon quantitative anomalies"},
    ]
    prompt = get_system_prompt("id", available_tools=tools_summary)
    assert "get_broker_summary" in prompt
    assert "get_filings" in prompt
    assert "skill_market_anomaly_recon" in prompt

