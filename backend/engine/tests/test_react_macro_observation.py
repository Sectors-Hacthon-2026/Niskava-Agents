"""Tests for observation compaction and macro guidance in react_agent."""

import json
from engine.agent.react_agent import _compact_tool_observation, get_system_prompt


def test_compact_tool_observation_sorts_news_by_date():
    news_items = [
        {"title": "Older News", "publication_date": "2026-09-10", "snippet": "Older"},
        {"title": "Latest Today News", "publication_date": "2026-09-24", "snippet": "Today"},
        {"title": "Middle News", "publication_date": "2026-09-18", "snippet": "Middle"},
        {"title": "Very Old News", "publication_date": "2026-08-01", "snippet": "Oldest"},
    ]
    res_str = _compact_tool_observation("search_news", news_items)
    compact = json.loads(res_str.split(" (summarized")[0])
    
    assert len(compact) == 3
    # Top 1 item must be the freshest (2026-09-24)
    assert compact[0]["title"] == "Latest Today News"
    assert compact[0]["date"] == "2026-09-24"


def test_system_prompt_includes_macro_sector_guidance():
    prompt = get_system_prompt()
    assert "subsectors" in prompt or "macro" in prompt.lower()
