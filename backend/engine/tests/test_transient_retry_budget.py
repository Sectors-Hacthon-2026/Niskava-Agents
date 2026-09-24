"""Test bahwa upstream transient retry (503/429) tidak menghabiskan iterasi ReAct.

Memverifikasi fix: ketika upstream provider mengembalikan error sementara
(e.g. 503 Overloaded, 429 Rate Limit), retry harus dilakukan dalam inner loop
tanpa mengurangi kuota MAX_REACT_ITERATIONS.
"""
import json
from unittest.mock import MagicMock, patch

import pytest

from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


def _make_agent(tmp_path) -> tuple[NiskavaReActAgent, list]:
    events: list = []
    registry = NiskavaToolRegistry(
        db_path=str(tmp_path / "retry_test.db"), mock_mode=True
    )
    agent = NiskavaReActAgent(
        tool_registry=registry, emitter=events.append, mock_mode=False
    )
    return agent, events


def test_transient_503_does_not_burn_react_iterations(tmp_path):
    """Upstream 503 harus di-retry tanpa mengurangi kuota iterasi ReAct.

    Skenario: 2x upstream 503 (transient), lalu sukses dengan tool call,
    lalu sukses dengan response. Total harus menggunakan hanya 2 ReAct iterasi.
    """
    agent, events = _make_agent(tmp_path)

    call_count = 0

    def mock_post(*args, **kwargs):
        nonlocal call_count
        call_count += 1
        mock_resp = MagicMock()

        if call_count <= 2:
            # First 2 calls: transient 503 error in JSON body
            mock_resp.status_code = 200
            mock_resp.text = json.dumps({
                "error": {
                    "message": "Upstream error from Nvidia: Service temporarily overloaded",
                    "code": 503,
                }
            })
            return mock_resp

        if call_count == 3:
            # Third call: success with tool call
            mock_resp.status_code = 200
            mock_resp.text = json.dumps({
                "choices": [{"message": {"content":
                    '<thought>Get data</thought>'
                    '<tool_call>{"name": "search_news", "arguments": {"ticker": ""}}</tool_call>'
                }}]
            })
            return mock_resp

        # Fourth call: success with response
        mock_resp.status_code = 200
        mock_resp.text = json.dumps({
            "choices": [{"message": {"content":
                '<thought>Data collected</thought>'
                '<response>Pasar hari ini menunjukkan pergerakan sideways.</response>'
            }}]
        })
        return mock_resp

    with patch("requests.post", side_effect=mock_post), \
         patch("time.sleep"):  # Skip actual sleep for speed
        result = agent._run_universal_chat_cycle(
            session_id="TEST-RETRY-001",
            user_prompt="kondisi market hari ini",
            history=[],
            start_time=0.0,
        )

    # Must succeed — transient errors should NOT cause ERROR status
    assert result.get("status") != "ERROR", (
        f"Transient retries should not exhaust ReAct budget. Got error: {result.get('error')}"
    )
    assert "pasar" in result.get("response", "").lower()

    # Verify retry thought events were emitted
    thought_events = [e for e in events if e.get("event") == "agent_thought"]
    retry_thoughts = [
        t for t in thought_events
        if "busy" in t.get("thought", "").lower()
        or "upstream" in t.get("thought", "").lower()
    ]
    assert len(retry_thoughts) >= 1, "Should emit thought about upstream retry"


def test_transient_retry_within_tight_iteration_budget(tmp_path):
    """Ketika MAX_REACT_ITERATIONS=2, 2x transient 503 di langkah pertama
    tidak boleh menghabiskan kuota 2 iterasi tersebut."""
    agent, events = _make_agent(tmp_path)

    call_count = 0

    def mock_post(*args, **kwargs):
        nonlocal call_count
        call_count += 1
        mock_resp = MagicMock()

        if call_count <= 2:
            # 2 transient 503s
            mock_resp.status_code = 200
            mock_resp.text = json.dumps({
                "error": {
                    "message": "Upstream error from Nvidia: Service temporarily overloaded",
                    "code": 503,
                }
            })
            return mock_resp

        if call_count == 3:
            # Iteration 1 succeeds with tool call
            mock_resp.status_code = 200
            mock_resp.text = json.dumps({
                "choices": [{"message": {"content":
                    '<thought>Need ANTM candles</thought>'
                    '<tool_call>{"name": "query_sectors", "arguments": {"domain": "candles", "ticker": "ANTM"}}</tool_call>'
                }}]
            })
            return mock_resp

        # Iteration 2 succeeds with final response
        mock_resp.status_code = 200
        mock_resp.text = json.dumps({
            "choices": [{"message": {"content":
                '<thought>Done</thought>'
                '<response>Analisis ANTM lengkap.</response>'
            }}]
        })
        return mock_resp

    with patch("engine.agent.react_agent.MAX_REACT_ITERATIONS", 2), \
         patch("requests.post", side_effect=mock_post), \
         patch("time.sleep"):
        result = agent._run_universal_chat_cycle(
            session_id="TEST-RETRY-002",
            user_prompt="analisis ANTM",
            history=[],
            start_time=0.0,
        )

    assert result.get("status") != "ERROR", (
        f"Expected success within 2 ReAct iterations, got ERROR: {result.get('error')}"
    )
    assert "ANTM" in result.get("response", "")

