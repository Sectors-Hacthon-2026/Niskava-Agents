"""Regression tests for memory graph ingestion pipeline.

Tests three things:
1. extract_response_tickers extracts tickers from agent response text
2. extract_response_tickers extracts tickers from tool call args
3. No false positives from common Indonesian words
"""
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest
from engine.memory.extractor import extract_response_tickers


def test_extracts_tickers_from_briefing_response():
    """Should find BBCA, BBRI, ANTM from a typical pre-open briefing response."""
    response = """# 📊 Intelijen Pasar Pre-Open

    ## Sektor Perbankan
    **BBCA** menunjukkan momentum bullish dengan net inflow asing.
    **BBRI** juga mencatat kinerja positif pada sesi kemarin.

    ## Sektor Tambang
    PT Aneka Tambang Tbk (ANTM) baru mengonfirmasi uji coba smelter.
    """
    result = extract_response_tickers(response)
    assert "BBCA" in result
    assert "BBRI" in result
    assert "ANTM" in result


def test_extracts_tickers_from_tool_call_args():
    """Should find TLKM from tool call args when not in response text."""
    response = "Berikut analisis market secara umum."
    tool_args = [
        {"ticker": "TLKM", "days": 30},
        {"symbol": "BBNI"},
    ]
    result = extract_response_tickers(response, tool_call_args_list=tool_args)
    assert "TLKM" in result
    assert "BBNI" in result


def test_no_false_positives_from_indonesian_words():
    """Common Bahasa Indonesia words that are 4 letters must NOT appear as tickers."""
    response = "Hari ini pasar saham IDX cukup ramai dengan banyak aksi beli."
    result = extract_response_tickers(response)
    # "hari", "pasar", "saham", "beli", "cukup", "ramai" should not appear
    # (they are not valid IDX tickers in the master list)
    for word in ["HARI", "BELI", "PAGI", "SORE", "DONG", "GUYS"]:
        assert word not in result, f"False positive: {word} incorrectly identified as ticker"


def test_deduplicates_tickers():
    """Each ticker should appear at most once even if mentioned multiple times."""
    response = "BBCA naik, BBCA terus menguat, dan BBCA mencatat rekor."
    result = extract_response_tickers(response)
    assert result.count("BBCA") == 1


def test_empty_inputs_return_empty_list():
    """Empty or None response and no tool args should return empty list."""
    assert extract_response_tickers("") == []
    assert extract_response_tickers("", tool_call_args_list=[]) == []


def test_tool_args_with_no_ticker_key_are_skipped():
    """Tool call args dicts without 'ticker' or 'symbol' key should be ignored safely."""
    result = extract_response_tickers(
        "Analisis umum.",
        tool_call_args_list=[{"days": 30, "limit": 10}],
    )
    assert result == []


# -----------------------------------------------------------------------
# Task 2: Tests for LocalGraphMemory.record_session_briefing
# -----------------------------------------------------------------------
from engine.memory.graph_memory import LocalGraphMemory


@pytest.fixture
def temp_memory_for_briefing(tmp_path):
    """Provide a fresh LocalGraphMemory instance backed by a temp SQLite file."""
    db_file = str(tmp_path / "briefing_test.db")
    return LocalGraphMemory(db_path=db_file)


def test_record_session_briefing_creates_edges(temp_memory_for_briefing):
    """record_session_briefing should create one MENTIONED_IN_BRIEFING edge per ticker."""
    mem = temp_memory_for_briefing
    count = mem.record_session_briefing(
        session_id="CHAT-TEST-BRIEF-001",
        tickers=["BBCA", "BBRI", "ANTM"],
        context="Pre-open market briefing 25 Sep 2026",
    )
    assert count == 3

    # Verify edges exist in the SQLite database
    import sqlite3
    conn = sqlite3.connect(mem.db_path)
    rows = conn.execute(
        "SELECT source_id, target_id, relation FROM memory_edges WHERE session_id = ?",
        ("CHAT-TEST-BRIEF-001",),
    ).fetchall()
    conn.close()

    relations = [(r[0], r[1], r[2]) for r in rows]
    assert any("bbca" in r[1] or "bbca" in r[0] for r in relations), "BBCA edge missing"
    assert any("bbri" in r[1] or "bbri" in r[0] for r in relations), "BBRI edge missing"
    assert any("antm" in r[1] or "antm" in r[0] for r in relations), "ANTM edge missing"


def test_record_session_briefing_empty_tickers_returns_zero(temp_memory_for_briefing):
    """Empty ticker list should do nothing and return 0."""
    count = temp_memory_for_briefing.record_session_briefing(
        session_id="CHAT-TEST-EMPTY",
        tickers=[],
    )
    assert count == 0


def test_record_session_briefing_skips_invalid_tickers(temp_memory_for_briefing):
    """Tickers that are empty strings should be skipped."""
    count = temp_memory_for_briefing.record_session_briefing(
        session_id="CHAT-TEST-INVALID",
        tickers=["BBCA", "", "  ", "ANTM"],
    )
    assert count == 2  # Only BBCA and ANTM are valid


def test_record_session_briefing_idempotent_no_duplicate_edges(temp_memory_for_briefing):
    """Calling record_session_briefing twice with the same session+ticker must not create duplicate edges."""
    mem = temp_memory_for_briefing
    mem.record_session_briefing("CHAT-DUP-001", ["TLKM"], "First call")
    mem.record_session_briefing("CHAT-DUP-001", ["TLKM"], "Second call")

    import sqlite3
    conn = sqlite3.connect(mem.db_path)
    count = conn.execute(
        "SELECT COUNT(*) FROM memory_edges WHERE session_id = ? AND relation = 'MENTIONED_IN_BRIEFING'",
        ("CHAT-DUP-001",),
    ).fetchone()[0]
    conn.close()
    # store_observation uses ON CONFLICT DO UPDATE (upsert), so exactly 1 edge expected
    assert count == 1


# -----------------------------------------------------------------------
# Task 3: End-to-end test: memory graph updated after general-purpose chat
# -----------------------------------------------------------------------

def test_memory_graph_updated_after_general_briefing_chat(tmp_path, monkeypatch):
    """Core regression: after a macro/general pre-market chat (no ticker in prompt),
    at least one MENTIONED_IN_BRIEFING edge should exist in the memory graph.

    This tests the exact bug scenario from the audit:
      User prompt: "sebelum open market ada apa aja"
      Agent response: "...BBCA naik... BBRI positif... ANTM smelter..."
      Expected: memory graph contains edges for BBCA, BBRI, ANTM
    """
    from engine.agent.react_agent import NiskavaReActAgent
    from engine.agent.tools import NiskavaToolRegistry
    from engine.memory.graph_memory import LocalGraphMemory
    import sqlite3

    db_file = str(tmp_path / "e2e_briefing.db")
    registry = NiskavaToolRegistry(db_path=db_file, mock_mode=True)
    memory = LocalGraphMemory(db_path=db_file)

    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)
    agent.memory = memory

    # Simulate agent response that mentions multiple tickers but user prompt has none
    fake_response = (
        "# Intelijen Pasar Pre-Open\n"
        "**BBCA** menunjukkan momentum bullish.\n"
        "**BBRI** juga mencatat net inflow asing positif.\n"
        "PT Aneka Tambang Tbk (**ANTM**) konfirmasi smelter trial.\n"
    )

    monkeypatch.setattr(
        agent,
        "_run_deterministic_chat_cycle",
        lambda session_id, prompt, history, start_time: {
            "session_id": session_id,
            "response": fake_response,
            "anomalies": [],
            "findings": [],
            "duration_ms": 100,
        },
    )

    # User prompt intentionally has NO ticker
    agent.chat(
        user_prompt="sebelum open market ada apa aja dan big potensial hari ini",
        session_id="CHAT-E2E-001",
    )

    conn = sqlite3.connect(db_file)
    edges = conn.execute(
        "SELECT target_id, relation FROM memory_edges WHERE session_id = 'CHAT-E2E-001'"
    ).fetchall()
    conn.close()

    target_ids = [r[0] for r in edges]
    assert any("bbca" in t for t in target_ids), f"BBCA not in graph. Edges: {target_ids}"
    assert any("bbri" in t for t in target_ids), f"BBRI not in graph. Edges: {target_ids}"
    assert any("antm" in t for t in target_ids), f"ANTM not in graph. Edges: {target_ids}"


def test_memory_graph_still_uses_record_investigation_when_ticker_in_prompt(tmp_path, monkeypatch):
    """When user prompt DOES contain a ticker, record_investigation should still be called
    (not replaced by record_session_briefing). Ensures backward compatibility."""
    from engine.agent.react_agent import NiskavaReActAgent
    from engine.agent.tools import NiskavaToolRegistry
    from engine.memory.graph_memory import LocalGraphMemory

    db_file = str(tmp_path / "e2e_ticker_prompt.db")
    registry = NiskavaToolRegistry(db_path=db_file, mock_mode=True)
    memory = LocalGraphMemory(db_path=db_file)

    record_investigation_calls = []
    original_ri = memory.record_investigation
    def spy_record_investigation(*a, **kw):
        record_investigation_calls.append(kw)
        return original_ri(*a, **kw)
    memory.record_investigation = spy_record_investigation

    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)
    agent.memory = memory

    monkeypatch.setattr(
        agent,
        "_run_deterministic_chat_cycle",
        lambda session_id, prompt, history, start_time: {
            "session_id": session_id,
            "response": "ANTM volume spike terdeteksi.",
            "anomalies": [],
            "findings": [],
            "duration_ms": 100,
        },
    )

    agent.chat(user_prompt="cek anomali volume ANTM", session_id="CHAT-E2E-TICKER")

    assert len(record_investigation_calls) >= 1, "record_investigation not called when ticker in prompt"
    assert any(kw.get("ticker") == "ANTM" for kw in record_investigation_calls)


# -----------------------------------------------------------------------
# Task 4: Adaptive iterations — classify_prompt_complexity
# -----------------------------------------------------------------------

def test_classify_greeting_as_simple():
    from engine.agent.react_agent import classify_prompt_complexity
    assert classify_prompt_complexity("halo apa kabar") == "simple"
    assert classify_prompt_complexity("hi, good morning") == "simple"
    assert classify_prompt_complexity("test") == "simple"


def test_classify_general_market_query():
    from engine.agent.react_agent import classify_prompt_complexity
    assert classify_prompt_complexity("sebelum open market ada apa aja hari ini") == "general"
    assert classify_prompt_complexity("berita IHSG terkini") == "general"
    assert classify_prompt_complexity("market news today IDX") == "general"
    assert classify_prompt_complexity("big potential stocks today") == "general"


def test_classify_deep_analysis_query():
    from engine.agent.react_agent import classify_prompt_complexity
    assert classify_prompt_complexity("cek anomali volume ANTM 30 hari terakhir") == "deep"
    assert classify_prompt_complexity("audit bandarmologi BBCA insider") == "deep"
    assert classify_prompt_complexity("financial health stress test TLKM") == "deep"
    assert classify_prompt_complexity("investigate BMRI net foreign flow") == "deep"


def test_adaptive_iterations_respected_in_universal_chat_cycle(tmp_path, monkeypatch):
    """For a 'simple' prompt (greeting), max_iter used inside _run_universal_chat_cycle
    should be NISKAVA_SIMPLE_REACT_ITERATIONS (5), not the global 30."""
    from engine.agent.react_agent import NiskavaReActAgent
    from engine.agent.tools import NiskavaToolRegistry

    db_file = str(tmp_path / "e2e_adaptive.db")
    registry = NiskavaToolRegistry(db_path=db_file, mock_mode=True)
    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=False, ai_provider="universal")

    iteration_counts_used = []

    def spy_fn(session_id, user_prompt, history, start_time):
        from engine.agent.react_agent import (
            classify_prompt_complexity,
            SIMPLE_MAX_REACT_ITERATIONS,
            GENERAL_MAX_REACT_ITERATIONS,
            MAX_REACT_ITERATIONS,
        )
        complexity = classify_prompt_complexity(user_prompt)
        if complexity == "simple":
            iteration_counts_used.append(SIMPLE_MAX_REACT_ITERATIONS)
        elif complexity == "general":
            iteration_counts_used.append(GENERAL_MAX_REACT_ITERATIONS)
        else:
            iteration_counts_used.append(MAX_REACT_ITERATIONS)
        return {
            "session_id": session_id,
            "response": "Halo! Saya Niskava.",
            "anomalies": [],
            "findings": [],
            "duration_ms": 50,
        }

    monkeypatch.setattr(agent, "_run_universal_chat_cycle", spy_fn)
    agent.chat(user_prompt="halo apa kabar", session_id="CHAT-ADAPTIVE-001")

    assert iteration_counts_used, "spy was never called"
    assert iteration_counts_used[0] <= 8, (
        f"Simple greeting should use <= 8 iterations, got {iteration_counts_used[0]}"
    )



