import pytest
from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry

@pytest.fixture
def agent(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    return NiskavaReActAgent(tool_registry=registry, mock_mode=True, language="id")

def test_apa_kabar_routes_to_greeting_not_news(agent):
    res = agent.chat("halo bro apa kabar")
    assert "Rangkuman Berita Pasar Modal Terkini" not in res["response"]
    assert any(w in res["response"].lower() for w in ["halo", "kabar", "bantu", "asisten"])

def test_interjection_hah_routes_to_conversational_clarification(agent):
    res = agent.chat("hah")
    # Must NOT run quantitative anomaly on any ticker
    assert len(res.get("anomalies", [])) == 0
    assert "VOLUME_AND_PRICE_SURGE" not in res["response"]
    assert any(w in res["response"].lower() for w in ["ada yang bisa", "bingung", "jelaskan", "bantu"])
