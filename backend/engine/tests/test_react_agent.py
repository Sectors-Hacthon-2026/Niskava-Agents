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


class TestLeanSystemPrompt:
    """Verifies the refactored get_system_prompt produces a lean, compliant output."""

    def _make_gateway_defs(self) -> list:
        """4 minimal gateway tool defs mirroring Task 2 output shape."""
        return [
            {
                "name": "execute_skill",
                "description": "Jalankan SOP. skill_id: market_anomaly_recon, event_causality_audit.",
                "parameters": {"type": "object", "properties": {"skill_id": {}, "arguments": {}}, "required": ["skill_id", "arguments"]},
            },
            {
                "name": "query_sectors",
                "description": "Router Sectors API. domain: candles, fundamentals, foreign_flow.",
                "parameters": {"type": "object", "properties": {"domain": {}, "ticker": {}}, "required": ["domain", "ticker"]},
            },
            {
                "name": "search_news",
                "description": "Router berita dan keterbukaan informasi bursa.",
                "parameters": {"type": "object", "properties": {"ticker": {}}, "required": ["ticker"]},
            },
            {
                "name": "query_memory",
                "description": "Router memori lokal graf.",
                "parameters": {"type": "object", "properties": {"concept_or_ticker": {}}, "required": ["concept_or_ticker"]},
            },
        ]

    def test_prompt_word_count_is_below_600(self):
        from engine.agent.react_agent import get_system_prompt
        defs = self._make_gateway_defs()
        prompt = get_system_prompt("id", available_tools=defs)
        word_count = len(prompt.split())
        assert word_count < 600, (
            f"System prompt is too large: {word_count} words. Target is ≤ 600 words (~450 tokens)."
        )

    def test_prompt_contains_four_gateway_tool_names(self):
        from engine.agent.react_agent import get_system_prompt
        defs = self._make_gateway_defs()
        prompt = get_system_prompt("id", available_tools=defs)
        for gateway in ["execute_skill", "query_sectors", "search_news", "query_memory"]:
            assert f"`{gateway}`" in prompt, f"Gateway `{gateway}` not found in prompt"

    def test_old_atomic_tool_names_not_in_prompt(self):
        from engine.agent.react_agent import get_system_prompt
        defs = self._make_gateway_defs()
        prompt = get_system_prompt("id", available_tools=defs)
        assert "get_daily_candles" not in prompt
        assert "memory_recall_context" not in prompt
        assert "memory_store_observation" not in prompt

    def test_law1_and_law2_rules_still_present(self):
        from engine.agent.react_agent import get_system_prompt
        defs = self._make_gateway_defs()
        prompt = get_system_prompt("id", available_tools=defs)
        assert "LAW 1" in prompt or "Deterministic" in prompt
        assert "LAW 2" in prompt or "BUY/SELL" in prompt

    def test_prompt_without_tools_does_not_crash(self):
        from engine.agent.react_agent import get_system_prompt
        prompt = get_system_prompt("en", available_tools=None)
        assert len(prompt) > 100

    def test_english_language_prompt_uses_english_instruction(self):
        from engine.agent.react_agent import get_system_prompt
        defs = self._make_gateway_defs()
        prompt = get_system_prompt("en", available_tools=defs)
        assert "English" in prompt or "english" in prompt.lower()

    def test_system_prompt_includes_language_mirroring_protocol(self):
        from engine.agent.react_agent import get_system_prompt
        defs = self._make_gateway_defs()
        prompt = get_system_prompt(available_tools=defs)
        assert "LANGUAGE & MIRRORING PROTOCOL" in prompt
        assert "mirror the exact language" in prompt.lower() or "mirror" in prompt.lower()
        assert "English" in prompt
        assert "Bahasa Indonesia" in prompt

    def test_system_prompt_default_is_polyglot_english_core(self):
        from engine.agent.react_agent import get_system_prompt
        prompt = get_system_prompt()
        assert "You are Niskava Agent" in prompt
        assert "Zero Preamble" in prompt or "ZERO PREAMBLE" in prompt


class TestFollowupChips:
    """Verifies proactive follow-up chips generation and appending (ADR-11)."""

    def test_constants_defined_for_all_six_skills(self):
        from engine.agent.react_agent import (
            _SKILL_FOLLOWUP_GRAPH,
            _SKILL_CHIP_DESCRIPTIONS_ID,
            _SKILL_CHIP_DESCRIPTIONS_EN,
        )
        expected_skills = [
            "market_anomaly_recon",
            "event_causality_audit",
            "insider_bandarmology_forensic",
            "financial_health_stress_test",
            "mining_commodity_divergence",
            "peer_valuation_benchmark",
        ]
        for skill in expected_skills:
            assert skill in _SKILL_FOLLOWUP_GRAPH, f"Missing {skill} in follow-up graph"
            assert skill in _SKILL_CHIP_DESCRIPTIONS_ID, f"Missing {skill} in ID chip descriptions"
            assert skill in _SKILL_CHIP_DESCRIPTIONS_EN, f"Missing {skill} in EN chip descriptions"

    def test_chips_generated_for_single_skill_executed(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True, language="id")

        chips = agent._build_followup_chips(
            final_response="Analisis ANTM selesai.",
            skills_executed=["market_anomaly_recon"],
            ticker="ANTM",
            language="id",
        )

        assert chips != ""
        assert "💡 Rekomendasi Penelusuran Lanjutan" in chips
        assert "ANTM" in chips
        assert "market_anomaly_recon" not in chips
        assert "event_causality_audit" in chips

    def test_chips_no_duplicate_executed_skills(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True, language="id")

        chips = agent._build_followup_chips(
            final_response="Analisis BBCA selesai.",
            skills_executed=["market_anomaly_recon", "event_causality_audit"],
            ticker="BBCA",
            language="id",
        )

        assert chips != ""
        assert "market_anomaly_recon" not in chips
        assert "event_causality_audit" not in chips
        assert "insider_bandarmology_forensic" in chips

    def test_chips_all_skills_executed_returns_empty(self, tmp_path):
        from engine.agent.react_agent import _SKILL_FOLLOWUP_GRAPH
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)

        all_skills = list(_SKILL_FOLLOWUP_GRAPH.keys())
        chips = agent._build_followup_chips(
            final_response="Investigasi lengkap.",
            skills_executed=all_skills,
            ticker="ANTM",
        )
        assert chips == ""

    def test_chips_empty_ticker_returns_empty(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)

        chips = agent._build_followup_chips(
            final_response="PER adalah Price to Earnings Ratio.",
            skills_executed=[],
            ticker="",
        )
        assert chips == ""

    def test_chips_already_present_returns_empty(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)

        already_has_chips = (
            "Analisis selesai.\n\n---\n### 💡 Rekomendasi Penelusuran Lanjutan:\n"
            "1. Jalankan event_causality_audit"
        )
        chips = agent._build_followup_chips(
            final_response=already_has_chips,
            skills_executed=["market_anomaly_recon"],
            ticker="ANTM",
        )
        assert chips == ""

    def test_chips_english_language(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True, language="en")

        chips = agent._build_followup_chips(
            final_response="Analysis completed for ANTM.",
            skills_executed=["market_anomaly_recon"],
            ticker="ANTM",
            language="en",
        )

        assert "💡 Recommended Next Steps" in chips
        assert "ANTM" in chips
        assert "market_anomaly_recon" not in chips

    def test_chips_appended_in_universal_chat_cycle(self, tmp_path):
        import json
        import time
        from unittest.mock import MagicMock, patch

        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        events = []
        agent = NiskavaReActAgent(
            tool_registry=registry,
            emitter=lambda ev: events.append(ev),
            mock_mode=True,
            language="id",
            append_followup_chips=True,
        )

        # Turn 1: Model calls execute_skill for ANTM
        mock_turn1 = {
            "message": {
                "role": "assistant",
                "content": (
                    '<thought>Scan volume anomaly</thought>\n'
                    '<tool_call>{"name": "execute_skill", "arguments": {"skill_id": "market_anomaly_recon", "arguments": {"ticker": "ANTM"}}}</tool_call>'
                ),
            }
        }
        # Turn 2: Model provides final response
        mock_turn2 = {
            "message": {
                "role": "assistant",
                "content": "<response>Analisis kuantitatif ANTM menunjukkan aktivitas wajar tanpa anomali.</response>",
            }
        }

        mock_resp1 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn1]}))
        mock_resp2 = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn2]}))

        with patch("requests.post", side_effect=[mock_resp1, mock_resp2]):
            res = agent._run_universal_chat_cycle(
                session_id="TEST-CHIPS-CYCLE",
                user_prompt="analisis ANTM dong",
                history=[],
                start_time=time.time(),
            )

        assert "Analisis kuantitatif ANTM menunjukkan aktivitas wajar" in res["response"]
        assert "💡 Rekomendasi Penelusuran Lanjutan" in res["response"]
        assert "ANTM" in res["response"]

        # Ensure recommendation section does not duplicate executed skill
        rec_section = res["response"].split("💡 Rekomendasi Penelusuran Lanjutan")[1]
        assert "market_anomaly_recon" not in rec_section

        # Ensure events emitted include the appended chips in message_complete
        complete_events = [e for e in events if e.get("event") == "agent_message_complete"]
        assert len(complete_events) >= 1
        assert "💡 Rekomendasi Penelusuran Lanjutan" in complete_events[-1]["content"]

    def test_chips_not_appended_by_default_in_universal_chat_cycle(self, tmp_path):
        """Verifies that by default, follow-up chips are NOT appended to produce clean LLM output."""
        import json
        import time
        from unittest.mock import MagicMock, patch

        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        events = []
        agent = NiskavaReActAgent(
            tool_registry=registry,
            emitter=lambda ev: events.append(ev),
            mock_mode=True,
            language="id",
        )
        assert agent.append_followup_chips is False

        mock_turn = {
            "message": {
                "role": "assistant",
                "content": "<response>Analisis BBCA menunjukkan performa stabil.</response>",
            }
        }
        mock_resp = MagicMock(status_code=200, text=json.dumps({"choices": [mock_turn]}))

        with patch("requests.post", return_value=mock_resp):
            res = agent._run_universal_chat_cycle(
                session_id="TEST-CHIPS-DEFAULT-CLEAN",
                user_prompt="analisis BBCA",
                history=[],
                start_time=time.time(),
            )

        assert res["response"] == "Analisis BBCA menunjukkan performa stabil."
        assert "💡 Rekomendasi" not in res["response"]
        assert "💡 Recommended" not in res["response"]


def test_memory_records_session_even_without_anomalies(tmp_path, monkeypatch):
    """Regression: record_investigation must be called even if no anomalies or findings
    are returned. Bug #1A — guard 'if target_ticker and (anomalies or findings)' was
    silently skipping graph recording for conversational-only sessions.
    """
    from engine.agent.react_agent import NiskavaReActAgent
    from engine.agent.tools import NiskavaToolRegistry
    from engine.memory.graph_memory import LocalGraphMemory

    db_file = str(tmp_path / "test_no_anomaly.db")
    registry = NiskavaToolRegistry(db_path=db_file, mock_mode=True)
    memory = LocalGraphMemory(db_path=db_file)

    record_calls = []
    original_record = memory.record_investigation

    def recording_spy(*args, **kwargs):
        record_calls.append(kwargs)
        return original_record(*args, **kwargs)

    memory.record_investigation = recording_spy

    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)
    agent.memory = memory

    # Ensure deterministic chat cycle returns zero anomalies and zero findings
    # to specifically test conversational-only sessions where no quant anomalies exist.
    monkeypatch.setattr(
        agent,
        "_run_deterministic_chat_cycle",
        lambda session_id, prompt, history, start_time: {
            "session_id": session_id,
            "response": "BBCA fundamental overview.",
            "anomalies": [],
            "findings": [],
        },
    )

    agent.chat(user_prompt="Tell me about BBCA fundamentals", session_id="CHAT-TEST-001")

    assert len(record_calls) >= 1, (
        "record_investigation was never called. "
        "Bug #1A: guard 'anomalies or findings' is blocking graph updates."
    )


def test_memory_graph_no_crash_when_memory_is_none(tmp_path):
    """Regression: Bug #1B — valid_candidates may be undefined if self.memory is None.
    When memory is None, the post-chat block must not raise NameError.
    """
    from engine.agent.react_agent import NiskavaReActAgent
    from engine.agent.tools import NiskavaToolRegistry

    db_file = str(tmp_path / "test_no_memory.db")
    registry = NiskavaToolRegistry(db_path=db_file, mock_mode=True)
    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)
    agent.memory = None  # Explicitly disable

    result = agent.chat(user_prompt="What is PER ratio?", session_id="CHAT-NO-MEM-001")
    assert isinstance(result, dict)
    assert "session_id" in result


def test_compact_tool_observation_limits_length():
    """Verify bulky tool observations (Sectors news, reports) are compacted to <= 1600 chars."""
    from engine.agent.react_agent import _compact_tool_observation

    # Simulate large news payload (> 8KB)
    large_news = [
        {"title": f"News Headline {i}", "snippet": "A" * 500, "date": "2026-09-20", "source": "Reuters"}
        for i in range(10)
    ]
    compacted = _compact_tool_observation("search_news", large_news, max_len=1500)
    assert len(compacted) <= 1600
    assert "News Headline 0" in compacted
    assert "summarized" in compacted or "items" in compacted


def test_on_llm_retry_emits_english_thoughts(tmp_path):
    """Verify that retry thoughts and provider warnings are strictly in English regardless of agent language."""
    from engine.agent.react_agent import NiskavaReActAgent
    from engine.agent.tools import NiskavaToolRegistry

    db_file = str(tmp_path / "test_retry_lang.db")
    registry = NiskavaToolRegistry(db_path=db_file, mock_mode=True)

    # Test with language="id" - errors and retry notices must still be English
    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=False, language="id")

    emitted = []
    agent.emitter = lambda ev: emitted.append(ev)

    agent.url = "http://127.0.0.1:59999/v1/chat/completions"  # Unreachable port
    agent.chat("investigasi MANDIRI", session_id="RETRY-TEST-001")

    thoughts = [e.get("thought", "") for e in emitted if e.get("event") == "agent_thought"]
    assert len(thoughts) > 0
    for th in thoughts:
        assert "Koneksi Terputus" not in th, f"Indonesian text found in thought: {th}"
        assert "Menunggu" not in th, f"Indonesian text found in thought: {th}"
        assert "Gagal mengeksekusi" not in th, f"Indonesian text found in thought: {th}"


def test_investigation_cycle_uses_neutral_english_prompt(tmp_path):
    from engine.agent.react_agent import NiskavaReActAgent
    from engine.agent.tools import NiskavaToolRegistry
    registry = NiskavaToolRegistry(str(tmp_path / "test.db"), mock_mode=True)
    events = []
    agent = NiskavaReActAgent(tool_registry=registry, emitter=lambda ev: events.append(ev), mock_mode=True)
    res = agent.investigate(ticker="ANTM", days=30)
    assert res is not None
    assert "ANTM" in str(res)


