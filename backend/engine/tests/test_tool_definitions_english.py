"""Tests verifying NiskavaToolRegistry gateway tool definitions are English."""

from engine.agent.tools import NiskavaToolRegistry


def test_gateway_tool_definitions_are_pure_english(tmp_path):
    registry = NiskavaToolRegistry(str(tmp_path / "test.db"), mock_mode=True)
    defs = registry.get_tool_definitions()

    assert len(defs) == 4
    tool_names = [d["name"] for d in defs]
    assert "execute_skill" in tool_names
    assert "query_sectors" in tool_names
    assert "search_news" in tool_names
    assert "query_memory" in tool_names

    for d in defs:
        desc = d["description"]
        # Must not contain prominent Indonesian prompt words
        assert "Jalankan" not in desc, f"Indonesian word 'Jalankan' in {d['name']}"
        assert "Memanen" not in desc, f"Indonesian word 'Memanen' in {d['name']}"
        assert "Ambil" not in desc, f"Indonesian word 'Ambil' in {d['name']}"
        assert "berita terkurasi" not in desc, f"Indonesian phrasing in {d['name']}"
