import pytest
from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry

def test_history_does_not_extract_tickers_from_assistant_news(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True, language="id")

    # History where assistant printed news mentioning "Saham Blue Chip", BBCA, ANTM
    history = [
        {"role": "user", "content": "halo bro apa kabar"},
        {"role": "assistant", "content": "1. IHSG Menguat Ditopang Arus Masuk Modal Asing dan Kinerja Saham Blue Chip\n2. BBCA Menguat"},
    ]

    # User responds with "hah"
    res = agent._run_deterministic_chat_cycle("sess-1", "hah", history, 0.0)

    # Must NOT have extracted BLUE or BBCA
    assert len(res.get("anomalies", [])) == 0
    assert "BLUE" not in res.get("response", "")
    assert "BBCA" not in res.get("response", "")

def test_history_recalls_ticker_on_explicit_followup(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True, language="id")

    history = [
        {"role": "user", "content": "analisis saham BBCA"},
        {"role": "assistant", "content": "Laporan BBCA selesai."},
    ]

    # User explicitly follows up with equity inquiry
    res = agent._run_deterministic_chat_cycle("sess-2", "bagaimana dengan volumenya?", history, 0.0)
    assert any(anom.get("ticker") == "BBCA" for anom in res.get("anomalies", [])) or "BBCA" in res.get("response", "")
