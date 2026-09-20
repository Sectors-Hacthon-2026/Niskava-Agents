"""Unit tests for IDX ticker validator and extractor."""

import pytest
from engine.sectors.tickers import extract_valid_tickers, is_valid_idx_ticker


def test_is_valid_idx_ticker():
    # Valid active IDX tickers
    assert is_valid_idx_ticker("BBCA") is True
    assert is_valid_idx_ticker("bbca") is True  # Case-insensitive
    assert is_valid_idx_ticker("ANTM") is True
    assert is_valid_idx_ticker("GOTO") is True
    assert is_valid_idx_ticker("TLKM") is True
    assert is_valid_idx_ticker("BUMI") is True

    # Invalid typos, slang, dictionary words not on IDX
    assert is_valid_idx_ticker("DOMG") is False  # User typo in screenshot
    assert is_valid_idx_ticker("DONG") is False
    assert is_valid_idx_ticker("GUYS") is False
    assert is_valid_idx_ticker("SUHU") is False
    assert is_valid_idx_ticker("WKWK") is False
    assert is_valid_idx_ticker("BROK") is False
    assert is_valid_idx_ticker("TEST") is False
    assert is_valid_idx_ticker("") is False
    assert is_valid_idx_ticker(None) is False


def test_extract_valid_tickers_no_false_positives():
    # The exact prompt from the user's screenshot
    prompt = "cek berita hari ini domg"
    extracted = extract_valid_tickers(prompt)
    assert extracted == []  # Crucial: Must NOT extract DOMG!

    # Casual greeting / chatter
    assert extract_valid_tickers("halo selamat pagi guys") == []
    assert extract_valid_tickers("gimana kabar pasar hari ini bro?") == []
    assert extract_valid_tickers("suhu tolong ajarin apa itu rasio DER") == []


def test_extract_valid_tickers_genuine_stocks():
    assert extract_valid_tickers("apakah volume ANTM kemarin anomali?") == ["ANTM"]
    assert extract_valid_tickers("tolong bandingkan BBCA dan BBRI") == ["BBCA", "BBRI"]
    assert extract_valid_tickers("Cek rumor default TINS") == ["TINS"]
    assert extract_valid_tickers("goto lagi rebound nih") == ["GOTO"]


def test_extract_valid_tickers_ambiguous_dictionary_words():
    # 'HALO' as greeting should NOT be extracted as ticker
    assert extract_valid_tickers("halo semua, apa kabar?") == []

    # 'HALO' with financial marker SHOULD be extracted
    assert extract_valid_tickers("pergerakan saham HALO hari ini") == ["HALO"]
    assert extract_valid_tickers("emiten GOLD mencatatkan kenaikan laba") == ["GOLD"]
