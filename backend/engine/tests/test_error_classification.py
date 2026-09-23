"""Test bahwa pesan error ReAct loop diklasifikasikan dengan akurat.

Memverifikasi fix: ketika loop exhaustion terjadi (bukan masalah koneksi),
pesan error harus menunjukkan 'ReAct Analysis Limit Reached' bukan
'Unable to Connect to AI Provider'.
"""
import json
from unittest.mock import MagicMock, patch

import pytest

from engine.agent.react_agent import MAX_REACT_ITERATIONS, NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


def _make_agent(tmp_path) -> tuple[NiskavaReActAgent, list]:
    events: list = []
    registry = NiskavaToolRegistry(
        db_path=str(tmp_path / "errclass_test.db"), mock_mode=True
    )
    agent = NiskavaReActAgent(
        tool_registry=registry, emitter=events.append, mock_mode=False
    )
    return agent, events


def test_loop_exhaustion_shows_analysis_limit_not_connection_error(tmp_path):
    """Ketika ReAct loop habis karena model tidak menghasilkan <response>,
    pesan error harus berisi 'Analysis Limit' bukan 'Unable to Connect'."""
    agent, events = _make_agent(tmp_path)

    # Model selalu mengembalikan tool call tanpa pernah memberi <response>
    eternal_tool_call = (
        '<thought>Need more data</thought>'
        '<tool_call>{"name": "query_sectors", "arguments": {"domain": "candles", "ticker": "ANTM"}}</tool_call>'
    )

    def mock_post(*args, **kwargs):
        mock_resp = MagicMock()
        mock_resp.status_code = 200
        mock_resp.text = json.dumps({
            "choices": [{"message": {"content": eternal_tool_call}}]
        })
        return mock_resp

    with patch("requests.post", side_effect=mock_post):
        result = agent._run_universal_chat_cycle(
            session_id="TEST-ERRCLASS-001",
            user_prompt="Analisis ANTM",
            history=[],
            start_time=0.0,
        )

    assert result.get("status") == "ERROR"
    response_text = result.get("response", "")

    # Must NOT contain the misleading connection error title
    assert "Unable to Connect to AI Provider" not in response_text, (
        f"Loop exhaustion should NOT show 'Unable to Connect'. Got: {response_text[:300]}"
    )

    # Must contain the accurate analysis limit message
    assert (
        "Analysis Limit Reached" in response_text
        or "Batas Analisis Tercapai" in response_text
        or "Batas Penalaran ReAct Tercapai" in response_text
    ), f"Loop exhaustion should show 'Analysis Limit Reached'. Got: {response_text[:300]}"


def test_real_connection_error_shows_unable_to_connect(tmp_path):
    """Ketika koneksi benar-benar gagal (ConnectionError), pesan harus
    menunjukkan 'Unable to Connect to AI Provider'."""
    agent, events = _make_agent(tmp_path)

    import requests as req_mod

    def mock_post(*args, **kwargs):
        raise req_mod.exceptions.ConnectionError("Connection refused to localhost:20128")

    with patch("requests.post", side_effect=mock_post):
        result = agent._run_universal_chat_cycle(
            session_id="TEST-ERRCLASS-002",
            user_prompt="Analisis BBCA",
            history=[],
            start_time=0.0,
        )

    assert result.get("status") == "ERROR"
    response_text = result.get("response", "")

    # Real connection error MUST show the connection error message
    assert "Unable to Connect to AI Provider" in response_text or "Gagal Terhubung" in response_text, (
        f"Real connection error should show 'Unable to Connect'. Got: {response_text[:300]}"
    )
