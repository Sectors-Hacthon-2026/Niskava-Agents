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
