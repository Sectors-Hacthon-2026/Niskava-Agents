"""Tests verifying ReAct loop internal nudges and notices are English."""

import inspect
from engine.agent.react_agent import NiskavaReActAgent


def test_react_system_notice_is_english(tmp_path):
    from engine.agent.tools import NiskavaToolRegistry
    registry = NiskavaToolRegistry(str(tmp_path / "test.db"), mock_mode=True)
    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)

    src = inspect.getsource(agent._run_universal_chat_cycle)
    assert "Hanya tersisa" not in src, "Found Indonesian notice 'Hanya tersisa' in chat cycle"
    assert "langkah penalaran" not in src, "Found Indonesian notice 'langkah penalaran' in chat cycle"
    assert "JANGAN panggil tool lagi" not in src, "Found Indonesian notice in chat cycle"
    assert "Anomali Volume" not in src, "Found Indonesian auto-finding title in chat cycle"
