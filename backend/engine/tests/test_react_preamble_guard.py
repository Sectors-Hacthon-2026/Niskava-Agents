"""Tests for ReAct prompt compliance, preamble handling, and native tool_calls extraction."""

import json
import time
from unittest.mock import MagicMock, patch
import pytest

from engine.agent.react_agent import NiskavaReActAgent, get_system_prompt
from engine.agent.tools import NiskavaToolRegistry


def test_system_prompt_is_pure_english():
    prompt_id = get_system_prompt("id")
    # All structural headers and instructions must be English
    assert "=== TARGET USER LANGUAGE ===" in prompt_id
    assert "=== GOLDEN OPERATIONAL RULES ===" in prompt_id
    assert "=== REACTION PROTOCOL & 1-SHOT DEMONSTRATION ===" in prompt_id
    assert "ZERO PREAMBLE TO USER" in prompt_id
    assert "THOUGHT ISOLATION" in prompt_id
    # Ensure no awkward Indonesian code-switching inside English prompt template
    assert "cek berita hari ini" not in prompt_id

def test_native_tool_calls_with_preamble_extracted(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    events = []
    agent = NiskavaReActAgent(
        tool_registry=registry,
        emitter=lambda ev: events.append(ev),
        mock_mode=True,
    )

    mock_choice_turn1 = {
        "message": {
            "role": "assistant",
            "content": "Mari kita lakukan analisis kuantitatif terhadap GOTO...",
            "tool_calls": [
                {
                    "id": "call_123",
                    "type": "function",
                    "function": {
                        "name": "compute_quant_anomalies",
                        "arguments": '{"ticker": "GOTO"}'
                    }
                }
            ]
        }
    }
    mock_choice_turn2 = {
        "message": {
            "role": "assistant",
            "content": "<thought>Data diterima</thought>\n<response>Analisis GOTO menunjukkan pergerakan normal.</response>"
        }
    }

    mock_resp1 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_choice_turn1]}))
    mock_resp2 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_choice_turn2]}))

    with patch("requests.post", side_effect=[mock_resp1, mock_resp2]):
        res = agent._run_universal_chat_cycle(
            session_id="TEST-SESSION",
            user_prompt="analisis GOTO coba",
            history=[],
            start_time=time.time(),
        )

    assert "Analisis GOTO" in res["response"]
    tool_events = [e for e in events if e.get("event") == "agent_tool_call"]
    assert len(tool_events) >= 1
    assert tool_events[0]["tool"] == "compute_quant_anomalies"

def test_preamble_without_tools_triggers_self_correction(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    events = []
    agent = NiskavaReActAgent(
        tool_registry=registry,
        emitter=lambda ev: events.append(ev),
        mock_mode=True,
    )

    mock_turn1 = {
        "message": {
            "role": "assistant",
            "content": "Mari kita lakukan analisis kuantitatif terhadap GOTO.\n### 🔍 Langkah 1: Ambil Data:",
        }
    }
    mock_turn2 = {
        "message": {
            "role": "assistant",
            "content": '<tool_call>{"name": "compute_quant_anomalies", "arguments": {"ticker": "GOTO"}}</tool_call>',
        }
    }
    mock_turn3 = {
        "message": {
            "role": "assistant",
            "content": '<response>Analisis GOTO selesai: tidak ada anomali.</response>',
        }
    }

    mock_resp1 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn1]}))
    mock_resp2 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn2]}))
    mock_resp3 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn3]}))

    with patch("requests.post", side_effect=[mock_resp1, mock_resp2, mock_resp3]):
        res = agent._run_universal_chat_cycle(
            session_id="TEST-NUDGE",
            user_prompt="analisis GOTO coba",
            history=[],
            start_time=time.time(),
        )

    assert "Analisis GOTO selesai" in res["response"]


def test_conversational_query_direct_response(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    events = []
    agent = NiskavaReActAgent(
        tool_registry=registry,
        emitter=lambda ev: events.append(ev),
        mock_mode=True,
    )

    mock_turn = {
        "message": {
            "role": "assistant",
            "content": "Price to Earnings Ratio (PER) adalah rasio valuasi pasar terhadap laba per saham.",
        }
    }
    mock_resp = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn]}))

    with patch("requests.post", return_value=mock_resp):
        res = agent._run_universal_chat_cycle(
            session_id="TEST-CONV",
            user_prompt="Apa itu PER?",
            history=[],
            start_time=time.time(),
        )

    assert "Price to Earnings Ratio" in res["response"]


def test_multiple_parallel_tool_calls_executed(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    events = []
    agent = NiskavaReActAgent(
        tool_registry=registry,
        emitter=lambda ev: events.append(ev),
        mock_mode=True,
    )

    # Turn 1: Model emits two tool calls in the same turn
    mock_turn1 = {
        "message": {
            "role": "assistant",
            "content": (
                '<thought>Need candles and quant anomalies.</thought>\n'
                '<tool_call>{"name": "get_daily_candles", "arguments": {"ticker": "BUMI", "days": 30}}</tool_call>\n'
                '<tool_call>{"name": "compute_quant_anomalies", "arguments": {"ticker": "BUMI"}}</tool_call>'
            ),
        }
    }
    mock_turn2 = {
        "message": {
            "role": "assistant",
            "content": '<response>Analisis BUMI: Data candles dan anomalies lengkap diterima.</response>',
        }
    }

    mock_resp1 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn1]}))
    mock_resp2 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn2]}))

    with patch("requests.post", side_effect=[mock_resp1, mock_resp2]):
        res = agent._run_universal_chat_cycle(
            session_id="TEST-PARALLEL",
            user_prompt="analisis BUMI",
            history=[],
            start_time=time.time(),
        )

    assert "Analisis BUMI" in res["response"]
    tool_events = [e for e in events if e.get("event") == "agent_tool_call"]
    tools_called = [e["tool"] for e in tool_events]
    assert "get_daily_candles" in tools_called
    assert "compute_quant_anomalies" in tools_called
    assert len(tool_events) == 2


def test_waiting_statement_triggers_nudge(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    events = []
    agent = NiskavaReActAgent(
        tool_registry=registry,
        emitter=lambda ev: events.append(ev),
        mock_mode=True,
    )

    # Turn 1: Model says "Let me wait for those results to come in..."
    mock_turn1 = {
        "message": {
            "role": "assistant",
            "content": "I see that the daily candles data was returned. Let me wait for the other tool results to come in before providing a comprehensive analysis of BUMI stock.",
        }
    }
    # Turn 2: After nudge, model synthesizes final response
    mock_turn2 = {
        "message": {
            "role": "assistant",
            "content": "<response>Here is the complete analysis of BUMI stock based on available data.</response>",
        }
    }

    mock_resp1 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn1]}))
    mock_resp2 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn2]}))

    with patch("requests.post", side_effect=[mock_resp1, mock_resp2]):
        res = agent._run_universal_chat_cycle(
            session_id="TEST-WAITING-NUDGE",
            user_prompt="mana udah belum?",
            history=[],
            start_time=time.time(),
        )

    assert "complete analysis of BUMI" in res["response"]
