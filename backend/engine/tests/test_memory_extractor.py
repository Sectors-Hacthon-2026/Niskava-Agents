"""Unit tests for automatic conversational dialogue triple extractor."""

import pytest
from engine.memory.extractor import extract_dialogue_observations


def test_extract_user_buy_intent():
    text = "Saya baru beli BBCA di 9800 kemarin untuk swing"
    observations = extract_dialogue_observations(text)
    assert len(observations) >= 1
    obs = observations[0]
    assert obs["ticker"] == "BBCA"
    assert obs["relation"] == "HOLDS_AT"
    assert "9800" in obs["target_label"]
    assert obs["target_type"] == "PRICE_LEVEL"


def test_extract_user_watchlist_intent():
    text = "Tolong pantau pergerakan saham ANTM ya"
    observations = extract_dialogue_observations(text)
    assert len(observations) >= 1
    obs = observations[0]
    assert obs["ticker"] == "ANTM"
    assert obs["relation"] == "WATCHES"
    assert obs["target_label"] == "ANTM"
    assert obs["target_type"] == "TICKER"


def test_extract_user_exit_intent():
    text = "Saya sudah take profit ANTM di 1650 tadi siang"
    observations = extract_dialogue_observations(text)
    assert len(observations) >= 1
    obs = observations[0]
    assert obs["ticker"] == "ANTM"
    assert obs["relation"] == "EXITED_AT"
    assert "1650" in obs["target_label"]


def test_no_extraction_for_general_chat():
    text = "Bagaimana prospek pasar IHSG pekan ini?"
    observations = extract_dialogue_observations(text)
    assert observations == []


def test_extractor_context_snippets_are_english():
    """Regression: All auto-generated context_snippet strings from extractor must be English (Bug #2C)."""
    obs_exit = extract_dialogue_observations("take profit ANTM di 1650")
    exit_obs = [o for o in obs_exit if o.get("relation") == "EXITED_AT"]
    assert len(exit_obs) >= 1

    obs_hold = extract_dialogue_observations("beli BBCA di 9800")
    hold_obs = [o for o in obs_hold if o.get("relation") == "HOLDS_AT"]
    assert len(hold_obs) >= 1

    obs_watch = extract_dialogue_observations("pantau ANTM")
    watch_obs = [o for o in obs_watch if o.get("relation") == "WATCHES"]
    assert len(watch_obs) >= 1

    for obs_list, label in [
        (exit_obs, "EXITED_AT"),
        (hold_obs, "HOLDS_AT"),
        (watch_obs, "WATCHES"),
    ]:
        snippet = obs_list[0].get("context_snippet", "")
        for marker in ["Realisasi posisi", "di level harga", "Kepemilikan posisi modal", "Pengguna memantau"]:
            assert marker not in snippet, (
                f"Indonesian string '{marker}' in {label} context_snippet: '{snippet}'"
            )

