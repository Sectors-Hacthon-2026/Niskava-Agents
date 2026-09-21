import pytest
from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


def test_react_agent_recovers_from_tool_execution_failure(tmp_path):
    emitted_events = []

    def mock_emitter(ev):
        emitted_events.append(ev)

    registry = NiskavaToolRegistry(db_path=str(tmp_path / "resilient.db"), mock_mode=True)
    agent = NiskavaReActAgent(tool_registry=registry, emitter=mock_emitter, mock_mode=True)

    # Test direct execution with unhandled/fictitious ticker does not crash ReAct loop
    res = agent.chat(
        user_prompt="Cek data saham emiten tidak terdaftar BOGUS123",
        session_id="TEST-RESILIENT-001",
    )

    assert res is not None
    assert len(res.get("response", "")) > 0

    session_complete = [e for e in emitted_events if e.get("event") == "session_complete"]
    assert len(session_complete) == 1
    assert session_complete[0].get("status") == "COMPLETED"

    # Ensure agent did not emit fatal session_error
    error_events = [e for e in emitted_events if e.get("event") == "session_error"]
    assert len(error_events) == 0, f"Expected 0 session_error events, got: {error_events}"


def test_react_agent_tool_dispatch_exception_does_not_cut_off(tmp_path):
    emitted_events = []

    def mock_emitter(ev):
        emitted_events.append(ev)

    registry = NiskavaToolRegistry(db_path=str(tmp_path / "resilient2.db"), mock_mode=True)
    agent = NiskavaReActAgent(tool_registry=registry, emitter=mock_emitter, mock_mode=True)

    # Calling an invalid tool should produce an error observation, not crash
    try:
        registry.execute_tool("non_existent_tool_xyz", {})
    except ValueError as exc:
        assert "tidak terdaftar" in str(exc)
