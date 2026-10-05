import pytest
from engine.sectors.tickers import extract_valid_tickers

def test_extract_valid_tickers_ignores_blue_chip():
    news_text = "IHSG Menguat Ditopang Arus Masuk Modal Asing dan Kinerja Saham Blue Chip"
    assert "BLUE" not in extract_valid_tickers(news_text)

def test_extract_valid_tickers_ignores_lowercase_and_compound_idioms():
    assert "BLUE" not in extract_valid_tickers("koleksi saham blue chip untuk investasi jangka panjang")
    assert "FAST" not in extract_valid_tickers("industri fast moving consumer goods tumbuh pesat")
    assert "REAL" not in extract_valid_tickers("sektor real estate mengalami perlambatan")

def test_extract_valid_tickers_detects_genuine_blue():
    assert "BLUE" in extract_valid_tickers("PT Berkah Prima Pratama Tbk (BLUE) mencatatkan kenaikan laba")
    assert "BLUE" in extract_valid_tickers("Cek transaksi saham BLUE hari ini")
