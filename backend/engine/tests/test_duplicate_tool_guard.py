"""Test bahwa ReAct loop mendeteksi dan memutus tool call duplikat berturut-turut.

Memverifikasi fix: ketika model memanggil tool yang sama dengan argumen identik
>= 2 kali berturut-turut, sistem harus menyuntikkan nudge paksa agar model
berhenti looping dan langsung menghasilkan <response>.
"""
import json
from unittest.mock import MagicMock, patch

import pytest

from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


def _make_agent(tmp_path) -> tuple[NiskavaReActAgent, list]:
    """Helper: buat agent mock + event collector."""
    events: list = []
    registry = NiskavaToolRegistry(
        db_path=str(tmp_path / "dup_guard_test.db"), mock_mode=True
    )
    agent = NiskavaReActAgent(
        tool_registry=registry, emitter=events.append, mock_mode=False
    )
    return agent, events


def test_duplicate_tool_guard_breaks_loop(tmp_path):
    """Ketika model memanggil tool identik 3x, guard harus menyuntik nudge
    sehingga pada iterasi berikutnya model menghasilkan <response>."""
    agent, events = _make_agent(tmp_path)

    # Model memanggil query_sectors(domain="", ticker="IDX") berulang kali,
    # lalu setelah mendapat nudge dari guard, akhirnya menghasilkan response.
    duplicate_call = (
        '<thought>Need market data</thought>'
        '<tool_call>{"name": "query_sectors", "arguments": {"ticker": "IDX"}}</tool_call>'
    )
    final_response = (
        '<thought>Got data from observations</thought>'
        '<response>Berikut kondisi pasar hari ini berdasarkan data terkini.</response>'
    )

    call_count = 0

    def mock_post(*args, **kwargs):
        nonlocal call_count
        call_count += 1
        mock_resp = MagicMock()
        mock_resp.status_code = 200
        # First 3 calls: identical tool call (would loop forever without guard)
        # After guard nudge: model produces final response
        content = duplicate_call if call_count <= 3 else final_response
        mock_resp.text = json.dumps({
            "choices": [{"message": {"content": content}}]
        })
        return mock_resp

    with patch("requests.post", side_effect=mock_post):
        result = agent._run_universal_chat_cycle(
            session_id="TEST-DUP-GUARD-001",
            user_prompt="analisa market hari ini",
            history=[],
            start_time=0.0,
        )

    # Must NOT be an error — guard should have broken the loop
    assert result.get("status") != "ERROR", (
        f"Expected successful response but got ERROR: {result.get('error')}"
    )
    assert "pasar" in result.get("response", "").lower() or "market" in result.get("response", "").lower()

    # Guard nudge should have been emitted as a thought event
    thought_events = [e for e in events if e.get("event") == "agent_thought"]
    nudge_thoughts = [
        t for t in thought_events
        if "duplikat" in t.get("thought", "").lower() or "duplicate" in t.get("thought", "").lower()
    ]
    assert len(nudge_thoughts) >= 1, (
        "Guard should emit a thought indicating duplicate tool call detected"
    )


def test_non_duplicate_calls_are_not_blocked(tmp_path):
    """Tool calls yang berbeda-beda tidak boleh diblokir oleh guard."""
    agent, events = _make_agent(tmp_path)

    call_count = 0

    def mock_post(*args, **kwargs):
        nonlocal call_count
        call_count += 1
        mock_resp = MagicMock()
        mock_resp.status_code = 200

        if call_count == 1:
            content = (
                '<thought>Get candles</thought>'
                '<tool_call>{"name": "query_sectors", "arguments": {"domain": "candles", "ticker": "BBCA"}}</tool_call>'
            )
        elif call_count == 2:
            content = (
                '<thought>Get news</thought>'
                '<tool_call>{"name": "search_news", "arguments": {"ticker": "BBCA"}}</tool_call>'
            )
        else:
            content = (
                '<thought>All data collected</thought>'
                '<response>BBCA menunjukkan pergerakan stabil hari ini.</response>'
            )

        mock_resp.text = json.dumps({
            "choices": [{"message": {"content": content}}]
        })
        return mock_resp

    with patch("requests.post", side_effect=mock_post):
        result = agent._run_universal_chat_cycle(
            session_id="TEST-DUP-GUARD-002",
            user_prompt="analisis BBCA",
            history=[],
            start_time=0.0,
        )

    assert result.get("status") != "ERROR"
    assert "BBCA" in result.get("response", "")

    # No duplicate guard nudge should have been emitted
    thought_events = [e for e in events if e.get("event") == "agent_thought"]
    nudge_thoughts = [
        t for t in thought_events
        if "duplikat" in t.get("thought", "").lower() or "duplicate" in t.get("thought", "").lower()
    ]
    assert len(nudge_thoughts) == 0, (
        "Guard should NOT fire for non-duplicate tool calls"
    )
