"""Tests for Evidence Translation & Anti-Contamination Protocol in System Prompt."""

from engine.agent.react_agent import get_system_prompt, detect_prompt_language


def test_system_prompt_has_anti_contamination_rule():
    prompt = get_system_prompt("en")
    assert "EVIDENCE TRANSLATION & ANTI-CONTAMINATION" in prompt
    assert "translate" in prompt.lower()
    assert "Bahasa Indonesia" in prompt


def test_system_prompt_does_not_force_indonesian_on_english_session():
    prompt = get_system_prompt("en")
    assert "Bahasa Indonesia requested" not in prompt
    assert "Active session preference: English requested." in prompt


def test_1_shot_example_is_english_aligned():
    prompt = get_system_prompt("en")
    assert 'User: "analyze ANTM"' in prompt or 'User: "analyze' in prompt
