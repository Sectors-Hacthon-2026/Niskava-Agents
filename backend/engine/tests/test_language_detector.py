"""Tests for lexical prompt language detection in Niskava ReAct Agent."""

import pytest
from engine.agent.react_agent import detect_prompt_language


def test_detect_english_prompts():
    assert detect_prompt_language("hi, who are you?") == "en"
    assert detect_prompt_language("analyze ANTM stock volume anomalies") == "en"
    assert detect_prompt_language("what is the current PE ratio of BBCA?") == "en"
    assert detect_prompt_language("check the latest market news today") == "en"


def test_detect_indonesian_prompts():
    assert detect_prompt_language("halo, siapa kamu?") == "id"
    assert detect_prompt_language("cek anomali volume saham ANTM dong") == "id"
    assert detect_prompt_language("bagaimana pergerakan IHSG hari ini?") == "id"
    assert detect_prompt_language("tolong analisa laporan keuangan BBRI") == "id"


def test_detect_ambiguous_short_prompts_uses_fallback():
    assert detect_prompt_language("ANTM", fallback="en") == "en"
    assert detect_prompt_language("BBCA 30", fallback="id") == "id"
    assert detect_prompt_language("", fallback="en") == "en"
