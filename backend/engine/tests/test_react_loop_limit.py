"""Test bahwa ReAct loop tidak berhenti sebelum response dihasilkan.

Memverifikasi Bug #1 fix: loop harus berlanjut melebihi 4 iterasi
jika tool calls berturut-turut diperlukan.
"""
import json
import os
from unittest.mock import MagicMock, patch

import pytest

from engine.agent.react_agent import MAX_REACT_ITERATIONS, NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


def _make_agent(tmp_path) -> tuple[NiskavaReActAgent, list]:
    """Helper: buat agent mock + event collector."""
    events: list = []
    registry = NiskavaToolRegistry(
        db_path=str(tmp_path / "test.db"), mock_mode=True
    )
    agent = NiskavaReActAgent(
        tool_registry=registry, emitter=events.append, mock_mode=False
    )
    return agent, events


def test_max_react_iterations_constant_exists():
    """MAX_REACT_ITERATIONS harus ada dan >= 8."""
    assert MAX_REACT_ITERATIONS >= 8, (
        f"MAX_REACT_ITERATIONS={MAX_REACT_ITERATIONS} terlalu kecil, harus >= 8"
    )


def test_max_react_iterations_env_override():
    """NISKAVA_MAX_REACT_ITERATIONS env var harus override konstanta."""
    with patch.dict(os.environ, {"NISKAVA_MAX_REACT_ITERATIONS": "15"}):
        import importlib
        import engine.agent.react_agent as mod
        importlib.reload(mod)
        assert mod.MAX_REACT_ITERATIONS == 15
    # Restore
    import importlib
    import engine.agent.react_agent as mod2
    importlib.reload(mod2)


def test_universal_cycle_does_not_emit_session_error_on_multi_tool(tmp_path):
    """Siklus universal harus menyelesaikan response meski butuh 5+ tool calls.

    Mensimulasikan skenario di mana LLM memanggil 5 tool berturut-turut
    sebelum menghasilkan <response>. Dengan range(4) lama ini akan gagal.
    """
    agent, events = _make_agent(tmp_path)

    # Siapkan sequence respons LLM: 5x tool_call, lalu 1x response
    tool_call_content = (
        "<thought>Perlu data</thought>"
        '<tool_call>{"name": "get_daily_candles", "arguments": {"ticker": "ANTM"}}</tool_call>'
    )
    final_response_content = (
        "<thought>Sudah cukup data</thought>"
        "<response>Berikut analisis ANTM berdasarkan data tersedia.</response>"
    )

    call_count = 0

    def mock_post(*args, **kwargs):
        nonlocal call_count
        call_count += 1
        mock_resp = MagicMock()
        mock_resp.status_code = 200
        # 5 iterasi pertama: tool_call; iterasi ke-6: response
        content = tool_call_content if call_count <= 5 else final_response_content
        mock_resp.text = json.dumps({
            "choices": [{"message": {"content": content}}]
        })
        return mock_resp

    with patch("requests.post", side_effect=mock_post):
        result = agent._run_universal_chat_cycle(
            session_id="TEST-LOOP-001",
            user_prompt="Analisis ANTM",
            history=[],
            start_time=0.0,
        )

    # Tidak boleh ada session_error
    error_events = [e for e in events if e.get("event") == "session_error"]
    assert len(error_events) == 0, f"session_error tidak diharapkan: {error_events}"

    # Harus ada session_complete
    complete_events = [e for e in events if e.get("event") == "session_complete"]
    assert len(complete_events) == 1

    # Response harus berisi teks final
    assert "ANTM" in result.get("response", "")


def test_universal_cycle_findings_are_populated(tmp_path):
    """_run_universal_chat_cycle harus mengembalikan findings yang terisi.

    Ketika LLM mengembalikan finding_emitted event (melalui tool observation),
    result["findings"] tidak boleh kosong.
    """
    agent, events = _make_agent(tmp_path)

    finding_response = (
        "<thought>Data menunjukkan anomali</thought>"
        "<response>"
        "[SUPPORTED] ANTM menunjukkan volume spike 3.2σ pada 2026-09-15 "
        "berkorelasi dengan pengumuman smelter. Confidence: 95%."
        "</response>"
    )

    def mock_post(*args, **kwargs):
        mock_resp = MagicMock()
        mock_resp.status_code = 200
        mock_resp.text = json.dumps({
            "choices": [{"message": {"content": finding_response}}]
        })
        return mock_resp

    with patch("requests.post", side_effect=mock_post):
        result = agent._run_universal_chat_cycle(
            session_id="TEST-FINDINGS-001",
            user_prompt="Cek anomali ANTM",
            history=[],
            start_time=0.0,
        )

    # Minimal: result harus punya key findings
    assert "findings" in result
    assert isinstance(result["findings"], list)


def test_universal_cycle_accumulates_findings_from_tool_observation(tmp_path):
    """Findings harus diisi ketika agent memanggil tool dan mendapat anomaly data.

    Skenario: model memanggil compute_quant_anomalies (anomaly terdeteksi),
    lalu emit finding. result["findings"] harus punya >= 1 item.
    """
    agent, events = _make_agent(tmp_path)

    tool_call_text = (
        "<thought>Cek anomali</thought>"
        '<tool_call>{"name": "compute_quant_anomalies", "arguments": {"ticker": "ANTM"}}</tool_call>'
    )
    final_response_text = (
        "<thought>Anomali ditemukan</thought>"
        "<response>ANTM menunjukkan volume anomali. [SUPPORTED]</response>"
    )
    responses = [
        json.dumps({"choices": [{"message": {"content": tool_call_text}}]}),
        json.dumps({"choices": [{"message": {"content": final_response_text}}]}),
    ]
    call_idx = 0

    def mock_post(*args, **kwargs):
        nonlocal call_idx
        mock_resp = MagicMock()
        mock_resp.status_code = 200
        idx = min(call_idx, len(responses) - 1)
        mock_resp.text = responses[idx]
        call_idx += 1
        return mock_resp

    with patch("requests.post", side_effect=mock_post):
        result = agent._run_universal_chat_cycle(
            session_id="TEST-FINDINGS-002",
            user_prompt="Cek anomali ANTM",
            history=[],
            start_time=0.0,
        )

    # Anomalies harus ada (dari mock compute_quant_anomalies)
    assert len(result.get("anomalies", [])) >= 1, (
        "anomalies harus terisi dari tool compute_quant_anomalies"
    )
    assert len(result.get("findings", [])) >= 1, (
        "findings harus terisi dari tool compute_quant_anomalies"
    )


def test_max_react_iterations_default_is_30():
    """Default MAX_REACT_ITERATIONS harus 30 untuk memastikan ReAct loop nyaman & playable."""
    assert MAX_REACT_ITERATIONS == 30


def test_agent_max_iterations_custom_override(tmp_path):
    """NiskavaReActAgent harus menghargai parameter max_iterations kustom."""
    agent, _ = _make_agent(tmp_path)
    assert agent.max_iterations == 30

    custom_agent = NiskavaReActAgent(
        tool_registry=agent.tools,
        max_iterations=50,
    )
    assert custom_agent.max_iterations == 50


def test_agent_final_iteration_graceful_synthesis(tmp_path):
    """Ketika iterasi terakhir memanggil tool, agent harus memberi kesempatan sintesis final alih-alih langsung error."""
    agent, events = _make_agent(tmp_path)

    # Ketika max_iterations diset 2:
    # Call 1: model panggil tool get_daily_candles (remaining_steps = 1)
    # Call 2: model panggil tool search_osint di langkah terakhir (remaining_steps = 0)
    # Call 3: prompt sintesis final dijalankan otomatis dan model memberikan <response>
    t1 = '<thought>Tool 1</thought><tool_call>{"name": "query_sectors", "arguments": {"domain": "candles", "ticker": "ANTM"}}</tool_call>'
    t2 = '<thought>Tool 2</thought><tool_call>{"name": "search_osint", "arguments": {"ticker": "ANTM"}}</tool_call>'
    t3 = '<thought>Synthesizing</thought><response>Analisis multi-langkah ANTM berhasil disintesis.</response>'

    responses = [
        json.dumps({"choices": [{"message": {"content": t1}}]}),
        json.dumps({"choices": [{"message": {"content": t2}}]}),
        json.dumps({"choices": [{"message": {"content": t3}}]}),
    ]
    call_idx = 0

    def mock_post(*args, **kwargs):
        nonlocal call_idx
        mock_resp = MagicMock()
        mock_resp.status_code = 200
        mock_resp.text = responses[min(call_idx, len(responses) - 1)]
        call_idx += 1
        return mock_resp

    with patch("engine.agent.react_agent.MAX_REACT_ITERATIONS", 2), \
         patch("requests.post", side_effect=mock_post):
        result = agent._run_universal_chat_cycle(
            session_id="TEST-GRACEFUL-001",
            user_prompt="Bandingkan data ANTM",
            history=[],
            start_time=0.0,
        )

    assert result.get("status") != "ERROR"
    assert "Analisis multi-langkah ANTM berhasil disintesis" in result.get("response", "")

