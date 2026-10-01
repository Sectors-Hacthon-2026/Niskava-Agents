"""Tests for ReAct Agent prompt guidance for market screener queries."""
import pytest
from engine.agent.react_agent import classify_prompt_complexity, get_system_prompt


def test_classify_prompt_complexity_screener_queries():
    assert classify_prompt_complexity("saham apa yang paling naik hari ini?") == "general"
    assert classify_prompt_complexity("top gainers hari ini siapa aja") == "general"
    assert classify_prompt_complexity("top losers bursa idx") == "general"
    assert classify_prompt_complexity("saham paling likuid paling ramai") == "general"


def test_system_prompt_contains_top_changes_guidance():
    prompt = get_system_prompt()
    assert "top_changes" in prompt
    assert "top_gainers" in prompt or "most_traded" in prompt
